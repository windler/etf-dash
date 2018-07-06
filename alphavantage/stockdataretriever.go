package alphavantage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"time"

	"github.com/windler/etf-dash/timeseries"

	"github.com/windler/etf-dash/stockdata"
)

//StockDataRecord represents stock data
type StockDataRecord struct {
	MetaData struct {
		Symbol string `json:"2. Symbol"`
	} `json:"Meta Data"`
	TimeSeries map[string]map[string]string `json:"Monthly Time Series"`
}

//StockDataRetriever retrieves stock data
type StockDataRetriever struct {
	From        time.Time
	To          time.Time
	cacheFolder string
	apiKey      string
}

//New creates a new alphavantage stock data retriever
func New(from, to time.Time, apiKey string) stockdata.DataRetriever {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	folder := usr.HomeDir + "/etf/json/"

	os.RemoveAll(folder)
	os.MkdirAll(folder, os.ModePerm)

	return StockDataRetriever{
		From:        from,
		To:          to,
		cacheFolder: folder,
		apiKey:      apiKey,
	}
}

//GetSeries retrieves stock data as timeseries.FloatSeries
func (dr StockDataRetriever) GetSeries(symbol string) timeseries.FloatSeries {
	timeSeries := timeseries.NewFloatSeries()

	record := dr.getRecord(symbol)

	for dateString, vals := range record.TimeSeries {
		val, _ := strconv.ParseFloat(vals["2. high"], 64)
		date, _ := time.Parse("2006-01-02", dateString)

		if date.Before(dr.From) || date.After(dr.To) {
			continue
		}

		timeSeries.Add(date, val)
	}

	return timeSeries
}

func (dr StockDataRetriever) getRecord(symbol string) StockDataRecord {
	var record StockDataRecord

	cacheFile := dr.cacheFolder + symbol + ".json"

	if _, err := os.Stat(cacheFile); err != nil {
		url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_MONTHLY&symbol=%s&apikey=%s", symbol, dr.apiKey)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal("NewRequest: ", err)
			return record
		}
		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Do: ", err)
			return record
		}

		defer resp.Body.Close()

		file, _ := os.Create(cacheFile)
		io.Copy(file, resp.Body)
	}

	r, _ := os.Open(cacheFile)

	if err := json.NewDecoder(bufio.NewReader(r)).Decode(&record); err != nil {
		log.Println(err)
	}

	return record
}
