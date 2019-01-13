package cryptobot

import (
	"bittrex"
	"coinigy"
	"strconv"
	"time"
)

var globalSettings GlobalSettings
var markets []bittrex.MarketSummary

func settings() {

	globalSettings = GetSettingsFromDB()

	//Get all current prices for all bittrex markets
	markets = bittrex.GetMarketSummaries()

	if globalSettings.Settings.GetAlerts == true {
		alerts()
	}

	if globalSettings.Settings.TradeOpportunities == true {

		if globalSettings.Settings.TradeLimit < 80 {
			tradeOpportunities()
		} else {
			CancelTrading()
			DeleteAllPotentialOrders()
		}
	}

	if globalSettings.Settings.CleanDB == true {
		//DeleteAllPotentialOrders()
		DeleteAllSoldOrders()
	}

	manageActiveTrades()
}

//Manage incoming alerts from Coinigy
func alerts() {
	//Request all alerts from Coinigy
	coinigyAlerts := coinigy.GetAlerts()

	//Check if we have received any alerts
	if len(coinigyAlerts.Data.OpenAlerts) != 0 {
		//Range over the received alerts and process them by adding them to DB trough AlertToDB() func
		for _, alerts := range coinigyAlerts.Data.OpenAlerts {
			AlertToDB(alerts, globalSettings)
		}
	}
}

func tradeOpportunities() {
	//Get All IsTradable == TRUE IsActive == FALSE orders from DB
	dbOrders := GetPotentialTradeOrders()

	for _, pair := range markets {
		for _, dbOrder := range dbOrders {
			//check if there are matching pairs between markets and dbOrders
			if pair.MarketName == dbOrder.MarketPair {
				//Calculate percentage on current pair from Base to Bid
				percent := CalculatePercent(pair.Ask, dbOrder.TradeBase)
				// check if current price is 10 percent above trade_base price and if true delete order from DB
				if 10 < percent && globalSettings.Settings.DeletePositiveOrder == true {
					DeletePositiveOrder(dbOrder.OrderID)
					if globalSettings.Settings.DeletePositiveOrderMsg == true {
						DeletePositiveOrderMsg(pair.MarketName)
					}
				} else if dbOrder.BuyInPercent > percent {
					//calculate 2 percent from globalBalance to start buy order
					quantity := CalculateBuyPrice(globalSettings.Settings.GlobalBalance, pair.Ask)
					//Open Buy trade
					orderUUID, message := bittrex.OpenBuyOrder(pair, quantity)
					//Update DB TradeOne with the order UUID
					if len(orderUUID.Uuid) == 0 && message == "INSUFFICIENT_FUNDS" {
						CancelTrading()
						DeleteAllPotentialOrders()
						InsufficientFundsMsg(dbOrder.MarketPair)
					} else if len(orderUUID.Uuid) == 0 && message == "MIN_TRADE_REQUIREMENT_NOT_MET" {
						getMarkets := bittrex.GetMarkets()

						for _, market := range getMarkets {
							if pair.MarketName == market.MarketName {

								orderUUID, message := bittrex.OpenBuyOrder(pair, market.MinTradeSize)

								if len(orderUUID.Uuid) == 0 && message == "INSUFFICIENT_FUNDS" {
									CancelTrading()
									DeleteAllPotentialOrders()
									InsufficientFundsMsg(dbOrder.MarketPair)
								} else if len(orderUUID.Uuid) == 0 && message != "INSUFFICIENT_FUNDS" {
									BadOrderMsg(message, dbOrder.MarketPair, "buying")
								} else if len(orderUUID.Uuid) != 0 {
									updateTradeLimit(globalSettings.Settings.TradeLimit + 1)
									UpdateTradeBuyID(orderUUID.Uuid, dbOrder.OrderID)
									NewTradeMsg(pair.MarketName, orderUUID.Uuid)
								}
							}
						}

					} else if len(orderUUID.Uuid) == 0 && message != "INSUFFICIENT_FUNDS" {
						BadOrderMsg(message, dbOrder.MarketPair, "buying")
					} else if len(orderUUID.Uuid) != 0 {
						updateTradeLimit(globalSettings.Settings.TradeLimit + 1)
						UpdateTradeBuyID(orderUUID.Uuid, dbOrder.OrderID)
						NewTradeMsg(pair.MarketName, orderUUID.Uuid)

						time.Sleep(1 * time.Second)
					}
					//notify if there is potential trades near buy target at 9 percent
				} else if dbOrder.BuyInPercent+1 > percent {
					if globalSettings.Settings.PotentialTradeMsg == true {
						PotentialTradeMsg(pair.MarketName, percent)
					}
				}
			}
		}
	}
}

