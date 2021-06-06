package bitmex

import (
	"github.com/fpChan/goex/types"
	"os"
	"testing"
	"time"
)

func TestNewSwapWs(t *testing.T) {
	os.Setenv("HTTPS_PROXY", "socks5://127.0.0.1:1080")
	ws := NewSwapWs()
	ws.DepthCallback(func(depth *types.Depth) {
		t.Log(depth)
	})
	ws.TickerCallback(func(ticker *types.FutureTicker) {
		t.Logf("%s %v", ticker.ContractType, ticker.Ticker)
	})
	//ws.SubscribeDepth(goex.NewCurrencyPair2("LTC_USD"), goex.SWAP_CONTRACT)
	ws.SubscribeTicker(types.LTC_USDT, types.SWAP_CONTRACT)

	time.Sleep(5 * time.Minute)
}
