package okex

import (
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"os"
	"testing"
	"time"
)

func init() {
	logger.SetLevel(logger.DEBUG)
}

func TestNewOKExSpotV3Ws(t *testing.T) {
	os.Setenv("HTTPS_PROXY", "socks5://127.0.0.1:1080")
	okexSpotV3Ws := okex.OKExV3SpotWs
	okexSpotV3Ws.TickerCallback(func(ticker *types.Ticker) {
		t.Log(ticker)
	})
	okexSpotV3Ws.DepthCallback(func(depth *types.Depth) {
		t.Log(depth)
	})
	okexSpotV3Ws.TradeCallback(func(trade *types.Trade) {
		t.Log(trade)
	})
	okexSpotV3Ws.KLineCallback(func(kline *types.Kline, period types.KlinePeriod) {
		t.Log(period, kline)
	})
	//okexSpotV3Ws.SubscribeDepth(goex.EOS_USDT, 5)
	//okexSpotV3Ws.SubscribeTrade(goex.EOS_USDT)
	//okexSpotV3Ws.SubscribeTicker(goex.EOS_USDT)
	okexSpotV3Ws.SubscribeKline(types.EOS_USDT, types.KLINE_PERIOD_1H)
	time.Sleep(time.Minute)
}
