package gossip

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils"
)

func TestConsensusCallback(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	const rounds = 30

	const validatorsNum = 3

	env := newTestEnv(validatorsNum)
	defer env.Close()

	accounts := validatorsNum

	// save start balances
	balances := make([]*big.Int, accounts)
	for i := range balances {
		balances[i] = env.State().GetBalance(env.Address(idx.ValidatorID(i + 1)))
	}

	for n := uint64(0); n < rounds; n++ {
		// transfers
		txs := make([]*types.Transaction, accounts)
		for i := range txs {
			from := (i)%accounts + 1
			to := 0 + 1
			txs[i] = env.Transfer(idx.ValidatorID(from), idx.ValidatorID(to), utils.ToFtm(100))
		}
		tm := sameEpoch
		if n%10 == 0 {
			tm = nextEpoch
		}
		rr := env.ApplyTxs(tm, txs...)
		// subtract fees
		for i, r := range rr {
			fee := big.NewInt(0).Mul(new(big.Int).SetUint64(r.GasUsed), txs[i].GasPrice())
			balances[i] = big.NewInt(0).Sub(balances[i], fee)
		}
		// balance movements
		balances[0].Add(balances[0], utils.ToFtm(200))
		balances[1].Sub(balances[1], utils.ToFtm(100))
		balances[2].Sub(balances[2], utils.ToFtm(100))
	}

	// check balances
	for i := range balances {
		require.Equal(
			balances[i],
			env.State().GetBalance(env.Address(idx.ValidatorID(i+1))),
			fmt.Sprintf("account%d", i),
		)
	}

}
