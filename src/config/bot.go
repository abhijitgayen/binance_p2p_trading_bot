package config

type ExtraFilter struct {
    Price        float64  `json:"price"`
    MinimumLimit float64  `json:"minimum_limit"`
    MaximumLimit float64  `json:"maximum_limit,omitempty"`
    ErrorCodes   string   `json:"error_codes"`
}

type BotConfig struct {
    Asset              string      `json:"asset"`
    Fiat               string      `json:"fiat"`
    Page               int         `json:"page"`
    Rows               int         `json:"rows"`
    TradeType          string      `json:"trade_type"`
    TotalAmountToInvest float64    `json:"total_amount_to_invest"`
    NoOfOrders         int         `json:"no_of_orders"`
    ExtraFilter        ExtraFilter `json:"extra_filter"`
    APIKey             string      `json:"api_key"`
    SecretKey          string      `json:"secret_key"`
}

var DefaultBotConfig = BotConfig{
    Asset:              "USDT",
    Fiat:               "INR",
    Page:               1,
    Rows:               20,
    TradeType:          "BUY",
    TotalAmountToInvest: 10000,
    NoOfOrders:         1,
    ExtraFilter: ExtraFilter{
        Price:        85,
        MinimumLimit: 1000,
        ErrorCodes:   "83999",
    },
    APIKey:    "",
    SecretKey: "",
}