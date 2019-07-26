package adapters_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/chainlink/core/adapters"
	"github.com/smartcontractkit/chainlink/core/internal/cltest"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/chainlink/core/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEthTxEncodeAdapter_Perform_ConfirmedWithJSON(t *testing.T) {

	address := cltest.NewAddress()
	fHash := models.HexToFunctionSelector("b3f98adc")
	dataPrefix := hexutil.Bytes(
		hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000045746736453745"))

	adapterUnderTest := adapters.EthTxEncode{
		Address:          address,
		DataPrefix:       dataPrefix,
		FunctionSelector: fHash,
		Types: map[string]string{
			"gammaX": "uint256", "gammaY": "uint256", "c": "uint256", "s": "uint256"},
		Order: []string{"gammaX", "gammaY", "c", "s"},
	}

	t.Parallel()
	app, cleanup := cltest.NewApplicationWithKey(t)
	defer cleanup()
	store := app.Store

	inputValue := `Should this be JSON???`

	ethMock, err := app.MockStartAndConnect()
	require.NoError(t, err)

	hash := cltest.NewHash()
	sentAt := uint64(23456)
	confirmed := sentAt + 1
	ethMock.Register("eth_sendRawTransaction", hash,
		func(_ interface{}, data ...interface{}) error {
			rlp := data[0].([]interface{})[0].(string)
			tx, err := utils.DecodeEthereumTx(rlp)
			assert.NoError(t, err)
			assert.Equal(t, address.String(), tx.To().String())
			wantData := "0xWhat should go here??"
			assert.Equal(t, wantData, hexutil.Encode(tx.Data()))
			return nil
		})
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(sentAt))
	receipt := models.TxReceipt{Hash: hash, BlockNumber: cltest.Int(confirmed)}
	ethMock.Register("eth_getTransactionReceipt", receipt)
	input := cltest.RunResultWithResult(inputValue)
	data := adapterUnderTest.Perform(input, store)
	assert.False(t, data.HasError())
	from := cltest.GetAccountAddress(t, store)
	txs, err := store.TxFrom(from)
	assert.NoError(t, err)
	require.Len(t, txs, 1)
	assert.Len(t, txs[0].Attempts, 1)
	ethMock.EventuallyAllCalled(t)
}
