package bitmex

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fpChan/goex/common/api"
	"github.com/fpChan/goex/common/exchange"
	"github.com/fpChan/goex/types"
	"net/url"
	"strings"
	"time"

	. "github.com/fpChan/goex/internal/logger"

	. "github.com/fpChan/goex"
)

const (
	baseUrl = "https://www.bitmex.com"
)

type bitmex struct {
	*types.APIConfig
}

func (bm *bitmex) GetFutureOrderHistory(pair types.CurrencyPair, contractType string, optional ...types.OptionalParameter) ([]types.FutureOrder, error) {
	panic("implement me")
}

func New(config *types.APIConfig) *bitmex {
	bm := &bitmex{config}
	if bm.Endpoint == "" {
		bm.Endpoint = baseUrl
	}
	if strings.HasSuffix(bm.Endpoint, "/") {
		bm.Endpoint = bm.Endpoint[0 : len(bm.Endpoint)-1]
	}
	Log.Debug("endpoint=", bm.Endpoint)
	return bm
}

func (bm *bitmex) generateSignature(httpMethod, uri, data, nonce string) string {
	payload := strings.ToUpper(httpMethod) + uri + nonce + data
	//println(payload)
	sign, _ := GetParamHmacSHA256Sign(bm.ApiSecretKey, payload)
	//println(sign)
	return sign
}

func (bm *bitmex) doAuthRequest(m, uri, param string, r interface{}) error {

	nonce := time.Now().UTC().Unix() + 3600
	sign := bm.generateSignature(m, uri, param, fmt.Sprint(nonce))

	resp, err := api.NewHttpRequest(bm.HttpClient, m, bm.Endpoint+uri, param, map[string]string{
		"User-Agent":    "github.com/fpChan/goex/bitmex",
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"api-expires":   fmt.Sprint(nonce),
		"api-key":       bm.ApiKey,
		"api-signature": sign})
	Log.Debug("response:", string(resp))
	if err != nil {
		return err
	} else {
		//println(string(resp))
		return json.Unmarshal(resp, &r)
	}
	return nil
}

func (bm *bitmex) toJson(param interface{}) string {
	dataJson, _ := json.Marshal(param)
	return string(dataJson)
}

func (bm *bitmex) GetFutureUserinfo(currencyPair ...types.CurrencyPair) (*types.FutureAccount, error) {
	uri := "/api/v1/user/margin?currency=XBt"
	var resp struct {
		Currency           string  `json:"currency"`
		RiskLimit          float64 `json:"riskLimit"`
		Amount             float64 `json:"amount"`
		MarginBalance      float64 `json:"marginBalance"`
		WalletBalance      float64 `json:"walletBalance"`
		AvailableMargin    float64 `json:"availableMargin"`
		WithdrawableMargin float64 `json:"withdrawableMargin"`
		InitMargin         float64 `json:"initMargin"`
		UnrealisedProfit   float64 `json:"unrealisedProfit"`
		UnrealisedPnl      float64 `json:"unrealisedPnl"`
		RealisedPnl        float64 `json:"realisedPnl"`
		RiskValue          float64 `json:"riskValue"`
	}

	err := bm.doAuthRequest("GET", uri, "", &resp)
	if err != nil {
		return nil, err
	}

	futureAcc := new(types.FutureAccount)
	futureAcc.FutureSubAccounts = make(map[types.Currency]types.FutureSubAccount, 1)
	futureAcc.FutureSubAccounts[types.BTC] = types.FutureSubAccount{
		Currency:      types.BTC,
		AccountRights: resp.MarginBalance / 100000000,
		KeepDeposit:   resp.InitMargin / 100000000,
		ProfitUnreal:  resp.UnrealisedPnl / 100000000,
		ProfitReal:    resp.RealisedPnl / 100000000,
		RiskRate:      resp.RiskValue}

	return futureAcc, nil
}

