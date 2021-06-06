package huobi

import (
	"github.com/fpChan/goex/types"
	"testing"
	"time"
)

var dm = NewHbdm(&types.APIConfig{
	Endpoint:     "https://api.hbdm.com",
	HttpClient:   httpProxyClient,
	ApiKey:       "",
	ApiSecretKey: ""})

func TestHbdm_GetFutureUserinfo(t *testing.T) {
	t.Log(dm.GetFutureUserinfo())
}

func TestHbdm_GetFuturePosition(t *testing.T) {
	t.Log(dm.GetFuturePosition(types.BTC_USD, types.QUARTER_CONTRACT))
}

func TestHbdm_PlaceFutureOrder(t *testing.T) {
	t.Log(dm.PlaceFutureOrder(types.BTC_USD, types.QUARTER_CONTRACT, "3800", "1", types.OPEN_BUY, 0, 20))
}

func TestHbdm_FutureCancelOrder(t *testing.T) {
	t.Log(dm.FutureCancelOrder(types.BTC_USD, types.QUARTER_CONTRACT, "6"))
}

func TestHbdm_GetUnfinishFutureOrders(t *testing.T) {
	t.Log(dm.GetUnfinishFutureOrders(types.BTC_USD, types.QUARTER_CONTRACT))
}

func TestHbdm_GetFutureOrders(t *testing.T) {
	t.Log(dm.GetFutureOrders([]string{"6", "5"}, types.BTC_USD, types.QUARTER_CONTRACT))
}

func TestHbdm_GetFutureOrder(t *testing.T) {
	t.Log(dm.GetFutureOrder("6", types.BTC_USD, types.QUARTER_CONTRACT))
}

func TestHbdm_GetFutureTicker(t *testing.T) {
	t.Log(dm.GetFutureTicker(types.EOS_USD, types.QUARTER_CONTRACT))
}

func TestHbdm_GetFutureDepth(t *testing.T) {
	dep, err := dm.GetFutureDepth(types.BTC_USD, types.QUARTER_CONTRACT, 0)
	t.Log(err)
	t.Logf("%+v\n%+v", dep.AskList, dep.BidList)
}
func TestHbdm_GetFutureIndex(t *testing.T) {
	t.Log(dm.GetFutureIndex(types.BTC_USD))
}

func TestHbdm_GetFutureEstimatedPrice(t *testing.T) {
	t.Log(dm.GetFutureEstimatedPrice(types.BTC_USD))
}

func TestHbdm_GetKlineRecords(t *testing.T) {
	klines, _ := dm.GetKlineRecords(types.QUARTER_CONTRACT, types.EOS_USD, types.KLINE_PERIOD_1MIN, 20)
	for _, k := range klines {
		tt := time.Unix(k.Timestamp, 0)
		t.Log(k.Pair, tt, k.Open, k.Close, k.High, k.Low, k.Vol, k.Vol2)
	}
}
