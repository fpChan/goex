package okex

import (
	"errors"
	"fmt"
	"github.com/fpChan/goex/types"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fpChan/goex/internal/logger"
)

//合约信息
type FutureContractInfo struct {
	InstrumentID    string  `json:"instrument_id"` //合约ID:如BTC-USD-180213
	UnderlyingIndex string  `json:"underlying_index"`
	QuoteCurrency   string  `json:"quote_currency"`
	TickSize        float64 `json:"tick_size,string"` //下单价格精度
	TradeIncrement  string  `json:"trade_increment"`  //数量精度
	ContractVal     string  `json:"contract_val"`     //合约面值(美元)
	Listing         string  `json:"listing"`
	Delivery        string  `json:"delivery"` //交割日期
	Alias           string  `json:"alias"`    //	本周 this_week 次周 next_week 季度 quarter
}

type AllFutureContractInfo struct {
	contractInfos []FutureContractInfo
	uTime         time.Time
}

type OKExFuture struct {
	*OKEx
	sync.Locker
	allContractInfo AllFutureContractInfo
}

func (ok *OKExFuture) GetExchangeName() string {
	return types.OKEX_FUTURE
}

//获取法币汇率
func (ok *OKExFuture) GetRate() (float64, error) {
	var response struct {
		Rate         float64   `json:"rate,string"`
		InstrumentId string    `json:"instrument_id"` //USD_CNY
		Timestamp    time.Time `json:"timestamp"`
	}
	err := ok.DoRequest("GET", "/api/futures/v3/rate", "", &response)
	if err != nil {
		return 0, err
	}

	return response.Rate, nil
}

func (ok *OKExFuture) GetFutureEstimatedPrice(currencyPair types.CurrencyPair) (float64, error) {
	urlPath := fmt.Sprintf("/api/futures/v3/instruments/%s/estimated_price", ok.GetFutureContractId(currencyPair, types.QUARTER_CONTRACT))
	var response struct {
		InstrumentId    string  `json:"instrument_id"`
		SettlementPrice float64 `json:"settlement_price,string"`
		Timestamp       string  `json:"timestamp"`
	}
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return 0, err
	}
	return response.SettlementPrice, nil
}