type BitmexOrder struct {
	Symbol      string    `json:"symbol"`
	OrderID     string    `json:"OrderID"`
	ClOrdID     string    `json:"clOrdID"`
	Price       float64   `json:"price,omitempty"`
	OrderQty    int       `json:"orderQty"`
	CumQty      int       `json:"cumQty"`
	AvgPx       float64   `json:"avgPx"`
	OrdType     string    `json:"ordType"`
	Text        string    `json:"text"`
	TimeInForce string    `json:"timeInForce,omitempty"`
	Side        string    `json:"side"`
	OrdStatus   string    `json:"ordStatus"`
	Timestamp   time.Time `json:"timestamp"`
}

func (bm *bitmex) PlaceFutureOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType, matchPrice int, leverRate float64) (string, error) {
	fOrder, err := bm.PlaceFutureOrder2(currencyPair, contractType, price, amount, openType, matchPrice, leverRate)
	return fOrder.OrderID2, err
}

func (bm *bitmex) PlaceFutureOrder2(currencyPair types.CurrencyPair, contractType, price, amount string, openType, matchPrice int, leverRate float64) (*types.FutureOrder, error) {
	var createOrderParameter BitmexOrder

	var resp struct {
		OrderId string `json:"orderID"`
	}

	createOrderParameter.Text = "github.com/fpChan/goex/tree/master/bitmex"
	createOrderParameter.Symbol = bm.adaptCurrencyPairToSymbol(currencyPair, contractType)
	createOrderParameter.OrdType = "Limit"
	createOrderParameter.TimeInForce = "GoodTillCancel"
	createOrderParameter.ClOrdID = GenerateOrderClientId(32)
	createOrderParameter.OrderQty = ToInt(amount)

	if matchPrice == 0 {
		createOrderParameter.Price = ToFloat64(price)
	} else {
		createOrderParameter.OrdType = "Market"
	}

	switch openType {
	case types.OPEN_BUY, types.CLOSE_SELL:
		createOrderParameter.Side = "Buy"
	case types.OPEN_SELL, types.CLOSE_BUY:
		createOrderParameter.Side = "Sell"
	}

	//if openType == CLOSE_BUY || openType == CLOSE_SELL {
	//	createOrderParameter.OrderQty = -ToInt(amount)
	//}

	fOrder := &types.FutureOrder{
		ClientOid:    createOrderParameter.ClOrdID,
		Currency:     currencyPair,
		Price:        ToFloat64(price),
		Amount:       ToFloat64(amount),
		OType:        openType,
		LeverRate:    leverRate,
		ContractName: contractType,
	}

	err := bm.doAuthRequest("POST", "/api/v1/order", bm.toJson(createOrderParameter), &resp)

	if err != nil {
		return fOrder, err
	}

	fOrder.OrderID2 = resp.OrderId

	return fOrder, nil
}

func (bm *bitmex) LimitFuturesOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType int, opt ...types.LimitOrderOptionalParameter) (*types.FutureOrder, error) {
	return bm.PlaceFutureOrder2(currencyPair, contractType, price, amount, openType, 0, 10)
}

func (bm *bitmex) MarketFuturesOrder(currencyPair types.CurrencyPair, contractType, amount string, openType int) (*types.FutureOrder, error) {
	return bm.PlaceFutureOrder2(currencyPair, contractType, "0", amount, openType, 1, 10)
}

