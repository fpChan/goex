package main

import (
	"fmt"
	"github.com/fpChan/goex/common/exchange"
	"github.com/fpChan/goex/monitor/telegram/config"
	"github.com/fpChan/goex/okex"
	"github.com/fpChan/goex/types"
	"log"
	"time"
)

const (
	okexURL     = "https://www.okex.com"
)

func main() {

	cfg := config.ProcessTGArgs()
	tgClient := config.NewClient(&cfg)
	var okEx = okex.NewOKEx(&types.APIConfig{
		Endpoint:      okexURL,
		HttpClient:    tgClient.HttpClient,
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
	})
	var symbols = []types.CurrencyPair{types.SHIB_USDT, types.SUSHI_USDT, types.UNI_USDT, types.LINCH_USDT, types.KSM_USDT, types.MATIC_USDT, types.THETA_USDT, types.BTC_USDT, types.ETH_USD}
	if err := StartPriceMonitor(tgClient, symbols, okEx.OKExFuture); err != nil {
		return
	}
}

func StartPriceMonitor(monitorClient config.MonitorClient, targetSymbols []types.CurrencyPair, futureClient exchange.ExpandFutureRestAPI) error {
	for {
		select {
		case msg, ok := <-monitorClient.GetMsgCh():
			if ok {
				monitorClient.SendMsg(msg)
			}

		case <-time.After(30 * time.Second):
			var msg = ""
			for _, symbol := range targetSymbols {
				candles, err := futureClient.GetKlineRecords(types.SWAP_CONTRACT, symbol, types.KLINE_PERIOD_1MIN, 0)
				if err != nil {
					log.Fatal(fmt.Sprintf("failed to get future %s ticker by", symbol), err)
				}
				var changePercent = (candles[0].Close - candles[0].Open) / candles[0].Open * 100
				fmt.Printf("changePercent %f \t", changePercent)
				if changePercent > -0.0001 && changePercent < 0.0001 {
					continue
				}
				msg = fmt.Sprintf("%s change percent: %0.2f %%\t high:%f\t low:%f\t price: %f\n time %s\n%s", symbol, changePercent, candles[0].High, candles[0].Low, candles[0].Close, time.Unix(candles[0].Timestamp, 0), msg)
			}
			monitorClient.GetMsgCh() <- msg

		}
	}
	return nil
}
