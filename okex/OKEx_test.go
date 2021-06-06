package okex

import (
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func init() {
	logger.Log.SetLevel(logger.DEBUG)
}

//
var config2 = &types.APIConfig{
	Endpoint:   "https://www.okex.me",
	HttpClient: http.DefaultClient,
}

var okex = NewOKEx(config2) //线上请用APIBuilder构建

func TestOKExSpot_GetAccount(t *testing.T) {
	t.Log(okex.GetAccount())
}

func TestOKExSpot_BatchPlaceOrders(t *testing.T) {
	t.Log(okex.OKExSpot.BatchPlaceOrders([]types.Order{
		types.Order{
			Cid:       okex.UUID(),
			Currency:  types.XRP_USD,
			Amount:    10,
			Price:     0.32,
			Side:      types.BUY,
			Type:      "limit",
			OrderType: types.ORDER_FEATURE_ORDINARY,
		},
		{
			Cid:       okex.UUID(),
			Currency:  types.EOS_USD,
			Amount:    1,
			Price:     5.2,
			Side:      types.BUY,
			OrderType: types.ORDER_FEATURE_ORDINARY,
		},
		types.Order{
			Cid:       okex.UUID(),
			Currency:  types.XRP_USD,
			Amount:    10,
			Price:     0.33,
			Side:      types.BUY,
			Type:      "limit",
			OrderType: types.ORDER_FEATURE_ORDINARY,
		}}))
}

func TestOKExSpot_LimitBuy(t *testing.T) {
	t.Log(okex.OKExSpot.LimitBuy("0.001", "9910", types.BTC_USD))
}

func TestOKExSpot_CancelOrder(t *testing.T) {
	t.Log(okex.OKExSpot.CancelOrder("2a647e51435647708b1c840802bf70e5", types.BTC_USD))

}

func TestOKExSpot_GetOneOrder(t *testing.T) {
	t.Log(okex.OKExSpot.GetOneOrder("5502594029936640", types.BTC_USD))
}

func TestOKExSpot_GetUnfinishOrders(t *testing.T) {
	t.Log(okex.OKExSpot.GetUnfinishOrders(types.EOS_USD))
}

func TestOKExSpot_GetTicker(t *testing.T) {
	t.Log(okex.OKExSpot.GetTicker(types.BTC_USD))
}

func TestOKExSpot_GetDepth(t *testing.T) {
	dep, err := okex.OKExSpot.GetDepth(2, types.EOS_USD)
	assert.Nil(t, err)
	t.Log(dep.AskList)
	t.Log(dep.BidList)
}

func TestOKExFuture_GetFutureTicker(t *testing.T) {
	t.Log(okex.OKExFuture.GetFutureTicker(types.BTC_USD, "BTC-USD-190927"))
	t.Log(okex.OKExFuture.GetFutureTicker(types.BTC_USD, types.QUARTER_CONTRACT))
	t.Log(okex.OKExFuture.GetFutureDepth(types.BTC_USD, types.QUARTER_CONTRACT, 2))
	t.Log(okex.OKExFuture.GetContractValue(types.XRP_USD))
	t.Log(okex.OKExFuture.GetFutureIndex(types.EOS_USD))
	t.Log(okex.OKExFuture.GetFutureEstimatedPrice(types.EOS_USD))
}

func TestOKExFuture_GetFutureUserinfo(t *testing.T) {
	t.Log(okex.OKExFuture.GetFutureUserinfo())
}

func TestOKExFuture_GetFuturePosition(t *testing.T) {
	t.Log(okex.OKExFuture.GetFuturePosition(types.EOS_USD, types.QUARTER_CONTRACT))
}

func TestOKExFuture_PlaceFutureOrder(t *testing.T) {
	t.Log(okex.OKExFuture.PlaceFutureOrder(types.EOS_USD, types.THIS_WEEK_CONTRACT, "5.8", "1", types.OPEN_BUY, 0, 10))
}

func TestOKExFuture_PlaceFutureOrder2(t *testing.T) {
	t.Log(okex.OKExFuture.PlaceFutureOrder2(0, &types.FutureOrder{
		Currency:     types.EOS_USD,
		ContractName: types.QUARTER_CONTRACT,
		OType:        types.OPEN_BUY,
		OrderType:    types.ORDER_FEATURE_ORDINARY,
		Price:        5.9,
		Amount:       10,
		LeverRate:    10}))
}

func TestOKExFuture_FutureCancelOrder(t *testing.T) {
	t.Log(okex.OKExFuture.FutureCancelOrder(types.EOS_USD, types.QUARTER_CONTRACT, "e88bd3361de94512b8acaf9aa154f95a"))
}

func TestOKExFuture_GetFutureOrder(t *testing.T) {
	t.Log(okex.OKExFuture.GetFutureOrder("3145664744431616", types.EOS_USD, types.QUARTER_CONTRACT))
}

func TestOKExFuture_GetUnfinishFutureOrders(t *testing.T) {
	t.Log(okex.OKExFuture.GetUnfinishFutureOrders(types.EOS_USD, types.QUARTER_CONTRACT))
}

func TestOKExFuture_MarketCloseAllPosition(t *testing.T) {
	t.Log(okex.OKExFuture.MarketCloseAllPosition(types.BTC_USD, types.THIS_WEEK_CONTRACT, types.CLOSE_BUY))
}

func TestOKExFuture_GetRate(t *testing.T) {
	t.Log(okex.OKExFuture.GetRate())
}

func TestOKExFuture_GetKlineRecords(t *testing.T) {
	time.Now().Add(-24 * time.Hour).Unix()
	kline, err := okex.OKExFuture.GetKlineRecords(types.QUARTER_CONTRACT, types.BTC_USD, types.KLINE_PERIOD_4H, 0)
	assert.Nil(t, err)
	for _, k := range kline {
		t.Logf("%+v", k.Kline)
	}
}

func TestOKExWallet_GetAccount(t *testing.T) {
	t.Log(okex.OKExWallet.GetAccount())
}

//func TestOKExWallet_Transfer(t *testing.T) {
//	t.Log(okex.OKExWallet.Transfer(TransferParameter{
//		Currency:     goex.EOS.Symbol,
//		From:         SPOT,
//		To:           SPOT_MARGIN,
//		Amount:       20,
//		InstrumentId: goex.EOS_USDT.ToLower().ToSymbol("-")}))
//}
//
//func TestOKExWallet_Withdrawal(t *testing.T) {
//	t.Log(okex.OKExWallet.Withdrawal(WithdrawParameter{
//		Currency:    goex.EOS.Symbol,
//		Amount:      100,
//		Destination: 2,
//		ToAddress:   "",
//		TradePwd:    "",
//		Fee:         "0.01",
//	}))
//}

func TestOKExWallet_GetDepositAddress(t *testing.T) {
	t.Log(okex.OKExWallet.GetDepositAddress(types.BTC))
}

func TestOKExWallet_GetWithDrawalFee(t *testing.T) {
	t.Log(okex.OKExWallet.GetWithDrawalFee(nil))
}

func TestOKExWallet_GetDepositHistory(t *testing.T) {
	t.Log(okex.OKExWallet.GetDepositHistory(&types.BTC))
}

func TestOKExWallet_GetWithDrawalHistory(t *testing.T) {
	//t.Log(okex.OKExWallet.GetWithDrawalHistory(&goex.XRP))
}

func TestOKExMargin_GetMarginAccount(t *testing.T) {
	t.Log(okex.OKExMargin.GetMarginAccount(types.EOS_USDT))
}

func TestOKExMargin_Borrow(t *testing.T) {
	t.Log(okex.OKExMargin.Borrow(types.BorrowParameter{
		Currency:     types.EOS,
		CurrencyPair: types.EOS_USDT,
		Amount:       10,
	}))
}

func TestOKExMargin_Repayment(t *testing.T) {
	t.Log(okex.OKExMargin.Repayment(types.RepaymentParameter{
		BorrowParameter: types.BorrowParameter{
			Currency:     types.EOS,
			CurrencyPair: types.EOS_USDT,
			Amount:       10},
		BorrowId: "123"}))
}

func TestOKExMargin_PlaceOrder(t *testing.T) {
	t.Log(okex.OKExMargin.PlaceOrder(&types.Order{
		Currency:  types.EOS_USDT,
		Amount:    0.2,
		Price:     6,
		Type:      "limit",
		OrderType: types.ORDER_FEATURE_ORDINARY,
		Side:      types.SELL,
	}))
}

func TestOKExMargin_GetUnfinishOrders(t *testing.T) {
	t.Log(okex.OKExMargin.GetUnfinishOrders(types.EOS_USDT))
}

func TestOKExMargin_CancelOrder(t *testing.T) {
	t.Log(okex.OKExMargin.CancelOrder("3174778420532224", types.EOS_USDT))
}

func TestOKExMargin_GetOneOrder(t *testing.T) {
	t.Log(okex.OKExMargin.GetOneOrder("3174778420532224", types.EOS_USDT))
}

func TestOKExSpot_GetCurrenciesPrecision(t *testing.T) {
	t.Log(okex.OKExSpot.GetCurrenciesPrecision())
}

func TestOKExSpot_GetOrderHistorys(t *testing.T) {
	orders, err := okex.OKExSpot.GetOrderHistorys(types.NewCurrencyPair2("DASH_USDT"))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(len(orders))
}
