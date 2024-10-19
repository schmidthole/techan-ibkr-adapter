package techanibkradapter

import (
	"fmt"
	"time"

	"github.com/schmidthole/ibkr-webapi-go/ibkr"
	"github.com/schmidthole/techan"
	"github.com/sdcoffey/big"
)

func GetAccountState(client *ibkr.IbkrWebClient, accountID string) (*techan.Account, error) {
	var err error

	retries := 0
	err = nil
	// required to call these before any other portfolio/account endpoints
	for retries < 3 {
		_, err = client.GetAccounts()
		if err != nil {
			retries += 1
			time.Sleep(time.Second * 1)
		} else {
			break
		}
	}
	if (retries == 3) && (err != nil) {
		return nil, fmt.Errorf("max retries for get account exceeded: %v", err)
	}

	retries = 0
	err = nil
	for retries < 3 {
		_, err = client.GetPortfolioSubaccounts()
		if err != nil {
			retries += 1
			time.Sleep(time.Second * 1)
		} else {
			break
		}
	}
	if (retries == 3) && (err != nil) {
		return nil, fmt.Errorf("max retries for get account exceeded: %v", err)
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

func GetMarketSnapshot(client *ibkr.IbkrWebClient, symbolConIdMap map[string]int) (*techan.MarketSnapshot, error) {
	symbolLookup := map[int]string{}
	conIds := []int{}

	for symbol, conid := range symbolConIdMap {
		conIds = append(conIds, conid)
		symbolLookup[conid] = symbol
	}

	retries := 0
	var err error
	var snapshots []ibkr.MarketDataSnapshot

	for retries < 5 {
		snapshots, err = client.MarketDataSnapshot(conIds)
		if err != nil {
			retries += 1
			time.Sleep(time.Second * 1)
		} else {
			break
		}
	}
	if retries == 5 {
		return nil, fmt.Errorf("market data snapshot retries exceeded: %v", err)
	}

	pricing := techan.Pricing{}
	tradingState := map[string]techan.TradingState{}
	for _, snapshot := range snapshots {
		symbol, exists := symbolLookup[snapshot.ConID]
		if !exists {
			return nil, fmt.Errorf("conid snapshot returned that is not in symbol lookup")
		}

		pricing[symbol] = big.NewDecimal(snapshot.LastPrice)

		var state techan.TradingState = techan.OPEN
		if !snapshot.TradingActive {
			state = techan.CLOSED
		}

		if snapshot.TradingHalted {
			state = techan.HALTED
		}

		tradingState[symbol] = state
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

	ibOrder := ibkr.Order{
		AccountId:   accountID,
		OrderType:   string(order.Type),
		Side:        side,
		TimeInForce: string(order.TimeInForce),
		Quantity:    order.Amount.Float(),
	}

	response, err := client.PlaceOrder(accountID, ibOrder)
	if err != nil {
		return "", err
	}

	return response.ID, nil
}

func GetOrders(client *ibkr.IbkrWebClient) ([]techan.Order, error) {
	orders := []techan.Order{}

	raw, err := client.GetLiveOrders()
	if err != nil {
		return nil, err
	}

	for _, o := range raw.Orders {
		var status techan.OrderStatus

		switch o.Status {
		case "Filled":
			status = techan.FILLED
		case "Inactive":
			fallthrough
		case "PendingSubmit":
			fallthrough
		case "PreSubmitted":
			fallthrough
		case "Submitted":
			status = techan.PENDING
		case "PendingCancel":
			fallthrough
		case "Cancelled":
			status = techan.CANCELLED
		default:
			status = techan.OTHER
		}

		order := techan.Order{
			Side:        techan.OrderSide(o.Side),
			Security:    o.Ticker,
			Type:        techan.OrderType(o.OrderType),
			Amount:      big.NewDecimal(o.RemainingQuantity + o.FilledQuantity),
			TimeInForce: techan.TimeInForce(o.TimeInForce),
			Status:      status,
		}

		orders = append(orders, order)
	}

	return orders, nil
}
