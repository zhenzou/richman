package stock

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func NewSina() Provider {
	return &sina{
		cli: resty.New().SetHostURL("https://hq.sinajs.cn"),
	}
}

type sina struct {
	cli *resty.Client
}

func (s *sina) List(ctx context.Context, codes ...string) ([]Stock, error) {
	response, err := s.cli.
		R().
		SetContext(ctx).
		Get("list=" + strings.Join(codes, ","))
	if err != nil {
		return nil, err
	}
	resp, err := simplifiedchinese.GB18030.NewDecoder().String(response.String())
	if err != nil {
		return nil, err
	}
	return s.parseRawResponse(resp, codes)
}

func (s *sina) parseRawResponse(resp string, codes []string) ([]Stock, error) {
	for _, code := range codes {
		resp = strings.ReplaceAll(resp, fmt.Sprintf("var hq_str_%s=", code), "")
	}

	lines := strings.Split(resp, "\n")

	stocks := make([]Stock, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimPrefix(line, "\"")
		line = strings.TrimSuffix(line, "\";")

		if strings.TrimSpace(line) == "" {
			continue
		}
		stocks = append(stocks, s.convertToStock(line))
	}

	return stocks, nil
}

func (s *sina) convertToStock(line string) Stock {

	units := strings.Split(line, ",")
	timeStr := units[30] + "T" + units[31] + "+08:00"

	ts, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	return Stock{
		Name:         units[0],
		OpenPrice:    parsePrice(units[1]),
		BeginPrice:   parsePrice(units[2]),
		NowPrice:     parsePrice(units[3]),
		HighestPrice: parsePrice(units[4]),
		LowestPrice:  parsePrice(units[5]),
		Time:         ts,
	}
}
