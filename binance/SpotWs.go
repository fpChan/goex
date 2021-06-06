package binance

import (
	json2 "encoding/json"
	"fmt"
	"github.com/fpChan/goex/common/api"
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type req struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int      `json:"id"`
}

type resp struct {
	Stream string           `json:"stream"`
	Data   json2.RawMessage `json:"data"`
}

type depthResp struct {
	LastUpdateId int             `json:"lastUpdateId"`
	Bids         [][]interface{} `json:"bids"`
	Asks         [][]interface{} `json:"asks"`
}

type SpotWs struct {
	c         *api.WsConn
	once      sync.Once
	wsBuilder *api.WsBuilder

	reqId int

	depthCallFn  func(depth *types.Depth)
	tickerCallFn func(ticker *types.Ticker)
	tradeCallFn  func(trade *types.Trade)
}

func NewSpotWs() *SpotWs {
	spotWs := &SpotWs{}
	logger.Debugf("proxy url: %s", os.Getenv("HTTPS_PROXY"))

	spotWs.wsBuilder = api.NewWsBuilder().
		WsUrl("wss://stream.binance.com:9443/stream?streams=depth/miniTicker/ticker/trade").
		ProxyUrl(os.Getenv("HTTPS_PROXY")).
		ProtoHandleFunc(spotWs.handle).AutoReconnect()

	spotWs.reqId = 1

	return spotWs
}

func (s *SpotWs) connect() {
	s.once.Do(func() {
		s.c = s.wsBuilder.Build()
	})
}

func (s *SpotWs) DepthCallback(f func(depth *types.Depth)) {
	s.depthCallFn = f
}

func (s *SpotWs) TickerCallback(f func(ticker *types.Ticker)) {
	s.tickerCallFn = f
}

func (s *SpotWs) TradeCallback(f func(trade *types.Trade)) {
	s.tradeCallFn = f
}

func (s *SpotWs) SubscribeDepth(pair types.CurrencyPair) error {
	defer func() {
		s.reqId++
	}()

	s.connect()

	return s.c.Subscribe(req{
		Method: "SUBSCRIBE",
		Params: []string{
			fmt.Sprintf("%s@depth10@100ms", pair.ToLower().ToSymbol("")),
		},
		Id: s.reqId,
	})
}

func (s *SpotWs) SubscribeTicker(pair types.CurrencyPair) error {
	defer func() {
		s.reqId++
	}()

	s.connect()

	return s.c.Subscribe(req{
		Method: "SUBSCRIBE",
		Params: []string{pair.ToLower().ToSymbol("") + "@ticker"},
		Id:     s.reqId,
	})
}

func (s *SpotWs) SubscribeTrade(pair types.CurrencyPair) error {
	panic("implement me")
}

func (s *SpotWs) handle(data []byte) error {
	var r resp
	err := json2.Unmarshal(data, &r)
	if err != nil {
		logger.Errorf("json unmarshal ws response error [%s] , response data = %s", err, string(data))
		return err
	}

	if strings.HasSuffix(r.Stream, "@depth10@100ms") {
		return s.depthHandle(r.Data, adaptStreamToCurrencyPair(r.Stream))
	}

	if strings.HasSuffix(r.Stream, "@ticker") {
		return s.tickerHandle(r.Data, adaptStreamToCurrencyPair(r.Stream))
	}

	logger.Warn("unknown ws response:", string(data))

	return nil
}

func (s *SpotWs) depthHandle(data json2.RawMessage, pair types.CurrencyPair) error {
	var (
		depthR depthResp
		dep    types.Depth
		err    error
	)

	err = json2.Unmarshal(data, &depthR)
	if err != nil {
		logger.Errorf("unmarshal depth response error %s[] , response data = %s", err, string(data))
		return err
	}

	dep.UTime = time.Now()
	dep.Pair = pair

	for _, bid := range depthR.Bids {
		dep.BidList = append(dep.BidList, types.DepthRecord{
			Price:  types.ToFloat64(bid[0]),
			Amount: types.ToFloat64(bid[1]),
		})
	}

	for _, ask := range depthR.Asks {
		dep.AskList = append(dep.AskList, types.DepthRecord{
			Price:  types.ToFloat64(ask[0]),
			Amount: types.ToFloat64(ask[1]),
		})
	}

	sort.Sort(sort.Reverse(dep.AskList))

	s.depthCallFn(&dep)

	return nil
}

func (s *SpotWs) tickerHandle(data json2.RawMessage, pair types.CurrencyPair) error {
	var (
		tickerData = make(map[string]interface{}, 4)
		ticker     types.Ticker
	)

	err := json2.Unmarshal(data, &tickerData)
	if err != nil {
		logger.Errorf("unmarshal ticker response data error [%s] , data = %s", err, string(data))
		return err
	}

	ticker.Pair = pair
	ticker.Vol = types.ToFloat64(tickerData["v"])
	ticker.Last = types.ToFloat64(tickerData["c"])
	ticker.Sell = types.ToFloat64(tickerData["a"])
	ticker.Buy = types.ToFloat64(tickerData["b"])
	ticker.High = types.ToFloat64(tickerData["h"])
	ticker.Low = types.ToFloat64(tickerData["l"])
	ticker.Date = types.ToUint64(tickerData["E"])

	s.tickerCallFn(&ticker)

	return nil
}
