package ethereum

import (
	"math/big"

	"github.com/diamnet/go/services/bifrost/common"
)

func (t Transaction) ValueToDiamNet() string {
	valueEth := new(big.Rat)
	valueEth.Quo(new(big.Rat).SetInt(t.ValueWei), weiInEth)
	return valueEth.FloatString(common.DiamNetAmountPrecision)
}
