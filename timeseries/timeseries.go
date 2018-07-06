package timeseries

import (
	"sort"
	"time"
)

//FloatSeries represents timepoints with floating point values
type FloatSeries struct {
	Data   []FloatPoint `json:"data"`
	sorted bool
}

//FloatPoint represents a floatingpoint at a specific time
type FloatPoint struct {
	Date time.Time `json:"date"`
	Val  float64   `json:"value"`
}

//NewFloatSeries creates a new FloatSeries
func NewFloatSeries() FloatSeries {
	return FloatSeries{
		Data: []FloatPoint{},
	}
}

//Add adds a new value to the series. If the date is already presents, the value will be ammended
func (s *FloatSeries) Add(date time.Time, val float64) {
	for i := 0; i < len(s.Data); i++ {
		d := s.Data[i]
		if d.Date.Year() == date.Year() && d.Date.Month() == date.Month() && d.Date.Day() == date.Day() {
			s.Data[i].Val += val
			return
		}
	}

	s.Data = append(s.Data, FloatPoint{
		Val:  val,
		Date: date,
	})

	s.sorted = false
}

//NearestTo gets the FloatPoint that is the most closest to date
func (s *FloatSeries) NearestTo(date time.Time) float64 {
	r := s.getDataUntil(date)

	if len(r) == 0 {
		return 0.0
	}

	return r[len(r)-1].Val
}

//Len gets the number of elemts in this series
func (s *FloatSeries) Len() int {
	return len(s.Data)
}

//First gets the first value (time based)
func (s *FloatSeries) First() float64 {
	return s.GetSortedFloatPoints()[0].Val
}

//FirstTimePoint gets the first date (time based)
func (s *FloatSeries) FirstTimePoint() time.Time {
	return s.GetSortedFloatPoints()[0].Date
}

//LastTimePoint gets the last date (time based)
func (s *FloatSeries) LastTimePoint() time.Time {
	r := s.GetSortedFloatPoints()
	return r[len(r)-1].Date
}

//ToCumulativeSeries creates a new FloatSeries based on this series where each value will be added with the predecessors value
func (s *FloatSeries) ToCumulativeSeries() *FloatSeries {
	newSeries := NewFloatSeries()

	lastVal := 0.0
	for _, fp := range s.GetSortedFloatPoints() {
		lastVal = (fp.Val + lastVal)
		newSeries.Add(fp.Date, lastVal)
	}

	return &newSeries
}

//Calc calculactes two FloatSeries (same length) by applying fn (float values)
func (s *FloatSeries) Calc(other FloatSeries, fn func(a, b float64) float64) *FloatSeries {
	calced := NewFloatSeries()

	fpA := s.GetSortedFloatPoints()
	fpB := other.GetSortedFloatPoints()

	for i := 0; i < len(fpA); i++ {
		calced.Add(fpA[i].Date, fn(fpA[i].Val, fpB[i].Val))
	}

	return &calced
}

//Calc calculactes two FloatSeries (same length) by applying fn
func Calc(fn func(a, b float64) float64, series ...FloatSeries) *FloatSeries {
	var total *FloatSeries
	for _, dv := range series {
		if total == nil {
			total = dv.Calc(dv, func(a, b float64) float64 {
				return a
			})
		} else {
			total = total.Calc(dv, func(a, b float64) float64 {
				return a + b
			})
		}
	}
	return total
}

//CalcFP calculactes two FloatSeries (same length) by applying fn (FloatPoint values)
func (s *FloatSeries) CalcFP(other FloatSeries, fn func(a, b FloatPoint) float64) *FloatSeries {
	calced := NewFloatSeries()

	fpA := s.GetSortedFloatPoints()
	fpB := other.GetSortedFloatPoints()

	for i := 0; i < len(fpA); i++ {
		calced.Add(fpA[i].Date, fn(fpA[i], fpB[i]))
	}

	return &calced
}

func (s *FloatSeries) getDataUntil(date time.Time) []FloatPoint {
	relevantData := []FloatPoint{}

	sortedData := s.GetSortedFloatPoints()
	for _, d := range sortedData {
		if d.Date.After(date) {
			break
		}
		relevantData = append(relevantData, d)
	}

	return relevantData
}

//GetSortedFloatPoints gets a slice of time sorted FloatPoints
func (s FloatSeries) GetSortedFloatPoints() []FloatPoint {
	if !s.sorted {
		sort.Slice(s.Data, func(i, j int) bool {
			return s.Data[i].Date.Before(s.Data[j].Date)
		})
	}

	return s.Data
}