func (bm *bitmex) FutureCancelOrder(currencyPair types.CurrencyPair, contractType, orderId string) (bool, error) {
	var param struct {
		OrderID string `json:"orderID,omitempty"`
		ClOrdID string `json:"clOrdID,omitempty"`
	}
	if strings.HasPrefix(orderId, "goex") {
		param.ClOrdID = orderId
	} else {
		param.OrderID = orderId
	}
	var response []interface{}
	err := bm.doAuthRequest("DELETE", "/api/v1/order", bm.toJson(param), &response)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (bm *bitmex) GetFuturePosition(currencyPair types.CurrencyPair, contractType string) ([]types.FuturePosition, error) {
	var (
		response []struct {
			Symbol            string    `json:"symbol"`
			CurrentQty        int       `json:"currentQty"`
			OpeningQty        int       `json:"openingQty"`
			AvgCostPrice      float64   `json:"avgCostPrice"`
			AvgEntryPrice     float64   `json:"avgEntryPrice"`
			UnrealisedPnl     float64   `json:"unrealisedPnl"`
			UnrealisedPnlPcnt float64   `json:"unrealisedPnlPcnt"`
			OpenOrderBuyQty   float64   `json:"openOrderBuyQty"`
			OpenOrderSellQty  float64   `json:"OpenOrderSellQty"`
			OpeningTimestamp  time.Time `json:"openingTimestamp"`
			LiquidationPrice  float64   `json:"liquidationPrice"`
			Leverage          float64   `json:"leverage"`
		}
		param = url.Values{}
	)
	param.Set("filter", fmt.Sprintf(`{"symbol":"%s"}`, bm.adaptCurrencyPairToSymbol(currencyPair, contractType)))
	er := bm.doAuthRequest("GET", "/api/v1/position?"+param.Encode(), "", &response)
	if er != nil {
		return nil, er
	}

	var postions []types.FuturePosition
	for _, p := range response {
		pos := types.FuturePosition{}
		pos.Symbol = currencyPair
		pos.ContractType = contractType
		pos.CreateDate = p.OpeningTimestamp.Unix()
		pos.ForceLiquPrice = p.LiquidationPrice
		pos.LeverRate = p.Leverage

		if p.CurrentQty < 0 {
			pos.SellAmount = float64(-p.CurrentQty)
			pos.SellAvailable = pos.SellAmount - p.OpenOrderBuyQty
			pos.SellPriceCost = p.AvgCostPrice
			pos.SellPriceAvg = p.AvgEntryPrice
			pos.SellProfitReal = p.UnrealisedPnlPcnt
		} else {
			pos.BuyAmount = float64(p.CurrentQty)
			pos.BuyPriceCost = p.AvgCostPrice
			pos.BuyPriceAvg = p.AvgEntryPrice
			pos.BuyProfitReal = p.UnrealisedPnlPcnt
			pos.BuyAvailable = pos.BuyAmount - p.OpenOrderSellQty
		}

		postions = append(postions, pos)
	}

	return postions, nil
}

func (bm *bitmex) GetFutureOrders(orderIds []string, currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error) {
	panic("no support")
}

func (bm *bitmex) GetFutureOrder(orderId string, currencyPair types.CurrencyPair, contractType string) (*types.FutureOrder, error) {
	var response []BitmexOrder
	filters := fmt.Sprintf(`{"orderID":"%s"}`, orderId)
	param := url.Values{}
	param.Set("symbol", bm.adaptCurrencyPairToSymbol(currencyPair, contractType))
	param.Set("filter", filters)
	uri := "/api/v1/order?" + param.Encode()
	err := bm.doAuthRequest("GET", uri, "", &response)
	if err != nil {
		return nil, err
	}
	if len(response) == 0 {
		return nil, errors.New("not find order")
	}
	ord := bm.adaptOrder(response[0])
	ord.ContractName = contractType
	ord.Currency = currencyPair
	return &ord, nil
}

func (bm *bitmex) GetUnfinishFutureOrders(currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error) {
	var response []BitmexOrder

	query := url.Values{}
	query.Set("symbol", bm.adaptCurrencyPairToSymbol(currencyPair, contractType))
	query.Set("filter", "{\"open\":true}")
	uri := "/api/v1/order?" + query.Encode()
	errr := bm.doAuthRequest("GET", uri, "", &response)
	if errr != nil {
		return nil, errr
	}

	var orders []types.FutureOrder
	for _, v := range response {
		ord := bm.adaptOrder(v)
		ord.Currency = currencyPair
		ord.ContractName = contractType
		orders = append(orders, ord)
	}

	return orders, nil
}

func (bm *bitmex) GetFee() (float64, error) {
	panic("no support")
}

func (bm *bitmex) GetFutureDepth(currencyPair types.CurrencyPair, contractType string, size int) (*types.Depth, error) {
	sym := bm.adaptCurrencyPairToSymbol(currencyPair, contractType)
	uri := fmt.Sprintf("/api/v1/orderBook/L2?symbol=%s&depth=%d", sym, size)

	resp, err := api.HttpGet3(bm.HttpClient, bm.Endpoint+uri, nil)
	if err != nil {
		return nil, exchange.HTTP_ERR_CODE.OriginErr(err.Error())
	}

	//log.Println(resp)

	dep := new(types.Depth)
	dep.UTime = time.Now()
	dep.Pair = currencyPair
	dep.ContractType = sym

	for _, r := range resp {
		rr := r.(map[string]interface{})
		switch strings.ToLower(rr["side"].(string)) {
		case "sell":
			dep.AskList = append(dep.AskList, types.DepthRecord{Price: ToFloat64(rr["price"]), Amount: ToFloat64(rr["size"])})
		case "buy":
			dep.BidList = append(dep.BidList, types.DepthRecord{Price: ToFloat64(rr["price"]), Amount: ToFloat64(rr["size"])})
		}
	}

	return dep, nil
}

func (bm *bitmex) GetFutureTicker(currencyPair types.CurrencyPair, contractType string) (*types.Ticker, error) {
	uri := fmt.Sprintf("/api/v1/instrument?symbol=%s", bm.adaptCurrencyPairToSymbol(currencyPair, contractType))
	resp, err := api.HttpGet3(bm.HttpClient, bm.Endpoint+uri, nil)
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, errors.New("get ticker response is null")
	}

	tickermap, isok := resp[0].(map[string]interface{})
	if !isok {
		return nil, errors.New(fmt.Sprintf("response format error [%s]", resp[0]))
	}

	date, _ := time.Parse(time.RFC3339, tickermap["timestamp"].(string))

	return &types.Ticker{
		Pair: currencyPair,
		Last: ToFloat64(tickermap["lastPrice"]),
		High: ToFloat64(tickermap["highPrice"]),
		Low:  ToFloat64(tickermap["lowPrice"]),
		Vol:  ToFloat64(tickermap["homeNotional24h"]),
		Sell: ToFloat64(tickermap["askPrice"]),
		Buy:  ToFloat64(tickermap["bidPrice"]),
		Date: uint64(date.Unix()),
	}, nil
}

