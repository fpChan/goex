package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fpChan/goex/common/api"
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type BaseResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type AccountResponse struct {
	FeeTier  int  `json:"feeTier"`
	CanTrade bool `json:"canTrade"`
	Assets   []struct {
		Asset            string  `json:"asset"`
		WalletBalance    float64 `json:"walletBalance,string"`
		MarginBalance    float64 `json:"marginBalance,string"`
		UnrealizedProfit float64 `json:"unrealizedProfit,string"`
		MaintMargin      float64 `json:"maintMargin,string"`
	} `json:"assets"`
}

type OrderInfoResponse struct {
	BaseResponse
	Symbol        string  `json:"symbol"`
	Pair          string  `json:"pair"`
	ClientOrderId string  `json:"clientOrderId"`
	OrderId       int64   `json:"orderId"`
	AvgPrice      float64 `json:"avgPrice,string"`
	ExecutedQty   float64 `json:"executedQty,string"`
	OrigQty       float64 `json:"origQty,string"`
	Price         float64 `json:"price,string"`
	Side          string  `json:"side"`
	PositionSide  string  `json:"positionSide"`
	Status        string  `json:"status"`
	Type          string  `json:"type"`
	Time          int64   `json:"time"`
	UpdateTime    int64   `json:"updateTime"`
}

type PositionRiskResponse struct {
	Symbol           string  `json:"symbol"`
	PositionAmt      float64 `json:"positionAmt,string"`
	EntryPrice       float64 `json:"entryPrice,string"`
	UnRealizedProfit float64 `json:"unRealizedProfit,string"`
	LiquidationPrice float64 `json:"liquidationPrice,string"`
	Leverage         float64 `json:"leverage,string"`
	MarginType       string  `json:"marginType"`
	PositionSide     string  `json:"positionSide"`
}

type SymbolInfo struct {
	Symbol         string
	Pair           string
	ContractType   string `json:"contractType"`
	DeliveryDate   int64  `json:"deliveryDate"`
	ContractStatus string `json:"contractStatus"`
	ContractSize   int    `json:"contractSize"`
	PricePrecision int    `json:"pricePrecision"`
}

type BinanceFutures struct {
	base         *Binance
	apikey       string
	exchangeInfo *struct {
		Symbols []SymbolInfo `json:"symbols"`
	}
}

func NewBinanceFutures(config *types.APIConfig) *BinanceFutures {
	if config.Endpoint == "" {
		config.Endpoint = "https://dapi.binance.com"
	}

	if config.HttpClient == nil {
		config.HttpClient = http.DefaultClient
	}

	bs := &BinanceFutures{
		apikey: config.ApiKey,
		base:   NewWithConfig(config),
	}

	bs.base.apiV1 = config.Endpoint + "/dapi/v1/"

	go bs.GetExchangeInfo()

	return bs
}

func (bs *BinanceFutures) SetBaseUri(uri string) {
	bs.base.baseUrl = uri
}

func (bs *BinanceFutures) GetExchangeName() string {
	return types.BINANCE_FUTURES
}

func (bs *BinanceFutures) GetFutureTicker(currencyPair types.CurrencyPair, contractType string) (*types.Ticker, error) {
	symbol, err := bs.adaptToSymbol(currencyPair, contractType)
	if err != nil {
		return nil, err
	}

	ticker24hrUri := bs.base.apiV1 + "ticker/24hr?symbol=" + symbol
	tickerBookUri := bs.base.apiV1 + "ticker/bookTicker?symbol=" + symbol

	var (
		ticker24HrResp []interface{}
		tickerBookResp []interface{}
		err1           error
		err2           error
		wg             = sync.WaitGroup{}
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		ticker24HrResp, err1 = api.HttpGet3(bs.base.httpClient, ticker24hrUri, map[string]string{})
	}()

	go func() {
		defer wg.Done()
		tickerBookResp, err2 = api.HttpGet3(bs.base.httpClient, tickerBookUri, map[string]string{})
	}()

	wg.Wait()

	if err1 != nil {
		return nil, err1
	}

	if err2 != nil {
		return nil, err2
	}

	if len(ticker24HrResp) == 0 {
		return nil, errors.New("response is empty")
	}

	if len(tickerBookResp) == 0 {
		return nil, errors.New("response is empty")
	}

	ticker24HrMap := ticker24HrResp[0].(map[string]interface{})
	tickerBookMap := tickerBookResp[0].(map[string]interface{})

	var ticker types.Ticker
	ticker.Pair = currencyPair
	ticker.Date = types.ToUint64(tickerBookMap["time"])
	ticker.Last = types.ToFloat64(ticker24HrMap["lastPrice"])
	ticker.Buy = types.ToFloat64(tickerBookMap["bidPrice"])
	ticker.Sell = types.ToFloat64(tickerBookMap["askPrice"])
	ticker.High = types.ToFloat64(ticker24HrMap["highPrice"])
	ticker.Low = types.ToFloat64(ticker24HrMap["lowPrice"])
	ticker.Vol = types.ToFloat64(ticker24HrMap["volume"])

	return &ticker, nil
}

