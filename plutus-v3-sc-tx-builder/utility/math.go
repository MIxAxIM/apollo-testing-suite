package utility

import "math/big"

func ToFraction(percentage float64) int64 {

	percentageBigInt := new(big.Float).SetFloat64(percentage)
	fraction := new(big.Float).SetInt64(1000)
	fraction.Quo(fraction, percentageBigInt)
	rounded, _ := fraction.Int64()

	return rounded
}