func manageActiveTrades() {

	dbOrders := GetActiveOrders()

	for _, pair := range markets {
		for _, dbOrder := range dbOrders {
			if pair.MarketName == dbOrder.MarketPair {
				if dbOrder.IsBuyFulfilled == false &&
					dbOrder.IsBuyPending == true {
					//Check if trade is fulfilled and update the DB
					checkTrade := bittrex.GetOrder(dbOrder.BuyID)
					//if trade is fulfilled update DB with the rest of the params from fulfilled trade
					if checkTrade.IsOpen == false && checkTrade.QuantityRemaining == 0 {

						UpdateBuyTrade(checkTrade, dbOrder, strconv.FormatInt(time.Now().Unix(), 10))
						BuyFulfilledMsg(checkTrade)
					}
				} else if dbOrder.IsBuyFulfilled == true {

					basePercent := CalculatePercent(pair.Ask, dbOrder.TradeBase)
					zeroProfitPercent := CalculatePercent(pair.Bid, dbOrder.ZeroProfit)
					takeProfitPercent := CalculatePercent(pair.Bid, dbOrder.TakeProfit)

					if zeroProfitPercent < 0 {
						if globalSettings.Settings.NegativeTradeMsg == true {
							NegativeTradeMsg(dbOrder, basePercent, zeroProfitPercent)
						}
					} else if zeroProfitPercent >= 0 {
						if globalSettings.Settings.PositiveTradeMsg == true {
							positiveTradeMsg(dbOrder, basePercent, zeroProfitPercent)
						}
					}

					if dbOrder.BuyInPercent > basePercent && globalSettings.Settings.MartingaleBuy == true {
						//calculate 2 percent from globalBalance to start buy order
						quantity := CalculateMartingalePrice(dbOrder.SpentBTC, pair.Ask)
						//Open Buy trade
						orderUUID, message := bittrex.OpenBuyOrder(pair, quantity)
						//Update DB TradeOne with the order UUID if we have good order
						if len(orderUUID.Uuid) == 0 && message == "INSUFFICIENT_FUNDS" {
							CancelTrading()
							InsufficientFundsMsg(dbOrder.MarketPair)
						} else if len(orderUUID.Uuid) == 0 && message != "INSUFFICIENT_FUNDS" {
							BadOrderMsg(message, dbOrder.MarketPair, "buying")
						} else if len(orderUUID.Uuid) != 0 {
							UpdateTradeBuyID(orderUUID.Uuid, dbOrder.OrderID)
							NewTradeMsg(pair.MarketName, orderUUID.Uuid)
						}
						//Check if condition is good to set up take profit
					} else if dbOrder.IsForSale == false && dbOrder.SellStep < zeroProfitPercent {
						setTakeProfitMsg(pair.MarketName)
						//calculate takeProfit price
						takeProfit := CalculateTakeProfit(dbOrder.ZeroProfit)
						//Set take profit in DB
						SetTakeProfit(takeProfit, dbOrder)
						//Check if current price is trading around trade base price and if true take profit
					} else if dbOrder.IsForSale == true && dbOrder.IsSellPending == false && pair.Bid >= dbOrder.TradeBase && globalSettings.Settings.SellAtBase == true {
						basePriceTakeProfitMsg(pair.MarketName)
						//Open sell trade and get the UUID
						orderUUID, message := bittrex.OpenSellOrder(pair)
						//Update UUID in DB
						if len(orderUUID.Uuid) == 0 && message == "QUANTITY_NOT_PROVIDED" {
							DeletePositiveOrder(dbOrder.OrderID)
							BadOrderMsg(message, dbOrder.MarketPair, "selling")
						} else if len(orderUUID.Uuid) == 0 && message != "QUANTITY_NOT_PROVIDED" {
							BadOrderMsg(message, dbOrder.MarketPair, "selling")
						} else {
							UpdateSellID(orderUUID.Uuid, dbOrder.OrderID)
							sellTradeMsg(pair.MarketName, orderUUID.Uuid)
						}
						//Check if condition is good to update take profit
					} else if dbOrder.IsForSale == true && dbOrder.IsSellPending == false && dbOrder.SellStep < takeProfitPercent {
						updateTakeProfitMsg(pair.MarketName)
						//calculate takeProfit price
						takeProfit := CalculateTakeProfit(dbOrder.TakeProfit)
						//Update take profit in DB
						UpdateTakeProfit(takeProfit, dbOrder)
						//Check if current Bid is bellow take profit price and if it is then start sell trade
					} else if dbOrder.IsForSale == true && dbOrder.IsSellPending == false && pair.Bid < dbOrder.TakeProfit {
						takeProfitMsg(pair.MarketName)
						//Open sell trade and get the UUID
						orderUUID, message := bittrex.OpenSellOrder(pair)
						//Update UUID in DB
						if len(orderUUID.Uuid) == 0 {
							BadOrderMsg(message, dbOrder.MarketPair, "selling")
						} else {
							UpdateSellID(orderUUID.Uuid, dbOrder.OrderID)
							sellTradeMsg(pair.MarketName, orderUUID.Uuid)
						}
						//Check if sell trade is fulfilled and if it is close trade in DB
					} else if dbOrder.IsForSale == true && dbOrder.IsSellPending == true {
						//Check if trade is fulfilled and update the DB
						checkTrade := bittrex.GetOrder(dbOrder.SellID)
						//if trade is fulfilled update DB with the rest of the params from fulfilled trade and close trade
						if checkTrade.IsOpen == false && checkTrade.QuantityRemaining == 0 {

							profit := calculateProfit(checkTrade.Price, dbOrder.SpentBTC)
							updateTradeLimit(globalSettings.Settings.TradeLimit - 1)
							UpdateSellTrade(checkTrade, dbOrder, strconv.FormatInt(time.Now().Unix(), 10))
							sellFulfilledMsg(pair.MarketName, profit)
						}
					}
				}
			}
		}
	}
}

func Start() {

	for {
		settings() // get global settings from DB
		time.Sleep(10 * time.Second)
	}
}
