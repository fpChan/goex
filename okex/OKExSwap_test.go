package okex

import (
	"github.com/fpChan/goex/types"
	"net/http"
	"net/url"
	"testing"
	"time"
)

var config = &types.APIConfig{
	HttpClient: &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return &url.URL{
					Scheme: "socks5",
					Host:   "127.0.0.1:1080"}, nil
			},
		},
	},
	Endpoint:      "https://www.okex.com",
	ApiKey:        "",
	ApiSecretKey:  "",
	ApiPassphrase: "",
}

var okExSwap = NewOKExSwap(config)

func TestOKExSwap_GetFutureUserinfo(t *testing.T) {
	t.Log(okExSwap.GetFutureUserinfo())
}

func TestOKExSwap_PlaceFutureOrder(t *testing.T) {
	t.Log(okExSwap.PlaceFutureOrder(types.BTC_USDT, types.SWAP_CONTRACT, "10000", "1", types.OPEN_BUY, 0, 0))
}

func TestOKExSwap_PlaceFutureOrder2(t *testing.T) {
	t.Log(okExSwap.PlaceFutureOrder2(types.BTC_USDT, types.SWAP_CONTRACT, "10000", "1", types.OPEN_BUY, 0, types.Ioc))
}

func TestOKExSwap_FutureCancelOrder(t *testing.T) {
	t.Log(okExSwap.FutureCancelOrder(types.BTC_USDT, types.SWAP_CONTRACT, "309935122485305344"))
}

func TestOKExSwap_GetFutureOrder(t *testing.T) {
	t.Log(okExSwap.GetFutureOrder("581084124456583168", types.BTC_USDT, types.SWAP_CONTRACT))
}

func TestOKExSwap_GetFuturePosition(t *testing.T) {
	t.Log(okExSwap.GetFuturePosition(types.BTC_USD, types.SWAP_CONTRACT))
}

func TestOKExSwap_GetFutureDepth(t *testing.T) {
	t.Log(okExSwap.GetFutureDepth(types.LTC_USD, types.SWAP_CONTRACT, 10))
}

func TestOKExSwap_GetFutureTicker(t *testing.T) {
	t.Log(okExSwap.GetFutureTicker(types.BTC_USD, types.SWAP_CONTRACT))
}

func TestOKExSwap_GetUnfinishFutureOrders(t *testing.T) {
	ords, _ := okExSwap.GetUnfinishFutureOrders(types.XRP_USD, types.SWAP_CONTRACT)
	for _, ord := range ords {
		t.Log(ord.OrderID2, ord.ClientOid)
	}

}

func TestOKExSwap_GetHistoricalFunding(t *testing.T) {
	for i := 1; ; i++ {
		funding, err := okExSwap.GetHistoricalFunding(types.SWAP_CONTRACT, types.BTC_USD, i)
		t.Log(err, len(funding))
	}
}

func TestOKExSwap_GetKlineRecords(t *testing.T) {
	time.Now().Add(-24 * time.Hour).Unix()
	kline, err := okExSwap.GetKlineRecords(types.SWAP_CONTRACT, types.BTC_USD, types.KLINE_PERIOD_4H, 0)
	t.Log(err, kline[0].Kline)
}

func TestOKExSwap_GetKlineRecords2(t *testing.T) {
	start := time.Now().Add(time.Minute * -30).UTC().Format(time.RFC3339)
	t.Log(start)
	kline, err := okExSwap.GetKlineRecords2(types.SWAP_CONTRACT, types.BTC_USDT, start, "", "900")
	t.Log(err, kline[0].Kline)
}

func TestOKExSwap_GetInstruments(t *testing.T) {
	t.Log(okExSwap.GetInstruments())
}

func TestOKExSwap_SetMarginLevel(t *testing.T) {
	t.Log(okExSwap.SetMarginLevel(types.EOS_USDT, 5, 3))
}

func TestOKExSwap_GetMarginLevel(t *testing.T) {
	t.Log(okExSwap.GetMarginLevel(types.EOS_USDT))
}

func TestOKExSwap_GetFutureAccountInfo(t *testing.T) {
	t.Log(okExSwap.GetFutureAccountInfo(types.BTC_USDT))
}

func TestOKExSwap_PlaceFutureAlgoOrder(t *testing.T) {
	ord := &types.FutureOrder{
		ContractName: types.SWAP_CONTRACT,
		Currency:     types.BTC_USD,
		OType:        2, //开空
		OrderType:    1, //1：止盈止损 2：跟踪委托 3：冰山委托 4：时间加权
		Price:        9877,
		Amount:       1,

		TriggerPrice: 9877,
		AlgoType:     1,
	}
	t.Log(okExSwap.PlaceFutureAlgoOrder(ord))
}

func TestOKExSwap_FutureCancelAlgoOrder(t *testing.T) {
	t.Log(okExSwap.FutureCancelAlgoOrder(types.BTC_USD, []string{"309935122485305344"}))

}

func TestOKExSwap_GetFutureAlgoOrders(t *testing.T) {
	t.Log(okExSwap.GetFutureAlgoOrders("", "2", types.BTC_USD))
}
