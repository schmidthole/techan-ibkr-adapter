package techanibkradapter

import (
	"time"

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

func GetMarketSnapshot(client *ibkr.IbkrWebClient, conIds []int) (*techan.MarketSnapshot, error) {
	snapshots, err := client.MarketDataSnapshot(conIds)
	if err != nil {
		return nil, err
	}

	pricing := techan.Pricing{}
	tradingState := map[string]techan.TradingState{}
	for _, snapshot := range snapshots {
		pricing[snapshot.Symbol] = big.NewDecimal(snapshot.LastPrice)

		var state techan.TradingState = techan.OPEN
		if !snapshot.TradingActive {
			state = techan.CLOSED
		}

		if snapshot.TradingHalted {
			state = techan.HALTED
		}

		tradingState[snapshot.Symbol] = state
	}

	return &techan.MarketSnapshot{
		Pricing:      pricing,
		TradingState: tradingState,
	}, nil
}

func GetTimeseries(client *ibkr.IbkrWebClient, conId int, period string, bar string) (*techan.TimeSeries, error) {
	data, err := client.MarketDataHistory(conId, period, bar)
	if err != nil {
		return nil, err
	}

	// bar length is fixed right now but should be parsed and updated from the bar param
	barLength := time.Hour * 24

	timeseries := techan.NewTimeSeries()
	for _, datum := range data.Data {
		barPeriod := techan.NewTimePeriod(time.Unix(int64(datum.T), 0), barLength)
		candle := techan.NewCandle(barPeriod)
		candle.OpenPrice = big.NewDecimal(datum.O)
		candle.ClosePrice = big.NewDecimal(datum.C)
		candle.MaxPrice = big.NewDecimal(datum.H)
		candle.MinPrice = big.NewDecimal(datum.L)
		candle.Volume = big.NewDecimal(datum.V)

		timeseries.AddCandle(candle)
	}

	return timeseries, nil
}

func ExecuteOrder(client *ibkr.IbkrWebClient, accountID string, order techan.Order) (string, error) {
	side := "BUY"
	if order.Side == techan.SELL {
		side = "SELL"
	}

	// some of the order fields are fixed for now. techan will need to be updated to support more
	// fields in the order object in the future to support real broker connections vs. data analysis
	ibOrder := ibkr.Order{
		AccountId:   accountID,
		OrderType:   "MKT",
		Side:        side,
		TimeInForce: "DAY",
		Quantity:    order.Amount.Float(),
	}

	response, err := client.PlaceOrder(accountID, ibOrder)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}