func (ok *OKExFuture) GetAllFutureContractInfo() ([]FutureContractInfo, error) {
	urlPath := "/api/futures/v3/instruments"
	var response []FutureContractInfo
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (ok *OKExFuture) GetContractInfo(contractId string) (*FutureContractInfo, error) {
	now := time.Now()
	if len(ok.allContractInfo.contractInfos) == 0 ||
		(ok.allContractInfo.uTime.Hour() < 16 && now.Hour() == 16 && now.Minute() <= 10) {
		ok.Lock()
		defer ok.Unlock()

		infos, err := ok.GetAllFutureContractInfo()
		if err != nil {
			logger.Errorf("Get All Futures Contract Infos Error=%s", err)
		} else {
			ok.allContractInfo.contractInfos = infos
			ok.allContractInfo.uTime = now
		}
	}

	for _, itm := range ok.allContractInfo.contractInfos {
		if itm.InstrumentID == contractId {
			return &itm, nil
		}
	}

	return nil, errors.New("unknown contract id " + contractId)
}

func (ok *OKExFuture) GetFutureContractId(pair types.CurrencyPair, contractAlias string) string {
	if contractAlias != types.QUARTER_CONTRACT &&
		contractAlias != types.NEXT_WEEK_CONTRACT &&
		contractAlias != types.THIS_WEEK_CONTRACT &&
		contractAlias != types.BI_QUARTER_CONTRACT { //传Alias，需要转为具体ContractId
		return contractAlias
	}

	now := time.Now()
	hour := now.Hour()
	minute := now.Minute()

	if ok.allContractInfo.uTime.IsZero() || (ok.allContractInfo.uTime.Hour() < 16 && hour == 16 && minute <= 11) {
		ok.Lock()
		defer ok.Unlock()

		contractInfo, err := ok.GetAllFutureContractInfo()
		if err == nil {
			ok.allContractInfo.uTime = time.Now()
			ok.allContractInfo.contractInfos = contractInfo
		} else {
			time.Sleep(120 * time.Millisecond) //retry
			contractInfo, err = ok.GetAllFutureContractInfo()
			if err != nil {
				logger.Errorf(fmt.Sprintf("Get Futures Contract Id Error [%s] ???", err.Error()))
			}
		}
	}

	contractId := ""
	for _, itm := range ok.allContractInfo.contractInfos {
		if itm.Alias == contractAlias &&
			itm.UnderlyingIndex == pair.CurrencyA.Symbol &&
			itm.QuoteCurrency == pair.CurrencyB.Symbol {
			contractId = itm.InstrumentID
			break
		}
	}

	return contractId
}

type tickerResponse struct {
	InstrumentId string  `json:"instrument_id"`
	Last         float64 `json:"last,string"`
	High24h      float64 `json:"high_24h,string"`
	Low24h       float64 `json:"low_24h,string"`
	BestBid      float64 `json:"best_bid,string"`
	BestAsk      float64 `json:"best_ask,string"`
	Volume24h    float64 `json:"volume_24h,string"`
	Timestamp    string  `json:"timestamp"`
}

func (ok *OKExFuture) GetFutureTicker(currencyPair types.CurrencyPair, contractType string) (*types.Ticker, error) {
	var (
		urlPath  = fmt.Sprintf("/api/swap/v3/instruments/%s-%s-%s/ticker", currencyPair.CurrencyA.Symbol, currencyPair.CurrencyB.Symbol, contractType)
		response tickerResponse
	)
	fmt.Printf("url: %s \n", urlPath)
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	date, _ := time.Parse(time.RFC3339, response.Timestamp)

	return &types.Ticker{
		Pair: currencyPair,
		Sell: response.BestAsk,
		Buy:  response.BestBid,
		Low:  response.Low24h,
		High: response.High24h,
		Last: response.Last,
		Vol:  response.Volume24h,
		Date: uint64(date.UnixNano() / int64(time.Millisecond))}, nil
}

func (ok *OKExFuture) GetFutureTrendTicker(currencyPair types.CurrencyPair, contractType string) (*types.Trend, error) {
	var (
		urlPath  = fmt.Sprintf("/v2/perpetual/pc/public/contracts/%s-%s-%s/ticker", currencyPair.CurrencyA.Symbol, currencyPair.CurrencyB.Symbol, strings.ToUpper(contractType))
		response types.OkTrendTicker
	)
	fmt.Printf("url: %s \n", urlPath)
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}
	if response.Code != 0 {
		logger.Error(fmt.Sprintf("failed to get ticker url %s error_msg is %s . msg is %s ", urlPath, response.Error_message, response.Msg))
	}

	return &types.Trend{
		ChangePercent: response.Data.ChangePercent,
		Close:         response.Data.Close,
		Contract:      response.Data.Contract,
		ContractId:    response.Data.ContractId,
		High:          response.Data.High,
		//HighLimit:     response.HighLimit,
		//HoldAmount:    response.HoldAmount,
		Low: response.Data.Low,
		//LowLimit:      response.LowLimit,
		Open: response.Data.Open,
	}, nil
}