func (bs *BinanceFutures) GetFutureDepth(currencyPair types.CurrencyPair, contractType string, size int) (*types.Depth, error) {
	symbol, err := bs.adaptToSymbol(currencyPair, contractType)
	if err != nil {
		return nil, err
	}

	limit := 5
	if size <= 5 {
		limit = 5
	} else if size <= 10 {
		limit = 10
	} else if size <= 20 {
		limit = 20
	} else if size <= 50 {
		limit = 50
	} else if size <= 100 {
		limit = 100
	} else if size <= 500 {
		limit = 500
	} else {
		limit = 1000
	}

	depthUri := bs.base.apiV1 + "depth?symbol=%s&limit=%d"

	ret, err := api.HttpGet(bs.base.httpClient, fmt.Sprintf(depthUri, symbol, limit))
	if err != nil {
		return nil, err
	}
	logger.Debug(ret)

	var dep types.Depth

	dep.ContractType = contractType
	dep.Pair = currencyPair
	eT := int64(ret["E"].(float64))
	dep.UTime = time.Unix(0, eT*int64(time.Millisecond))

	for _, item := range ret["asks"].([]interface{}) {
		ask := item.([]interface{})
		dep.AskList = append(dep.AskList, types.DepthRecord{
			Price:  types.ToFloat64(ask[0]),
			Amount: types.ToFloat64(ask[1]),
		})
	}

	for _, item := range ret["bids"].([]interface{}) {
		bid := item.([]interface{})
		dep.BidList = append(dep.BidList, types.DepthRecord{
			Price:  types.ToFloat64(bid[0]),
			Amount: types.ToFloat64(bid[1]),
		})
	}

	sort.Sort(sort.Reverse(dep.AskList))

	return &dep, nil
}

func (bs *BinanceFutures) GetFutureIndex(currencyPair types.CurrencyPair) (float64, error) {
	panic("not supported.")
}

func (bs *BinanceFutures) GetFutureUserinfo(currencyPair ...types.CurrencyPair) (*types.FutureAccount, error) {
	accountUri := bs.base.apiV1 + "account"
	param := url.Values{}
	bs.base.buildParamsSigned(&param)

	respData, err := api.HttpGet5(bs.base.httpClient, accountUri+"?"+param.Encode(), map[string]string{
		"X-MBX-APIKEY": bs.apikey})

	if err != nil {
		return nil, err
	}

	logger.Debug(string(respData))

	var (
		accountResp    AccountResponse
		futureAccounts types.FutureAccount
	)

	err = json.Unmarshal(respData, &accountResp)
	if err != nil {
		return nil, fmt.Errorf("response body: %s , %w", string(respData), err)
	}

	futureAccounts.FutureSubAccounts = make(map[types.Currency]types.FutureSubAccount, 4)
	for _, asset := range accountResp.Assets {
		currency := types.NewCurrency(asset.Asset, "")
		futureAccounts.FutureSubAccounts[currency] = types.FutureSubAccount{
			Currency:      types.NewCurrency(asset.Asset, ""),
			AccountRights: asset.MarginBalance,
			KeepDeposit:   asset.MaintMargin,
			ProfitReal:    0,
			ProfitUnreal:  asset.UnrealizedProfit,
			RiskRate:      0,
		}
	}

	return &futureAccounts, nil
}

