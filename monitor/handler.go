package monitor

import (
	"encoding/json"
	"github.com/fpChan/goex"
	"github.com/fpChan/goex/monitor/wechat"
	"github.com/fpChan/goex/okex"
	"log"
	"net/http"
	"net/url"
)

//func (ok *OKEx) DoRequest(httpMethod, uri, reqBody string, response interface{}) error {
//	url := ok.config.Endpoint + uri
//	//sign, timestamp := ok.doParamSign(httpMethod, uri, reqBody)
//	//logger.Log.Debug("timestamp=", timestamp, ", sign=", sign)
//	resp, err := goex.NewHttpRequest(ok.config.HttpClient, httpMethod, url, reqBody, map[string]string{
//		CONTENT_TYPE: APPLICATION_JSON_UTF8,
//		ACCEPT:       APPLICATION_JSON,
//		//COOKIE:               LOCALE + "en_US",
//		OK_ACCESS_KEY:        ok.config.ApiKey,
//		OK_ACCESS_PASSPHRASE: ok.config.ApiPassphrase,
//		OK_ACCESS_SIGN:       sign,
//		OK_ACCESS_TIMESTAMP:  fmt.Sprint(timestamp)})
//	if err != nil {
//		//log.Println(err)
//		return err
//	} else {
//		logger.Log.Debug(string(resp))
//		return json.Unmarshal(resp, &response)
//	}
//}

func getFuturesPriceBySymbol(symbol string) {

	var okex = okex.NewOKEx(&goex.APIConfig{
		Endpoint: "https://www.okex.com",
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return &url.URL{
						Scheme: "socks5",
						Host:   "127.0.0.1:10000"}, nil
				},
			},
		},
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
	})
	var (
		okexSpot   = okex.OKExSpot
		okexSwap   = okex.OKExSwap   //永续合约实现
		okexFuture = okex.OKExFuture //交割合约实现
		//okexWallet =okex.OKExWallet //资金账户（钱包）操作
	)

	swapTickers, err := okexSwap.GetFutureAllTicker()
	if err != nil {
		log.Fatal("failed to get swap price by", err)
	}
	tickersBytes, err := json.MarshalIndent(swapTickers, "", "   ")
	log.Println(string(tickersBytes))

	//接口调用,更多接口调用请看代码
	log.Println(okexSpot.GetAccount()) //获取账户资产信息
	//okexSpot.BatchPlaceOrders([]goex.Order{...}) //批量下单,单个交易对同时最大只能下10笔
	log.Println()                               //获取账户权益信息
	log.Println(okexFuture.GetFutureUserinfo()) //获取账户权益信息

}

func main() {
	getFuturesPriceBySymbol("1")
	wechat.CreateWeChat()
}
