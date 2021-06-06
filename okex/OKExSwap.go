package okex

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/fpChan/goex"
)

const (
	/*
	  http headers
	*/
	OK_ACCESS_KEY        = "OK-ACCESS-KEY"
	OK_ACCESS_SIGN       = "OK-ACCESS-SIGN"
	OK_ACCESS_TIMESTAMP  = "OK-ACCESS-TIMESTAMP"
	OK_ACCESS_PASSPHRASE = "OK-ACCESS-PASSPHRASE"

	/**
	  paging params
	*/
	OK_FROM  = "OK-FROM"
	OK_TO    = "OK-TO"
	OK_LIMIT = "OK-LIMIT"

	CONTENT_TYPE = "Content-Type"
	ACCEPT       = "Accept"
	COOKIE       = "Cookie"
	LOCALE       = "locale="

	APPLICATION_JSON      = "application/json"
	APPLICATION_JSON_UTF8 = "application/json; charset=UTF-8"

	/*
	  i18n: internationalization
	*/
	ENGLISH            = "en_US"
	SIMPLIFIED_CHINESE = "zh_CN"
	//zh_TW || zh_HK
	TRADITIONAL_CHINESE = "zh_HK"

	/*
	  http methods
	*/
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"

	/*
	 others
	*/
	ResultDataJsonString = "resultDataJsonString"
	ResultPageJsonString = "resultPageJsonString"

	BTC_USD_SWAP = "BTC-USD-SWAP"
	LTC_USD_SWAP = "LTC-USD-SWAP"
	ETH_USD_SWAP = "ETH-USD-SWAP"
	ETC_USD_SWAP = "ETC-USD-SWAP"
	BCH_USD_SWAP = "BCH-USD-SWAP"
	BSV_USD_SWAP = "BSV-USD-SWAP"
	EOS_USD_SWAP = "EOS-USD-SWAP"
	XRP_USD_SWAP = "XRP-USD-SWAP"

	/*Rest Endpoint*/
	Endpoint              = "https://www.okex.com"
	GET_ACCOUNTS          = "/api/swap/v3/accounts"
	PLACE_ORDER           = "/api/swap/v3/order"
	CANCEL_ORDER          = "/api/swap/v3/cancel_order/%s/%s"
	GET_ORDER             = "/api/swap/v3/orders/%s/%s"
	GET_POSITION          = "/api/swap/v3/%s/position"
	GET_DEPTH             = "/api/swap/v3/instruments/%s/depth?size=%d"
	GET_TICKER            = "/api/swap/v3/instruments/%s/ticker"
	GET_ALL_TICKER        = "/api/swap/v3/instruments/ticker"
	GET_UNFINISHED_ORDERS = "/api/swap/v3/orders/%s?status=%d&limit=%d"
	PLACE_ALGO_ORDER      = "/api/swap/v3/order_algo"
	CANCEL_ALGO_ORDER     = "/api/swap/v3/cancel_algos"
	GET_ALGO_ORDER        = "/api/swap/v3/order_algo/%s?order_type=%d&"
)

type BaseResponse struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Result       bool   `json:"result,string"`
}

type OKExSwap struct {
	*OKEx
	config *types.APIConfig
}

func NewOKExSwap(config *types.APIConfig) *OKExSwap {
	return &OKExSwap{OKEx: &OKEx{config: config}, config: config}
}

func (ok *OKExSwap) GetExchangeName() string {
	return types.OKEX_SWAP
}

