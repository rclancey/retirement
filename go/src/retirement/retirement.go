package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"sim"
)

var runCount = flag.Int("n", 1000, "Number of Monte Carlo simulations")
var configFile = flag.String("config", "-", "Configuration file")
var resultsDir = flag.String("output", "results", "Output directory")

type Transaction struct {
	Date time.Time `json:"date"`
	Amount float64 `json:"amount"`
	Memo string `json:"memo"`
	Balance float64 `json:"balance"`
}

type IndexedFloat struct {
	Index int `json:"ix"`
	Value float64 `json:"v"`
}

type IndexedFloats []*IndexedFloat

func NewIndexedFloats() *IndexedFloats {
	x := IndexedFloats([]*IndexedFloat{})
	return &x
}

func (x *IndexedFloats) Add(v float64) {
	ix := len(*x)
	*x = append(*x, &IndexedFloat{ix,v})
}

func (x *IndexedFloats) Len() int {
	return len(*x)
}

func (x *IndexedFloats) Swap(i, j int) {
	(*x)[i], (*x)[j] = (*x)[j], (*x)[i]
}

func (x *IndexedFloats) Less(i, j int) bool {
	return (*x)[i].Value < (*x)[j].Value
}

type Result struct {
	Age *IndexedFloat `json:"age"`
	Balance *IndexedFloat `json:"balance"`
	Market *IndexedFloat `json:"market"`
}

type Results struct {
	Count int `json:"count"`
	EarlyDeaths []int `json:"early_deaths"`
	LiquidityCrises []int `json:"liquidity_crises"`
	Bankruptcies []int `json:"bankruptcies"`
	Runs []*sim.Results `json:"runs"`
	Start float64 `json:"start"`
	Worst *Result `json:"worst"`
	Worst95 *Result `json:"worst95"`
	Worst75 *Result `json:"worst75"`
	Median *Result `json:"median"`
	Best75 *Result `json:"best75"`
	Best95 *Result `json:"best95"`
	Best *Result `json:"best"`
	Mean *sim.Results `json:"mean"`
}

func writeBalances(res *sim.Results) error {
	fn := filepath.Join(*resultsDir, fmt.Sprintf("balances-%06d.csv", res.Index));
	f, err := os.Create(fn)
	if err != nil {
		fmt.Println("error opening results file:", err)
		return err
	}
	w := csv.NewWriter(f)
	accts := []string{}
	for k := range res.Balances[0].Balances {
		accts = append(accts, k)
	}
	sort.Strings(accts)
	header := append([]string{"date"}, accts...)
	err = w.Write(header)
	if err != nil {
		fmt.Println("error writing results header:", err)
		return err
	}
	crow := make([]string, len(header))
	for _, row := range res.Balances {
		crow[0] = row.Date
		for j, k := range accts {
			v, ok := row.Balances[k]
			if ok {
				crow[j+1] = strconv.FormatFloat(v, 'f', 2, 64)
			} else {
				crow[j+1] = ""
			}
		}
		err = w.Write(crow)
		if err != nil {
			fmt.Println("error writing results record:", err)
			return err
		}
	}
	w.Flush()
	err = f.Close()
	if err != nil {
		fmt.Println("error writing results file:", err)
		return err
	}
	return nil
}

func writeEvents(res *sim.Results) error {
	fn := filepath.Join(*resultsDir, fmt.Sprintf("events-%06d.csv", res.Index));
	f, err := os.Create(fn)
	if err != nil {
		fmt.Println("error opening events file:", err)
		return err
	}
	w := csv.NewWriter(f)
	err = w.Write([]string{"date","severity","value"})
	if err != nil {
		fmt.Println("error writing events header:", err)
		return err
	}
	for _, row := range *res.Events {
		err = w.Write([]string{
			row.Date.Format("2006-01-02"),
			strconv.Itoa(row.Severity),
			row.Value,
		})
		if err != nil {
			fmt.Println("error writing event record:", err)
			return err
		}
	}
	w.Flush()
	err = f.Close()
	if err != nil {
		fmt.Println("error writing events file:", err)
		return err
	}
	return nil
}

