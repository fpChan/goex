package huobi

import (
	"fmt"
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"sort"
	"strings"
)

func ParseDepthFromResponse(r DepthResponse) types.Depth {
	var dep types.Depth
	for _, bid := range r.Bids {
		dep.BidList = append(dep.BidList, types.DepthRecord{Price: bid[0], Amount: bid[1]})
	}

	for _, ask := range r.Asks {
		dep.AskList = append(dep.AskList, types.DepthRecord{Price: ask[0], Amount: ask[1]})
	}

	sort.Sort(sort.Reverse(dep.BidList))
	sort.Sort(sort.Reverse(dep.AskList))
	return dep
}

func ParseCurrencyPairFromSpotWsCh(ch string) types.CurrencyPair {
	meta := strings.Split(ch, ".")
	if len(meta) < 2 {
		logger.Errorf("parse error, ch=%s", ch)
		return types.UNKNOWN_PAIR
	}

	currencyPairStr := meta[1]
	if strings.HasSuffix(currencyPairStr, "usdt") {
		currencyA := strings.TrimSuffix(currencyPairStr, "usdt")
		return types.NewCurrencyPair2(fmt.Sprintf("%s_usdt", currencyA))
	}

	if strings.HasSuffix(currencyPairStr, "btc") {
		currencyA := strings.TrimSuffix(currencyPairStr, "btc")
		return types.NewCurrencyPair2(fmt.Sprintf("%s_btc", currencyA))
	}

	if strings.HasSuffix(currencyPairStr, "eth") {
		currencyA := strings.TrimSuffix(currencyPairStr, "eth")
		return types.NewCurrencyPair2(fmt.Sprintf("%s_eth", currencyA))
	}

	if strings.HasSuffix(currencyPairStr, "husd") {
		currencyA := strings.TrimSuffix(currencyPairStr, "husd")
		return types.NewCurrencyPair2(fmt.Sprintf("%s_husd", currencyA))
	}

	if strings.HasSuffix(currencyPairStr, "ht") {
		currencyA := strings.TrimSuffix(currencyPairStr, "ht")
		return types.NewCurrencyPair2(fmt.Sprintf("%s_ht", currencyA))
	}

	if strings.HasSuffix(currencyPairStr, "trx") {
		currencyA := strings.TrimSuffix(currencyPairStr, "trx")
		return types.NewCurrencyPair2(fmt.Sprintf("%s_trx", currencyA))
	}

	return types.UNKNOWN_PAIR
}
