package unionpay

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"fmt"
	"io/ioutil"
	"net/url"
	"encoding/base64"
	"crypto"
)


type UnifyOrderClient struct {
	TN 		string	`json:"tn"`
}


const bodyType = "application/x-www-form-urlencoded;charset=utf-8"

// 客户端
type UnionClient struct {
	stdClient *http.Client
	tlsClient *http.Client

	MerId  			string
	PfxPath         string
	PfxPassword     string
	ReqUrl          string
	Notify_Url		string
}

// 实例化API客户端
func NewClient(merid, pfxpath, pfxpassword, requrl, notify_url string) *UnionClient {
	return &UnionClient{
		stdClient: 	&http.Client{},
		MerId:     	merid,
		PfxPath: 	pfxpath,
		PfxPassword: pfxpassword,
		ReqUrl:     requrl,
		Notify_Url: notify_url,
	}
}

// 设置请求超时时间
func (c *UnionClient) SetTimeout(d time.Duration) {
	c.stdClient.Timeout = d
	if c.tlsClient != nil {
		c.tlsClient.Timeout = d
	}
}


// 发送请求
func (c *UnionClient) Post(url string, params Params, tls bool) (Params, error) {
	var httpc *http.Client
	httpc = c.stdClient

	resp, err := httpc.Post(url, bodyType, c.Encode(params))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	v, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(v))

	return c.DecodeRequest(v), nil
}


func (c *UnionClient) DecodeRequest(bytes []byte) Params {

	src := string(bytes)
	//global.GetLogMgr().Output(src)

	params := make(Params)
	values := strings.Split(src, "&")
	for _, v := range values {
		kv := strings.SplitN(v, "=", 2)
		params.SetString(kv[0], kv[1])
	}
	return params
}

func (c *UnionClient) DecodeNotify(bytes []byte) Params {
	params := make(Params)
	//global.GetLogMgr().Output(string(bytes))
	src, err := url.QueryUnescape(string(bytes))
	if err != nil {
		return params
		//global.GetLogMgr().Output(err.Error())
	}

	values := strings.Split(src, "&")
	for _, v := range values {
		kv := strings.SplitN(v, "=", 2)
		params.SetString(kv[0], kv[1])
	}
	return params
}

func (c *UnionClient) Encode(params Params) io.Reader {
	var buf bytes.Buffer

	var keys = make([]string, 0, len(params))
	for k, _ := range params {

		keys = append(keys, k)
	}
	sort.Strings(keys)


	for i, k := range keys {
		if len(params.GetString(k)) > 0 {
			buf.WriteString(k)
			buf.WriteString(`=`)

			buf.WriteString(url.QueryEscape(params.GetString(k)))

			if i != len(keys) - 1 {
				buf.WriteString(`&`)
			}

		}
	}

	//fmt.Println(buf.String())
	return &buf
}

// 验证签名
func (c *UnionClient) verifySign(params Params) (ok bool, err error) {
	signbase64 := params.GetString("signature")
	signpublickey := params.GetString("signPubKeyCert")
	sign, err := base64.StdEncoding.DecodeString(signbase64)
	if err != nil {
		return false, err
	}

	var keys = make([]string, 0, len(params))
	for k, _ := range params {
		if k != "signature" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for i, k := range keys {
		if len(params.GetString(k)) > 0 {
			buf.WriteString(k)
			buf.WriteString(`=`)
			buf.WriteString(params.GetString(k))
			if i != len(keys) - 1 {
				buf.WriteString(`&`)
			}
		}
	}

	err = VerifyPKCS1v15(buf.Bytes(), sign, []byte(signpublickey), crypto.SHA256)

	if err != nil {
		return false, err
	}
	return true, nil
}


// 生成签名
func (c *UnionClient) Sign(params Params) string {
	var keys = make([]string, 0, len(params))
	for k, _ := range params {
		if k != "signature" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for i, k := range keys {
		if len(params.GetString(k)) > 0 {
			buf.WriteString(k)
			buf.WriteString(`=`)
			buf.WriteString(params.GetString(k))
			if i != len(keys) - 1 {
				buf.WriteString(`&`)
			}
		}
	}

	sign, err := PfxSign(c.PfxPath, c.PfxPassword, buf.String())
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	return sign
}


func (c *UnionClient) UnionpayUnifyOrder(txnamt int, attach string) (*UnifyOrderClient, error) {
	params := make(Params)


	params.SetString("version", "5.1.0")
	params.SetString("encoding", "utf-8")
	params.SetString("certId", GetCertId(c.PfxPath, c.PfxPassword))
	params.SetString("txnType", "01")
	params.SetString("txnSubType", "01")
	params.SetString("bizType", "000201")
	params.SetString("channelType", "08")
	//params.SetString("frontUrl", c.Notify_Url)

	params.SetString("backUrl", c.Notify_Url)
	params.SetString("accessType", "0")
	params.SetString("merId", c.MerId)
	params.SetString("orderId", UniqueId())
	//params.SetString("orderId", "1321321412414511144144")
	params.SetString("txnTime", time.Now().Format("20060102150405"))
	//params.SetString("txnTime", "20060102150405")
	params.SetString("txnAmt", fmt.Sprintf("%d", txnamt))
	params.SetString("accType", "01")
	params.SetString("currencyCode", "156")
	params.SetString("reqReserved", base64.StdEncoding.EncodeToString([]byte(attach)))
	params.SetString("signMethod", "01")
	params.SetString("signature", c.Sign(params))


	// 调用统一下单接口请求
	ret, err := c.Post(c.ReqUrl, params, false)
	if err != nil {
		return nil, err
	}
	if ret.GetString("respCode") != "00" {

		return nil, errors.New(fmt.Sprintf("银联下单不成功，原因:%s", ret.GetString("respMsg")))
	}

	ok, err := c.verifySign(ret)
	if !ok {
		return nil, errors.New(fmt.Sprintf("银联下单签名校验失败，原因:%s", err.Error()))

	}


	var out UnifyOrderClient
	out.TN = ret.GetString("tn")

	return &out, nil
}