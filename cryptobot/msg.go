package cryptobot

import (
	"bittrex"
	"coinigy"
	"fmt"
	"github.com/fatih/color"
	"time"
)

var message = color.New(color.FgHiBlue, color.Bold).SprintFunc()
var errorMsg = color.New(color.FgHiRed, color.Bold).SprintFunc()
var orderMsg = color.New(color.FgHiYellow, color.Bold).SprintFunc()
var positiveTrade = color.New(color.FgHiGreen, color.Bold).SprintFunc()
var negativeTrade = color.New(color.FgHiRed, color.Bold).SprintFunc()
var value = color.New(color.FgHiMagenta, color.Bold).SprintFunc()
var redColor = color.New(color.FgHiRed, color.Bold).SprintFunc()

func NewOrderMsg(alert coinigy.OpenAlert) {
	fmt.Printf("%s New Alert with ID: %s found on %s, and will be added as a potential order to DB. \n", message("Message:"), alert.AlertID, alert.MktName)
}

func ExistingDeletableOrderMsg(dbOrder DBOrders, alert coinigy.OpenAlert) {
	fmt.Printf("%s Existing untraded Order found on %s with ID: %s. This order will be deleted and new order with ID: %s will be added to DB. \n", message("Message:"), dbOrder.DBOrder.MarketPair, dbOrder.DBOrder.OrderID, alert.AlertID)
}

func DeletePositiveOrderMsg(pair string) {
	currentTime := time.Now().Local()

	fmt.Printf("%v %s Positive untraded order found on %s. This order will be deleted from DB. \n", currentTime.Format("15:04:05"), message("Message:"), redColor(pair))
}

func PotentialTradeMsg(pair string, percent float64) {
	currentTime := time.Now().Local()

	fmt.Printf("%v %s Potential trade developing on %s. The pair is currently trading at %v%s from our basePrice. \n", currentTime.Format("15:04:05"), message("Message:"), pair, percent, "%")
}

func ExistingNonDeletableOrderMsg(dbOrder DBOrders) {
	fmt.Printf("%s Active trade found on %s with ID: %s. This trade cannot be modified or deleted. \n", errorMsg("ERROR:"), dbOrder.DBOrder.MarketPair, dbOrder.DBOrder.OrderID)
}

func InsufficientFundsMsg(pair string) {
	currentTime := time.Now().Local()
	fmt.Printf("%v %s while buying %s All potential and martingale trades will be desabled. \n", currentTime.Format("15:04:05"), errorMsg("INSUFFICIENT FUNDS"), pair)
}

func BadOrderMsg(message, pair, tradeType string) {
	currentTime := time.Now().Local()
	fmt.Printf("%v Ooooops something went wrong when %s %s: %s. \n", currentTime.Format("15:04:05"), tradeType, pair, errorMsg(message))
}

func NewTradeMsg(pair, uuid string) {
	currentTime := time.Now().Local()
	fmt.Printf("%v %s with UUID: %s oppened on %s. \n", currentTime.Format("15:04:05"), orderMsg("NEW TRADE"), uuid, pair)
}

func sellTradeMsg(pair, uuid string) {
	currentTime := time.Now().Local()
	fmt.Printf("%v %s with UUID: %s oppened on %s. \n", currentTime.Format("15:04:05"), orderMsg("SELL TRADE"), uuid, redColor(pair))
}

func BuyFulfilledMsg(trade bittrex.Order2) {
	currentTime := time.Now().Local()
	fmt.Printf("%v %s on pair %s. \n", currentTime.Format("15:04:05"), orderMsg("BUY FULFILLED"), trade.Exchange)
}

func NegativeTradeMsg(order DBOrder, basePercent, zeroPercent float64) {
	fmt.Printf("%s PAIR: %s BASE: %v%s ZERO: %v%s, LEVEL: %v, BTC: %v BUYIN: %v. \n", negativeTrade("NEGATIVE TRADE"), value(order.MarketPair), value(basePercent), value("%"), value(zeroPercent), value("%"), value(order.ZeroProfit), value(order.SpentBTC), value(order.BuyInPercent))
}

func positiveTradeMsg(order DBOrder, basePercent, zeroPercent float64) {
	fmt.Printf("%s PAIR: %s BASE: %v%s ZERO: %v%s BTC: %v BUYIN: %v. \n", positiveTrade("POSITIVE TRADE"), value(order.MarketPair), value(basePercent), value("%"), value(zeroPercent), value("%"), value(order.SpentBTC), value(order.BuyInPercent))
}

func setTakeProfitMsg(pair string) {
	currentTime := time.Now().Local()
	fmt.Printf("%v %s on %s. The pair is trading abouve SELL STEP. \n", currentTime.Format("15:04:05"), orderMsg("TAKE PROFIT SETUP"), pair)
}

func updateTakeProfitMsg(pair string) {
	currentTime := time.Now().Local()
	fmt.Printf("%v %s on %s. The pair is trading abouve SELL STEP. \n", currentTime.Format("15:04:05"), orderMsg("TAKE PROFIT UPDATE"), pair)
}

func takeProfitMsg(pair string) {
	currentTime := time.Now().Local()
	fmt.Printf("%v %s on %s. Executing SELL ORDER. \n", currentTime.Format("15:04:05"), orderMsg("TAKE PROFIT HIT"), pair)
}

func basePriceTakeProfitMsg(pair string) {
	currentTime := time.Now().Local()
	fmt.Printf("%v %s BASE PRICE HIT on %s. Taking profit before it goes down :0). \n", currentTime.Format("15:04:05"), orderMsg("WHOOOOOOO"), pair)
}

func sellFulfilledMsg(pair string, profit float64) {
	currentTime := time.Now().Local()
	fmt.Printf("%v %s on %s. Our profit is %v .\n", currentTime.Format("15:04:05"), orderMsg("SELL FULFILLED"), pair, value(profit))
}