func writeTransactions(res *sim.Results) error {
	fn := filepath.Join(*resultsDir, fmt.Sprintf("transactions-%06d.csv", res.Index));
	f, err := os.Create(fn)
	if err != nil {
		fmt.Println("error opening transactions file:", err)
		return err
	}
	w := csv.NewWriter(f)
	err = w.Write([]string{"date","amount","memo"})
	if err != nil {
		fmt.Println("error writing transactions header:", err)
		return err
	}
	for _, row := range *res.Transactions {
		err = w.Write([]string{
			row.Date.Format("2006-01-02"),
			strconv.FormatFloat(row.Amount, 'f', 2, 64),
			row.Memo,
		})
		if err != nil {
			fmt.Println("error writing transaction record:", err)
			return err
		}
	}
	w.Flush()
	err = f.Close()
	if err != nil {
		fmt.Println("error writing transaction file:", err)
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	var err error
	var cfgBytes []byte
	if *configFile == "-" {
		cfgBytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		cfgBytes, err = ioutil.ReadFile(*configFile)
	}
	if err != nil {
		fmt.Println("error reading input:", err)
		return
	}
	cfg := &sim.SimConfig{}
	err = json.Unmarshal(cfgBytes, cfg)
	if err != nil {
		fmt.Println("error parsing config:", err)
		return
	}
	n := *runCount
	runs := make([]*sim.Results, n)
	mean := &sim.Results{}
	age := NewIndexedFloats()
	cash := NewIndexedFloats()
	market := NewIndexedFloats()
	nf := float64(n)
	var start float64
	err = os.MkdirAll(*resultsDir, os.FileMode(0777))
	if err != nil {
		fmt.Println("error creating results directory", *resultsDir, err)
		return
	}
	earlyDeaths := []int{}
	liquidityCrises := []int{}
	bankruptcies := []int{}
	for i := 0; i < n; i++ {
		os.Stderr.WriteString(fmt.Sprintf("\rsim run %d", i))
		s := sim.NewSimulation(i, cfg)
		start = s.Balance() - s.Liabilities(s.StartDate())
		res := s.Results()
		res.Index = i
		mean.Age += res.Age / nf
		mean.Balance += res.Balance / nf
		mean.Market += res.Market / nf
		runs[i] = &sim.Results{
			Index: i,
			Age: res.Age,
			Balance: res.Balance,
			Market: res.Market,
		}
		age.Add(res.Age)
		cash.Add(res.Balance)
		market.Add(res.Market)
		if res.Balance <= 0.0 {
			os.Stderr.WriteString(fmt.Sprintf("\rsim run %d BUSTED                              \n", i))
		}
		if res.Age < s.RetirementAge() {
			earlyDeaths = append(earlyDeaths, i)
		}
		if res.Events.Has("Liquidity Crisis") {
			liquidityCrises = append(liquidityCrises, i)
		}
		if res.Events.Has("Bankruptcy") {
			bankruptcies = append(bankruptcies, i)
		}
		err := writeBalances(res)
		if err != nil {
			return
		}
		/*
		err = writeTransactions(res)
		if err != nil {
			return
		}
		*/
		err = writeEvents(res)
		if err != nil {
			return
		}
	}
	os.Stderr.WriteString("\nSummarizing...")
	mid := n / 2
	min95 := n / 20
	min75 := n / 4
	max75 := 3 * n / 4
	max95 := 19 * n / 20
	sort.Sort(age)
	sort.Sort(cash)
	sort.Sort(market)
	res := &Results{
		Count: *runCount,
		EarlyDeaths: earlyDeaths,
		LiquidityCrises: liquidityCrises,
		Bankruptcies: bankruptcies,
		Runs: runs,
		Start: start,
		Mean: mean,
		Best: &Result{
			Age: (*age)[n-1],
			Balance: (*cash)[n-1],
			Market: (*market)[n-1],
		},
		Best95: &Result{
			Age: (*age)[max95],
			Balance: (*cash)[max95],
			Market: (*market)[max95],
		},
		Best75: &Result{
			Age: (*age)[max75],
			Balance: (*cash)[max75],
			Market: (*market)[max75],
		},
		Median: &Result{
			Age: (*age)[mid],
			Balance: (*cash)[mid],
			Market: (*market)[mid],
		},
		Worst75: &Result{
			Age: (*age)[min75],
			Balance: (*cash)[min75],
			Market: (*market)[min75],
		},
		Worst95: &Result{
			Age: (*age)[min95],
			Balance: (*cash)[min95],
			Market: (*market)[min95],
		},
		Worst: &Result{
			Age: (*age)[0],
			Balance: (*cash)[0],
			Market: (*market)[0],
		},
	}

	os.Stderr.WriteString("\nFormatting...")
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Println(err)
	} else {
		fn := filepath.Join(*resultsDir, "index.json")
		ioutil.WriteFile(fn, out, os.FileMode(0666))
	}
	os.Stderr.WriteString("\nDone\n")
}

