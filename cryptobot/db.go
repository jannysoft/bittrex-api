package cryptobot

import (
	"bittrex"
	"coinigy"
	"encoding/json"
	"github.com/couchbase/gocb"
	"strconv"
	"strings"
)

var bucket *gocb.Bucket

func init() {
	//Connect to Couchbase
	cluster, _ := gocb.Connect("couchbase://192.168.1.85")
	bucket, _ = cluster.OpenBucket("crypto", "")
}

func UpdateGlobalBalanceSettings(settings GlobalSettings) {

	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, "0-settings")
	n1qlParams = append(n1qlParams, settings.Settings.GlobalBalance)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto USE KEYS $1 SET global_balance = $2")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()
}

func UpdateGlobalBotSettings(settings GlobalSettings) {

	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, "0-settings")
	n1qlParams = append(n1qlParams, settings.Settings.GetAlerts)
	n1qlParams = append(n1qlParams, settings.Settings.MartingaleBuy)
	n1qlParams = append(n1qlParams, settings.Settings.TradeOpportunities)
	n1qlParams = append(n1qlParams, settings.Settings.DeletePositiveOrder)
	n1qlParams = append(n1qlParams, settings.Settings.SellAtBase)
	n1qlParams = append(n1qlParams, settings.Settings.CleanDB)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto USE KEYS $1 SET get_alerts = $2, martingale_buy = $3, trade_opportunities = $4, delete_positive_order = $5, sell_at_base = $6, clean_db = $7")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()
}

func UpdateGlobalMsgSettings(settings GlobalSettings) {

	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, "0-settings")
	n1qlParams = append(n1qlParams, settings.Settings.DeletePositiveOrderMsg)
	n1qlParams = append(n1qlParams, settings.Settings.GetAlertsMsg)
	n1qlParams = append(n1qlParams, settings.Settings.PotentialTradeMsg)
	n1qlParams = append(n1qlParams, settings.Settings.NegativeTradeMsg)
	n1qlParams = append(n1qlParams, settings.Settings.PositiveTradeMsg)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto USE KEYS $1 SET delete_positive_order_msg = $2, get_alerts_msg = $3, potential_trade_msg = $4, negative_trade_msg = $5, positive_trade_msg = $6")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()
}

func GetSettingsFromDB() (settings GlobalSettings) {

	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, "0-settings")

	dbQuery := gocb.NewN1qlQuery("SELECT * FROM crypto AS settings USE KEYS $1")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	json.Unmarshal(dbResult.NextBytes(), &settings)

	dbResult.Close()
	return
}

//Get specific order from DB based on pair which has been passed from AlertToDB
func DBOrderByPair(pair string) (dbOrder DBOrders) {

	var n1qlParams []interface{}
	n1qlParams = append(n1qlParams, pair)

	dbQuery := gocb.NewN1qlQuery("SELECT * FROM crypto AS db_order WHERE market_pair = $1")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	json.Unmarshal(dbResult.NextBytes(), &dbOrder)

	dbResult.Close()
	return
}

func DeletePositiveOrder(dbOrder string) {
	var n1qlParams []interface{}
	n1qlParams = append(n1qlParams, dbOrder)

	dbQuery := gocb.NewN1qlQuery("DELETE FROM crypto WHERE order_id = $1")

	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()
}

func DeleteAllPotentialOrders() {

	dbQuery := gocb.NewN1qlQuery("DELETE FROM crypto WHERE is_tradable = true AND is_active = false")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, nil)

	dbResult.Close()

}

func DeleteAllSoldOrders() {

	dbQuery := gocb.NewN1qlQuery("DELETE FROM crypto WHERE is_tradable = false AND is_active = false")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, nil)

	dbResult.Close()

}

//Process all incoming alerts
func AlertToDB(alert coinigy.OpenAlert, settings GlobalSettings) {

	marketNameSplit := strings.Split(alert.MktName, "/")

	var order DBOrder

	order.OrderID = string(alert.AlertID)
	order.MarketPair = marketNameSplit[1] + "-" + marketNameSplit[0]
	order.IsDeletable = true
	order.IsTradable = true
	order.IsActive = false
	order.TradeBase, _ = strconv.ParseFloat(alert.Price, 64)
	order.BuyInPercent = -10
	order.SellStep = 8

	//Check if there is existing order in DB
	dbOrder := DBOrderByPair(marketNameSplit[1] + "-" + marketNameSplit[0])

	//If there is existing order on the current pair but is not traded yet delete the order and create new order with the fresh alert params
	if dbOrder.DBOrder.IsDeletable == true && dbOrder.DBOrder.IsTradable == true {

		var n1qlParams []interface{}
		n1qlParams = append(n1qlParams, dbOrder.DBOrder.OrderID)

		dbQuery := gocb.NewN1qlQuery("DELETE FROM crypto WHERE order_id = $1")

		dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

		bucket.Upsert(order.OrderID, order, 0)

		coinigy.DeleteAlert(order.OrderID)

		dbResult.Close()

		if settings.Settings.GetAlertsMsg == true {
			ExistingDeletableOrderMsg(dbOrder, alert)
		}

		//If there is existing order on the current pair but is traded and is currently active ignore the fresh alert
	} else if dbOrder.DBOrder.IsDeletable == false && dbOrder.DBOrder.IsTradable == true {

		coinigy.DeleteAlert(order.OrderID)

		if settings.Settings.GetAlertsMsg == true {
			ExistingNonDeletableOrderMsg(dbOrder)
		}

		//If there is no order on the current pair add received alert as a order to DB
	} else {

		bucket.Upsert(order.OrderID, order, 0)

		coinigy.DeleteAlert(order.OrderID)

		if settings.Settings.GetAlertsMsg == true {
			NewOrderMsg(alert)
		}
	}
}

