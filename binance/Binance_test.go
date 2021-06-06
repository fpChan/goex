package binance

import (
	"fmt"
	"github.com/fpChan/goex/common/exchange"
	"github.com/fpChan/goex/types"
	"net/http"
	"testing"
	"time"
)

var ba = NewWithConfig(
	&types.APIConfig{
		HttpClient: http.DefaultClient,
		Endpoint:   "https://api.binancezh.pro",
	})

func TestBinance_GetTicker(t *testing.T) {
	ticker, err := ba.GetTicker(types.NewCurrencyPair2("USDT_USD"))
	t.Log(ticker, err)
}

func TestBinance_LimitBuy(t *testing.T) {
	order, err := ba.LimitBuy("3", "68.5", types.LTC_USDT)
	t.Log(order, err)
}

func TestBinance_LimitSell(t *testing.T) {
	order, err := ba.LimitSell("1", "90", types.LTC_USDT)
	t.Log(order, err)
}

func TestBinance_CancelOrder(t *testing.T) {
	r, er := ba.CancelOrder("3848718241", types.BTC_USDT)
	if !r {
		t.Log((er.(exchange.ApiError)).ErrCode)
	}
}

func TestBinance_GetOneOrder(t *testing.T) {
	odr, err := ba.GetOneOrder("3874087228", types.BTC_USDT)
	t.Log(err, odr)
}

func TestBinance_GetDepth(t *testing.T) {
	//return
	dep, err := ba.GetDepth(5, types.NewCurrencyPair2("BTC_USDT"))
	t.Log(err)
	if err == nil {
		t.Log(dep.AskList)
		t.Log(dep.BidList)
	}
}

func TestBinance_GetAccount(t *testing.T) {
	account, err := ba.GetAccount()
	t.Log(err, account)
}

func TestBinance_GetUnfinishOrders(t *testing.T) {
	orders, err := ba.GetUnfinishOrders(types.NewCurrencyPair2("BTC_USDT"))
	t.Log(orders, err)
}

func TestBinance_GetKlineRecords(t *testing.T) {
	startTime := time.Now().Add(-24*time.Hour).Unix() * 1000
	endTime := time.Now().Add(-5*time.Hour).Unix() * 1000

	kline, _ := ba.GetKlineRecords(types.ETH_USD, types.KLINE_PERIOD_5MIN, 100,
		types.OptionalParameter{}.Optional("startTime", fmt.Sprint(startTime)).Optional("endTime", fmt.Sprint(endTime)))

	for _, k := range kline {
		tt := time.Unix(k.Timestamp, 0)
		t.Log(tt, k.Open, k.Close, k.High, k.Low, k.Vol)
	}
}

func TestBinance_GetTrades(t *testing.T) {
	t.Log(ba.GetTrades(types.BTC_USDT, 0))
}

func TestBinance_GetTradeSymbols(t *testing.T) {
	t.Log(ba.GetTradeSymbol(types.BTC_USDT))
}

func TestBinance_SetTimeOffset(t *testing.T) {
	t.Log(ba.setTimeOffset())
	t.Log(ba.timeOffset)
}

func TestBinance_GetOrderHistorys(t *testing.T) {
	t.Log(ba.GetOrderHistorys(types.BTC_USDT,
		types.OptionalParameter{}.
			Optional("startTime", "1607656034333").
			Optional("limit", "5")))
}
