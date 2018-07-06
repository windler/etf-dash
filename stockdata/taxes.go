package stockdata

import (
	"time"

	"github.com/windler/etf-dash/timeseries"
)

//TaxCalculator calculates taxes
type TaxCalculator struct{}

//Taxes represents taxes
type Taxes struct {
	Annual  float64
	Selling float64
}

//Calculate calculates taxes regarding german tax law
func (tc TaxCalculator) Calculate(date time.Time, valueAtYearBegin, valueCurrent float64) Taxes {
	taxTotal := 0.0
	sellingTaxCosts := 0.0

	if date.Month() == time.December {
		//german taxes. currently assume that all stocks are reinvesting
		basisertrag := valueAtYearBegin * 0.011 * 0.7
		vorabpauschale := basisertrag - 0
		taxAusschuettung := 0.0 //reinvesting stock
		taxVorabpauschale := vorabpauschale * 0.7 * 0.26375

		basisErtragSelling := 0.0
		if taxVorabpauschale > 0 {
			taxTotal = taxAusschuettung + taxVorabpauschale
			basisErtragSelling = basisertrag
		}

		sellingTaxCosts = (valueCurrent - valueAtYearBegin - basisErtragSelling) * 0.7 * 0.26375
		if sellingTaxCosts < 0 {
			sellingTaxCosts = 0
		}
	}

	return Taxes{
		Annual:  taxTotal,
		Selling: sellingTaxCosts,
	}

}

func (tc TaxCalculator) taxFn(depotValueSeries timeseries.FloatSeries, annual bool) func(a, b timeseries.FloatPoint) float64 {
	return func(a, b timeseries.FloatPoint) float64 {
		curDate := a.Date
		compDate := time.Date(curDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		valBeginnging := depotValueSeries.NearestTo(compDate)
		if valBeginnging == 0.0 {
			valBeginnging = depotValueSeries.First()
		}

		taxes := tc.Calculate(a.Date, valBeginnging, a.Val)

		if annual {
			return taxes.Annual
		}
		return taxes.Selling
	}
}
