package main

import (
	"github.com/fpChan/goex"
	"github.com/fpChan/goex/monitor"
	"github.com/fpChan/goex/okex"
	"net/http"
	"net/url"
)

const (
	tgURL       = "https://api.telegram.org/bot1830414088:AAGMc_sB_XWcqmY7AebZX2eW1SpGu3pAZOE/sendMessage"
	tgChatID    = "879754066"
	okexURL     = "https://www.okex.com"
	proxyScheme = "socks5"
	proxyHost   = "127.0.0.1:10000"
)

func main() {

	var monitorTool monitor.Monitor
	var okEx = okex.NewOKEx(&goex.APIConfig{
		Endpoint: okexURL,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return &url.URL{
						Scheme: proxyScheme,
						Host:   proxyHost}, nil
				},
			},
		},
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
	})
	var symbols = []goex.CurrencyPair{goex.SHIB_USDT, goex.SUSHI_USDT, goex.UNI_USDT, goex.LINCH_USDT, goex.KSM_USDT, goex.MATIC_USDT, goex.THETA_USDT, goex.BTC_USDT, goex.ETH_USD}
	monitorTool = monitor.NewTelegramMonitor(okEx.OKExFuture, tgURL, tgChatID, proxyScheme, proxyHost, symbols)
	if err := monitorTool.Start(); err != nil {
		return
	}
}
