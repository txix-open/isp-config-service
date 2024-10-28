package rqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/rqlite/rqlite/v8/auth"
	"github.com/txix-open/isp-kit/http/httpcli"
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
	cfg                      *config.Config
	internalClientCredential *httpcli.BasicAuth
	credentials              []auth.Credential

	localHttpAddr string
	store         *store.Store
	closed        chan struct{}
	cancel        context.CancelFunc
}

func New(
	cfg *config.Config,
	internalClientCredential *httpcli.BasicAuth,
	credentials []auth.Credential,
) *Rqlite {
	return &Rqlite{
		cfg:                      cfg,
		internalClientCredential: internalClientCredential,
		credentials:              credentials,
		closed:                   make(chan struct{}),
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

	db, err := sql.Open("rqlite", r.Dsn())
	if err != nil {
		return nil, errors.WithMessage(err, "open sql db")
	}
	return db, nil
}

func (r *Rqlite) Dsn() string {
	return fmt.Sprintf(
		"http://%s:%s@%s",
		r.internalClientCredential.Username,
		r.internalClientCredential.Password,
		r.localHttpAddr,
	)
}

func (r *Rqlite) LocalHttpAddr() string {
	return r.localHttpAddr
}

func (r *Rqlite) InternalClientCredential() *httpcli.BasicAuth {
	return r.internalClientCredential
}

func (r *Rqlite) Close() error {
	r.cancel()
	<-r.closed
	return nil
}
