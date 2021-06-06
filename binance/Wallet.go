package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/fpChan/goex"
	"github.com/fpChan/goex/common/api"
	"github.com/fpChan/goex/internal/logger"
	"github.com/fpChan/goex/types"
	"net/url"
)

type Wallet struct {
	ba   *Binance
	conf *types.APIConfig
}

func NewWallet(c *types.APIConfig) *Wallet {
	return &Wallet{ba: NewWithConfig(c), conf: c}
}

func (w *Wallet) GetAccount() (*types.Account, error) {
	return nil, errors.New("not implement")
}

func (w *Wallet) Withdrawal(param types.WithdrawParameter) (withdrawId string, err error) {
	return "", errors.New("not implement")
}

func (w *Wallet) Transfer(param types.TransferParameter) error {
	transferUrl := w.conf.Endpoint + "/sapi/v1/futures/transfer"

	postParam := url.Values{}
	postParam.Set("asset", param.Currency)
	postParam.Set("amount", fmt.Sprint(param.Amount))

	if param.From == types.SPOT && param.To == types.SWAP_USDT {
		postParam.Set("type", "1")
	}

	if param.From == types.SWAP_USDT && param.To == types.SPOT {
		postParam.Set("type", "2")
	}

	if param.From == types.SPOT && param.To == types.FUTURE {
		postParam.Set("type", "3")
	}

	if param.From == types.FUTURE && param.To == types.SPOT {
		postParam.Set("type", "4")
	}

	w.ba.buildParamsSigned(&postParam)

	resp, err := api.HttpPostForm2(w.ba.httpClient, transferUrl, postParam,
		map[string]string{"X-MBX-APIKEY": w.ba.accessKey})

	if err != nil {
		return err
	}

	respmap := make(map[string]interface{})
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		return err
	}

	if respmap["tranId"] != nil && ToInt64(respmap["tranId"]) > 0 {
		return nil
	}

	return errors.New(string(resp))
}

func (w *Wallet) GetWithDrawHistory(currency *types.Currency) ([]types.DepositWithdrawHistory, error) {
	//historyUrl := w.conf.Endpoint + "/wapi/v3/withdrawHistory.html"
	historyUrl := w.conf.Endpoint + "/sapi/v1/accountSnapshot"
	postParam := url.Values{}
	postParam.Set("type", "SPOT")
	w.ba.buildParamsSigned(&postParam)

	resp, err := api.HttpGet5(w.ba.httpClient, historyUrl+"?"+postParam.Encode(),
		map[string]string{"X-MBX-APIKEY": w.ba.accessKey})

	if err != nil {
		return nil, err
	}
	logger.Debugf("response body: %s", string(resp))
	respmap := make(map[string]interface{})
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (w *Wallet) GetDepositHistory(currency *types.Currency) ([]types.DepositWithdrawHistory, error) {
	historyUrl := w.conf.Endpoint + "/wapi/v3/depositHistory.html"
	postParam := url.Values{}
	postParam.Set("asset", currency.Symbol)
	w.ba.buildParamsSigned(&postParam)

	resp, err := api.HttpGet5(w.ba.httpClient, historyUrl+"?"+postParam.Encode(),
		map[string]string{"X-MBX-APIKEY": w.ba.accessKey})

	if err != nil {
		return nil, err
	}
	logger.Debugf("response body: %s", string(resp))
	respmap := make(map[string]interface{})
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
