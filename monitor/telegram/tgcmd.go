package main

import (
	"github.com/fpChan/goex/monitor"
	"github.com/fpChan/goex/okex"
	"github.com/fpChan/goex/types"
	"net/http"
	"net/url"
)

const (
	tgURL       = "https://api.telegram.org/bot1830414088:AAGMc_sB_XWcqmY7AebZX2eW1SpGu3pAZOE/sendMessage"
	tgChatID    = "879754066"
	okexURL     = "https://www.okex.com"
	proxyScheme = "" //"socks5"
	proxyHost   = "127.0.0.1:10000"
)

func main() {
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

	var monitorTool monitor.Monitor
	var okEx = okex.NewOKEx(&types.APIConfig{
		Endpoint:      okexURL,
		HttpClient:    client,
		ApiKey:        "",
		ApiSecretKey:  "",
		ApiPassphrase: "",
	})
	var symbols = []types.CurrencyPair{types.SHIB_USDT, types.SUSHI_USDT, types.UNI_USDT, types.LINCH_USDT, types.KSM_USDT, types.MATIC_USDT, types.THETA_USDT, types.BTC_USDT, types.ETH_USD}
	monitorTool = monitor.NewTelegramMonitor(okEx.OKExFuture, tgURL, tgChatID, proxyScheme, proxyHost, symbols)
	if err := monitorTool.Start(); err != nil {
		return
	}
}
