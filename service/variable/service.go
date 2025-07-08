package variable

import (
	"context"
	"github.com/pkg/errors"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/entity/xtypes"
	"regexp"
	"strings"
	"time"
)

type Repo interface {
	All(ctx context.Context) ([]entity.Variable, error)
	GetByName(ctx context.Context, name string) (*entity.Variable, error)
	Insert(ctx context.Context, variable entity.Variable) error
	Update(ctx context.Context, variable entity.Variable) (bool, error)
	Upsert(ctx context.Context, variables []entity.Variable) error
	Delete(ctx context.Context, name string) (bool, error)
	UpsertLinks(ctx context.Context, configId string, links []entity.ConfigHasVariable) error
}

type ConfigRepo interface {
	GetMetaByVariables(ctx context.Context, variables []string) (map[string][]entity.ConfigMetaWithVariable, error)
}

type EventRepo interface {
	Insert(ctx context.Context, event entity.Event) error
}

type Service struct {
	repo              Repo
	configRepo        ConfigRepo
	eventRepo         EventRepo
	stringToIntRegexp *regexp.Regexp
	variableRegexp    *regexp.Regexp
}

func NewService(
	repo Repo,
	configRepo ConfigRepo,
	eventRepo EventRepo,
) Service {
	return Service{
		repo:              repo,
		configRepo:        configRepo,
		eventRepo:         eventRepo,
		stringToIntRegexp: regexp.MustCompile(`"_ToInt\((\d+)\)"`),
		variableRegexp:    regexp.MustCompile(`_Var\((.*?)\)`),
	}
}

func (s Service) All(ctx context.Context) ([]domain.Variable, error) {
	variables, err := s.repo.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all variables")
	}
	if len(variables) == 0 {
		return []domain.Variable{}, nil
	}

	varNames := make([]string, 0, len(variables))
	for _, v := range variables {
		varNames = append(varNames, v.Name)
	}
	metaByVariables, err := s.configRepo.GetMetaByVariables(ctx, varNames)
	if err != nil {
		return nil, errors.WithMessage(err, "get meta by variables")
	}

	result := make([]domain.Variable, 0, len(variables))
	for _, v := range variables {
		result = append(result, s.toDto(v, metaByVariables))
	}

	return result, nil
}

func (s Service) GetByName(ctx context.Context, name string) (*domain.Variable, error) {
	variable, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, errors.WithMessage(err, "get variable by name")
	}
	if variable == nil {
		return nil, entity.ErrVariableNotFound
	}

	metaByVariables, err := s.configRepo.GetMetaByVariables(ctx, []string{name})
	if err != nil {
		return nil, errors.WithMessage(err, "get meta by variables")
	}

	result := s.toDto(*variable, metaByVariables)
	return &result, nil
}

func (s Service) Create(ctx context.Context, req domain.CreateVariableRequest) error {
	variable := entity.Variable{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Value:       req.Value,
	}
	err := s.repo.Insert(ctx, variable)
	if err != nil {
		return errors.WithMessage(err, "insert variable")
	}
	return nil
}

func (s Service) Update(ctx context.Context, req domain.UpdateVariableRequest) error {
	variable := entity.Variable{
		Name:        req.Name,
		Description: req.Description,
		Value:       req.Value,
	}
	updated, err := s.repo.Update(ctx, variable)
	if err != nil {
		return errors.WithMessage(err, "update variable")
	}
	if !updated {
		return entity.ErrVariableNotFound
	}

	err = s.syncConfigsByVariables(ctx, []string{variable.Name})
	if err != nil {
		return errors.WithMessage(err, "sync configs by variables")
	}
	return nil
}

func (s Service) Upsert(ctx context.Context, req []domain.UpsertVariableRequest) error {
	if len(req) == 0 {
		return nil
	}

	variables := make([]entity.Variable, 0, len(req))
	variableNames := make([]string, 0, len(req))
	for _, request := range req {
		variables = append(variables, entity.Variable{
			Name:        request.Name,
			Description: request.Description,
			Type:        request.Type,
			Value:       request.Value,
		})
		variableNames = append(variableNames, request.Name)
	}
	err := s.repo.Upsert(ctx, variables)
	if err != nil {
		return errors.WithMessage(err, "upsert variables")
	}

	err = s.syncConfigsByVariables(ctx, variableNames)
	if err != nil {
		return errors.WithMessage(err, "sync configs by variables")
	}

	return nil
}

func (s Service) Delete(ctx context.Context, name string) error {
	metaByVariables, err := s.configRepo.GetMetaByVariables(ctx, []string{name})
	if err != nil {
		return errors.WithMessage(err, "get meta by variables")
	}
	if len(metaByVariables[name]) > 0 {
		return entity.ErrVariableUsedInConfigs
	}

	deleted, err := s.repo.Delete(ctx, name)
	if err != nil {
		return errors.WithMessage(err, "delete variable")
	}
	if !deleted {
		return entity.ErrVariableNotFound
	}

	return nil
}

