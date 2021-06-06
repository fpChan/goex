package huobi

import (
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"testing"
	"time"
)

func TestNewHbdmSwapWs(t *testing.T) {
	logger.SetLevel(logger.DEBUG)

	ws := NewHbdmSwapWs()

	ws.DepthCallback(func(depth *types.Depth) {
		t.Log(depth)
	})
	ws.TickerCallback(func(ticker *types.FutureTicker) {
		t.Log(ticker.Date, ticker.Last, ticker.Buy, ticker.Sell, ticker.High, ticker.Low, ticker.Vol)
	})
	ws.TradeCallback(func(trade *types.Trade, contract string) {
		t.Log(trade, contract)
	})

	//t.Log(ws.SubscribeDepth(goex.BTC_USD, goex.SWAP_CONTRACT))
	//t.Log(ws.SubscribeTicker(goex.BTC_USD, goex.SWAP_CONTRACT))
	t.Log(ws.SubscribeTrade(types.BTC_USD, types.SWAP_CONTRACT))

	time.Sleep(time.Minute)
}