func (ok *OKExSwap) GetFutureTicker(currencyPair types.CurrencyPair, contractType string) (*types.Ticker, error) {
	var resp struct {
		InstrumentId string  `json:"instrument_id"`
		Last         float64 `json:"last,string"`
		High24h      float64 `json:"high_24h,string"`
		Low24h       float64 `json:"low_24h,string"`
		BestBid      float64 `json:"best_bid,string"`
		BestAsk      float64 `json:"best_ask,string"`
		Volume24h    float64 `json:"volume_24h,string"`
		Timestamp    string  `json:"timestamp"`
	}
	contractType = ok.adaptContractType(currencyPair)
	err := ok.DoRequest("GET", fmt.Sprintf(GET_TICKER, contractType), "", &resp)
	if err != nil {
		return nil, err
	}

	date, _ := time.Parse(time.RFC3339, resp.Timestamp)
	return &types.Ticker{
		Pair: currencyPair,
		Last: resp.Last,
		Low:  resp.Low24h,
		High: resp.High24h,
		Vol:  resp.Volume24h,
		Buy:  resp.BestBid,
		Sell: resp.BestAsk,
		Date: uint64(date.UnixNano() / int64(time.Millisecond))}, nil
}

func (ok *OKExSwap) GetFutureAllTicker() (*[]types.FutureTicker, error) {
	var resp SwapTickerList
	err := ok.DoRequest("GET", GET_ALL_TICKER, "", &resp)
	if err != nil {
		return nil, err
	}

	var tickers []types.FutureTicker
	for _, t := range resp {
		date, _ := time.Parse(time.RFC3339, t.Timestamp)
		tickers = append(tickers, types.FutureTicker{
			ContractType: t.InstrumentId,
			Ticker: &types.Ticker{
				Pair: types.NewCurrencyPair3(t.InstrumentId, "-"),
				Sell: t.BestAsk,
				Buy:  t.BestBid,
				Low:  t.Low24h,
				High: t.High24h,
				Last: t.Last,
				Vol:  t.Volume24h,
				Date: uint64(date.UnixNano() / int64(time.Millisecond))}})
	}

	return &tickers, nil
}

func (ok *OKExSwap) GetFutureDepth(currencyPair types.CurrencyPair, contractType string, size int) (*types.Depth, error) {
	var resp SwapInstrumentDepth
	contractType = ok.adaptContractType(currencyPair)

	err := ok.DoRequest("GET", fmt.Sprintf(GET_DEPTH, contractType, size), "", &resp)
	if err != nil {
		return nil, err
	}

	var dep types.Depth
	dep.ContractType = contractType
	dep.Pair = currencyPair
	dep.UTime, _ = time.Parse(time.RFC3339, resp.Timestamp)

	for _, v := range resp.Bids {
		dep.BidList = append(dep.BidList, types.DepthRecord{
			Price:  ToFloat64(v[0]),
			Amount: ToFloat64(v[1])})
	}

	for i := len(resp.Asks) - 1; i >= 0; i-- {
		dep.AskList = append(dep.AskList, types.DepthRecord{
			Price:  ToFloat64(resp.Asks[i][0]),
			Amount: ToFloat64(resp.Asks[i][1])})
	}

	return &dep, nil
}

func (ok *OKExSwap) GetFutureUserinfo(currencyPair ...types.CurrencyPair) (*types.FutureAccount, error) {
	var (
		err   error
		infos SwapAccounts
	)

	if len(currencyPair) == 1 {
		accountInfo, err := ok.GetFutureAccountInfo(currencyPair[0])
		if err != nil {
			return nil, err
		}

		if accountInfo == nil {
			return nil, errors.New("api return info is empty")
		}

		infos.Info = append(infos.Info, *accountInfo)

		goto wrapperF
	}

	err = ok.OKEx.DoRequest("GET", GET_ACCOUNTS, "", &infos)
	if err != nil {
		return nil, err
	}

	//log.Println(infos)
wrapperF:
	acc := types.FutureAccount{}
	acc.FutureSubAccounts = make(map[types.Currency]types.FutureSubAccount, 2)

	for _, account := range infos.Info {
		subAcc := types.FutureSubAccount{AccountRights: account.Equity,
			KeepDeposit: account.Margin, ProfitReal: account.RealizedPnl,
			ProfitUnreal: account.UnrealizedPnl, RiskRate: account.MarginRatio}
		meta := strings.Split(account.InstrumentId, "-")
		if len(meta) > 0 {
			subAcc.Currency = types.NewCurrency(meta[0], "")
		}
		acc.FutureSubAccounts[subAcc.Currency] = subAcc
	}

	return &acc, nil
}

