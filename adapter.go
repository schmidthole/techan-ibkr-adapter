package techanibkradapter

import (
	"github.com/schmidthole/ibkr-webapi-go/ibkr"
	"github.com/schmidthole/techan"
)

func GetAccountState(client *ibkr.IbkrWebClient) (*techan.Account, error) {
	return nil, nil
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
