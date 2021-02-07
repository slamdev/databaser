package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/slamdev/databaser/pkg/clickhouse"
	"github.com/slamdev/databaser/pkg/postgres"
	"net/url"
)

func CreatePostgresSqlConnection(ctx context.Context, params postgres.Params) (*sql.DB, error) {
	d, u := postgres.DSN(params)
	return createSqlConnection(ctx, d, u)
}

func CreateClickhouseSqlConnection(ctx context.Context, params clickhouse.Params) (*sql.DB, error) {
	d, u := clickhouse.DSN(params)
	return createSqlConnection(ctx, d, u)
}

func createSqlConnection(ctx context.Context, driver string, dsn url.URL) (*sql.DB, error) {
	c, err := sql.Open(driver, dsn.String())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to instance; %w", err)
	}
	if err := c.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failde to ping connection; %w", err)
	}
	return c, nil
}
