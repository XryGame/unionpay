package unionpay



type UnionTradeNotification struct {

	Total_Fee       int	   `json:"total_fee"`       // 总金额
	Cash_Fee        int	   `json:"cash_fee"`        // 现金支付金额
	Out_Trade_No    string `json:"out_trade_no"`    // 商户订单号
	Attach          string `json:"attach"`          // 商家数据包
	Time_End      	string `json:"time_end"`      	// 支付完成时间

}
