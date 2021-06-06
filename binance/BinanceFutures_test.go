package binance

import (
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"net/http"
	"testing"
)

var baDapi = NewBinanceFutures(&types.APIConfig{
	HttpClient:   http.DefaultClient,
	ApiKey:       "",
	ApiSecretKey: "",
})

func init() {
	logger.SetLevel(logger.DEBUG)
}

func TestBinanceFutures_GetFutureDepth(t *testing.T) {
	t.Log(baDapi.GetFutureDepth(types.ETH_USD, types.QUARTER_CONTRACT, 10))
}

func TestBinanceSwap_GetFutureTicker(t *testing.T) {
	ticker, err := baDapi.GetFutureTicker(types.LTC_USD, types.SWAP_CONTRACT)
	t.Log(err)
	t.Logf("%+v", ticker)
}

func TestBinance_GetExchangeInfo(t *testing.T) {
	baDapi.GetExchangeInfo()
}

func TestBinanceFutures_GetFutureUserinfo(t *testing.T) {
	t.Log(baDapi.GetFutureUserinfo())
}

func TestBinanceFutures_PlaceFutureOrder(t *testing.T) {
	//1044675677
	t.Log(baDapi.PlaceFutureOrder(types.BTC_USD, types.QUARTER_CONTRACT, "19990", "2", types.OPEN_SELL, 0, 10))
}

func TestBinanceFutures_LimitFuturesOrder(t *testing.T) {
	t.Log(baDapi.LimitFuturesOrder(types.BTC_USD, types.QUARTER_CONTRACT, "20001", "2", types.OPEN_SELL))
}

func TestBinanceFutures_MarketFuturesOrder(t *testing.T) {
	t.Log(baDapi.MarketFuturesOrder(types.BTC_USD, types.QUARTER_CONTRACT, "2", types.OPEN_SELL))
}

func TestBinanceFutures_GetFutureOrder(t *testing.T) {
	t.Log(baDapi.GetFutureOrder("1045208666", types.BTC_USD, types.QUARTER_CONTRACT))
}

func TestBinanceFutures_FutureCancelOrder(t *testing.T) {
	t.Log(baDapi.FutureCancelOrder(types.BTC_USD, types.QUARTER_CONTRACT, "1045328328"))
}

func TestBinanceFutures_GetFuturePosition(t *testing.T) {
	t.Log(baDapi.GetFuturePosition(types.BTC_USD, types.QUARTER_CONTRACT))
}

func TestBinanceFutures_GetUnfinishFutureOrders(t *testing.T) {
	t.Log(baDapi.GetUnfinishFutureOrders(types.BTC_USD, types.QUARTER_CONTRACT))
}
