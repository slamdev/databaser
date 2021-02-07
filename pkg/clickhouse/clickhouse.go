package clickhouse

import (
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	"net/url"
)

type Params struct {
	User     string
	Password string
	Host     string
	Port     int
}

func DSN(params Params) (string, url.URL) {
	query := url.Values{}
	if params.User != "" {
		query.Set("username", params.User)
	}
	if params.Password != "" {
		query.Set("password", params.Password)
	}
	dbUrl := url.URL{
		Scheme:   "tcp",
		Host:     fmt.Sprintf("%s:%d", params.Host, params.Port),
		RawQuery: query.Encode(),
	}
	return "clickhouse", dbUrl
}