//Request all orders is_tradible == true from DB
func GetPotentialTradeOrders() []DBOrder {
	var dbOrders []DBOrders
	var dbOrder DBOrders

	dbQuery := gocb.NewN1qlQuery("SELECT * FROM crypto AS db_order WHERE is_tradable = true AND is_active = false")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, nil)

	for dbResult.Next(&dbOrder) {
		dbOrders = append(dbOrders, dbOrder)
	}

	var orders []DBOrder

	for _, order := range dbOrders {
		orders = append(orders, order.DBOrder)
	}
	dbResult.Close()
	return orders
}

func GetActiveOrders() []DBOrder {
	var dbOrders []DBOrders
	var dbOrder DBOrders

	dbQuery := gocb.NewN1qlQuery("SELECT * FROM crypto AS db_order WHERE is_tradable = true AND is_active = true")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, nil)

	for dbResult.Next(&dbOrder) {
		dbOrders = append(dbOrders, dbOrder)
	}

	var orders []DBOrder

	for _, order := range dbOrders {
		orders = append(orders, order.DBOrder)
	}
	dbResult.Close()
	return orders
}

func UpdateTradeBuyID(uuid, orderID string) {
	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, uuid)
	n1qlParams = append(n1qlParams, orderID)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto SET buy_id = $1, is_buy_pending = true, is_buy_fulfilled = false, is_deletable = false, is_active = true WHERE order_id = $2")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)
	dbResult.Close()
}

func updateTradeLimit(limit float64) {

	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, "0-settings")
	n1qlParams = append(n1qlParams, limit)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto USE KEYS $1 SET trade_limit = $2")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()
}

func CancelTrading() {

	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, "0-settings")

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto USE KEYS $1 SET trade_opportunities = false")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()

}

func UpdateBuyTrade(bittrexOrder bittrex.Order2, dbOrder DBOrder, unixTime string) {

	calculateZeroProfit := CalculateZeroProfit(dbOrder, bittrexOrder)

	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, bittrexOrder.Quantity+dbOrder.AltCoinQuantity)
	n1qlParams = append(n1qlParams, bittrexOrder.Price+dbOrder.SpentBTC)
	n1qlParams = append(n1qlParams, bittrexOrder.PricePerUnit)
	n1qlParams = append(n1qlParams, unixTime)
	n1qlParams = append(n1qlParams, calculateZeroProfit)
	n1qlParams = append(n1qlParams, dbOrder.BuyInPercent*2)
	n1qlParams = append(n1qlParams, dbOrder.OrderID)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto SET is_buy_fulfilled = true, is_buy_pending = false, alt_coin_quantity = $1, spent_btc = $2, trade_at_price = $3, buy_time_stamp = $4, zero_profit = $5, buy_in_percent = $6 WHERE order_id = $7")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()
}

func SetTakeProfit(takeProfit float64, dbOrder DBOrder) {
	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, takeProfit)
	n1qlParams = append(n1qlParams, dbOrder.OrderID)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto SET is_for_sale = true, take_profit = $1 WHERE order_id = $2")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()

}

func UpdateTakeProfit(takeProfit float64, dbOrder DBOrder) {

	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, takeProfit)
	n1qlParams = append(n1qlParams, dbOrder.OrderID)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto SET take_profit = $1 WHERE order_id = $2")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()

}

func UpdateSellID(orderUuid string, dbOrder string) {
	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, orderUuid)
	n1qlParams = append(n1qlParams, dbOrder)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto SET is_sell_pending = true, sell_id = $1 WHERE order_id = $2")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()

}

func UpdateSellTrade(bittrexOrder bittrex.Order2, dbOrder DBOrder, unixTime string) {

	var n1qlParams []interface{}

	n1qlParams = append(n1qlParams, unixTime)
	n1qlParams = append(n1qlParams, bittrexOrder.Price)
	n1qlParams = append(n1qlParams, dbOrder.OrderID)

	dbQuery := gocb.NewN1qlQuery("UPDATE crypto SET is_sell_pending = false, is_sell_fulfilled = true, is_tradable = false, is_active = false, is_for_sale = false, sell_time_stamp = $1, sold_price = $2 WHERE order_id = $3")
	dbResult, _ := bucket.ExecuteN1qlQuery(dbQuery, n1qlParams)

	dbResult.Close()

}
