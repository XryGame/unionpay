package unionpay

import (
	"errors"

	"net/http"
	"io/ioutil"
	"global"
	"encoding/base64"
)




func (c *UnionClient) GetTradeNotification(req *http.Request) (noti *UnionTradeNotification, err error) {
	if req == nil {
		return nil, errors.New("request 参数不能为空")
	}

	defer req.Body.Close()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	params := c.DecodeNotify(data)

	global.GetLogMgr().Output(global.HF_JtoA(&params))

	noti = &UnionTradeNotification{}

	noti.Total_Fee = int(params.GetInt64("txnAmt"))
	noti.Cash_Fee = int(params.GetInt64("txnAmt"))
	noti.Out_Trade_No = params.GetString("orderId")
	attach, err := base64.StdEncoding.DecodeString(params.GetString("reqReserved"))
	if err != nil {
		return nil, err
	}
	noti.Attach = string(attach)
	noti.Time_End = params.GetString("txnTime")


	if params.GetString("respCode") != "00" {
		return nil, errors.New("交易失败")
	}
	ok, err := c.verifySign(params)
	if !ok {
		return  nil, err
	}

	return noti, nil
}
