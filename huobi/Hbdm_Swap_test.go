package huobi

import (
	"github.com/fpChan/goex/types"
	"net/http"
	"testing"
	"time"
)

var swap *HbdmSwap

func init() {
	swap = NewHbdmSwap(&types.APIConfig{
		HttpClient:   http.DefaultClient,
		Endpoint:     "https://api.btcgateway.pro",
		ApiKey:       "",
		ApiSecretKey: "",
		Lever:        5,
	})
}

func TestHbdmSwap_GetFutureTicker(t *testing.T) {
	t.Log(swap.GetFutureTicker(types.BTC_USD, types.SWAP_CONTRACT))
}

func TestHbdmSwap_GetFutureDepth(t *testing.T) {
	dep, err := swap.GetFutureDepth(types.BTC_USD, types.SWAP_CONTRACT, 5)
	t.Log(err)
	t.Log(dep.AskList)
	t.Log(dep.BidList)
}

func TestHbdmSwap_GetFutureUserinfo(t *testing.T) {
	t.Log(swap.GetFutureUserinfo(types.NewCurrencyPair2("DOT_USD")))
}

func TestHbdmSwap_GetFuturePosition(t *testing.T) {
	t.Log(swap.GetFuturePosition(types.NewCurrencyPair2("DOT_USD"), types.SWAP_CONTRACT))
}

func TestHbdmSwap_LimitFuturesOrder(t *testing.T) {
	//784115347040780289
	t.Log(swap.LimitFuturesOrder(types.NewCurrencyPair2("DOT_USD"), types.SWAP_CONTRACT, "6.5", "1", types.OPEN_SELL))
}

func TestHbdmSwap_FutureCancelOrder(t *testing.T) {
	t.Log(swap.FutureCancelOrder(types.NewCurrencyPair2("DOT_USD"), types.SWAP_CONTRACT, "784118017750929408"))
}

func TestHbdmSwap_GetUnfinishFutureOrders(t *testing.T) {
	t.Log(swap.GetUnfinishFutureOrders(types.NewCurrencyPair2("DOT_USD"), types.SWAP_CONTRACT))
}

func TestHbdmSwap_GetFutureOrder(t *testing.T) {
	t.Log(swap.GetFutureOrder("784118017750929408", types.NewCurrencyPair2("DOT_USD"), types.SWAP_CONTRACT))
}

func TestHbdmSwap_GetFutureOrderHistory(t *testing.T) {
	t.Log(swap.GetFutureOrderHistory(types.NewCurrencyPair2("KSM_USD"), types.SWAP_CONTRACT,
		types.OptionalParameter{}.Optional("start_time", time.Now().Add(-5*24*time.Hour).Unix()*1000),
		types.OptionalParameter{}.Optional("end_time", time.Now().Unix()*1000)))
}