func (ok *OKExFuture) GetFutureAllTicker() (*[]types.FutureTicker, error) {
	var urlPath = "/api/futures/v3/instruments/ticker"

	var response []tickerResponse

	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	var tickers []types.FutureTicker
	for _, t := range response {
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

type depthResponse struct {
	Bids         [][4]interface{} `json:"bids"`
	Asks         [][4]interface{} `json:"asks"`
	InstrumentId string           `json:"instrument_id"`
	Timestamp    string           `json:"timestamp"`
}

func (ok *OKExFuture) GetFutureDepth(currencyPair types.CurrencyPair, contractType string, size int) (*types.Depth, error) {
	var (
		response depthResponse
		dep      types.Depth
	)
	urlPath := fmt.Sprintf("/api/futures/v3/instruments/%s/book?size=%d", ok.GetFutureContractId(currencyPair, contractType), size)
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}
	dep.Pair = currencyPair
	dep.ContractType = contractType
	dep.UTime, _ = time.Parse(time.RFC3339, response.Timestamp)
	for _, itm := range response.Asks {
		dep.AskList = append(dep.AskList, types.DepthRecord{
			Price:  types.ToFloat64(itm[0]),
			Amount: types.ToFloat64(itm[1])})
	}
	for _, itm := range response.Bids {
		dep.BidList = append(dep.BidList, types.DepthRecord{
			Price:  types.ToFloat64(itm[0]),
			Amount: types.ToFloat64(itm[1])})
	}
	sort.Sort(sort.Reverse(dep.AskList))
	return &dep, nil
}

func (ok *OKExFuture) GetFutureIndex(currencyPair types.CurrencyPair) (float64, error) {
	//统一交易对，当周，次周，季度指数一样的
	urlPath := fmt.Sprintf("/api/futures/v3/instruments/%s/index", ok.GetFutureContractId(currencyPair, types.QUARTER_CONTRACT))
	var response struct {
		InstrumentId string  `json:"instrument_id"`
		Index        float64 `json:"index,string"`
		Timestamp    string  `json:"timestamp"`
	}
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return 0, nil
	}
	return response.Index, nil
}

type CrossedAccountInfo struct {
	MarginMode       string  `json:"margin_mode"`
	Equity           float64 `json:"equity,string"`
	RealizedPnl      float64 `json:"realized_pnl,string"`
	UnrealizedPnl    float64 `json:"unrealized_pnl,string"`
	MarginFrozen     float64 `json:"margin_frozen,string"`
	MarginRatio      float64 `json:"margin_ratio,string"`
	MaintMarginRatio float64 `json:"maint_margin_ratio,string"`
}

func (ok *OKExFuture) GetAccounts(currencyPair types.CurrencyPair) (*types.FutureAccount, error) {
	urlPath := "/api/futures/v3/accounts/" + currencyPair.ToLower().ToSymbol("-")
	var response CrossedAccountInfo

	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	acc := new(types.FutureAccount)
	acc.FutureSubAccounts = make(map[types.Currency]types.FutureSubAccount, 1)
	if response.MarginMode == "crossed" {
		acc.FutureSubAccounts[currencyPair.CurrencyA] = types.FutureSubAccount{
			Currency:      currencyPair.CurrencyA,
			AccountRights: response.Equity,
			ProfitReal:    response.RealizedPnl,
			ProfitUnreal:  response.UnrealizedPnl,
			KeepDeposit:   response.MarginFrozen,
			RiskRate:      response.MarginRatio,
		}
	} else {
		//todo 逐仓模式
		return nil, errors.New("goex unsupported  fixed margin mode")
	}

	return acc, nil
}

//基本上已经报废，OK限制10s一次，但是基本上都会返回error：{"code":30014,"message":"Too Many Requests"}
//加入currency  pair救活了这个接口
func (ok *OKExFuture) GetFutureUserinfo(currencyPair ...types.CurrencyPair) (*types.FutureAccount, error) {
	if len(currencyPair) == 1 {
		return ok.GetAccounts(currencyPair[0])
	}

	urlPath := "/api/futures/v3/accounts"
	var response struct {
		Info map[string]map[string]interface{}
	}

	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	acc := new(types.FutureAccount)
	acc.FutureSubAccounts = make(map[types.Currency]types.FutureSubAccount, 2)
	for c, info := range response.Info {
		if info["margin_mode"] == "crossed" {
			currency := types.NewCurrency(c, "")
			acc.FutureSubAccounts[currency] = types.FutureSubAccount{
				Currency:      currency,
				AccountRights: types.ToFloat64(info["equity"]),
				ProfitReal:    types.ToFloat64(info["realized_pnl"]),
				ProfitUnreal:  types.ToFloat64(info["unrealized_pnl"]),
				KeepDeposit:   types.ToFloat64(info["margin_frozen"]),
				RiskRate:      types.ToFloat64(info["margin_ratio"]),
			}
		} else {
			//todo 逐仓模式
		}
	}
	return acc, nil
}