func (bm *bitmex) GetIndicativeFundingRate(symbol string) (float64, *time.Time, error) {
	//indicativeFundingRate
	uri := fmt.Sprintf("/api/v1/instrument?symbol=%s", symbol)
	resp, err := api.HttpGet3(bm.HttpClient, bm.Endpoint+uri, nil)
	if err != nil {
		return 0, nil, err
	}

	if len(resp) == 0 {
		return 0, nil, errors.New(" response is null")
	}

	retmap, isok := resp[0].(map[string]interface{})
	if !isok {
		return 0, nil, errors.New(fmt.Sprintf("response format error [%s]", resp[0]))
	}

	t, _ := time.Parse(time.RFC3339, retmap["fundingTimestamp"].(string))

	return ToFloat64(retmap["indicativeFundingRate"]), &t, nil
}

func (bm *bitmex) GetExchangeName() string {
	return types.BITMEX
}

func (bm *bitmex) GetFutureIndex(currencyPair types.CurrencyPair) (float64, error) {
	panic("no support")
}

func (bm *bitmex) GetContractValue(currencyPair types.CurrencyPair) (float64, error) {
	return 1.0, nil
}

func (bm *bitmex) GetDeliveryTime() (int, int, int, int) {
	panic("no support")
}

func (bm *bitmex) GetFutureEstimatedPrice(currencyPair types.CurrencyPair) (float64, error) {
	panic("no support")
}

