//nolint:lll
package api

import (
	"context"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/apierrors"
	_ "github.com/txix-open/isp-kit/grpc/apierrors"
	"isp-config-service/domain"
	"isp-config-service/entity"
)

type Service interface {
	All(ctx context.Context) ([]domain.Variable, error)
	GetByName(ctx context.Context, name string) (*domain.Variable, error)
	Create(ctx context.Context, req domain.CreateVariableRequest) error
	Update(ctx context.Context, req domain.UpdateVariableRequest) error
	Upsert(ctx context.Context, req []domain.UpsertVariableRequest) error
	Delete(ctx context.Context, name string) error
}

type Variable struct {
	service Service
}

func NewVariable(service Service) Variable {
	return Variable{
		service: service,
	}
}

// All
// @Summary Метод получения списка всех переменных
// @Tags Переменные
// @Accept json
// @Produce json
// @Success 200 {array} domain.Variable
// @Failure 500 {object} apierrors.Error
// @Router /variable/all [POST]
func (c Variable) All(ctx context.Context) ([]domain.Variable, error) {
	return c.service.All(ctx)
}

// GetByName
// @Summary Метод получения переменной по точному совпадению имени
// @Tags Переменные
// @Accept json
// @Produce json
// @Param body body domain.VariableByNameRequest true "тело запроса"
// @Success 200 {object} domain.Variable
// @Failure 400 {object} apierrors.Error "`errorCode: 2006` - переменная по имени не найдена<br/>"
// @Failure 500 {object} apierrors.Error
// @Router /variable/get_by_name [POST]
func (c Variable) GetByName(ctx context.Context, req domain.VariableByNameRequest) (*domain.Variable, error) {
	v, err := c.service.GetByName(ctx, req.Name)
	switch {
	case errors.Is(err, entity.ErrVariableNotFound):
		return nil, apierrors.NewBusinessError(domain.ErrorCodeVariableNotFound, "variable by name not found", err)
	case err != nil:
		return nil, apierrors.NewInternalServiceError(err)
	}
	return v, nil
}

// Create
// @Summary Метод создания новой переменной
// @Description Создает новую переменную с указанным наименованием и значением
// @Tags Переменные
// @Accept json
// @Produce json
// @Param body body domain.CreateVariableRequest true "тело запроса"
// @Success 200
// @Failure 400 {object} apierrors.Error "`errorCode: 2007` - переменная с таким именем уже существует<br/>"
// @Failure 500 {object} apierrors.Error
// @Router /variable/create [POST]
func (c Variable) Create(ctx context.Context, req domain.CreateVariableRequest) error {
	err := c.service.Create(ctx, req)
	switch {
	case errors.Is(err, entity.ErrVariableAlreadyExists):
		return apierrors.NewBusinessError(domain.ErrorCodeVariableAlreadyExists, "variable with the same name already exists", err)
	case err != nil:
		return apierrors.NewInternalServiceError(err)
	}
	return nil
}

// Update
// @Summary Метод обновления переменной
// @Description Выполняет обновление переменной по ее названию
// @Tags Переменные
// @Accept json
// @Produce json
// @Param body body domain.UpdateVariableRequest true "тело запроса"
// @Success 200
// @Failure 400 {object} apierrors.Error "`errorCode: 2006` - переменная по имени не найдена<br/>"
// @Failure 500 {object} apierrors.Error
// @Router /variable/update [POST]
func (c Variable) Update(ctx context.Context, req domain.UpdateVariableRequest) error {
	err := c.service.Update(ctx, req)
	switch {
	case errors.Is(err, entity.ErrVariableNotFound):
		return apierrors.NewBusinessError(domain.ErrorCodeVariableNotFound, "variable by name not found", err)
	case err != nil:
		return apierrors.NewInternalServiceError(err)
	}
	return nil
}

// Upsert
// @Summary Метод добавления/обновления переменных
// @Description Выполняет безусловное добавление/обновление указанных в теле запроса переменных
// @Description Операция выполняется не в транзакции
// @Tags Переменные
// @Accept json
// @Produce json
// @Param body body []domain.UpsertVariableRequest true "тело запроса"
// @Success 200
// @Failure 500 {object} apierrors.Error
// @Router /variable/upsert [POST]
func (c Variable) Upsert(ctx context.Context, req []domain.UpsertVariableRequest) error {
	err := c.service.Upsert(ctx, req)
	if err != nil {
		return apierrors.NewInternalServiceError(err)
	}
	return nil
}

// Delete
// @Summary Метод удаления переменной
// @Description Переменная связанная с конфигурациями, не может быть удалена
// @Tags Переменные
// @Accept json
// @Produce json
// @Param body body domain.VariableByNameRequest true "тело запроса"
// @Success 200
// @Failure 400 {object} apierrors.Error "`errorCode: 2006` - переменная по имени не найдена<br/>`errorCode: 2008` - переменная используется в конфигурациях<br/>"
// @Failure 500 {object} apierrors.Error
// @Router /variable/delete [POST]
func (c Variable) Delete(ctx context.Context, req domain.VariableByNameRequest) error {
	err := c.service.Delete(ctx, req.Name)
	switch {
	case errors.Is(err, entity.ErrVariableNotFound):
		return apierrors.NewBusinessError(domain.ErrorCodeVariableNotFound, "variable by name not found", err)
	case errors.Is(err, entity.ErrVariableUsedInConfigs):
		return apierrors.NewBusinessError(domain.ErrorCodeVariableUsedInConfigs, "variable used in configs", err)
	case err != nil:
		return apierrors.NewInternalServiceError(err)
	}
	return nil
}
