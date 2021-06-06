package binance

import (
	"encoding/json"
	"errors"
	"github.com/fpChan/goex"
	"github.com/fpChan/goex/common/api"
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type FuturesWs struct {
	base  *BinanceFutures
	fOnce sync.Once
	dOnce sync.Once

	wsBuilder *api.WsBuilder
	f         *api.WsConn
	d         *api.WsConn

	depthCallFn  func(depth *types.Depth)
	tickerCallFn func(ticker *types.FutureTicker)
	tradeCalFn   func(trade *types.Trade, contract string)
}

func NewFuturesWs() *FuturesWs {
	futuresWs := new(FuturesWs)

	futuresWs.wsBuilder = api.NewWsBuilder().
		ProxyUrl(os.Getenv("HTTPS_PROXY")).
		ProtoHandleFunc(futuresWs.handle).AutoReconnect()

	httpCli := &http.Client{
		Timeout: 10 * time.Second,
	}

	if os.Getenv("HTTPS_PROXY") != "" {
		httpCli = &http.Client{
			Transport: &http.Transport{
				Proxy: func(r *http.Request) (*url.URL, error) {
					return url.Parse(os.Getenv("HTTPS_PROXY"))
				},
			},
			Timeout: 10 * time.Second,
		}
	}

	futuresWs.base = NewBinanceFutures(&types.APIConfig{
		HttpClient: httpCli,
	})

	return futuresWs
}

func (s *FuturesWs) connectUsdtFutures() {
	s.fOnce.Do(func() {
		s.f = s.wsBuilder.WsUrl("wss://fstream.binance.com/ws").Build()
	})
}

func (s *FuturesWs) connectFutures() {
	s.dOnce.Do(func() {
		s.d = s.wsBuilder.WsUrl("wss://dstream.binance.com/ws").Build()
	})
}

func (s *FuturesWs) DepthCallback(f func(depth *types.Depth)) {
	s.depthCallFn = f
}

func (s *FuturesWs) TickerCallback(f func(ticker *types.FutureTicker)) {
	s.tickerCallFn = f
}

func (s *FuturesWs) TradeCallback(f func(trade *types.Trade, contract string)) {
	s.tradeCalFn = f
}

func (s *FuturesWs) SubscribeDepth(pair types.CurrencyPair, contractType string) error {
	switch contractType {
	case types.SWAP_USDT_CONTRACT:
		return s.f.Subscribe(req{
			Method: "SUBSCRIBE",
			Params: []string{pair.AdaptUsdToUsdt().ToLower().ToSymbol("") + "@depth10@100ms"},
			Id:     1,
		})
	default:
		sym, _ := s.base.adaptToSymbol(pair.AdaptUsdtToUsd(), contractType)
		return s.d.Subscribe(req{
			Method: "SUBSCRIBE",
			Params: []string{strings.ToLower(sym) + "@depth10@100ms"},
			Id:     2,
		})
	}
	return errors.New("contract is error")
}

func (s *FuturesWs) SubscribeTicker(pair types.CurrencyPair, contractType string) error {
	switch contractType {
	case types.SWAP_USDT_CONTRACT:
		s.connectUsdtFutures()
		return s.f.Subscribe(req{
			Method: "SUBSCRIBE",
			Params: []string{pair.AdaptUsdToUsdt().ToLower().ToSymbol("") + "@ticker"},
			Id:     1,
		})
	default:
		s.connectFutures()
		sym, _ := s.base.adaptToSymbol(pair.AdaptUsdtToUsd(), contractType)
		return s.d.Subscribe(req{
			Method: "SUBSCRIBE",
			Params: []string{strings.ToLower(sym) + "@ticker"},
			Id:     2,
		})
	}
	return errors.New("contract is error")
}

func (s *FuturesWs) SubscribeTrade(pair types.CurrencyPair, contractType string) error {
	panic("implement me")
}

func (s *FuturesWs) handle(data []byte) error {
	var m = make(map[string]interface{}, 4)
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	if e, ok := m["e"].(string); ok && e == "depthUpdate" {
		dep := s.depthHandle(m["b"].([]interface{}), m["a"].([]interface{}))
		dep.ContractType = m["s"].(string)
		symbol, ok := m["ps"].(string)

		if ok {
			dep.Pair = adaptSymbolToCurrencyPair(symbol)
		} else {
			dep.Pair = adaptSymbolToCurrencyPair(dep.ContractType) //usdt swap
		}

		dep.UTime = time.Unix(0, goex.ToInt64(m["T"])*int64(time.Millisecond))
		s.depthCallFn(dep)

		return nil
	}

	if e, ok := m["e"].(string); ok && e == "24hrTicker" {
		s.tickerCallFn(s.tickerHandle(m))
		return nil
	}

	logger.Warn("unknown ws response:", string(data))

	return nil
}

func (s *FuturesWs) depthHandle(bids []interface{}, asks []interface{}) *types.Depth {
	var dep types.Depth

	for _, item := range bids {
		bid := item.([]interface{})
		dep.BidList = append(dep.BidList,
			types.DepthRecord{
				Price:  goex.ToFloat64(bid[0]),
				Amount: goex.ToFloat64(bid[1]),
			})
	}

	for _, item := range asks {
		ask := item.([]interface{})
		dep.AskList = append(dep.AskList, types.DepthRecord{
			Price:  goex.ToFloat64(ask[0]),
			Amount: goex.ToFloat64(ask[1]),
		})
	}

	sort.Sort(sort.Reverse(dep.AskList))

	return &dep
}

func (s *FuturesWs) tickerHandle(m map[string]interface{}) *types.FutureTicker {
	var ticker types.FutureTicker
	ticker.Ticker = new(types.Ticker)

	symbol, ok := m["ps"].(string)
	if ok {
		ticker.Pair = adaptSymbolToCurrencyPair(symbol)
	} else {
		ticker.Pair = adaptSymbolToCurrencyPair(m["s"].(string)) //usdt swap
	}

	ticker.ContractType = m["s"].(string)
	ticker.Date = goex.ToUint64(m["E"])
	ticker.High = goex.ToFloat64(m["h"])
	ticker.Low = goex.ToFloat64(m["l"])
	ticker.Last = goex.ToFloat64(m["c"])
	ticker.Vol = goex.ToFloat64(m["v"])

	return &ticker
}
