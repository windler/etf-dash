package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/windler/etf-dash/stockdata"

	"github.com/windler/etf-dash/alphavantage"
)

var apiKey *string

//GetDataRequest represents getData api request
type GetDataRequest struct {
	Savings                   []stockdata.MonthlySaving `json:"savings"`
	Allowance                 float64                   `json:"allowance"`
	TransactionCostPercentage float64                   `json:"transactionCostPercentage"`
	From                      string                    `json:"from"`
	To                        string                    `json:"to"`
}

//GetDataResponse represents getData api response
type GetDataResponse struct {
	AnalyzedData *stockdata.Analyzed `json:"analyzedData"`
}

func main() {
	apiKey = flag.String("apiKey", "demo", "alphavantage api key")
	flag.Parse()

	http.HandleFunc("/", drawChart)
	http.HandleFunc("/getdata", getData)
	http.HandleFunc("/favico.ico", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte{})
	})
	http.ListenAndServe(":1234", nil)
}

func getData(res http.ResponseWriter, req *http.Request) {
	var getDataReq GetDataRequest

	if err := json.NewDecoder(req.Body).Decode(&getDataReq); err != nil {
		log.Println(err)
	}

	from, _ := time.Parse("2006-01", getDataReq.From)
	to, _ := time.Parse("2006-01", getDataReq.To)
	stockDataRetriever := alphavantage.New(from, to, *apiKey)

	calculator := &stockdata.Calculator{
		TransactionCostPercentage: getDataReq.TransactionCostPercentage,
		Allowance:                 getDataReq.Allowance,
		StockDataRetriever:        &stockDataRetriever,
	}

	analyzedData := calculator.Analyze(getDataReq.Savings)

	json.NewEncoder(res).Encode(&GetDataResponse{
		AnalyzedData: analyzedData,
	})
}

func drawChart(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, "tmpl/overview.html")
}