func (bm *bitmex) GetKlineRecords(contract_type string, currency types.CurrencyPair, period types.KlinePeriod, size int, optional ...types.OptionalParameter) ([]types.FutureKline, error) {
	urlPath := "/api/v1/trade/bucketed?binSize=%s&partial=false&symbol=%s&count=%d&startTime=%s&reverse=true"
	contractId := bm.adaptCurrencyPairToSymbol(currency, contract_type)

	var granularity string
	switch period {
	case types.KLINE_PERIOD_1MIN:
		granularity = "1m"
	case types.KLINE_PERIOD_5MIN:
		granularity = "5m"
	case types.KLINE_PERIOD_1H, types.KLINE_PERIOD_60MIN:
		granularity = "1h"
	case types.KLINE_PERIOD_1DAY:
		granularity = "1d"
	default:
		granularity = "5m"
	}

	sinceTime := time.Now()
	if len(optional) > 0 && optional[0].GetTime("startTime") != nil {
		sinceTime = *optional[0].GetTime("startTime")
	}

	uri := fmt.Sprintf(urlPath, granularity, contractId, size, sinceTime.Format(time.RFC3339))
	response, err := api.HttpGet3(bm.HttpClient, bm.Endpoint+uri, nil)
	if err != nil {
		return nil, err
	}

	var klines []types.FutureKline
	for _, record := range response {
		r := record.(map[string]interface{})
		t, _ := time.Parse(time.RFC3339, fmt.Sprint(r["timestamp"]))
		klines = append(klines, types.FutureKline{
			Kline: &types.Kline{
				Timestamp: t.Unix(),
				Pair:      currency,
				Open:      ToFloat64(r["open"]),
				High:      ToFloat64(r["high"]),
				Low:       ToFloat64(r["low"]),
				Close:     ToFloat64(r["close"]),
				Vol:       ToFloat64(r["volume"])}})
	}

	return klines, nil
}

func (bm *bitmex) GetTrades(contract_type string, currency types.CurrencyPair, since int64) ([]types.Trade, error) {
	var urlPath = "/api/v1/trade?symbol=%s&startTime=%s&reverse=true"
	contractId := bm.adaptCurrencyPairToSymbol(currency, contract_type)
	sinceTime := time.Unix(int64(since), 0).UTC()

	if since/int64(time.Second) != 1 {
		sinceTime = time.Unix(int64(since)/int64(time.Second), 0).UTC()
	}

	uri := fmt.Sprintf(urlPath, contractId, sinceTime.Format(time.RFC3339))
	response, err := api.HttpGet3(bm.HttpClient, bm.Endpoint+uri, nil)
	if err != nil {
		return nil, err
	}

	trades := make([]types.Trade, 0)
	for _, v := range response {
		vv := v.(map[string]interface{})
		side := types.BUY
		if vv["side"] == "Sell" {
			side = types.SELL
		}
		timestamp, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", vv["timestamp"]))
		trades = append(trades, types.Trade{
			Tid:    ToInt64(vv["trdMatchID"]),
			Type:   side,
			Amount: ToFloat64(vv["size"]),
			Price:  ToFloat64(vv["price"]),
			Date:   timestamp.Unix(),
			Pair:   currency,
		})
	}

	return trades, nil
}

func (bm *bitmex) adaptCurrencyPairToSymbol(pair types.CurrencyPair, contract string) string {
	if contract == "" || contract == types.SWAP_CONTRACT {
		if pair.CurrencyA.Eq(types.BTC) {
			pair = types.NewCurrencyPair(types.XBT, types.USD)
		}
		if pair.CurrencyB.Eq(types.BTC) {
			pair = types.NewCurrencyPair(pair.CurrencyA, types.XBT)
		}
		return pair.AdaptUsdtToUsd().ToSymbol("")
	}

	coin := pair.CurrencyA.Symbol
	if pair.CurrencyA.Eq(types.BTC) {
		coin = types.XBT.Symbol
	}
	return fmt.Sprintf("%s%s", coin, strings.ToUpper(contract))
}

func (bm *bitmex) adaptOrder(o BitmexOrder) types.FutureOrder {
	status := types.ORDER_UNFINISH
	if o.OrdStatus == "Filled" {
		status = types.ORDER_FINISH
	} else if o.OrdStatus == "Canceled" {
		status = types.ORDER_CANCEL
	}
	return types.FutureOrder{
		OrderID2:   o.OrderID,
		ClientOid:  o.ClOrdID,
		Amount:     float64(o.OrderQty),
		Price:      o.Price,
		DealAmount: float64(o.CumQty),
		AvgPrice:   o.AvgPx,
		Status:     status,
		OrderTime:  o.Timestamp.Unix()}
}