func (ok *OKExSwap) GetFutureAccountInfo(currency types.CurrencyPair) (*SwapAccountInfo, error) {
	var infos struct {
		Info SwapAccountInfo `json:"info"`
	}

	err := ok.OKEx.DoRequest("GET", fmt.Sprintf("/api/swap/v3/%s/accounts", ok.adaptContractType(currency)), "", &infos)
	if err != nil {
		return nil, err
	}

	return &infos.Info, nil
}

/*
 OKEX swap api parameter's definition
 @author Lingting Fu
 @date 2018-12-27
 @version 1.0.0
*/

type BasePlaceOrderInfo struct {
	ClientOid  string `json:"client_oid"`
	Price      string `json:"price"`
	MatchPrice string `json:"match_price"`
	Type       string `json:"type"`
	Size       string `json:"size"`
	OrderType  string `json:"order_type"`
}

type PlaceOrderInfo struct {
	BasePlaceOrderInfo
	InstrumentId string `json:"instrument_id"`
}

type PlaceOrdersInfo struct {
	InstrumentId string                `json:"instrument_id"`
	OrderData    []*BasePlaceOrderInfo `json:"order_data"`
}

func (ok *OKExSwap) PlaceFutureOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType, matchPrice int, leverRate float64) (string, error) {
	fOrder, err := ok.PlaceFutureOrder2(currencyPair, contractType, price, amount, openType, matchPrice)
	return fOrder.OrderID2, err
}

func (ok *OKExSwap) PlaceFutureOrder2(currencyPair types.CurrencyPair, contractType, price, amount string, openType, matchPrice int, opt ...types.LimitOrderOptionalParameter) (*types.FutureOrder, error) {
	cid := GenerateOrderClientId(32)
	param := PlaceOrderInfo{
		BasePlaceOrderInfo{
			ClientOid:  cid,
			Price:      price,
			MatchPrice: fmt.Sprint(matchPrice),
			Type:       fmt.Sprint(openType),
			Size:       amount,
			OrderType:  "0",
		},
		ok.adaptContractType(currencyPair),
	}

	if len(opt) > 0 {
		switch opt[0] {
		case types.PostOnly:
			param.OrderType = "1"
		case types.Fok:
			param.OrderType = "2"
		case types.Ioc:
			param.OrderType = "3"
		}
	}

	reqBody, _, _ := ok.OKEx.BuildRequestBody(param)

	fOrder := &types.FutureOrder{
		ClientOid:    cid,
		Currency:     currencyPair,
		ContractName: contractType,
		OType:        openType,
		Price:        ToFloat64(price),
		Amount:       ToFloat64(amount),
	}

	var resp struct {
		BaseResponse
		OrderID   string `json:"order_id"`
		ClientOid string `json:"client_oid"`
	}

	err := ok.DoRequest("POST", PLACE_ORDER, reqBody, &resp)
	if err != nil {
		logger.Errorf("[param] %s", param)
		return fOrder, err
	}

	if resp.ErrorMessage != "" {
		logger.Errorf("[param] %s", param)
		return fOrder, errors.New(fmt.Sprintf("%s:%s", resp.ErrorCode, resp.ErrorMessage))
	}

	fOrder.OrderID2 = resp.OrderID

	return fOrder, nil
}

func (ok *OKExSwap) LimitFuturesOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType int, opt ...types.LimitOrderOptionalParameter) (*types.FutureOrder, error) {
	return ok.PlaceFutureOrder2(currencyPair, contractType, price, amount, openType, 0, opt...)
}

func (ok *OKExSwap) MarketFuturesOrder(currencyPair types.CurrencyPair, contractType, amount string, openType int) (*types.FutureOrder, error) {
	return ok.PlaceFutureOrder2(currencyPair, contractType, "0", amount, openType, 1)
}

