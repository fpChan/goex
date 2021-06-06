package bitmex

import (
	"fmt"
	"github.com/fpChan/goex/types"
	"strings"
)

func AdaptCurrencyPairToSymbol(pair types.CurrencyPair, contract string) string {
	if contract == "" || contract == types.SWAP_CONTRACT {
		if pair.CurrencyA.Eq(types.BTC) {
			pair = types.NewCurrencyPair(types.XBT, types.USD)
		}
		if pair.CurrencyB.Eq(types.BTC) {
			pair = types.NewCurrencyPair(pair.CurrencyA, types.XBT)
		}
		return pair.AdaptUsdtToUsd().ToSymbol("")
	}

	coin := pair.CurrencyA.Symbol
	if pair.CurrencyA.Eq(types.BTC) {
		coin = types.XBT.Symbol
	}
	return fmt.Sprintf("%s%s", coin, strings.ToUpper(contract))
}

func AdaptWsSymbol(symbol string) (pair types.CurrencyPair, contract string) {
	symbol = strings.ToUpper(symbol)

	if symbol == "XBTCUSD" {
		return types.BTC_USD, types.SWAP_CONTRACT
	}

	if symbol == "BCHUSD" {
		return types.BCH_USD, types.SWAP_CONTRACT
	}

	if symbol == "ETHUSD" {
		return types.ETH_USD, types.SWAP_CONTRACT
	}

	if symbol == "LTCUSD" {
		return types.LTC_USD, types.SWAP_CONTRACT
	}

	if symbol == "LINKUSDT" {
		return types.NewCurrencyPair2("LINK_USDT"), types.SWAP_CONTRACT
	}

	pair = types.NewCurrencyPair(types.NewCurrency(symbol[0:3], ""), types.USDT)
	contract = symbol[3:]
	if pair.CurrencyA.Eq(types.XBT) {
		return types.NewCurrencyPair(types.BTC, types.USDT), contract
	}

	return pair, contract
}
