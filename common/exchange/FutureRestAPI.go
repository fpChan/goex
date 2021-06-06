package exchange

import (
	"github.com/fpChan/goex/types"
)

type FutureRestAPI interface {
	/**
	 *获取交易所名字
	 */
	GetExchangeName() string

	/**
	 *获取交割预估价
	 */
	GetFutureEstimatedPrice(currencyPair types.CurrencyPair) (float64, error)

	/**
	 * 期货行情
	 * @param currency_pair   btc_usd:比特币    ltc_usd :莱特币
	 * @param contractType  合约类型: this_week:当周   next_week:下周   month:当月   quarter:季度
	 */
	GetFutureTicker(currencyPair types.CurrencyPair, contractType string) (*types.Ticker, error)

	/**
	 * 期货深度
	 * @param currencyPair  btc_usd:比特币    ltc_usd :莱特币
	 * @param contractType  合约类型: this_week:当周   next_week:下周   month:当月   quarter:季度
	 * @param size 获取深度档数
	 * @return
	 */
	GetFutureDepth(currencyPair types.CurrencyPair, contractType string, size int) (*types.Depth, error)

	/**
	 * 期货指数
	 * @param currencyPair   btc_usd:比特币    ltc_usd :莱特币
	 */
	GetFutureIndex(currencyPair types.CurrencyPair) (float64, error)

	/**
	 *全仓账户
	 *@param currency
	 */
	GetFutureUserinfo(currencyPair ...types.CurrencyPair) (*types.FutureAccount, error)

	/**
	 * @deprecated
	 * 期货下单
	 * @param currencyPair   btc_usd:比特币    ltc_usd :莱特币
	 * @param contractType   合约类型: this_week:当周   next_week:下周   month:当月   quarter:季度
	 * @param price  价格
	 * @param amount  委托数量
	 * @param openType   1:开多   2:开空   3:平多   4:平空
	 * @param matchPrice  是否为对手价 0:不是    1:是   ,当取值为1时,price无效
	 */
	PlaceFutureOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType, matchPrice int, leverRate float64) (string, error)

	LimitFuturesOrder(currencyPair types.CurrencyPair, contractType, price, amount string, openType int, opt ...types.LimitOrderOptionalParameter) (*types.FutureOrder, error)

	//对手价下单
	MarketFuturesOrder(currencyPair types.CurrencyPair, contractType, amount string, openType int) (*types.FutureOrder, error)

	/**
	 * 取消订单
	 * @param symbol   btc_usd:比特币    ltc_usd :莱特币
	 * @param contractType    合约类型: this_week:当周   next_week:下周   month:当月   quarter:季度
	 * @param orderId   订单ID

	 */
	FutureCancelOrder(currencyPair types.CurrencyPair, contractType, orderId string) (bool, error)

	/**
	 * 用户持仓查询
	 * @param symbol   btc_usd:比特币    ltc_usd :莱特币
	 * @param contractType   合约类型: this_week:当周   next_week:下周   month:当月   quarter:季度
	 * @return
	 */
	GetFuturePosition(currencyPair types.CurrencyPair, contractType string) ([]types.FuturePosition, error)

	/**
	 *获取订单信息
	 */
	GetFutureOrders(orderIds []string, currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error)

	/**
	 *获取单个订单信息
	 */
	GetFutureOrder(orderId string, currencyPair types.CurrencyPair, contractType string) (*types.FutureOrder, error)

	/**
	 *获取未完成订单信息
	 */
	GetUnfinishFutureOrders(currencyPair types.CurrencyPair, contractType string) ([]types.FutureOrder, error)

	/**
	 * 获取个人订单历史,默认获取最近的订单历史列表，返回多少条订单数据，需根据平台接口定义而定
	 */
	GetFutureOrderHistory(pair types.CurrencyPair, contractType string, optional ...types.OptionalParameter) ([]types.FutureOrder, error)

	/**
	 *获取交易费
	 */
	GetFee() (float64, error)

	/**
	 *获取交易所的美元人民币汇率
	 */
	//GetExchangeRate() (float64, error)

	/**
	 *获取每张合约价值
	 */
	GetContractValue(currencyPair types.CurrencyPair) (float64, error)

	/**
	 *获取交割时间 星期(0,1,2,3,4,5,6)，小时，分，秒
	 */
	GetDeliveryTime() (int, int, int, int)

	/**
	 * 获取K线数据
	 */
	GetKlineRecords(contractType string, currency types.CurrencyPair, period types.KlinePeriod, size int, optional ...types.OptionalParameter) ([]types.FutureKline, error)

	/**
	 * 获取Trade数据
	 */
	GetTrades(contractType string, currencyPair types.CurrencyPair, since int64) ([]types.Trade, error)
}

type ExpandFutureRestAPI interface {
	FutureRestAPI
	GetFutureTrendTicker(currencyPair types.CurrencyPair, contractType string) (*types.Trend, error)
}