func (ok *OKExSwap) FutureCancelOrder(currencyPair types.CurrencyPair, contractType, orderId string) (bool, error) {
	var cancelParam struct {
		OrderId      string `json:"order_id"`
		InstrumentId string `json:"instrument_id"`
	}

	var resp SwapCancelOrderResult

	cancelParam.InstrumentId = contractType
	cancelParam.OrderId = orderId

	//req, _, _ := BuildRequestBody(cancelParam)

	err := ok.DoRequest("POST", fmt.Sprintf(CANCEL_ORDER, ok.adaptContractType(currencyPair), orderId), "", &resp)
	if err != nil {
		return false, err
	}

	return resp.Result, nil
}

func (ok *OKExSwap) GetFutureOrderHistory(pair types.CurrencyPair, contractType string, optional ...types.OptionalParameter) ([]types.FutureOrder, error) {
	urlPath := fmt.Sprintf("/api/swap/v3/orders/%s?", ok.adaptContractType(pair))

	param := url.Values{}
	param.Set("limit", "100")
	param.Set("state", "7")
	MergeOptionalParameter(&param, optional...)

	var response SwapOrdersInfo

	err := ok.DoRequest("GET", urlPath+param.Encode(), "", &response)
	if err != nil {
		return nil, err
	}

	orders := make([]types.FutureOrder, 0, 100)
	for _, info := range response.OrderInfo {
		ord := ok.parseOrder(info)
		ord.Currency = pair
		ord.ContractName = contractType
		orders = append(orders, ord)
	}

	return orders, nil
}

func (ok *OKExSwap) parseOrder(ord BaseOrderInfo) types.FutureOrder {
	oTime, _ := time.Parse(time.RFC3339, ord.Timestamp)
	return types.FutureOrder{
		ClientOid:  ord.ClientOid,
		OrderID2:   ord.OrderId,
		Amount:     ord.Size,
		Price:      ord.Price,
		DealAmount: ord.FilledQty,
		AvgPrice:   ord.PriceAvg,
		OType:      ord.Type,
		Status:     ok.AdaptTradeStatus(ord.Status),
		Fee:        ord.Fee,
		OrderTime:  oTime.UnixNano() / int64(time.Millisecond)}
}

func (ok *OKExSwap) GetUnfinishFutureOrders(currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error) {
	var (
		resp SwapOrdersInfo
	)
	contractType = ok.adaptContractType(currencyPair)
	err := ok.DoRequest("GET", fmt.Sprintf(GET_UNFINISHED_ORDERS, contractType, 6, 100), "", &resp)
	if err != nil {
		return nil, err
	}

	if resp.Message != "" {
		return nil, errors.New(fmt.Sprintf("{\"ErrCode\":%d,\"ErrMessage\":\"%s\"", resp.Code, resp.Message))
	}

	var orders []types.FutureOrder
	for _, info := range resp.OrderInfo {
		ord := ok.parseOrder(info)
		ord.Currency = currencyPair
		ord.ContractName = contractType
		orders = append(orders, ord)
	}

	//log.Println(len(orders))
	return orders, nil
}

/**
 *获取订单信息
 */
func (ok *OKExSwap) GetFutureOrders(orderIds []string, currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error) {
	panic("")
}

/**
 *获取单个订单信息
 */
func (ok *OKExSwap) GetFutureOrder(orderId string, currencyPair types.CurrencyPair, contractType string) (*types.FutureOrder, error) {
	var getOrderParam struct {
		OrderId      string `json:"order_id"`
		InstrumentId string `json:"instrument_id"`
	}

	var resp struct {
		BizWarmTips
		BaseOrderInfo
	}

	contractType = ok.adaptContractType(currencyPair)

	getOrderParam.OrderId = orderId
	getOrderParam.InstrumentId = contractType

	//reqBody, _, _ := BuildRequestBody(getOrderParam)

	err := ok.DoRequest("GET", fmt.Sprintf(GET_ORDER, contractType, orderId), "", &resp)
	if err != nil {
		return nil, err
	}

	if resp.Message != "" {
		return nil, errors.New(fmt.Sprintf("{\"ErrCode\":%d,\"ErrMessage\":\"%s\"}", resp.Code, resp.Message))
	}

	oTime, err := time.Parse(time.RFC3339, resp.Timestamp)

	return &types.FutureOrder{
		ClientOid:    resp.ClientOid,
		Currency:     currencyPair,
		ContractName: contractType,
		OrderID2:     resp.OrderId,
		Amount:       resp.Size,
		Price:        resp.Price,
		DealAmount:   resp.FilledQty,
		AvgPrice:     resp.PriceAvg,
		OType:        resp.Type,
		Fee:          resp.Fee,
		Status:       ok.AdaptTradeStatus(resp.Status),
		OrderTime:    oTime.UnixNano() / int64(time.Millisecond),
	}, nil
}

