package binance

import (
	"github.com/fpChan/goex/types"
	"net/http"
	"testing"
)

var wallet *Wallet

func init() {
	wallet = NewWallet(&types.APIConfig{
		HttpClient:   http.DefaultClient,
		ApiKey:       "",
		ApiSecretKey: "",
	})
}

func TestWallet_Transfer(t *testing.T) {
	t.Log(wallet.Transfer(types.TransferParameter{
		Currency: "USDT",
		From:     types.SPOT,
		To:       types.SWAP_USDT,
		Amount:   100,
	}))
}
