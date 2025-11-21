package rqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rqlite/rqlite/v9/auth"
	"github.com/txix-open/isp-kit/http/httpcli"

	"github.com/pkg/errors"
	_ "github.com/rqlite/gorqlite/stdlib"
	"github.com/rqlite/rqlite/v9/store"
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
	storePtr      *atomic.Pointer[store.Store]
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
		storePtr:                 &atomic.Pointer[store.Store]{},
	}
}

func (r *Rqlite) Run(ctx context.Context) error {
	ctx, r.cancel = context.WithCancel(ctx)
	defer close(r.closed)

	return main(ctx, r)
}

func (r *Rqlite) WaitForLeader(timeout time.Duration) error {
	if r.storePtr.Load() == nil {
		return ErrNotRun
	}
	_, err := r.storePtr.Load().WaitForLeader(timeout)
	if err != nil {
		return errors.WithMessage(err, "wait for leader in rqlite store")
	}
	return nil
}

func (r *Rqlite) IsLeader() bool {
	if r.storePtr.Load() == nil {
		return false
	}
	return r.storePtr.Load().IsLeader()
}

func (r *Rqlite) SqlDB() (*sql.DB, error) {
	if r.storePtr.Load() == nil {
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
	return fmt.Sprintf("http://%s", r.localHttpAddr)
}

func (r *Rqlite) InternalClientCredential() *httpcli.BasicAuth {
	return r.internalClientCredential
}

func (r *Rqlite) Close() error {
	r.cancel()
	<-r.closed
	return nil
}
