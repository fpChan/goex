package okex

import (
	"github.com/fpChan/goex/types"
	"time"
)

func adaptKLinePeriod(period types.KlinePeriod) int {
	granularity := -1
	switch period {
	case types.KLINE_PERIOD_1MIN:
		granularity = 60
	case types.KLINE_PERIOD_3MIN:
		granularity = 180
	case types.KLINE_PERIOD_5MIN:
		granularity = 300
	case types.KLINE_PERIOD_15MIN:
		granularity = 900
	case types.KLINE_PERIOD_30MIN:
		granularity = 1800
	case types.KLINE_PERIOD_1H, types.KLINE_PERIOD_60MIN:
		granularity = 3600
	case types.KLINE_PERIOD_2H:
		granularity = 7200
	case types.KLINE_PERIOD_4H:
		granularity = 14400
	case types.KLINE_PERIOD_6H:
		granularity = 21600
	case types.KLINE_PERIOD_1DAY:
		granularity = 86400
	case types.KLINE_PERIOD_1WEEK:
		granularity = 604800
	}
	return granularity
}

func adaptSecondsToKlinePeriod(seconds int) types.KlinePeriod {
	var p types.KlinePeriod
	switch seconds {
	case 60:
		p = types.KLINE_PERIOD_1MIN
	case 180:
		p = types.KLINE_PERIOD_3MIN
	case 300:
		p = types.KLINE_PERIOD_5MIN
	case 900:
		p = types.KLINE_PERIOD_15MIN
	case 1800:
		p = types.KLINE_PERIOD_30MIN
	case 3600:
		p = types.KLINE_PERIOD_1H
	case 7200:
		p = types.KLINE_PERIOD_2H
	case 14400:
		p = types.KLINE_PERIOD_4H
	case 21600:
		p = types.KLINE_PERIOD_6H
	case 86400:
		p = types.KLINE_PERIOD_1DAY
	case 604800:
		p = types.KLINE_PERIOD_1WEEK
	}
	return p
}

func timeStringToInt64(t string) (int64, error) {
	timestamp, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return 0, err
	}
	return timestamp.UnixNano() / int64(time.Millisecond), nil
}
