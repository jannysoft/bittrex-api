package cryptobot

import (
	"bittrex"
	"math"
)

func fixFloat(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func CalculatePercent(ask, base float64) (percent float64) {

	difference := ask - base
	percent = fixFloat(difference/base*100, 2)

	return
}

func CalculateBuyPrice(globalBalance float64, ask float64) (quantity float64) {
	twoPercentOfGlobal := fixFloat(globalBalance-globalBalance*0.996, 8)
	quantity = fixFloat(twoPercentOfGlobal/ask, 8)

	return
}

func CalculateMartingalePrice(spentBTC float64, ask float64) (quantity float64) {

	quantity = fixFloat(spentBTC/ask, 8)

	return
}

func CalculateZeroProfit(dbOrder DBOrder, bittrexOrder bittrex.Order2) (zeroProfit float64) {

	commission := fixFloat(bittrexOrder.CommissionPaid*2, 8)

	btcSpent := bittrexOrder.Price + dbOrder.SpentBTC + commission
	quantity := bittrexOrder.Quantity + dbOrder.AltCoinQuantity

	zeroProfit = fixFloat(btcSpent/quantity, 8)

	return
}

func CalculateTakeProfit(price float64) (profit float64) {

	profit = fixFloat(price*1.02, 8)
	return
}

func calculateProfit(soldPrice, spentBTC float64) (profit float64) {

	profit = fixFloat(soldPrice-spentBTC, 8)
	return
}
