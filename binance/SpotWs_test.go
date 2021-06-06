package binance

import (
	"github.com/fpChan/goex/types"
	"log"
	"os"
	"testing"
	"time"
)

var spotWs *SpotWs

func createSpotWs() {
	os.Setenv("HTTPS_PROXY", "socks5://127.0.0.1:1080")
	spotWs = NewSpotWs()
	spotWs.DepthCallback(func(depth *types.Depth) {
		log.Println(depth)
	})
	spotWs.TickerCallback(func(ticker *types.Ticker) {
		log.Println(ticker)
	})
}

func TestSpotWs_DepthCallback(t *testing.T) {
	createSpotWs()

	spotWs.SubscribeDepth(types.BTC_USDT)
	spotWs.SubscribeTicker(types.LTC_USDT)
	time.Sleep(11 * time.Minute)
}

func TestSpotWs_SubscribeTicker(t *testing.T) {
	createSpotWs()

	spotWs.SubscribeTicker(types.LTC_USDT)
	time.Sleep(30 * time.Minute)
}
