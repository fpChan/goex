package okex

import (
	"errors"
	"fmt"
	"github.com/fpChan/goex/types"
	"strings"
)

type OKExMargin struct {
	*OKEx
}

func (ok *OKExMargin) GetMarginAccount(pair types.CurrencyPair) (*types.MarginAccount, error) {
	urlPath := fmt.Sprintf("/api/margin/v3/accounts/%s", pair.ToSymbol("-"))
	var response map[string]interface{}
	err := ok.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	acc := types.MarginAccount{}
	acc.Sub = make(map[types.Currency]types.MarginSubAccount, 2)

	acc.LiquidationPrice = types.ToFloat64(response["liquidation_price"])
	acc.RiskRate = types.ToFloat64(response["risk_rate"])
	acc.MarginRatio = types.ToFloat64(response["margin_ratio"])

	for k, v := range response {
		if strings.Contains(k, "currency") {
			c := types.NewCurrency(strings.Split(k, ":")[1], "")
			vv := v.(map[string]interface{})
			if err != nil {
				return nil, err
			}

			acc.Sub[c] = types.MarginSubAccount{
				Balance:     types.ToFloat64(vv["balance"]),
				Frozen:      types.ToFloat64(vv["frozen"]),
				Available:   types.ToFloat64(vv["available"]),
				CanWithdraw: types.ToFloat64(vv["can_withdraw"]),
				Loan:        types.ToFloat64(vv["borrowed"]),
				LendingFee:  types.ToFloat64(vv["lending_fee"])}
		}
	}

	return &acc, nil
}

/**
  杠杆交易区借币，
  pair ： 操作的交易对
  currency： 需要借的币种
  amount : 借的金额
*/
func (ok *OKExMargin) Borrow(parameter types.BorrowParameter) (borrowId string, err error) {
	var param = struct {
		InstrumentId string `json:"instrument_id"`
		Currency     string `json:"currency"`
		Amount       string `json:"amount"`
	}{
		InstrumentId: parameter.CurrencyPair.ToSymbol("-"),
		Currency:     parameter.Currency.Symbol,
		Amount:       types.FloatToString(parameter.Amount, 8)}

	reqBody, _, _ := ok.BuildRequestBody(param)
	println(reqBody)
	var response struct {
		BorrowId     string `json:"borrow_id"`
		Result       bool   `json:"result"`
		ErrorCode    string `json:"code"`
		ErrorMessage string `json:"message"`
	}

	err = ok.DoRequest("POST", "/api/margin/v3/accounts/borrow", reqBody, &response)
	if err != nil {
		return "", err
	}

	if response.ErrorMessage != "" {
		return "", errors.New(response.ErrorMessage)
	}

	return response.BorrowId, nil
}

func (ok *OKExMargin) Repayment(parameter types.RepaymentParameter) (repaymentId string, err error) {
	urlPath := "/api/margin/v3/accounts/repayment"
	param := struct {
		BorrowId     string `json:"borrow_id,omitempty"`
		InstrumentId string `json:"instrument_id"`
		Currency     string `json:"currency"`
		Amount       string `json:"amount"`
	}{
		parameter.BorrowId,
		parameter.CurrencyPair.ToSymbol("-"),
		parameter.Currency.Symbol,
		types.FloatToString(parameter.Amount, 8)}

	reqBody, _, _ := ok.BuildRequestBody(param)
	println(reqBody)
	var response struct {
		RepaymentId string `json:"repayment_id"`
		Result      bool   `json:"result"`
		Code        string `json:"code"`
		Message     string `json:"message"`
	}
	err = ok.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return "", err
	}

	if !response.Result {
		return "", errors.New(response.Message)
	}

	return response.RepaymentId, nil
}

func (ok *OKExMargin) PlaceOrder(ord *types.Order) (*types.Order, error) {
	param := PlaceOrderParam{
		ClientOid:     ok.UUID(),
		InstrumentId:  ord.Currency.AdaptUsdToUsdt().ToLower().ToSymbol("-"),
		Type:          ord.Type,
		OrderType:     ord.OrderType,
		MarginTrading: "2"}

	var response PlaceOrderResponse

	switch ord.Side {
	case types.BUY, types.SELL:
		param.Side = strings.ToLower(ord.Side.String())
		param.Price = ord.Price
		param.Size = ord.Amount
	case types.SELL_MARKET:
		param.Side = "sell"
		param.Size = ord.Amount
	case types.BUY_MARKET:
		param.Side = "buy"
		param.Notional = ord.Price
	default:
		param.Size = ord.Amount
		param.Price = ord.Price
	}

	jsonStr, _, _ := ok.OKEx.BuildRequestBody(param)
	err := ok.OKEx.DoRequest("POST", "/api/margin/v3/orders", jsonStr, &response)
	if err != nil {
		return nil, err
	}

	if !response.Result {
		return nil, errors.New(response.ErrorMessage)
	}

	ord.Cid = response.ClientOid
	ord.OrderID2 = response.OrderId

	return ord, nil
}

func (ok *OKExMargin) GetUnfinishOrders(currency types.CurrencyPair) ([]types.Order, error) {
	var response []OrderResponse
	err := ok.DoRequest("GET", fmt.Sprintf("/api/margin/v3/orders_pending?instrument_id=%s", currency.AdaptUsdToUsdt().ToSymbol("-")), "", &response)
	if err != nil {
		return nil, err
	}

	var orders []types.Order

	for _, info := range response {
		ord := ok.OKExSpot.adaptOrder(info)
		ord.Currency = currency
		orders = append(orders, *ord)
	}

	return orders, nil
}

func (ok *OKExMargin) CancelOrder(orderId string, currency types.CurrencyPair) (bool, error) {
	urlPath := fmt.Sprintf("/api/margin/v3/cancel_orders/%s", orderId)
	reqBody, _, _ := ok.BuildRequestBody(map[string]string{"instrument_id": currency.AdaptUsdToUsdt().ToSymbol("-")})
	var response struct {
		ClientOid    string `json:"client_oid"`
		OrderId      string `json:"order_id"`
		Result       bool   `json:"result"`
		ErrorCode    string `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	}
	err := ok.DoRequest("POST", urlPath, reqBody, &response)
	if err != nil {
		return false, err
	}

	if !response.Result {
		return false, errors.New(response.ErrorMessage)
	}

	return true, nil
}

//orderId can set client oid or orderId
func (ok *OKExMargin) GetOneOrder(orderId string, currency types.CurrencyPair) (*types.Order, error) {
	urlPath := "/api/margin/v3/orders/" + orderId + "?instrument_id=" + currency.AdaptUsdToUsdt().ToSymbol("-")
	//param := struct {
	//	InstrumentId string `json:"instrument_id"`
	//}{currency.AdaptUsdToUsdt().ToLower().ToSymbol("-")}
	//reqBody, _, _ := ok.BuildRequestBody(param)
	var response OrderResponse
	err := ok.OKEx.DoRequest("GET", urlPath, "", &response)
	if err != nil {
		return nil, err
	}

	ordInfo := ok.OKExSpot.adaptOrder(response)
	ordInfo.Currency = currency

	return ordInfo, nil
}
