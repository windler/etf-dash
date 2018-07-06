package stockdata

import (
	"github.com/windler/etf-dash/timeseries"
)

//DataRetriever gets stockdata as timeseries.FloatSeries
type DataRetriever interface {
	GetSeries(symbol string) timeseries.FloatSeries
}

//Analyzed reepresents analyzed stock data result
type Analyzed struct {
	DepotValues  map[string]timeseries.FloatSeries `json:"depotValues"`
	Stocks       map[string]timeseries.FloatSeries `json:"stocks"`
	Spendings    map[string]timeseries.FloatSeries `json:"spendings"`
	WinLoss      map[string]timeseries.FloatSeries `json:"winLoss"`
	AnnualTaxes  map[string]timeseries.FloatSeries `json:"annualTaxes"`
	SellingTaxes map[string]timeseries.FloatSeries `json:"sellingTaxes"`
	Allocations  map[string]timeseries.FloatSeries `json:"allocations"`
	Amounts      map[string]timeseries.FloatSeries `json:"amounts"`
}

//MonthlySaving are monthly savings that apply to analyzeation
type MonthlySaving struct {
	Symbol        string  `json:"symbol"`
	InitialAmount float64 `json:"initialAmount"`
	Saving        float64 `json:"monthlySavings"`
}

//Calculator analyzes stock progression
type Calculator struct {
	TransactionCostPercentage float64
	Allowance                 float64
	StockDataRetriever        *DataRetriever
}

//Analyze analyzes stock progression
func (c Calculator) Analyze(monthlySavings []MonthlySaving) *Analyzed {
	stocks := map[string]timeseries.FloatSeries{}
	depotValues := map[string]timeseries.FloatSeries{}
	spendings := map[string]timeseries.FloatSeries{}
	winLoss := map[string]timeseries.FloatSeries{}
	annualTaxes := map[string]timeseries.FloatSeries{}
	sellingTaxes := map[string]timeseries.FloatSeries{}
	allocations := map[string]timeseries.FloatSeries{}
	amounts := map[string]timeseries.FloatSeries{}

	taxCalculator := &TaxCalculator{}

	for _, monthlySaving := range monthlySavings {
		series := (*c.StockDataRetriever).GetSeries(monthlySaving.Symbol)
		stocks[monthlySaving.Symbol] = series

		amountSeries := series.Calc(series, func(a, b float64) float64 {
			return monthlySaving.Saving
		})
		amountSeries = amountSeries.Calc(series, func(a, b float64) float64 {
			return a / b
		})
		amountSeries.Add(series.FirstTimePoint(), monthlySaving.InitialAmount)
		amountSeries = amountSeries.ToCumulativeSeries()
		amounts[monthlySaving.Symbol] = *amountSeries

		depotValueSeries := amountSeries.Calc(series, func(a, b float64) float64 {
			return a * b
		})

		annualTaxes[monthlySaving.Symbol] = *(depotValueSeries.CalcFP(*depotValueSeries, taxCalculator.taxFn(*depotValueSeries, true)))
		sellingTaxes[monthlySaving.Symbol] = *(depotValueSeries.CalcFP(*depotValueSeries, taxCalculator.taxFn(*depotValueSeries, false)))

		spendingsSeries := series.Calc(series, func(a, b float64) float64 {
			return (monthlySaving.Saving * (1 + c.TransactionCostPercentage))
		})
		spendingsSeries.Add(series.FirstTimePoint(), (depotValueSeries.First() - monthlySaving.Saving*(1+c.TransactionCostPercentage)))
		spendingsSeries = spendingsSeries.ToCumulativeSeries()

		depotValues[monthlySaving.Symbol] = *depotValueSeries
		spendings[monthlySaving.Symbol] = *spendingsSeries
		winLoss[monthlySaving.Symbol] = *(depotValueSeries.Calc(*spendingsSeries, func(a, b float64) float64 {
			return a - b
		}))
	}

	depotValues["Total"] = *(timeseries.Calc(func(a, b float64) float64 {
		return a + b
	}, getTimeSeries(depotValues)...))

	total := depotValues["Total"]
	for symbol, depotValSeries := range depotValues {
		if symbol == "Total" {
			continue
		}

		allocations[symbol] = *(total.Calc(depotValSeries, func(a, b float64) float64 {
			return b / a
		}))
	}

	winLoss["Total"] = *(timeseries.Calc(func(a, b float64) float64 {
		return a + b
	}, getTimeSeries(winLoss)...))

	spendings["Total"] = *(timeseries.Calc(func(a, b float64) float64 {
		return a + b
	}, getTimeSeries(spendings)...))

	annualTaxes["Total"] = *(timeseries.Calc(func(a, b float64) float64 {
		return a + b
	}, getTimeSeries(annualTaxes)...))

	sellingTaxes["Total"] = *(timeseries.Calc(func(a, b float64) float64 {
		return a + b
	}, getTimeSeries(sellingTaxes)...))

	return &Analyzed{
		Stocks:       stocks,
		DepotValues:  depotValues,
		WinLoss:      winLoss,
		Spendings:    spendings,
		AnnualTaxes:  annualTaxes,
		SellingTaxes: sellingTaxes,
		Allocations:  allocations,
		Amounts:      amounts,
	}
}

func getTimeSeries(seriesMap map[string]timeseries.FloatSeries) []timeseries.FloatSeries {
	m := make([]timeseries.FloatSeries, 0, len(seriesMap))
	for _, val := range seriesMap {
		m = append(m, val)
	}
	return m
}
