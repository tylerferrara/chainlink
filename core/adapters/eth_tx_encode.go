package adapters

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	strpkg "github.com/smartcontractkit/chainlink/core/store"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/chainlink/core/utils"
	"gopkg.in/guregu/null.v3"
)

// EthTx holds the Address to send the result to and the FunctionSelector
// to execute.
type EthTxEncode struct {
	Address          common.Address          `json:"address"`
	FunctionSelector models.FunctionSelector `json:"functionSelector"`
	DataPrefix       hexutil.Bytes           `json:"dataPrefix"`
	DataFormat       string                  `json:"format"`
	GasPrice         *models.Big             `json:"gasPrice" gorm:"type:numeric"`
	GasLimit         uint64                  `json:"gasLimit"`
}

// Perform creates the run result for the transaction if the existing run result
// is not currently pending. Then it confirms the transaction was confirmed on
// the blockchain.
func (etx *EthTxEncode) Perform(
	input models.RunResult, store *strpkg.Store) models.RunResult {
	if !store.TxManager.Connected() {
		input.MarkPendingConnection()
		return input
	}

	if !input.Status.PendingConfirmations() {
		createTxEncodeRunResult(etx, &input, store)
		return input
	}
	ensureTxRunResult(&input, store)
	return input
}

func encodeData(typ string, value string) ([]byte, error) {
	switch typ {
	case "uint256":
		return common.LeftPadBytes(common.HexToHash(value).Bytes(), 32), nil
	default:
		return nil, fmt.Errorf("unencodable type: %s", typ)
	}
}

// getTxData returns the data to save against the callback encoded according to
// the `types` and `order` fields of the job.
func getTxEncodeData(e *EthTxEncode, input *models.RunResult) ([]byte, error) {
	result := input.Result()
	types := result.Get("types").Map()
	order := result.Get("order").Array()
	values := result.Map()
	// Initially assign inputs to array of byte arrays, to avoid lots of array
	// reallocations during construction of the final return value
	rv := make([][]byte, len(order))
	totalLength := 0
	for idx, name := range(order) {
		encoding, err := encodeData(types[name.Str].Str, values[name.Str].Str)
		if err != nil {
			return nil, errors.Wrap(err,
				fmt.Sprintf("while attempting to encode element %d, %s", idx, name.Str))
		}
		rv = append(rv, encoding)
		totalLength += len(encoding)
	}
	return utils.ConcatBytes(rv...), nil
}

// XXX: ``
func createTxEncodeRunResult(
	e *EthTxEncode,
	input *models.RunResult,
	store *strpkg.Store,
) {
	value, err := getTxEncodeData(e, input)
	if err != nil {
		input.SetError(err)
		return
	}

	data := utils.ConcatBytes(e.FunctionSelector.Bytes(), e.DataPrefix, value)
	tx, err := store.TxManager.CreateTxWithGas(
		null.StringFrom(input.CachedJobRunID),
		e.Address,
		data,
		e.GasPrice.ToInt(),
		e.GasLimit,
	)
	if IsClientRetriable(err) {
		input.MarkPendingConnection()
		return
	} else if err != nil {
		input.SetError(err)
		return
	}

	input.ApplyResult(tx.Hash.String())

	txAttempt := tx.Attempts[0]
	logger.Debugw(
		fmt.Sprintf("Tx #0 checking on-chain state"),
		"txHash", txAttempt.Hash.String(),
		"txID", txAttempt.TxID,
	)

	receipt, state, err := store.TxManager.CheckAttempt(txAttempt, tx.SentAt)
	if IsClientRetriable(err) {
		input.MarkPendingConnection()
		return
	} else if err != nil {
		input.SetError(err)
		return
	}

	logger.Debugw(
		fmt.Sprintf("Tx #0 is %s", state),
		"txHash", txAttempt.Hash.String(),
		"txID", txAttempt.TxID,
		"receiptBlockNumber", receipt.BlockNumber.ToInt(),
		"currentBlockNumber", tx.SentAt,
		"receiptHash", receipt.Hash.Hex(),
	)

	if state != strpkg.Safe {
		input.MarkPendingConfirmations()
		return
	}

	addReceiptToResult(receipt, input)
}

