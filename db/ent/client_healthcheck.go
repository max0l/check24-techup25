package ent

import (
	"context"
	"errors"

	"entgo.io/ent/dialect/sql"
)

var ErrInvalidDriver = errors.New("invalid driver")

// Ping checks if the database is reachable.
// It returns an error if the database is not reachable.
func (c *Client) Ping(ctx context.Context) error {
	driver, ok := c.driver.(*sql.Driver)
	if !ok {
		return ErrInvalidDriver
	}

	return driver.DB().PingContext(ctx)
}
