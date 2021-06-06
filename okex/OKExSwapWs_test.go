package okex

import (
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"net/http"
	"os"
	"testing"
	"time"
)

func init() {
	logger.SetLevel(logger.DEBUG)
}

func TestNewOKExV3SwapWs(t *testing.T) {
	os.Setenv("HTTPS_PROXY", "socks5://127.0.0.1:1080")
	ok := NewOKEx(&types.APIConfig{
		HttpClient: http.DefaultClient,
	})
	ok.OKExV3SwapWs.TickerCallback(func(ticker *types.FutureTicker) {
		t.Log(ticker.Ticker, ticker.ContractType)
	})
	ok.OKExV3SwapWs.DepthCallback(func(depth *types.Depth) {
		t.Log(depth)
	})
	ok.OKExV3SwapWs.TradeCallback(func(trade *types.Trade, s string) {
		t.Log(s, trade)
	})
	ok.OKExV3SwapWs.SubscribeTicker(types.BTC_USDT, types.SWAP_CONTRACT)
	time.Sleep(1 * time.Minute)
}
