package utility

func CalculateFee(precision int64, orderAmount, orderThreshold, feePercentage, minFee int64) int64 {
	if orderAmount <= orderThreshold {
		return minFee
	} else {
		result := ((orderAmount * int64(precision)) / feePercentage)
		return result
	}
}