func (ok *OKExSwap) GetFuturePosition(currencyPair types.CurrencyPair, contractType string) ([]types.FuturePosition, error) {
	var resp SwapPosition
	contractType = ok.adaptContractType(currencyPair)

	err := ok.DoRequest("GET", fmt.Sprintf(GET_POSITION, contractType), "", &resp)
	if err != nil {
		return nil, err
	}

	var positions []types.FuturePosition

	positions = append(positions, types.FuturePosition{
		ContractType: contractType,
		Symbol:       currencyPair})

	var (
		buyPosition  SwapPositionHolding
		sellPosition SwapPositionHolding
	)

	if len(resp.Holding) > 0 {
		if resp.Holding[0].Side == "long" {
			buyPosition = resp.Holding[0]
			if len(resp.Holding) == 2 {
				sellPosition = resp.Holding[1]
			}
		} else {
			sellPosition = resp.Holding[0]
			if len(resp.Holding) == 2 {
				buyPosition = resp.Holding[1]
			}
		}

		positions[0].ForceLiquPrice = buyPosition.LiquidationPrice
		positions[0].BuyAmount = buyPosition.Position
		positions[0].BuyAvailable = buyPosition.AvailPosition
		positions[0].BuyPriceAvg = buyPosition.AvgCost
		positions[0].BuyProfitReal = buyPosition.RealizedPnl
		positions[0].BuyPriceCost = buyPosition.SettlementPrice

		positions[0].ForceLiquPrice = sellPosition.LiquidationPrice
		positions[0].SellAmount = sellPosition.Position
		positions[0].SellAvailable = sellPosition.AvailPosition
		positions[0].SellPriceAvg = sellPosition.AvgCost
		positions[0].SellProfitReal = sellPosition.RealizedPnl
		positions[0].SellPriceCost = sellPosition.SettlementPrice

		positions[0].LeverRate = ToFloat64(sellPosition.Leverage)
	}
	return positions, nil
}

/**
 * BTC: 100美元一张合约
 * LTC/ETH/ETC/BCH: 10美元一张合约
 */
func (ok *OKExSwap) GetContractValue(currencyPair types.CurrencyPair) (float64, error) {
	if currencyPair.CurrencyA.Eq(types.BTC) {
		return 100, nil
	}
	return 10, nil
}

func (ok *OKExSwap) GetFee() (float64, error) {
	panic("not support")
}

func (ok *OKExSwap) GetFutureEstimatedPrice(currencyPair types.CurrencyPair) (float64, error) {
	panic("not support")
}

func (ok *OKExSwap) GetFutureIndex(currencyPair types.CurrencyPair) (float64, error) {
	panic("not support")
}

func (ok *OKExSwap) GetDeliveryTime() (int, int, int, int) {
	panic("not support")
}

func (ok *OKExSwap) GetKlineRecords(contractType string, currency types.CurrencyPair, period types.KlinePeriod, size int, opt ...types.OptionalParameter) ([]types.FutureKline, error) {
	granularity := adaptKLinePeriod(types.KlinePeriod(period))
	if granularity == -1 {
		return nil, errors.New("kline period parameter is error")
	}
	return ok.GetKlineRecords2(contractType, currency, "", "", strconv.Itoa(granularity))
}

