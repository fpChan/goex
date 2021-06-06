package exchange

import (
	"github.com/fpChan/goex/types"
)

// api interface

type API interface {
	LimitBuy(amount, price string, currency types.CurrencyPair, opt ...types.LimitOrderOptionalParameter) (*types.Order, error)
	LimitSell(amount, price string, currency types.CurrencyPair, opt ...types.LimitOrderOptionalParameter) (*types.Order, error)
	MarketBuy(amount, price string, currency types.CurrencyPair) (*types.Order, error)
	MarketSell(amount, price string, currency types.CurrencyPair) (*types.Order, error)
	CancelOrder(orderId string, currency types.CurrencyPair) (bool, error)
	GetOneOrder(orderId string, currency types.CurrencyPair) (*types.Order, error)
	GetUnfinishOrders(currency types.CurrencyPair) ([]types.Order, error)
	GetOrderHistorys(currency types.CurrencyPair, opt ...types.OptionalParameter) ([]types.Order, error)
	GetAccount() (*types.Account, error)

	GetTicker(currency types.CurrencyPair) (*types.Ticker, error)
	GetDepth(size int, currency types.CurrencyPair) (*types.Depth, error)
	GetKlineRecords(currency types.CurrencyPair, period types.KlinePeriod, size int, optional ...types.OptionalParameter) ([]types.Kline, error)
	//非个人，整个交易所的交易记录
	GetTrades(currencyPair types.CurrencyPair, since int64) ([]types.Trade, error)

	GetExchangeName() string
}
