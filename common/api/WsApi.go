package api

import "github.com/fpChan/goex/types"

type FuturesWsApi interface {
	DepthCallback(func(depth *types.Depth))
	TickerCallback(func(ticker *types.FutureTicker))
	TradeCallback(func(trade *types.Trade, contract string))
	//OrderCallback(func(order *FutureOrder))
	//PositionCallback(func(position *FuturePosition))
	//AccountCallback(func(account *FutureAccount))

	SubscribeDepth(pair types.CurrencyPair, contractType string) error
	SubscribeTicker(pair types.CurrencyPair, contractType string) error
	SubscribeTrade(pair types.CurrencyPair, contractType string) error

	//Login() error
	//SubscribeOrder(pair CurrencyPair, contractType string) error
	//SubscribePosition(pair CurrencyPair, contractType string) error
	//SubscribeAccount(pair CurrencyPair) error
}

type SpotWsApi interface {
	DepthCallback(func(depth *types.Depth))
	TickerCallback(func(ticker *types.Ticker))
	TradeCallback(func(trade *types.Trade))
	//OrderCallback(func(order *Order))
	//AccountCallback(func(account *Account))

	SubscribeDepth(pair types.CurrencyPair) error
	SubscribeTicker(pair types.CurrencyPair) error
	SubscribeTrade(pair types.CurrencyPair) error

	//Login() error
	//SubscribeOrder(pair CurrencyPair) error
	//SubscribeAccount(pair CurrencyPair) error
}
