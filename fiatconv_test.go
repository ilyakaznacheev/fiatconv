package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"golang.org/x/text/currency"
)

var testErr = errors.New("test")

type testExchangeClientMock struct {
	rate float64
	err  error
}

func (m *testExchangeClientMock) getExchangeRate(scr, dst currency.Unit) (float64, error) {
	return m.rate, m.err
}

func Test_convertCurrency(t *testing.T) {
	type args struct {
		c exchangeClient
	}
	tests := []struct {
		name     string
		rs       requestSet
		mockRate float64
		mockErr  error
		want     float64
		wantErr  bool
	}{
		{
			name: "normal case",
			rs: requestSet{
				amount: 2.0,
			},
			mockRate: 3.0,
			mockErr:  nil,
			want:     6.0,
			wantErr:  false,
		},

		{
			name: "errpr case",
			rs: requestSet{
				amount: 2.0,
			},
			mockRate: 1.0,
			mockErr:  testErr,
			want:     0.0,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testExchangeClientMock{
				rate: tt.mockRate,
				err:  tt.mockErr,
			}
			got, err := convertCurrency(&client, tt.rs)
			if (err != nil) != tt.wantErr {
				t.Errorf("wrong error behavior %v, want %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("wrong rate %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_currencyClient_getExchangeRate(t *testing.T) {

	mustParse := func(c string) currency.Unit {
		curr, _ := currency.ParseISO(c)
		return curr
	}

	tests := []struct {
		name       string
		respString string
		respErr    error
		scr        currency.Unit
		dst        currency.Unit
		want       float64
		wantErr    bool
	}{
		{
			name:       "normal case",
			respString: `{"rates":{"EUR":0.89}}`,
			scr:        mustParse("USD"),
			dst:        mustParse("EUR"),
			want:       0.89,
			wantErr:    false,
		},

		{
			name:    "error case",
			respErr: testErr,
			scr:     mustParse("USD"),
			dst:     mustParse("EUR"),
			want:    0.0,
			wantErr: true,
		},

		{
			name:       "wrong response",
			respString: `{"rates":{"JPY":123.0}}`,
			scr:        mustParse("USD"),
			dst:        mustParse("EUR"),
			want:       0.0,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.respErr != nil {
					http.Error(w, tt.respErr.Error(), http.StatusInternalServerError)
					return
				}
				w.Write([]byte(tt.respString))
			}))
			defer server.Close()

			clntURL, _ := url.Parse(server.URL)

			cc := currencyClient{
				apiURL: clntURL,
				client: server.Client(),
			}
			got, err := cc.getExchangeRate(tt.scr, tt.dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("wrong error behavior %v, want %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("wrong currency %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseInput(t *testing.T) {
	mustParse := func(c string) currency.Unit {
		curr, _ := currency.ParseISO(c)
		return curr
	}

	tests := []struct {
		name    string
		amount  float64
		src     string
		dst     string
		want    requestSet
		wantErr bool
	}{
		{
			name:   "normal case",
			amount: 1.23,
			src:    "USD",
			dst:    "EUR",
			want: requestSet{
				amount:  1.23,
				currSrc: mustParse("USD"),
				currDst: mustParse("EUR"),
			},
			wantErr: false,
		},

		{
			name:    "error case A",
			amount:  1.23,
			src:     "---",
			dst:     "EUR",
			wantErr: true,
		},

		{
			name:    "error case B",
			amount:  1.23,
			src:     "USD",
			dst:     "---",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInput(tt.amount, tt.src, tt.dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("wrong error behavior %v, want %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("wrong parsing result %v, want %v", *got, tt.want)
			}
		})
	}
}
