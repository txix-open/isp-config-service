package repository

import (
	"context"

	"github.com/txix-open/isp-kit/db"
	"isp-config-service/entity"
)

type Module struct {
	db db.DB
}

func NewModule(db db.DB) Module {
	return Module{
		db: db,
	}
}

func (r Module) Upsert(ctx context.Context, module entity.Module) error {

}