/**
  since : 单位秒,开始时间
  to : 单位秒,结束时间
*/
func (ok *OKExSwap) GetKlineRecordsByRange(currency types.CurrencyPair, period, since, to int) ([]types.FutureKline, error) {
	urlPath := "/api/swap/v3/instruments/%s/candles?start=%s&end=%s&granularity=%d"
	sinceTime := time.Unix(int64(since), 0).UTC().Format(time.RFC3339)
	toTime := time.Unix(int64(to), 0).UTC().Format(time.RFC3339)
	contractId := ok.adaptContractType(currency)
	granularity := adaptKLinePeriod(types.KlinePeriod(period))
	if granularity == -1 {
		return nil, errors.New("kline period parameter is error")
	}

	var response [][]interface{}
	err := ok.DoRequest("GET", fmt.Sprintf(urlPath, contractId, sinceTime, toTime, granularity), "", &response)
	if err != nil {
		return nil, err
	}

	var klines []types.FutureKline
	for _, itm := range response {
		t, _ := time.Parse(time.RFC3339, fmt.Sprint(itm[0]))
		klines = append(klines, types.FutureKline{
			Kline: &types.Kline{
				Timestamp: t.Unix(),
				Pair:      currency,
				Open:      ToFloat64(itm[1]),
				High:      ToFloat64(itm[2]),
				Low:       ToFloat64(itm[3]),
				Close:     ToFloat64(itm[4]),
				Vol:       ToFloat64(itm[5])},
			Vol2: ToFloat64(itm[6])})
	}

	return klines, nil
}

/**
  since : 单位秒,开始时间
*/
func (ok *OKExSwap) GetKlineRecords2(contractType string, currency types.CurrencyPair, start, end, period string) ([]types.FutureKline, error) {
	urlPath := "/api/swap/v3/instruments/%s/candles?%s"
	params := url.Values{}
	if start != "" {
		params.Set("start", start)
	}
	if end != "" {
		params.Set("end", end)
	}
	if period != "" {
		params.Set("granularity", period)
	}
	contractId := ok.adaptContractType(currency)

	var response [][]interface{}
	err := ok.DoRequest("GET", fmt.Sprintf(urlPath, contractId, params.Encode()), "", &response)
	if err != nil {
		return nil, err
	}

	var kline []types.FutureKline
	for _, itm := range response {
		t, _ := time.Parse(time.RFC3339, fmt.Sprint(itm[0]))
		kline = append(kline, types.FutureKline{
			Kline: &types.Kline{
				Timestamp: t.Unix(),
				Pair:      currency,
				Open:      ToFloat64(itm[1]),
				High:      ToFloat64(itm[2]),
				Low:       ToFloat64(itm[3]),
				Close:     ToFloat64(itm[4]),
				Vol:       ToFloat64(itm[5])},
			Vol2: ToFloat64(itm[6])})
	}

	return kline, nil
}

func (ok *OKExSwap) GetTrades(contractType string, currencyPair types.CurrencyPair, since int64) ([]types.Trade, error) {
	panic("not support")
}

func (ok *OKExSwap) GetExchangeRate() (float64, error) {
	panic("not support")
}