func (ok *OKExFuture) normalizePrice(price float64, pair types.CurrencyPair) string {
	for _, info := range ok.allContractInfo.contractInfos {
		if info.UnderlyingIndex == pair.CurrencyA.Symbol && info.QuoteCurrency == pair.CurrencyB.Symbol {
			var bit = 0
			for info.TickSize < 1 {
				bit++
				info.TickSize *= 10
			}
			return types.FloatToString(price, bit)
		}
	}
	return types.FloatToString(price, 2)
}

//matchPrice:是否以对手价下单(0:不是 1:是)，默认为0;当取值为1时,price字段无效，当以对手价下单，order_type只能选择0:普通委托
func (ok *OKExFuture) PlaceFutureOrder2(matchPrice int, ord *types.FutureOrder) (*types.FutureOrder, error) {
	urlPath := "/api/futures/v3/order"
	var param struct {
		ClientOid    string `json:"client_oid"`
		InstrumentId string `json:"instrument_id"`
		Type         int    `json:"type"`
		OrderType    int    `json:"order_type"`
		Price        string `json:"price"`
		Size         string `json:"size"`
		MatchPrice   int    `json:"match_price"`
		//Leverage     int    `json:"leverage"` //v3 api 已废弃
	}

	var response struct {
		Result       bool   `json:"result"`
		ErrorMessage string `json:"error_message"`
		ErrorCode    string `json:"error_code"`
		ClientOid    string `json:"client_oid"`
		OrderId      string `json:"order_id"`
	}
	if ord == nil {
		return nil, errors.New("ord param is nil")
	}
	param.InstrumentId = ok.GetFutureContractId(ord.Currency, ord.ContractName)
	param.ClientOid = types.GenerateOrderClientId(32)
	param.Type = ord.OType
	param.OrderType = ord.OrderType
	param.Price = ok.normalizePrice(ord.Price, ord.Currency)
	param.Size = fmt.Sprint(ord.Amount)
	//param.Leverage = ord.LeverRate
	param.MatchPrice = matchPrice

	//当matchPrice=1以对手价下单，order_type只能选择0:普通委托
	if param.MatchPrice == 1 {
		logger.Warn("注意:当matchPrice=1以对手价下单时，order_type只能选择0:普通委托")
		param.OrderType = types.ORDER_FEATURE_ORDINARY
	}

	reqBody, _, _ := ok.BuildRequestBody(param)
	err := ok.DoRequest("POST", urlPath, reqBody, &response)

	if err != nil {
		return ord, err
	}

	ord.ClientOid = response.ClientOid
	ord.OrderID2 = response.OrderId
	ord.OrderTime = time.Now().UnixNano() / int64(time.Millisecond)

	return ord, nil
}

func (ok *OKExFuture) PlaceFutureOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType, matchPrice int, leverRate float64) (string, error) {
	fOrder, err := ok.PlaceFutureOrder2(matchPrice, &types.FutureOrder{
		Price:        types.ToFloat64(price),
		Amount:       types.ToFloat64(amount),
		OType:        openType,
		ContractName: contractType,
		Currency:     currencyPair,
	})
	return fOrder.OrderID2, err
}

func (ok *OKExFuture) LimitFuturesOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType int, opt ...types.LimitOrderOptionalParameter) (*types.FutureOrder, error) {
	ord := &types.FutureOrder{
		Currency:     currencyPair,
		Price:        types.ToFloat64(price),
		Amount:       types.ToFloat64(amount),
		OType:        openType,
		ContractName: contractType,
	}

	if len(opt) > 0 {
		switch opt[0] {
		case types.PostOnly:
			ord.OrderType = 1
		case types.Fok:
			ord.OrderType = 2
		case types.Ioc:
			ord.OrderType = 3
		}
	}

	return ok.PlaceFutureOrder2(0, ord)
}

