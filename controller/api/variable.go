package api

import (
	"context"
	_ "github.com/txix-open/isp-kit/grpc/apierrors"
	"isp-config-service/domain"
)

type Variable struct {
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

}

// Create
// @Summary Метод создания новой переменной
// @Description Создает новую переменную с указанным наименованием и значением
// @Tags Переменные
// @Accept json
// @Produce json
// @Param body body domain.CreateVariableRequest true "тело запроса"
// @Success 200 {object} domain.Variable
// @Failure 400 {object} apierrors.Error "`errorCode: 2007` - переменная с таким именем уже существует<br/>"
// @Failure 500 {object} apierrors.Error
// @Router /variable/create [POST]
func (c Variable) Create(ctx context.Context, req domain.CreateVariableRequest) (*domain.Variable, error) {

}

// Update
// @Summary Метод обновления переменной
// @Description Выполняет обновление переменной по ее названию
// @Tags Переменные
// @Accept json
// @Produce json
// @Param body body domain.UpdateVariableRequest true "тело запроса"
// @Success 200 {object} domain.Variable
// @Failure 500 {object} apierrors.Error
// @Router /variable/update [POST]
func (c Variable) Update(ctx context.Context, req domain.UpdateVariableRequest) (*domain.Variable, error) {

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
func (c Variable) Delete(ctx context.Context, req domain.VariableByNameRequest) (*domain.Variable, error) {

}
