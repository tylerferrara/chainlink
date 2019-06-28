package adapters

import (
	"crypto/rand"
	"math/big"

	"github.com/smartcontractkit/chainlink/core/store"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

// Random adapter generates a number between 0 and 2**256-1.
//
// WARNING: The random number adapter as implemented is not verifiable.
// We intend to either improve it in the future, or introduce a verifiable alternative.
// For now it is provided as an alternative to making web requests for random numbers,
// which is similarly unverifiable and has additional possible points of failure.
type Random struct{}

// Perform returns a random uint256 number in 0 | 2**256-1 range
func (ra *Random) Perform(input models.RunResult, _ *store.Store) models.RunResult {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		input.SetError(err)
		return input
	}
	ran := new(big.Int).SetBytes(b)
	input.ApplyResult(ran.String())
	return input
}
