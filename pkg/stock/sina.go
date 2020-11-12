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
	// var hq_str_sh601006="大秦铁路,6.480,6.490,6.520,6.550,6.450,6.520,6.530,26048333,169706151.000,456838,6.520,843440,6.510,305800,6.500,334600,6.490,158700,6.480,395300,6.530,1535656,6.540,2976299,6.550,1109280,6.560,807600,6.570,2020-11-11,15:00:02,00,";
	response, err := s.cli.R().Get("list=" + strings.Join(codes, ","))
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

	lines := strings.Split(resp, ";")

	stocks := make([]Stock, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		stocks = append(stocks, s.convertToStock(line))
	}

	return stocks, nil
}

func (s *sina) convertToStock(line string) Stock {

	line = strings.TrimPrefix(line, "\"")
	line = strings.TrimSuffix(line, ",\"")

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
