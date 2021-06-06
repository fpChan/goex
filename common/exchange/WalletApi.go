package exchange

import "github.com/fpChan/goex/types"

type WalletApi interface {
	//获取钱包资产
	GetAccount() (*types.Account, error)
	//提币
	Withdrawal(param types.WithdrawParameter) (withdrawId string, err error)
	//划转资产
	Transfer(param types.TransferParameter) error
	//获取提币记录
	GetWithDrawHistory(currency *types.Currency) ([]types.DepositWithdrawHistory, error)
	//获取充值记录
	GetDepositHistory(currency *types.Currency) ([]types.DepositWithdrawHistory, error)
}
