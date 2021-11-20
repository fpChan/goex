package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

type MonitorClient interface {
	SendMsg(msg string) error
	GetMsgCh() chan string
}

type TelegramMonitor struct {
	tgURL      string
	chatid     string
	httpClient *http.Client
	msgCh      chan string
}

func NewTelegramMonitor(tgURL, chatid, proxyScheme, proxyHost string) TelegramMonitor {
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
		tgURL:      tgURL,
		chatid:     chatid,
		httpClient: client,
		msgCh:      make(chan string, 1),
	}
}

func (tgMonitor TelegramMonitor) SendMsg(msg string) error {
	params := url.Values{}
	Url, _ := url.Parse(tgMonitor.tgURL)

	params.Set("chat_id", tgMonitor.chatid)
	params.Set("text", msg)
	Url.RawQuery = params.Encode()
	urlPath := Url.String()

	response, err := tgMonitor.httpClient.Get(urlPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"step": "send msg"}).Error(err.Error())
		return err
	}
	logrus.Info(response)
	return nil
}

func (tgMonitor TelegramMonitor) GetMsgCh() chan string {
	return tgMonitor.msgCh
}