func (ok *OKExFuture) MarketFuturesOrder(currencyPair types.CurrencyPair, contractType, amount string, openType int) (*types.FutureOrder, error) {
	return ok.PlaceFutureOrder2(1, &types.FutureOrder{
		Currency:     currencyPair,
		Amount:       types.ToFloat64(amount),
		OType:        openType,
		ContractName: contractType,
	})
}

func (ok *OKExFuture) FutureCancelOrder(currencyPair types.CurrencyPair, contractType, orderId string) (bool, error) {
	urlPath := fmt.Sprintf("/api/futures/v3/cancel_order/%s/%s", ok.GetFutureContractId(currencyPair, contractType), orderId)
	var response struct {
		Result       bool   `json:"result"`
		OrderId      string `json:"order_id"`
		ClientOid    string `json:"client_oid"`
		InstrumentId string `json:"instrument_id"`
	}
	err := ok.DoRequest("POST", urlPath, "", &response)
	if err != nil {
		return false, err
	}
	return response.Result, nil
}

func (ok *OKExFuture) GetFuturePosition(currencyPair types.CurrencyPair, contractType string) ([]types.FuturePosition, error) {
	urlPath := fmt.Sprintf("/api/futures/v3/%s/position", ok.GetFutureContractId(currencyPair, contractType))
	var response struct {
		Result     bool   `json:"result"`
		MarginMode string `json:"margin_mode"`
		Holding    []struct {
			InstrumentId         string    `json:"instrument_id"`
			LongQty              float64   `json:"long_qty,string"` //多
			LongAvailQty         float64   `json:"long_avail_qty,string"`
			LongAvgCost          float64   `json:"long_avg_cost,string"`
			LongSettlementPrice  float64   `json:"long_settlement_price,string"`
			LongMargin           float64   `json:"long_margin,string"`
			LongPnl              float64   `json:"long_pnl,string"`
			LongPnlRatio         float64   `json:"long_pnl_ratio,string"`
			LongUnrealisedPnl    float64   `json:"long_unrealised_pnl,string"`
			RealisedPnl          float64   `json:"realised_pnl,string"`
			Leverage             float64   `json:"leverage,string"`
			ShortQty             float64   `json:"short_qty,string"`
			ShortAvailQty        float64   `json:"short_avail_qty,string"`
			ShortAvgCost         float64   `json:"short_avg_cost,string"`
			ShortSettlementPrice float64   `json:"short_settlement_price,string"`
			ShortMargin          float64   `json:"short_margin,string"`
			ShortPnl             float64   `json:"short_pnl,string"`
			ShortPnlRatio        float64   `json:"short_pnl_ratio,string"`
			ShortUnrealisedPnl   float64   `json:"short_unrealised_pnl,string"`
			LiquidationPrice     float64   `json:"liquidation_price,string"`
			CreatedAt            time.Time `json:"created_at,string"`
			UpdatedAt            time.Time `json:"updated_at"`
		}
	}
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	var postions []types.FuturePosition

	if !response.Result {
		return nil, errors.New("unknown error")
	}

	if response.MarginMode == "fixed" {
		panic("not support the fix future")
	}

	for _, pos := range response.Holding {
		postions = append(postions, types.FuturePosition{
			Symbol:         currencyPair,
			ContractType:   contractType,
			ContractId:     types.ToInt64(pos.InstrumentId[8:]),
			BuyAmount:      pos.LongQty,
			BuyAvailable:   pos.LongAvailQty,
			BuyPriceAvg:    pos.LongAvgCost,
			BuyPriceCost:   pos.LongAvgCost,
			BuyProfitReal:  pos.LongPnl,
			SellAmount:     pos.ShortQty,
			SellAvailable:  pos.ShortAvailQty,
			SellPriceAvg:   pos.ShortAvgCost,
			SellPriceCost:  pos.ShortAvgCost,
			SellProfitReal: pos.ShortPnl,
			ForceLiquPrice: pos.LiquidationPrice,
			LeverRate:      pos.Leverage,
			CreateDate:     pos.CreatedAt.Unix(),
			ShortPnlRatio:  pos.ShortPnlRatio,
			LongPnlRatio:   pos.LongPnlRatio,
		})
	}

	return postions, nil
}

