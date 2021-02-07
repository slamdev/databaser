package postgres

import (
	"fmt"
	_ "github.com/lib/pq"
	"net/url"
)

type Params struct {
	User     string
	Password string
	Host     string
	Port     int
	AuthDB   string
}

func DSN(params Params) (string, url.URL) {
	query := url.Values{}
	dbUrl := url.URL{
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%d", params.Host, params.Port),
		RawQuery: query.Encode(),
	}
	if params.Password != "" {
		dbUrl.User = url.UserPassword(params.User, params.Password)
	} else if params.User != "" {
		dbUrl.User = url.User(params.User)
	}
	if params.AuthDB != "" {
		dbUrl.Path = params.AuthDB
	}
	return "postgres", dbUrl
}
