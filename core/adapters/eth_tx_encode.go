package adapters

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/store"
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
				fmt.Sprint("while attempting to encode element %i, %s", idx, name.Str))
		}
		rv = append(rv, encoding)
		totalLength += len(encoding)
	}
	return utils.ConcatBytes(rv...), nil
}

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
	m

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

func ensureTxRunResult(input *models.RunResult, str *strpkg.Store) {
	val, err := input.ResultString()
	if err != nil {
		input.SetError(err)
		return
	}

	hash := common.HexToHash(val)
	if err != nil {
		input.SetError(err)
		return
	}

	receipt, state, err := str.TxManager.BumpGasUntilSafe(hash)
	if err != nil {
		if state == strpkg.Unknown {
			input.SetError(err)
			return
		}

		// We failed to get one of the TxAttempt receipts, so we won't mark this
		// run as errored in order to try again
		logger.Warn("EthTx Adapter Perform Resuming: ", err)
	}

	recordLatestTxHash(receipt, input)
	if state != strpkg.Safe {
		input.MarkPendingConfirmations()
		return
	}

	addReceiptToResult(receipt, input)
}

var zero = common.Hash{}

// recordLatestTxHash adds the current tx hash to the run result
func recordLatestTxHash(receipt *models.TxReceipt, in *models.RunResult) {
	if receipt == nil || receipt.Unconfirmed() {
		return
	}
	hex := receipt.Hash.String()
	in.ApplyResult(hex)
	in.Add("latestOutgoingTxHash", hex)
}

func addReceiptToResult(receipt *models.TxReceipt, in *models.RunResult) {
	receipts := []models.TxReceipt{}

	if !in.Get("ethereumReceipts").IsArray() {
		in.Add("ethereumReceipts", receipts)
	}

	if err := json.Unmarshal([]byte(in.Get("ethereumReceipts").String()), &receipts); err != nil {
		logger.Error(fmt.Errorf("EthTx Adapter unmarshaling ethereum Receipts: %v", err))
	}

	receipts = append(receipts, *receipt)
	in.Add("ethereumReceipts", receipts)
	in.CompleteWithResult(receipt.Hash.String())
}

// IsClientRetriable does its best effort to see if an error indicates one that
// might have a different outcome if we retried the operation
func IsClientRetriable(err error) bool {
	if err == nil {
		return false
	}

	if err, ok := err.(net.Error); ok {
		return err.Timeout() || err.Temporary()
	} else if errors.Cause(err) == store.ErrPendingConnection {
		return true
	}

	return false
}
