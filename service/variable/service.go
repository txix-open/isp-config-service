package variable

import (
	"context"
	"isp-config-service/domain"
)

type Repo interface {
}

type Service struct {
}

func NewService() Service {

}

func (s Service) All(ctx context.Context) ([]domain.Variable, error) {

}

func (s Service) GetByName(ctx context.Context, name string) (*domain.Variable, error) {

}

func (s Service) Create(ctx context.Context, req domain.CreateVariableRequest) (*domain.Variable, error) {

}

func (s Service) Update(ctx context.Context, req domain.UpdateVariableRequest) (*domain.Variable, error) {

}

func (s Service) Upsert(ctx context.Context, req []domain.UpsertVariableRequest) error {

}

func (s Service) Delete(ctx context.Context, name string) (*domain.Variable, error) {

}