func (bs *BinanceFutures) PlaceFutureOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType, matchPrice int, leverRate float64) (string, error) {
	apiPath := "order"
	symbol, err := bs.adaptToSymbol(currencyPair, contractType)
	if err != nil {
		return "", err
	}

	param := url.Values{}
	param.Set("symbol", symbol)
	param.Set("newClientOrderId", types.GenerateOrderClientId(32))
	param.Set("quantity", amount)
	param.Set("newOrderRespType", "ACK")

	if matchPrice == 0 {
		param.Set("type", "LIMIT")
		param.Set("timeInForce", "GTC")
		param.Set("price", price)
	} else {
		param.Set("type", "MARKET")
	}

	switch openType {
	case types.OPEN_BUY, types.CLOSE_SELL:
		param.Set("side", "BUY")
	case types.OPEN_SELL, types.CLOSE_BUY:
		param.Set("side", "SELL")
	}

	bs.base.buildParamsSigned(&param)

	resp, err := api.HttpPostForm2(bs.base.httpClient, fmt.Sprintf("%s%s", bs.base.apiV1, apiPath), param,
		map[string]string{"X-MBX-APIKEY": bs.apikey})

	if err != nil {
		return "", err
	}

	logger.Debug(string(resp))

	var response struct {
		BaseResponse
		OrderId int64 `json:"orderId"`
	}

	err = json.Unmarshal(resp, &response)
	if err != nil {
		return "", err
	}

	if response.Code == 0 {
		return fmt.Sprint(response.OrderId), nil
	}

	return "", errors.New(response.Msg)
}

func (bs *BinanceFutures) LimitFuturesOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType int, opt ...types.LimitOrderOptionalParameter) (*types.FutureOrder, error) {
	orderId, err := bs.PlaceFutureOrder(currencyPair, contractType, price, amount, openType, 0, 10)
	return &types.FutureOrder{
		OrderID2:     orderId,
		Currency:     currencyPair,
		ContractName: contractType,
		Amount:       types.ToFloat64(amount),
		Price:        types.ToFloat64(price),
		OType:        openType,
	}, err
}

func (bs *BinanceFutures) MarketFuturesOrder(currencyPair types.CurrencyPair, contractType, amount string, openType int) (*types.FutureOrder, error) {
	orderId, err := bs.PlaceFutureOrder(currencyPair, contractType, "", amount, openType, 1, 10)
	return &types.FutureOrder{
		OrderID2:     orderId,
		Currency:     currencyPair,
		ContractName: contractType,
		Amount:       types.ToFloat64(amount),
		OType:        openType,
	}, err
}

func (bs *BinanceFutures) FutureCancelOrder(currencyPair types.CurrencyPair, contractType, orderId string) (bool, error) {
	apiPath := "order"
	symbol, err := bs.adaptToSymbol(currencyPair, contractType)
	if err != nil {
		return false, err
	}

	param := url.Values{}
	param.Set("symbol", symbol)
	if strings.HasPrefix(orderId, "goex") {
		param.Set("origClientOrderId", orderId)
	} else {
		param.Set("orderId", orderId)
	}

	bs.base.buildParamsSigned(&param)

	reqUrl := fmt.Sprintf("%s%s?%s", bs.base.apiV1, apiPath, param.Encode())
	resp, err := api.HttpDeleteForm(bs.base.httpClient, reqUrl, url.Values{}, map[string]string{"X-MBX-APIKEY": bs.apikey})
	if err != nil {
		logger.Errorf("request url: %s", reqUrl)
		return false, err
	}

	logger.Debug(string(resp))

	return true, nil
}

func (bs *BinanceFutures) GetFuturePosition(currencyPair types.CurrencyPair, contractType string) ([]types.FuturePosition, error) {
	symbol, err := bs.adaptToSymbol(currencyPair, contractType)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	bs.base.buildParamsSigned(&params)
	path := bs.base.apiV1 + "positionRisk?" + params.Encode()

	respBody, err := api.HttpGet5(bs.base.httpClient, path, map[string]string{"X-MBX-APIKEY": bs.apikey})
	if err != nil {
		return nil, err
	}
	logger.Debug(string(respBody))

	var (
		positionRiskResponse []PositionRiskResponse
		positions            []types.FuturePosition
	)

	err = json.Unmarshal(respBody, &positionRiskResponse)
	if err != nil {
		logger.Errorf("response body: %s", string(respBody))
		return nil, err
	}

	for _, info := range positionRiskResponse {
		if info.Symbol != symbol {
			continue
		}

		p := types.FuturePosition{
			LeverRate:      info.Leverage,
			Symbol:         currencyPair,
			ForceLiquPrice: info.LiquidationPrice,
		}

		if info.PositionAmt > 0 {
			p.BuyAmount = info.PositionAmt
			p.BuyAvailable = info.PositionAmt
			p.BuyPriceAvg = info.EntryPrice
			p.BuyPriceCost = info.EntryPrice
			p.BuyProfit = info.UnRealizedProfit
			p.BuyProfitReal = info.UnRealizedProfit
		} else if info.PositionAmt < 0 {
			p.SellAmount = math.Abs(info.PositionAmt)
			p.SellAvailable = math.Abs(info.PositionAmt)
			p.SellPriceAvg = info.EntryPrice
			p.SellPriceCost = info.EntryPrice
			p.SellProfit = info.UnRealizedProfit
			p.SellProfitReal = info.UnRealizedProfit
		}

		positions = append(positions, p)
	}

	return positions, nil
}

