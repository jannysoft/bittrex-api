package cryptobot

type Settings struct {
	GlobalBalance          float64 `json:"global_balance"`
	TradeLimit             float64 `json:"trade_limit"`
	TradeOpportunities     bool    `json:"trade_opportunities"`
	TradeOpportunitiesMsg  bool    `json:"trade_opportunities_msg"`
	GetAlerts              bool    `json:"get_alerts"`
	GetAlertsMsg           bool    `json:"get_alerts_msg"`
	DeletePositiveOrderMsg bool    `json:"delete_positive_order_msg"`
	DeletePositiveOrder    bool    `json:"delete_positive_order"`
	MartingaleBuy          bool    `json:"martingale_buy"`
	PotentialTradeMsg      bool    `json:"potential_trade_msg"`
	NegativeTradeMsg       bool    `json:"negative_trade_msg"`
	PositiveTradeMsg       bool    `json:"positive_trade_msg"`
	SellAtBase             bool    `json:"sell_at_base"`
	CleanDB                bool    `json:"clean_db"`
}

type GlobalSettings struct {
	Settings Settings `json:"settings"`
}

type DBOrder struct {
	OrderID         string  `json:"order_id"`
	MarketPair      string  `json:"market_pair"`
	IsDeletable     bool    `json:"is_deletable"`
	IsTradable      bool    `json:"is_tradable"`
	IsActive        bool    `json:"is_active"`
	IsForSale       bool    `json:"is_for_sale"`
	IsSellPending   bool    `json:"is_sell_pending"`
	IsSellFulfilled bool    `json:"is_sell_fulfilled"`
	TradeBase       float64 `json:"trade_base"`
	SellStep        float64 `json:"sell_step"`
	ZeroProfit      float64 `json:"zero_profit"`
	TakeProfit      float64 `json:"take_profit"`
	SoldPrice       float64 `json:"sold_price"`
	SellID          string  `json:"sell_id"`
	SellTimeStamp   string  `json:"sell_time_stamp"`
	IsBuyPending    bool    `json:"is_buy_pending"`
	IsBuyFulfilled  bool    `json:"is_buy_fulfilled"`
	BuyInPercent    float64 `json:"buy_in_percent"`
	BuyID           string  `json:"buy_id"`
	SpentBTC        float64 `json:"spent_btc"`
	AltCoinQuantity float64 `json:"alt_coin_quantity"`
	BuyTimeStamp    string  `json:"buy_time_stamp"`
}

type DBOrders struct {
	DBOrder DBOrder `json:"db_order"`
}
