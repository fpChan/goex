package binance

import (
	"github.com/fpChan/goex/types"
	"log"
	"os"
	"testing"
	"time"
)

var futuresWs *FuturesWs

func createFuturesWs() {
	os.Setenv("HTTPS_PROXY", "socks5://127.0.0.1:1080")

	futuresWs = NewFuturesWs()

	futuresWs.DepthCallback(func(depth *types.Depth) {
		log.Println(depth)
	})

	futuresWs.TickerCallback(func(ticker *types.FutureTicker) {
		log.Println(ticker.Ticker, ticker.ContractType)
	})
}

func TestFuturesWs_DepthCallback(t *testing.T) {
	createFuturesWs()

	futuresWs.SubscribeDepth(types.LTC_USDT, types.SWAP_USDT_CONTRACT)
	futuresWs.SubscribeDepth(types.LTC_USDT, types.SWAP_CONTRACT)
	futuresWs.SubscribeDepth(types.LTC_USDT, types.QUARTER_CONTRACT)

	time.Sleep(30 * time.Second)
}

func TestFuturesWs_SubscribeTicker(t *testing.T) {
	createFuturesWs()

	futuresWs.SubscribeTicker(types.BTC_USDT, types.SWAP_USDT_CONTRACT)
	futuresWs.SubscribeTicker(types.BTC_USDT, types.SWAP_CONTRACT)
	futuresWs.SubscribeTicker(types.BTC_USDT, types.QUARTER_CONTRACT)

	time.Sleep(30 * time.Second)
}