func (bs *BinanceFutures) GetFutureOrders(orderIds []string, currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error) {
	panic("not supported.")
}

func (bs *BinanceFutures) GetFutureOrder(orderId string, currencyPair types.CurrencyPair, contractType string) (*types.FutureOrder, error) {
	apiPath := "order"
	symbol, err := bs.adaptToSymbol(currencyPair, contractType)
	if err != nil {
		return nil, err
	}

	param := url.Values{}
	param.Set("symbol", symbol)
	param.Set("orderId", orderId)

	bs.base.buildParamsSigned(&param)

	reqUrl := fmt.Sprintf("%s%s?%s", bs.base.apiV1, apiPath, param.Encode())
	resp, err := api.HttpGet5(bs.base.httpClient, reqUrl, map[string]string{"X-MBX-APIKEY": bs.apikey})
	if err != nil {
		logger.Errorf("request url: %s", reqUrl)
		return nil, err
	}

	logger.Debug(string(resp))

	var getOrderInfoResponse OrderInfoResponse
	err = json.Unmarshal(resp, &getOrderInfoResponse)
	if err != nil {
		logger.Errorf("response body: %s", string(resp))
		return nil, err
	}

	return &types.FutureOrder{
		Currency:     currencyPair,
		ClientOid:    getOrderInfoResponse.ClientOrderId,
		OrderID2:     fmt.Sprint(getOrderInfoResponse.OrderId),
		Price:        getOrderInfoResponse.Price,
		Amount:       getOrderInfoResponse.OrigQty,
		AvgPrice:     getOrderInfoResponse.AvgPrice,
		DealAmount:   getOrderInfoResponse.ExecutedQty,
		OrderTime:    getOrderInfoResponse.Time / 1000,
		Status:       bs.adaptStatus(getOrderInfoResponse.Status),
		OType:        bs.adaptOType(getOrderInfoResponse.Side, getOrderInfoResponse.PositionSide),
		ContractName: contractType,
		FinishedTime: getOrderInfoResponse.UpdateTime / 1000,
	}, nil
}

func (bs *BinanceFutures) GetUnfinishFutureOrders(currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error) {
	apiPath := "openOrders"
	param := url.Values{}

	symbol, err := bs.adaptToSymbol(currencyPair, contractType)
	if err != nil {
		return nil, err
	}

	param.Set("symbol", symbol)
	bs.base.buildParamsSigned(&param)

	respbody, err := api.HttpGet5(bs.base.httpClient, fmt.Sprintf("%s%s?%s", bs.base.apiV1, apiPath, param.Encode()),
		map[string]string{
			"X-MBX-APIKEY": bs.apikey,
		})
	if err != nil {
		return nil, err
	}
	logger.Debug(string(respbody))

	var (
		openOrderResponse []OrderInfoResponse
		orders            []types.FutureOrder
	)

	err = json.Unmarshal(respbody, &openOrderResponse)
	if err != nil {
		return nil, err
	}

	for _, ord := range openOrderResponse {
		orders = append(orders, types.FutureOrder{
			Currency:     currencyPair,
			ClientOid:    ord.ClientOrderId,
			OrderID:      ord.OrderId,
			OrderID2:     fmt.Sprint(ord.OrderId),
			Price:        ord.Price,
			Amount:       ord.OrigQty,
			AvgPrice:     ord.AvgPrice,
			DealAmount:   ord.ExecutedQty,
			Status:       bs.adaptStatus(ord.Status),
			OType:        bs.adaptOType(ord.Side, ord.PositionSide),
			ContractName: contractType,
			FinishedTime: ord.UpdateTime / 1000,
			OrderTime:    ord.Time / 1000,
		})
	}

	return orders, nil
}

func (bs *BinanceFutures) GetFutureOrderHistory(pair types.CurrencyPair, contractType string, optional ...types.OptionalParameter) ([]types.FutureOrder, error) {
	panic("implement me")
}

func (bs *BinanceFutures) GetFee() (float64, error) {
	panic("not supported.")
}