//nolint:funlen
func (s Service) RenderConfig(ctx context.Context, input []byte) ([]byte, error) {
	variables, err := s.repo.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all variables")
	}
	if len(variables) == 0 {
		return input, nil
	}
	varsMap := make(map[string]string, len(variables))
	for _, variable := range variables {
		varsMap[variable.Name] = variable.Value
	}

	undefinedVariables := make([]string, 0)
	withReplacedVars := s.variableRegexp.ReplaceAllStringFunc(string(input), func(group string) string {
		varName := s.variableRegexp.FindStringSubmatch(group)[1]

		value, ok := varsMap[varName]
		if !ok {
			undefinedVariables = append(undefinedVariables, varName)
			return varName
		}
		return value
	})
	if len(undefinedVariables) > 0 {
		return nil, errors.Errorf("undefined variables: [%s]", strings.Join(undefinedVariables, ", "))
	}

	result := s.stringToIntRegexp.ReplaceAll([]byte(withReplacedVars), []byte("$1"))

	return result, nil
}

func (s Service) ExtractVariables(ctx context.Context, configId string, input []byte) (*domain.VariableExtractionResult, error) {
	matches := s.variableRegexp.FindAllStringSubmatch(string(input), -1)
	varNames := make([]string, len(matches))
	for i, match := range matches {
		varNames[i] = match[1]
	}
	if len(varNames) == 0 {
		return &domain.VariableExtractionResult{
			VariableLinks:   []entity.ConfigHasVariable{},
			AbsentVariables: []string{},
		}, nil
	}

	allVariables, err := s.repo.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all variables")
	}
	allVariablesMap := make(map[string]bool, len(allVariables))
	for _, variable := range allVariables {
		allVariablesMap[variable.Name] = true
	}

	absentVariables := make([]string, 0)
	variableLinks := make([]entity.ConfigHasVariable, 0)
	for _, name := range varNames {
		if !allVariablesMap[name] {
			absentVariables = append(absentVariables, name)
			continue
		}
		variableLinks = append(variableLinks, entity.ConfigHasVariable{
			ConfigId:     configId,
			VariableName: name,
		})
	}

	return &domain.VariableExtractionResult{
		VariableLinks:   variableLinks,
		AbsentVariables: absentVariables,
	}, nil
}

func (s Service) SaveVariableLinks(ctx context.Context, configId string, links []entity.ConfigHasVariable) error {
	err := s.repo.UpsertLinks(ctx, configId, links)
	if err != nil {
		return errors.WithMessage(err, "upsert links")
	}

	return nil
}

func (s Service) toDto(variable entity.Variable, metaByVariables map[string][]entity.ConfigMetaWithVariable) domain.Variable {
	value := variable.Value
	if variable.Type == "SECRET" {
		value = ""
	}
	containsInConfigs := []domain.LinkedConfig{}
	for _, metaWithVariable := range metaByVariables[variable.Name] {
		containsInConfigs = append(containsInConfigs, domain.LinkedConfig{
			Id:       metaWithVariable.Id,
			ModuleId: metaWithVariable.ModuleId,
			Name:     metaWithVariable.Name,
		})
	}
	return domain.Variable{
		Name:              variable.Name,
		Description:       variable.Description,
		Type:              variable.Type,
		Value:             value,
		CreatedAt:         time.Time(variable.CreatedAt),
		UpdatedAt:         time.Time(variable.UpdatedAt),
		ContainsInConfigs: containsInConfigs,
	}
}

func (s Service) syncConfigsByVariables(ctx context.Context, variables []string) error {
	metaByVariables, err := s.configRepo.GetMetaByVariables(ctx, variables)
	if err != nil {
		return errors.WithMessage(err, "get meta by variables")
	}
	modulesToSync := make(map[string]bool)
	for _, variable := range variables {
		for _, withVariable := range metaByVariables[variable] {
			if !withVariable.Active {
				continue
			}
			modulesToSync[withVariable.ModuleId] = true
		}
	}

	for moduleId := range modulesToSync {
		err := s.eventRepo.Insert(ctx, entity.Event{
			Payload: xtypes.Json[entity.EventPayload]{
				Value: entity.EventPayload{
					ConfigUpdated: &entity.PayloadConfigUpdated{
						ModuleId: moduleId,
					},
				},
			},
		})
		if err != nil {
			return errors.WithMessagef(err, "insert change config event for moduleId: %s", moduleId)
		}
	}

	return nil
}
