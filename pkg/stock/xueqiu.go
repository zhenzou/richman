package stock

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shopspring/decimal"
)

func NewXueQiu() Provider {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return &xueqiu{
		cli: resty.New().
			SetHostURL("").
			SetCookieJar(jar),
		cache: map[string]string{},
	}
}

type xueqiu struct {
	once  sync.Once
	cli   *resty.Client
	cache map[string]string
}

type xueqiuQuote struct {
	Symbol             string          `json:"symbol"`
	Current            decimal.Decimal `json:"current"`
	Percent            decimal.Decimal `json:"percent"`
	Chg                decimal.Decimal `json:"chg"`
	Timestamp          int64           `json:"timestamp"`
	Volume             int             `json:"volume"`
	Amount             decimal.Decimal `json:"amount"`
	MarketCapital      decimal.Decimal `json:"market_capital"`
	FloatMarketCapital decimal.Decimal `json:"float_market_capital"`
	TurnoverRate       decimal.Decimal `json:"turnover_rate"`
	Amplitude          decimal.Decimal `json:"amplitude"`
	Open               decimal.Decimal `json:"open"`
	LastClose          decimal.Decimal `json:"last_close"`
	High               decimal.Decimal `json:"high"`
	Low                decimal.Decimal `json:"low"`
	AvgPrice           decimal.Decimal `json:"avg_price"`
	TradeVolume        decimal.Decimal `json:"trade_volume"`
	Side               decimal.Decimal `json:"side"`
	IsTrade            bool            `json:"is_trade"`
	Level              decimal.Decimal `json:"level"`
	CurrentYearPercent decimal.Decimal `json:"current_year_percent"`
	TradeUniqueID      string          `json:"trade_unique_id"`
	Type               int             `json:"type"`
}

type xueqiuRealTimeResponse struct {
	Data             []xueqiuQuote `json:"data"`
	ErrorCode        int           `json:"error_code"`
	ErrorDescription interface{}   `json:"error_description"`
}

func (s *xueqiu) getCookie(ctx context.Context) error {
	resp, err := s.cli.
		R().
		SetContext(ctx).
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36").
		Get("https://xueqiu.com")

	if resp.StatusCode() != http.StatusOK {
		return errors.New(resp.Status())
	}
	return err
}

func (s *xueqiu) List(ctx context.Context, codes ...string) ([]Stock, error) {

	var err error
	s.once.Do(func() {
		err = s.getCookie(ctx)
	})

	if err != nil {
		return nil, err
	}

	var resp xueqiuRealTimeResponse
	_, err = s.cli.
		R().
		SetContext(ctx).
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36").
		SetQueryParam("symbol", strings.ToUpper(strings.Join(codes, ","))).
		SetResult(&resp).
		Get("https://stock.xueqiu.com/v5/stock/realtime/quotec.json")
	if err != nil {
		return nil, err
	}

	stocks := make([]Stock, 0, len(resp.Data))

	for _, datum := range resp.Data {
		name, err := s.getStockName(ctx, datum.Symbol)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, s.convert(datum, name))
	}
	return stocks, nil
}

type xueqiuDetailResponse struct {
	Data struct {
		Quote struct {
			Symbol string `json:"symbol"`
			Code   string `json:"code"`
			Name   string `json:"name"`
		} `json:"quote"`
	} `json:"data"`
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func (s *xueqiu) getStockName(ctx context.Context, code string) (string, error) {
	name, ok := s.cache[code]
	if ok {
		return name, nil
	}
	var resp xueqiuDetailResponse
	_, err := s.cli.
		R().
		SetContext(ctx).
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.99 Safari/537.36").
		SetHeader("Content-Type", "application/json;charset=UTF-8").
		SetHeader("Accept-Encoding", "deflate, gzip").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("Upgrade-Insecure-Requests", "1").
		SetQueryParam("symbol", strings.TrimSpace(code)).
		SetResult(&resp).
		Get("https://stock.xueqiu.com/v5/stock/quote.json")

	if err != nil {
		return "", err
	}

	s.cache[code] = resp.Data.Quote.Name
	return resp.Data.Quote.Name, nil
}

func (s *xueqiu) convert(datum xueqiuQuote, name string) Stock {
	return Stock{
		Name:         name,
		OpenPrice:    datum.Open,
		BeginPrice:   datum.LastClose,
		NowPrice:     datum.Current,
		HighestPrice: datum.High,
		LowestPrice:  datum.Low,
		Time:         time.Unix(datum.Timestamp, 0),
	}
}