func (bs *BinanceFutures) GetContractValue(currencyPair types.CurrencyPair) (float64, error) {
	switch currencyPair {
	case types.BTC_USD:
		return 100, nil
	default:
		return 10, nil
	}
}

func (bs *BinanceFutures) GetDeliveryTime() (int, int, int, int) {
	panic("not supported.")
}

func (bs *BinanceFutures) GetKlineRecords(contractType string, currency types.CurrencyPair, period types.KlinePeriod, size int, opt ...types.OptionalParameter) ([]types.FutureKline, error) {
	panic("not supported.")
}

func (bs *BinanceFutures) GetTrades(contractType string, currencyPair types.CurrencyPair, since int64) ([]types.Trade, error) {
	panic("not supported.")
}

func (bs *BinanceFutures) GetFutureEstimatedPrice(currencyPair types.CurrencyPair) (float64, error) {
	panic("not supported.")
}

func (bs *BinanceFutures) GetExchangeInfo() {
	exchangeInfoUri := bs.base.apiV1 + "exchangeInfo"
	ret, err := api.HttpGet5(bs.base.httpClient, exchangeInfoUri, map[string]string{})
	if err != nil {
		logger.Error("[exchangeInfo] Http Error", err)
		return
	}

	err = json.Unmarshal(ret, &bs.exchangeInfo)
	if err != nil {
		logger.Error("json unmarshal response content error , content= ", string(ret))
		return
	}

	logger.Debug("[ExchangeInfo]", bs.exchangeInfo)
}

func (bs *BinanceFutures) adaptToSymbol(pair types.CurrencyPair, contractType string) (string, error) {
	if contractType == types.THIS_WEEK_CONTRACT || contractType == types.NEXT_WEEK_CONTRACT {
		return "", errors.New("binance only support contract quarter or bi_quarter")
	}

	if contractType == types.SWAP_CONTRACT {
		return fmt.Sprintf("%s_PERP", pair.AdaptUsdtToUsd().ToSymbol("")), nil
	}

	if bs.exchangeInfo == nil || len(bs.exchangeInfo.Symbols) == 0 {
		bs.GetExchangeInfo()
	}

	for _, info := range bs.exchangeInfo.Symbols {
		if info.ContractType != "PERPETUAL" &&
			info.ContractStatus == "TRADING" &&
			info.DeliveryDate <= time.Now().Unix()*1000 {
			logger.Debugf("pair=%s , contractType=%s, delivery date = %d ,  now= %d", info.Pair, info.ContractType, info.DeliveryDate, time.Now().Unix()*1000)
			bs.GetExchangeInfo()
		}

		if info.Pair == pair.ToSymbol("") {
			if info.ContractStatus != "TRADING" {
				return "", errors.New("contract status " + info.ContractStatus)
			}

			if info.ContractType == "CURRENT_QUARTER" && contractType == types.QUARTER_CONTRACT {
				return info.Symbol, nil
			}

			if info.ContractType == "NEXT_QUARTER" && contractType == types.BI_QUARTER_CONTRACT {
				return info.Symbol, nil
			}

			if info.Symbol == contractType {
				return info.Symbol, nil
			}
		}
	}

	return "", errors.New("binance not support " + pair.ToSymbol("") + " " + contractType)
}

func (bs *BinanceFutures) adaptStatus(status string) types.TradeStatus {
	switch status {
	case "NEW":
		return types.ORDER_UNFINISH
	case "CANCELED":
		return types.ORDER_CANCEL
	case "FILLED":
		return types.ORDER_FINISH
	case "PARTIALLY_FILLED":
		return types.ORDER_PART_FINISH
	case "PENDING_CANCEL":
		return types.ORDER_CANCEL_ING
	case "REJECTED":
		return types.ORDER_REJECT
	default:
		return types.ORDER_UNFINISH
	}
}

func (bs *BinanceFutures) adaptOType(side string, positionSide string) int {
	if positionSide == "BOTH" && side == "SELL" {
		return types.OPEN_SELL
	}

	if positionSide == "BOTH" && side == "BUY" {
		return types.OPEN_BUY
	}

	if positionSide == "LONG" {
		switch side {
		case "BUY":
			return types.OPEN_BUY
		default:
			return types.CLOSE_BUY
		}
	}

	if positionSide == "SHORT" {
		switch side {
		case "SELL":
			return types.OPEN_SELL
		default:
			return types.CLOSE_SELL
		}
	}

	return 0
}
