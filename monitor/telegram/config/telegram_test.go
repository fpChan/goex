package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestMakeNewPair(t *testing.T) {
	var cfg = TelegramConfig {
		TgURL:       "https://api.telegram.org/bot5109509737:AAEDZS0frdLbY8OPnPl8oBKNDmxlOfjZ2qI/sendMessage",
		TgChatID:    "-626079521",
		ProxyScheme: "",
		ProxyHost:   "",
		Users:       []string{},
	}

	cfgByte, _ := json.Marshal(cfg)
	err := ioutil.WriteFile("okex_price.json", cfgByte, os.ModeAppend)
	if err != nil {
		fmt.Println(err)
	}
}
