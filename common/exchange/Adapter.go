package exchange

import (
	"github.com/fpChan/goex/types"
	"strings"
)

func AdaptTradeSide(side string) types.TradeSide {
	side2 := strings.ToUpper(side)
	switch side2 {
	case "SELL":
		return types.SELL
	case "BUY":
		return types.BUY
	case "BUY_MARKET":
		return types.BUY_MARKET
	case "SELL_MARKET":
		return types.SELL_MARKET
	default:
		return -1
	}
}

func AdaptKlinePeriodForOKEx(period int) string {
	switch period {
	case types.KLINE_PERIOD_1MIN:
		return "1min"
	case types.KLINE_PERIOD_5MIN:
		return "5min"
	case types.KLINE_PERIOD_15MIN:
		return "15min"
	case types.KLINE_PERIOD_30MIN:
		return "30min"
	case types.KLINE_PERIOD_1H:
		return "1hour"
	case types.KLINE_PERIOD_4H:
		return "4hour"
	case types.KLINE_PERIOD_1DAY:
		return "day"
	case types.KLINE_PERIOD_2H:
		return "2hour"
	case types.KLINE_PERIOD_1WEEK:
		return "week"
	default:
		return "day"
	}
}
