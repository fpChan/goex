package okex

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/fpChan/goex"
	"github.com/fpChan/goex/common/api"
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"github.com/google/uuid"
	"strings"
	"sync"
	"time"
)

const baseUrl = "https://www.okex.com"

type OKEx struct {
	config          *types.APIConfig
	OKExSpot        *OKExSpot
	OKExFuture      *OKExFuture
	OKExSwap        *OKExSwap
	OKExWallet      *OKExWallet
	OKExMargin      *OKExMargin
	OKExV3FuturesWs *OKExV3FuturesWs
	OKExV3SpotWs    *OKExV3SpotWs
	OKExV3SwapWs    *OKExV3SwapWs
}

func NewOKEx(config *types.APIConfig) *OKEx {
	if config.Endpoint == "" {
		config.Endpoint = baseUrl
	}
	okex := &OKEx{config: config}
	okex.OKExSpot = &OKExSpot{okex}
	okex.OKExFuture = &OKExFuture{OKEx: okex, Locker: new(sync.Mutex)}
	okex.OKExWallet = &OKExWallet{okex}
	okex.OKExMargin = &OKExMargin{okex}
	okex.OKExSwap = &OKExSwap{okex, config}
	okex.OKExV3FuturesWs = NewOKExV3FuturesWs(okex)
	okex.OKExV3SpotWs = NewOKExSpotV3Ws(okex)
	okex.OKExV3SwapWs = NewOKExV3SwapWs(okex)
	return okex
}

func (ok *OKEx) GetExchangeName() string {
	return types.OKEX
}

func (ok *OKEx) UUID() string {
	return strings.Replace(uuid.New().String(), "-", "", 32)
}

func (ok *OKEx) DoRequest(httpMethod, uri, reqBody string, response interface{}) error {
	url := ok.config.Endpoint + uri
	sign, timestamp := ok.doParamSign(httpMethod, uri, reqBody)
	//fmt.Printf("url %s\n", url)
	//logger.Log.Debug("timestamp=", timestamp, ", sign=", sign)
	resp, err := api.NewHttpRequest(ok.config.HttpClient, httpMethod, url, reqBody, map[string]string{
		CONTENT_TYPE: APPLICATION_JSON_UTF8,
		ACCEPT:       APPLICATION_JSON,
		//COOKIE:               LOCALE + "en_US",
		OK_ACCESS_KEY:        ok.config.ApiKey,
		OK_ACCESS_PASSPHRASE: ok.config.ApiPassphrase,
		OK_ACCESS_SIGN:       sign,
		OK_ACCESS_TIMESTAMP:  fmt.Sprint(timestamp)})
	if err != nil {
		//log.Println(err)
		return err
	} else {
		logger.Log.Debug(string(resp))
		return json.Unmarshal(resp, &response)
	}
}

func (ok *OKEx) adaptOrderState(state int) types.TradeStatus {
	switch state {
	case -2:
		return types.ORDER_FAIL
	case -1:
		return types.ORDER_CANCEL
	case 0:
		return types.ORDER_UNFINISH
	case 1:
		return types.ORDER_PART_FINISH
	case 2:
		return types.ORDER_FINISH
	case 3:
		return types.ORDER_UNFINISH
	case 4:
		return types.ORDER_CANCEL_ING
	}
	return types.ORDER_UNFINISH
}

/*
 Get a http request body is a json string and a byte array.
*/
func (ok *OKEx) BuildRequestBody(params interface{}) (string, *bytes.Reader, error) {
	if params == nil {
		return "", nil, errors.New("illegal parameter")
	}
	data, err := json.Marshal(params)
	if err != nil {
		//log.Println(err)
		return "", nil, errors.New("json convert string error")
	}

	jsonBody := string(data)
	binBody := bytes.NewReader(data)

	return jsonBody, binBody, nil
}

func (ok *OKEx) doParamSign(httpMethod, uri, requestBody string) (string, string) {
	timestamp := ok.IsoTime()
	preText := fmt.Sprintf("%s%s%s%s", timestamp, strings.ToUpper(httpMethod), uri, requestBody)
	//log.Println("preHash", preText)
	sign, _ := GetParamHmacSHA256Base64Sign(ok.config.ApiSecretKey, preText)
	return sign, timestamp
}

/*
 Get a iso time
  eg: 2018-03-16T18:02:48.284Z
*/
func (ok *OKEx) IsoTime() string {
	utcTime := time.Now().UTC()
	iso := utcTime.String()
	isoBytes := []byte(iso)
	iso = string(isoBytes[:10]) + "T" + string(isoBytes[11:23]) + "Z"
	return iso
}

func (ok *OKEx) LimitBuy(amount, price string, currency types.CurrencyPair, opt ...types.LimitOrderOptionalParameter) (*types.Order, error) {
	return ok.OKExSpot.LimitBuy(amount, price, currency, opt...)
}

func (ok *OKEx) LimitSell(amount, price string, currency types.CurrencyPair, opt ...types.LimitOrderOptionalParameter) (*types.Order, error) {
	return ok.OKExSpot.LimitSell(amount, price, currency, opt...)
}

func (ok *OKEx) MarketBuy(amount, price string, currency types.CurrencyPair) (*types.Order, error) {
	return ok.OKExSpot.MarketBuy(amount, price, currency)
}

func (ok *OKEx) MarketSell(amount, price string, currency types.CurrencyPair) (*types.Order, error) {
	return ok.OKExSpot.MarketSell(amount, price, currency)
}

func (ok *OKEx) CancelOrder(orderId string, currency types.CurrencyPair) (bool, error) {
	return ok.OKExSpot.OKExSpot.CancelOrder(orderId, currency)
}

func (ok *OKEx) GetOneOrder(orderId string, currency types.CurrencyPair) (*types.Order, error) {
	return ok.OKExSpot.GetOneOrder(orderId, currency)
}

func (ok *OKEx) GetUnfinishOrders(currency types.CurrencyPair) ([]types.Order, error) {
	return ok.OKExSpot.GetUnfinishOrders(currency)
}

func (ok *OKEx) GetOrderHistorys(currency types.CurrencyPair, opt ...types.OptionalParameter) ([]types.Order, error) {
	return ok.OKExSpot.GetOrderHistorys(currency, opt...)
}

func (ok *OKEx) GetAccount() (*types.Account, error) {
	return ok.OKExSpot.GetAccount()
}

func (ok *OKEx) GetTicker(currency types.CurrencyPair) (*types.Ticker, error) {
	return ok.OKExSpot.GetTicker(currency)
}

func (ok *OKEx) GetDepth(size int, currency types.CurrencyPair) (*types.Depth, error) {
	return ok.OKExSpot.GetDepth(size, currency)
}

func (ok *OKEx) GetKlineRecords(currency types.CurrencyPair, period types.KlinePeriod, size int, optional ...types.OptionalParameter) ([]types.Kline, error) {
	return ok.OKExSpot.GetKlineRecords(currency, period, size, optional...)
}

func (ok *OKEx) GetTrades(currencyPair types.CurrencyPair, since int64) ([]types.Trade, error) {
	return ok.OKExSpot.GetTrades(currencyPair, since)
}
