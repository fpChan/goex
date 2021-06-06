package binance

import (
	"fmt"
	"github.com/fpChan/goex/types"
	"strings"
)

func adaptStreamToCurrencyPair(stream string) types.CurrencyPair {
	symbol := strings.Split(stream, "@")[0]

	if strings.HasSuffix(symbol, "usdt") {
		return types.NewCurrencyPair2(fmt.Sprintf("%s_usdt", strings.TrimSuffix(symbol, "usdt")))
	}

	if strings.HasSuffix(symbol, "usd") {
		return types.NewCurrencyPair2(fmt.Sprintf("%s_usd", strings.TrimSuffix(symbol, "usd")))
	}

	if strings.HasSuffix(symbol, "btc") {
		return types.NewCurrencyPair2(fmt.Sprintf("%s_btc", strings.TrimSuffix(symbol, "btc")))
	}

	return types.UNKNOWN_PAIR
}

func adaptSymbolToCurrencyPair(symbol string) types.CurrencyPair {
	symbol = strings.ToUpper(symbol)

	if strings.HasSuffix(symbol, "USD") {
		return types.NewCurrencyPair2(fmt.Sprintf("%s_USD", strings.TrimSuffix(symbol, "USD")))
	}

	if strings.HasSuffix(symbol, "USDT") {
		return types.NewCurrencyPair2(fmt.Sprintf("%s_USDT", strings.TrimSuffix(symbol, "USDT")))
	}

	if strings.HasSuffix(symbol, "PAX") {
		return types.NewCurrencyPair2(fmt.Sprintf("%s_PAX", strings.TrimSuffix(symbol, "PAX")))
	}

	if strings.HasSuffix(symbol, "BTC") {
		return types.NewCurrencyPair2(fmt.Sprintf("%s_BTC", strings.TrimSuffix(symbol, "BTC")))
	}

	return types.UNKNOWN_PAIR
}

func adaptOrderStatus(status string) types.TradeStatus {
	var tradeStatus types.TradeStatus
	switch status {
	case "NEW":
		tradeStatus = types.ORDER_UNFINISH
	case "FILLED":
		tradeStatus = types.ORDER_FINISH
	case "PARTIALLY_FILLED":
		tradeStatus = types.ORDER_PART_FINISH
	case "CANCELED":
		tradeStatus = types.ORDER_CANCEL
	case "PENDING_CANCEL":
		tradeStatus = types.ORDER_CANCEL_ING
	case "REJECTED":
		tradeStatus = types.ORDER_REJECT
	}
	return tradeStatus
}
