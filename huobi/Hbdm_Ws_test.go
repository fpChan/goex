package huobi

import (
	"github.com/fpChan/goex/types"
	"log"
	"testing"
	"time"
)

func TestNewHbdmWs(t *testing.T) {
	ws := NewHbdmWs()
	ws.ProxyUrl("socks5://127.0.0.1:1080")

	ws.SetCallbacks(func(ticker *types.FutureTicker) {
		log.Println(ticker.Ticker)
	}, func(depth *types.Depth) {
		log.Println(">>>>>>>>>>>>>>>")
		log.Println(depth.ContractType, depth.Pair)
		log.Println(depth.BidList)
		log.Println(depth.AskList)
		log.Println("<<<<<<<<<<<<<<")
	}, func(trade *types.Trade, s string) {
		log.Println(s, trade)
	})

	t.Log(ws.SubscribeTicker(types.BTC_USD, types.QUARTER_CONTRACT))
	t.Log(ws.SubscribeDepth(types.BTC_USD, types.NEXT_WEEK_CONTRACT))
	t.Log(ws.SubscribeTrade(types.LTC_USD, types.THIS_WEEK_CONTRACT))
	time.Sleep(time.Minute)
}
