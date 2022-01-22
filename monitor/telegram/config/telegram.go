package config

import (
	"encoding/json"
	"flag"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
)

type MonitorClient interface {
	SendMsg(msg string) error
	GetMsgCh() chan string
}

type TelegramMonitor struct {
	TgURL  string
	Chatid string
	HttpClient *http.Client
	MsgCh      chan string
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
		TgURL:      tgURL,
		Chatid:     chatid,
		HttpClient: client,
		MsgCh:      make(chan string, 1),
	}
}

func (tgMonitor TelegramMonitor) SendMsg(msg string) error {
	params := url.Values{}
	Url, _ := url.Parse(tgMonitor.TgURL)

	params.Set("chat_id", tgMonitor.Chatid)
	params.Set("text", msg)
	Url.RawQuery = params.Encode()
	urlPath := Url.String()

	_, err := tgMonitor.HttpClient.Get(urlPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"step": "send msg"}).Error(err.Error())
		return err
	}
	return nil
}

func (tgMonitor TelegramMonitor) GetMsgCh() chan string {
	return tgMonitor.MsgCh
}

type TelegramConfig struct {
	TgURL       string   `json:"tg_url"`
	TgChatID    string   `json:"tg_chat_id"`
	ProxyScheme string   `json:"proxy_scheme"`
	ProxyHost   string   `json:"proxy_host"`
	Users       []string `json:"users"`
}


type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	if err = json.Unmarshal(data, v); err != nil {
		return
	}
}

func ProcessTGArgs() TelegramConfig {
	var cfgStr string
	flag.StringVar(&cfgStr, "config", "", "config file that include rpc urls")
	flag.Parse()
	JsonParse := NewJsonStruct()
	cfg := TelegramConfig{}
	JsonParse.Load(cfgStr, &cfg)
	return cfg
}


func NewClient(cfg *TelegramConfig) (tgClient TelegramMonitor){
	return NewTelegramMonitor(cfg.TgURL, cfg.TgChatID, cfg.ProxyScheme, cfg.ProxyHost)
}