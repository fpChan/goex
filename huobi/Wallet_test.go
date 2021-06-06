package huobi

import (
	"github.com/fpChan/goex/types"
	"testing"
)

var wallet *Wallet

func init() {
	wallet = NewWallet(&types.APIConfig{
		HttpClient:   httpProxyClient,
		ApiKey:       "",
		ApiSecretKey: "",
	})
}

func TestWallet_Transfer(t *testing.T) {
	t.Log(wallet.Transfer(types.TransferParameter{
		Currency: "BTC",
		From:     types.SWAP_USDT,
		To:       types.SPOT,
		Amount:   11,
	}))
}