func (ok *OKExSwap) GetHistoricalFunding(contractType string, currencyPair types.CurrencyPair, page int) ([]types.HistoricalFunding, error) {
	var resp []types.HistoricalFunding
	uri := fmt.Sprintf("/api/swap/v3/instruments/%s/historical_funding_rate?from=%d", ok.adaptContractType(currencyPair), page)
	err := ok.DoRequest("GET", uri, "", &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (ok *OKExSwap) AdaptTradeStatus(status int) types.TradeStatus {
	switch status {
	case -1:
		return types.ORDER_CANCEL
	case 0:
		return types.ORDER_UNFINISH
	case 1:
		return types.ORDER_PART_FINISH
	case 2:
		return types.ORDER_FINISH
	default:
		return types.ORDER_UNFINISH
	}
}

func (ok *OKExSwap) adaptContractType(currencyPair types.CurrencyPair) string {
	return fmt.Sprintf("%s-SWAP", currencyPair.ToSymbol("-"))
}

type Instrument struct {
	InstrumentID        string    `json:"instrument_id"`
	UnderlyingIndex     string    `json:"underlying_index"`
	QuoteCurrency       string    `json:"quote_currency"`
	Coin                string    `json:"coin"`
	ContractVal         float64   `json:"contract_val,string"`
	Listing             time.Time `json:"listing"`
	Delivery            time.Time `json:"delivery"`
	SizeIncrement       int       `json:"size_increment,string"`
	TickSize            float64   `json:"tick_size,string"`
	BaseCurrency        string    `json:"base_currency"`
	Underlying          string    `json:"underlying"`
	SettlementCurrency  string    `json:"settlement_currency"`
	IsInverse           bool      `json:"is_inverse,string"`
	ContractValCurrency string    `json:"contract_val_currency"`
}

func (ok *OKExSwap) GetInstruments() ([]Instrument, error) {
	var resp []Instrument
	err := ok.DoRequest("GET", "/api/swap/v3/instruments", "", &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type MarginLeverage struct {
	LongLeverage  float64 `json:"long_leverage,string"`
	MarginMode    string  `json:"margin_mode"`
	ShortLeverage float64 `json:"short_leverage,string"`
	InstrumentId  string  `json:"instrument_id"`
}

func (ok *OKExSwap) GetMarginLevel(currencyPair types.CurrencyPair) (*MarginLeverage, error) {
	var resp MarginLeverage
	uri := fmt.Sprintf("/api/swap/v3/accounts/%s/settings", ok.adaptContractType(currencyPair))

	err := ok.DoRequest("GET", uri, "", &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil

}

// marginmode
//1:逐仓-多仓
//2:逐仓-空仓
//3:全仓
func (ok *OKExSwap) SetMarginLevel(currencyPair types.CurrencyPair, level, marginMode int) (*MarginLeverage, error) {
	var resp MarginLeverage
	uri := fmt.Sprintf("/api/swap/v3/accounts/%s/leverage", ok.adaptContractType(currencyPair))

	reqBody := make(map[string]string)
	reqBody["leverage"] = strconv.Itoa(level)
	reqBody["side"] = strconv.Itoa(marginMode)
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	err = ok.DoRequest("POST", uri, string(data), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

//委托策略下单 algo_type 1:限价 2:市场价；触发价格类型，默认是限价；为市场价时，委托价格不必填；
func (ok *OKExSwap) PlaceFutureAlgoOrder(ord *types.FutureOrder) (*types.FutureOrder, error) {
	var param struct {
		InstrumentId string `json:"instrument_id"`
		Type         int    `json:"type"`
		OrderType    int    `json:"order_type"` //1：止盈止损 2：跟踪委托 3：冰山委托 4：时间加权
		Size         string `json:"size"`
		TriggerPrice string `json:"trigger_price"`
		AlgoPrice    string `json:"algo_price"`
		AlgoType     string `json:"algo_type"`
	}

	var response struct {
		ErrorMessage string `json:"error_message"`
		ErrorCode    string `json:"error_code"`
		DetailMsg    string `json:"detail_msg"`

		Data struct {
			Result       string `json:"result"`
			ErrorMessage string `json:"error_message"`
			ErrorCode    string `json:"error_code"`
			AlgoId       string `json:"algo_id"`
			InstrumentId string `json:"instrument_id"`
			OrderType    int    `json:"order_type"`
		} `json:"data"`
	}
	if ord == nil {
		return nil, errors.New("ord param is nil")
	}
	param.InstrumentId = ok.adaptContractType(ord.Currency)
	param.Type = ord.OType
	param.OrderType = ord.OrderType
	param.AlgoType = fmt.Sprint(ord.AlgoType)
	param.TriggerPrice = fmt.Sprint(ord.TriggerPrice)
	param.AlgoPrice = fmt.Sprint(ToFloat64(ord.Price))
	param.Size = fmt.Sprint(ord.Amount)

	reqBody, _, _ := ok.BuildRequestBody(param)
	err := ok.DoRequest("POST", PLACE_ALGO_ORDER, reqBody, &response)

	if err != nil {
		return ord, err
	}

	ord.OrderID2 = response.Data.AlgoId
	ord.OrderTime = time.Now().UnixNano() / int64(time.Millisecond)

	return ord, nil
}

//委托策略撤单
func (ok *OKExSwap) FutureCancelAlgoOrder(currencyPair types.CurrencyPair, orderId []string) (bool, error) {
	if len(orderId) == 0 {
		return false, errors.New("invalid order id")
	}
	var cancelParam struct {
		InstrumentId string   `json:"instrument_id"`
		AlgoIds      []string `json:"algo_ids"`
		OrderType    string   `json:"order_type"`
	}

	var resp struct {
		ErrorMessage string `json:"error_message"`
		ErrorCode    string `json:"error_code"`
		DetailMsg    string `json:"detailMsg"`
		Data         struct {
			Result       string `json:"result"`
			AlgoIds      string `json:"algo_ids"`
			InstrumentID string `json:"instrument_id"`
			OrderType    string `json:"order_type"`
		} `json:"data"`
	}

	cancelParam.InstrumentId = ok.adaptContractType(currencyPair)
	cancelParam.OrderType = "1"
	cancelParam.AlgoIds = orderId

	reqBody, _, _ := ok.BuildRequestBody(cancelParam)

	err := ok.DoRequest("POST", CANCEL_ALGO_ORDER, reqBody, &resp)
	if err != nil {
		return false, err
	}

	return resp.Data.Result == "success", nil
}

//获取委托单列表, status和algo_id必填且只能填其一
func (ok *OKExSwap) GetFutureAlgoOrders(algo_id string, status string, currencyPair types.CurrencyPair) ([]types.FutureOrder, error) {
	uri := fmt.Sprintf(GET_ALGO_ORDER, ok.adaptContractType(currencyPair), 1)
	if algo_id != "" {
		uri += "algo_id=" + algo_id
	} else if status != "" {
		uri += "status=" + status
	} else {
		return nil, errors.New("status or algo_id is needed")
	}

	var resp struct {
		OrderStrategyVOS []struct {
			AlgoId       string `json:"algo_id"`
			AlgoPrice    string `json:"algo_price"`
			InstrumentId string `json:"instrument_id"`
			Leverage     string `json:"leverage"`
			OrderType    string `json:"order_type"`
			RealAmount   string `json:"real_amount"`
			RealPrice    string `json:"real_price"`
			Size         string `json:"size"`
			Status       string `json:"status"`
			Timestamp    string `json:"timestamp"`
			TriggerPrice string `json:"trigger_price"`
			Type         string `json:"type"`
		} `json:"orderStrategyVOS"`
	}

	err := ok.DoRequest("GET", uri, "", &resp)
	if err != nil {
		return nil, err
	}

	var orders []types.FutureOrder
	for _, info := range resp.OrderStrategyVOS {
		oTime, _ := time.Parse(time.RFC3339, info.Timestamp)

		ord := types.FutureOrder{
			OrderID2:     info.AlgoId,
			Price:        ToFloat64(info.AlgoPrice),
			Amount:       ToFloat64(info.Size),
			AvgPrice:     ToFloat64(info.RealPrice),
			DealAmount:   ToFloat64(info.RealAmount),
			OrderTime:    oTime.UnixNano() / int64(time.Millisecond),
			Status:       ok.AdaptTradeStatus(ToInt(info.Status)),
			Currency:     types.CurrencyPair{},
			OrderType:    ToInt(info.OrderType),
			OType:        ToInt(info.Type),
			TriggerPrice: ToFloat64(info.TriggerPrice),
		}
		ord.Currency = currencyPair
		orders = append(orders, ord)
	}

	//log.Println(len(orders))
	return orders, nil
}