func (ok *OKExFuture) GetFutureOrders(orderIds []string, currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error) {
	panic("")
}

func (ok *OKExFuture) GetFutureOrderHistory(pair types.CurrencyPair, contractType string, optional ...types.OptionalParameter) ([]types.FutureOrder, error) {
	urlPath := fmt.Sprintf("/api/futures/v3/orders/%s?", ok.GetFutureContractId(pair, contractType))

	param := url.Values{}
	param.Set("limit", "100")
	param.Set("state", "7")
	types.MergeOptionalParameter(&param, optional...)
	urlPath += param.Encode()

	var response struct {
		Result    bool
		OrderInfo []futureOrderResponse `json:"order_info"`
	}

	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	if !response.Result {
		return nil, errors.New(fmt.Sprintf("%v", response))
	}

	orders := make([]types.FutureOrder, 0, 100)
	for _, info := range response.OrderInfo {
		ord := ok.adaptOrder(info)
		ord.Currency = pair
		orders = append(orders, ord)
	}

	return orders, nil
}

type futureOrderResponse struct {
	InstrumentId string    `json:"instrument_id"`
	ClientOid    string    `json:"client_oid"`
	OrderId      string    `json:"order_id"`
	Size         float64   `json:"size,string"`
	Price        float64   `json:"price,string"`
	FilledQty    float64   `json:"filled_qty,string"`
	PriceAvg     float64   `json:"price_avg,string"`
	Fee          float64   `json:"fee,string"`
	Type         int       `json:"type,string"`
	OrderType    int       `json:"order_type,string"`
	Pnl          float64   `json:"pnl,string"`
	Leverage     int       `json:"leverage,string"`
	ContractVal  float64   `json:"contract_val,string"`
	State        int       `json:"state,string"`
	Timestamp    time.Time `json:"timestamp,string"`
}

func (ok *OKExFuture) adaptOrder(response futureOrderResponse) types.FutureOrder {
	return types.FutureOrder{
		ContractName: response.InstrumentId,
		OrderID2:     response.OrderId,
		ClientOid:    response.ClientOid,
		Amount:       response.Size,
		Price:        response.Price,
		DealAmount:   response.FilledQty,
		AvgPrice:     response.PriceAvg,
		OType:        response.Type,
		OrderType:    response.OrderType,
		Status:       ok.adaptOrderState(response.State),
		Fee:          response.Fee,
		OrderTime:    response.Timestamp.UnixNano() / int64(time.Millisecond),
	}
}

func (ok *OKExFuture) GetFutureOrder(orderId string, currencyPair types.CurrencyPair, contractType string) (*types.FutureOrder, error) {
	urlPath := fmt.Sprintf("/api/futures/v3/orders/%s/%s", ok.GetFutureContractId(currencyPair, contractType), orderId)
	var response futureOrderResponse
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	ord := ok.adaptOrder(response)
	ord.Currency = currencyPair

	return &ord, nil
}

func (ok *OKExFuture) GetUnfinishFutureOrders(currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error) {
	urlPath := fmt.Sprintf("/api/futures/v3/orders/%s?state=6&limit=100", ok.GetFutureContractId(currencyPair, contractType))
	var response struct {
		Result    bool                  `json:"result"`
		OrderInfo []futureOrderResponse `json:"order_info"`
	}
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}
	if !response.Result {
		return nil, errors.New("error")
	}

	var ords []types.FutureOrder
	for _, itm := range response.OrderInfo {
		ord := ok.adaptOrder(itm)
		ord.Currency = currencyPair
		ords = append(ords, ord)
	}

	return ords, nil
}

func (ok *OKExFuture) GetFee() (float64, error) { panic("") }

