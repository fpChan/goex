package main

import (
	"fmt"
	exchange2 "github.com/fpChan/goex/common/exchange"
	"github.com/fpChan/goex/types"
	"log"
	"net/http"
	"net/url"
	"time"
)

type TelegramMonitor struct {
	tgURL         string
	chatid        string
	httpClient    *http.Client
	futureClient  exchange2.ExpandFutureRestAPI
	targetSymbols []types.CurrencyPair
	msgCh         chan string
}

func NewTelegramMonitor(futureClient exchange2.ExpandFutureRestAPI, tgURL, chatid, proxyScheme, proxyHost string, targetSymbols []types.CurrencyPair) TelegramMonitor {
	var client = &http.Client{
		Transport: &http.Transport{},
	}

	if proxyScheme != "" {
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return &url.URL{
						Scheme: proxyScheme,
						Host:   proxyHost,
					}, nil
				},
			},
		}
	}

	return TelegramMonitor{
		tgURL:         tgURL,
		chatid:        chatid,
		httpClient:    client,
		futureClient:  futureClient,
		targetSymbols: targetSymbols,
		msgCh:         make(chan string, 1),
	}
}

func (tgMonitor TelegramMonitor) Start() error {
	for {
		select {
		case msg, ok := <-tgMonitor.msgCh:
			if ok {
				tgMonitor.send(msg)
			}

		case <-time.After(30 * time.Second):
			var msg = ""
			for _, symbol := range tgMonitor.targetSymbols {
				candles, err := tgMonitor.futureClient.GetKlineRecords(types.SWAP_CONTRACT, symbol, types.KLINE_PERIOD_1MIN, 0)
				if err != nil {
					log.Fatal(fmt.Sprintf("failed to get future %s ticker by", symbol), err)
				}
				var changePercent = (candles[0].Low - candles[0].High) / candles[0].High * 100
				fmt.Printf("changePercent %f \t", changePercent)
				if changePercent > -1 && changePercent < 1 {
					continue
				}
				//msg = fmt.Sprintf("%s change percent: %f\t high:%f\t low:%f\t price: %f\n%s", symbol , ticker.Last,  msg)
				msg = fmt.Sprintf("%s change percent: %0.2f %%\t high:%f\t low:%f\t price: %f\n time %s\n%s", symbol, changePercent, candles[0].High, candles[0].Low, candles[0].Close, time.Unix(candles[0].Timestamp, 0), msg)

				//msg = fmt.Sprintf("%s change percent: %s\t high:%s\t low:%s\t price: %s\n%s", ticker.Contract ,ticker.ChangePercent, ticker.High,ticker.Low,ticker.Close,  msg)
			}
			tgMonitor.msgCh <- msg

		}
	}
	return nil
}

func (tgMonitor TelegramMonitor) send(msg string) error {
	params := url.Values{}
	Url, err := url.Parse(tgMonitor.tgURL)
	if err != nil {
		log.Fatal("failed to parse url msg by", err.Error())
		return err
	}
	params.Set("chat_id", tgMonitor.chatid)
	params.Set("text", msg)
	Url.RawQuery = params.Encode()
	urlPath := Url.String()

	_, err = tgMonitor.httpClient.Get(urlPath)
	if err != nil {
		log.Fatal("failed to send msg by", err.Error())
		return err
	}

	//defer resp.Body.Close()
	//ioutil.ReadAll(resp.Body)
	return nil
}
