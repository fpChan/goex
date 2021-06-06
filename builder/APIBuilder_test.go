package builder

import (
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

var builder = NewAPIBuilder()

func init() {
	logger.SetLevel(logger.INFO)
}

func TestAPIBuilder_Build(t *testing.T) {
	assert.Equal(t, builder.APIKey("").APISecretkey("").Build(types.OKCOIN_COM).GetExchangeName(), types.OKCOIN_COM)
	assert.Equal(t, builder.APIKey("").APISecretkey("").Build(types.HUOBI_PRO).GetExchangeName(), types.HUOBI_PRO)
	//assert.Equal(t, builder.APIKey("").APISecretkey("").BuildFuture(types.HBDM).GetExchangeName(), goex.HBDM)
}

func TestAPIBuilder_BuildSpotWs(t *testing.T) {
	//os.Setenv("HTTPS_PROXY" , "socks5://127.0.0.1:1080")
	wsApi, _ := builder.BuildSpotWs(types.OKEX_V3)
	wsApi.DepthCallback(func(depth *types.Depth) {
		log.Println(depth)
	})
	wsApi.SubscribeDepth(types.BTC_USDT)
	time.Sleep(time.Minute)
}

func TestAPIBuilder_BuildFuturesWs(t *testing.T) {
	//os.Setenv("HTTPS_PROXY" , "socks5://127.0.0.1:1080")
	wsApi, _ := builder.BuildFuturesWs(types.OKEX_V3)
	wsApi.DepthCallback(func(depth *types.Depth) {
		log.Println(depth)
	})
	wsApi.SubscribeDepth(types.BTC_USD, types.QUARTER_CONTRACT)
	time.Sleep(time.Minute)
}
