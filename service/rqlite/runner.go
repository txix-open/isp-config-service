package rqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	_ "github.com/rqlite/gorqlite/stdlib"
	"github.com/rqlite/rqlite/v8/store"
	"github.com/txix-open/isp-kit/config"
)

var (
	ErrNotRun = errors.New("rqlite runner is not initialized")
)

type Rqlite struct {
	cfg *config.Config

	localHttpAddr string
	store         *store.Store
	closed        chan struct{}
	cancel        context.CancelFunc
}

func New(cfg *config.Config) *Rqlite {
	return &Rqlite{
		cfg:    cfg,
		closed: make(chan struct{}),
	}
}

func (r *Rqlite) Run(ctx context.Context) error {
	ctx, r.cancel = context.WithCancel(ctx)
	defer close(r.closed)

	return main(ctx, r)
}

func (r *Rqlite) WaitForLeader(timeout time.Duration) error {
	if r.store == nil {
		return ErrNotRun
	}
	_, err := r.store.WaitForLeader(timeout)
	if err != nil {
		return errors.WithMessage(err, "wait for leader in rqlite store")
	}
	return nil
}

func (r *Rqlite) IsLeader() bool {
	if r.store == nil {
		return false
	}
	return r.store.IsLeader()
}

func (r *Rqlite) SqlDB() (*sql.DB, error) {
	if r.store == nil {
		return nil, ErrNotRun
	}

	return sql.Open("rqlite", r.Dsn())
}

func (r *Rqlite) Dsn() string {
	return fmt.Sprintf("http://%s", r.localHttpAddr)
}

func (r *Rqlite) Close() error {
	r.cancel()
	<-r.closed
	return nil
}
