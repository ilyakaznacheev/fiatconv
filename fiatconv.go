package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/jessevdk/go-flags"
	"golang.org/x/text/currency"
)

const defaultAPIurl = "https://api.exchangeratesapi.io/"

type parameters struct {
	Positional struct {
		Amount      float64 `positional-arg-name:"AMOUNT" description:"decimal amount of source currency"`
		CurrencySrc string  `positional-arg-name:"SRC" description:"ISO currency code of source currency"`
		CurrencyDst string  `positional-arg-name:"DST" description:"ISO currency code of destination currency"`
	} `positional-args:"yes"`
	ExchangeAPIurl string  `long:"api-url" description:"Exchange API address" default:"https://api.exchangeratesapi.io/"`
	ProxyPath      *string `long:"proxy" description:"Optional proxy path [url:port]"`
}

func main() {
	var param parameters

	_, err := flags.ParseArgs(&param, os.Args[1:])

	reqSet, err := parseInput(param.Positional.Amount, param.Positional.CurrencySrc, param.Positional.CurrencyDst)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	var hc *http.Client

	if param.ProxyPath == nil {
		hc = http.DefaultClient
	} else {
		proxyURL, err := url.Parse(*param.ProxyPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		hc = &http.Client{
			Transport: transport,
		}
	}

	cc := newCurrencyClient(param.ExchangeAPIurl, hc)

	res, err := convertCurrency(&cc, *reqSet)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	ffl := func(f float64) string {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}

	fmt.Printf("%s %s -> %s %s\n", reqSet.currSrc, ffl(reqSet.amount), reqSet.currDst, ffl(res))
}

type requestSet struct {
	amount  float64
	currSrc currency.Unit
	currDst currency.Unit
}

func parseInput(amount float64, src, dst string) (*requestSet, error) {
	srcISO, err := currency.ParseISO(src)
	if err != nil {
		return nil, err
	}

	dstISO, err := currency.ParseISO(dst)
	if err != nil {
		return nil, err
	}

	return &requestSet{
		amount:  amount,
		currSrc: srcISO,
		currDst: dstISO,
	}, nil
}

type currencyClient struct {
	apiURL string
	client *http.Client
}

func newCurrencyClient(url string, client *http.Client) currencyClient {
	return currencyClient{
		apiURL: url,
		client: client,
	}
}

// getExchangeRate returns exchange rate with sourche currency as base
func (cc currencyClient) getExchangeRate(scr, dst currency.Unit) (float64, error) {
	params := fmt.Sprintf("latest?base=%s&symbols=%s,%s", scr, scr, dst)
	request, err := http.NewRequest("GET", cc.apiURL+params, nil)
	if err != nil {
		return 0.0, err
	}
	resp, err := cc.client.Do(request)
	if err != nil {
		return 0.0, err
	}

	var respData response
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return 0.0, err
	}

	rate, ok := respData.Rates[dst.String()]
	if !ok {
		return 0.0, fmt.Errorf("exchange rate for %s not found", dst)
	}
	return rate, nil
}

type response struct {
	Rates map[string]float64 `json:"rates"`
}

type exchangeClient interface {
	getExchangeRate(scr, dst currency.Unit) (float64, error)
}

func convertCurrency(c exchangeClient, rs requestSet) (float64, error) {
	rate, err := c.getExchangeRate(rs.currSrc, rs.currDst)
	if err != nil {
		return 0.0, err
	}
	return rs.amount * rate, nil
}
