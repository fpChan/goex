package okex

import (
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	client *http.Client
)

func init() {
	logger.SetLevel(logger.DEBUG)
}

func TestNewOKExV3FuturesWs(t *testing.T) {
	os.Setenv("HTTPS_PROXY", "socks5://127.0.0.1:1080")
	ok := NewOKEx(&types.APIConfig{
		HttpClient: http.DefaultClient,
	})
	ok.OKExV3FuturesWs.TickerCallback(func(ticker *types.FutureTicker) {
		t.Log(ticker.Ticker, ticker.ContractType)
	})
	ok.OKExV3FuturesWs.DepthCallback(func(depth *types.Depth) {
		t.Log(depth)
	})
	ok.OKExV3FuturesWs.TradeCallback(func(trade *types.Trade, s string) {
		t.Log(s, trade)
	})
	//ok.OKExV3FuturesWs.SubscribeTicker(goex.EOS_USD, goex.QUARTER_CONTRACT)
	ok.OKExV3FuturesWs.SubscribeDepth(types.EOS_USD, types.QUARTER_CONTRACT)
	//ok.OKExV3FuturesWs.SubscribeTrade(goex.EOS_USD, goex.QUARTER_CONTRACT)
	time.Sleep(1 * time.Minute)
}
