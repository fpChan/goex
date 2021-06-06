package huobi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fpChan/goex/common/api"
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"net/url"
	"strings"
)

type Wallet struct {
	pro *HuoBiPro
}

func NewWallet(c *types.APIConfig) *Wallet {
	return &Wallet{pro: NewHuobiWithConfig(c)}
}

//获取钱包资产
func (w *Wallet) GetAccount() (*types.Account, error) {
	return nil, errors.New("not implement")
}

func (w *Wallet) Withdrawal(param types.WithdrawParameter) (withdrawId string, err error) {
	return "", errors.New("not implement")
}

func (w *Wallet) Transfer(param types.TransferParameter) error {
	if param.From == types.SUB_ACCOUNT || param.To == types.SUB_ACCOUNT ||
		param.From == types.SPOT_MARGIN || param.To == types.SPOT_MARGIN {
		return errors.New("not implements")
	}

	httpParam := url.Values{}
	httpParam.Set("currency", strings.ToLower(param.Currency))
	httpParam.Set("amount", types.FloatToString(param.Amount, 8))

	path := ""

	if (param.From == types.SPOT && param.To == types.FUTURE) ||
		(param.From == types.FUTURE && param.To == types.SPOT) {
		path = "/v1/futures/transfer"
	}

	if param.From == types.SWAP || param.From == types.SWAP_USDT ||
		param.To == types.SWAP || param.To == types.SWAP_USDT {
		path = "/v2/account/transfer"
	}

	if param.From == types.SPOT && param.To == types.FUTURE {
		httpParam.Set("type", "pro-to-futures")
	}

	if param.From == types.FUTURE && param.To == types.SPOT {
		httpParam.Set("type", "futures-to-pro")
	}

	if param.From == types.SPOT && param.To == types.SWAP {
		httpParam.Set("from", "spot")
		httpParam.Set("to", "swap")
	}

	if param.From == types.SPOT && param.To == types.SWAP_USDT {
		httpParam.Set("currency", "usdt")
		httpParam.Set("from", "spot")
		httpParam.Set("to", "linear-swap")
		httpParam.Set("margin-account", fmt.Sprintf("%s-usdt", strings.ToLower(param.Currency)))
	}

	if param.From == types.SWAP && param.To == types.SPOT {
		httpParam.Set("from", "swap")
		httpParam.Set("to", "spot")
	}

	if param.From == types.SWAP_USDT && param.To == types.SPOT {
		httpParam.Set("currency", "usdt")
		httpParam.Set("from", "linear-swap")
		httpParam.Set("to", "spot")
		httpParam.Set("margin-account",
			fmt.Sprintf("%s-usdt", strings.ToLower(param.Currency)))
	}

	w.pro.buildPostForm("POST", path, &httpParam)

	postJsonParam, _ := types.ValuesToJson(httpParam)
	responseBody, err := api.HttpPostForm3(w.pro.httpClient,
		fmt.Sprintf("%s%s?%s", w.pro.baseUrl, path, httpParam.Encode()),
		string(postJsonParam),
		map[string]string{"Content-Type": "application/json", "Accept-Language": "zh-cn"})

	if err != nil {
		return err
	}

	logger.Debugf("[response body] %s", string(responseBody))

	var responseRet map[string]interface{}

	err = json.Unmarshal(responseBody, &responseRet)
	if err != nil {
		return err
	}

	if responseRet["status"] != nil &&
		responseRet["status"].(string) == "ok" {
		return nil
	}

	if responseRet["code"] != nil && responseRet["code"].(float64) == 200 {
		return nil
	}

	return errors.New(string(responseBody))
}

func (w *Wallet) GetWithDrawHistory(currency *types.Currency) ([]types.DepositWithdrawHistory, error) {
	return nil, errors.New("not implement")
}

func (w *Wallet) GetDepositHistory(currency *types.Currency) ([]types.DepositWithdrawHistory, error) {
	return nil, errors.New("not implement")
}
