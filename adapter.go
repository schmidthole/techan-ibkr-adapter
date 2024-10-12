package techanibkradapter

import (
	"github.com/schmidthole/ibkr-webapi-go/ibkr"
	"github.com/schmidthole/techan"
	"github.com/sdcoffey/big"
)

func GetAccountState(client *ibkr.IbkrWebClient, accountID string) (*techan.Account, error) {
	// required to call this before any other portfolio/account endpoints
	_, err := client.GetPortfolioSubaccounts()
	if err != nil {
		return nil, err
	}

	ibAccountLedger, err := client.GetPortfolioAccountLedger(accountID)
	if err != nil {
		return nil, err
	}

	cash := ibAccountLedger.Base.CashBalance

	ibPositions, err := client.GetPositions(accountID, 0)
	if err != nil {
		return nil, err
	}

	account := techan.NewAccount()
	account.Deposit(big.NewDecimal(cash))

	for _, ibPosition := range ibPositions {
		position := techan.Position{
			Security: ibPosition.Ticker,
			Amount:   big.NewDecimal(ibPosition.Position),
			Price:    big.NewDecimal(ibPosition.MarketPrice),
		}

		account.Positions[ibPosition.Ticker] = &position
	}

	return account, nil
}

func GetPricing(client *ibkr.IbkrWebClient, symbols []string) (techan.Pricing, error) {
	return nil, nil
}

func GetTimeseries(client *ibkr.IbkrWebClient, symbol string, duration string, bar string) (*techan.TimeSeries, error) {
	return nil, nil
}

func ExecuteOrder(client *ibkr.IbkrWebClient, order techan.Order) error {
	return nil
}