func (ok *OKExFuture) GetContractValue(currencyPair types.CurrencyPair) (float64, error) {
	//for _, info := range ok.allContractInfo.contractInfos {
	//	if info.UnderlyingIndex == currencyPair.CurrencyA.Symbol && info.QuoteCurrency == currencyPair.CurrencyB.Symbol {
	//		return ToFloat64(info.ContractVal), nil
	//	}
	//}
	if currencyPair.CurrencyA.Eq(types.BTC) {
		return 100, nil
	}

	return 10, nil
}

func (ok *OKExFuture) GetDeliveryTime() (int, int, int, int) {
	return 4, 16, 0, 0 //星期五，下午4点交割
}

func (ok *OKExFuture) GetKlineRecords(contractType string, currency types.CurrencyPair, period types.KlinePeriod, size int, opt ...types.OptionalParameter) ([]types.FutureKline, error) {
	urlPath := "/api/swap/v3/instruments/%s-%s-%s/candles?granularity=%d"
	granularity := adaptKLinePeriod(types.KlinePeriod(period))
	if granularity == -1 {
		return nil, errors.New("kline period parameter is error")
	}

	var response [][]interface{}
	err := ok.DoRequest("GET", fmt.Sprintf(urlPath, currency.CurrencyA.Symbol, currency.CurrencyB.Symbol, strings.ToUpper(contractType), granularity), "", &response)
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
				Open:      types.ToFloat64(itm[1]),
				High:      types.ToFloat64(itm[2]),
				Low:       types.ToFloat64(itm[3]),
				Close:     types.ToFloat64(itm[4]),
				Vol:       types.ToFloat64(itm[5])},
			Vol2: types.ToFloat64(itm[6])})
	}

	return klines, nil
}

/**
  since : 单位秒,开始时间
  to : 单位秒,结束时间
*/
func (ok *OKExFuture) GetKlineRecordsByRange(contractType string, currency types.CurrencyPair, period, since, to int) ([]types.FutureKline, error) {
	urlPath := "/api/futures/v3/instruments/%s/candles?start=%s&end=%s&granularity=%d"
	contractId := ok.GetFutureContractId(currency, contractType)
	sinceTime := time.Unix(int64(since), 0).UTC()
	toTime := time.Unix(int64(to), 0).UTC()

	granularity := adaptKLinePeriod(types.KlinePeriod(period))
	if granularity == -1 {
		return nil, errors.New("kline period parameter is error")
	}

	var response [][]interface{}
	err := ok.DoRequest("GET", fmt.Sprintf(urlPath, contractId, sinceTime.Format(time.RFC3339), toTime.Format(time.RFC3339), granularity), "", &response)
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
				Open:      types.ToFloat64(itm[1]),
				High:      types.ToFloat64(itm[2]),
				Low:       types.ToFloat64(itm[3]),
				Close:     types.ToFloat64(itm[4]),
				Vol:       types.ToFloat64(itm[5])},
			Vol2: types.ToFloat64(itm[6])})
	}

	return klines, nil
}

func (ok *OKExFuture) GetTrades(contractType string, currencyPair types.CurrencyPair, since int64) ([]types.Trade, error) {
	panic("")
}

//特殊接口
/*
 市价全平仓
 contract:合约ID
 oType：平仓方向：CLOSE_SELL平空，CLOSE_BUY平多
*/
func (ok *OKExFuture) MarketCloseAllPosition(currency types.CurrencyPair, contract string, oType int) (bool, error) {
	urlPath := "/api/futures/v3/close_position"
	var response struct {
		InstrumentId string `json:"instrument_id"`
		Result       bool   `json:"result"`
		Message      string `json:"message"`
		Code         int    `json:"code"`
	}

	var param struct {
		InstrumentId string `json:"instrument_id"`
		Direction    string `json:"direction"`
	}

	param.InstrumentId = ok.GetFutureContractId(currency, contract)
	if oType == types.CLOSE_BUY {
		param.Direction = "long"
	} else {
		param.Direction = "short"
	}
	reqBody, _, _ := ok.BuildRequestBody(param)
	err := ok.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return false, err
	}

	if !response.Result {
		return false, errors.New(response.Message)
	}

	return true, nil
}
