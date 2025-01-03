package api

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/json"
	"golang.org/x/sync/errgroup"
	"isp-config-service/domain"
	"isp-config-service/entity"
	"isp-config-service/helpers"
)

type ModuleRepo interface {
	All(ctx context.Context) ([]entity.Module, error)
	Delete(ctx context.Context, id string) error
	GetByNames(ctx context.Context, names []string) ([]entity.Module, error)
}

type BackendsRepo interface {
	All(ctx context.Context) ([]entity.Backend, error)
}

type SchemaRepo interface {
	All(ctx context.Context) ([]entity.ConfigSchema, error)
	GetByModuleId(ctx context.Context, moduleId string) (*entity.ConfigSchema, error)
}

type Module struct {
	moduleRepo   ModuleRepo
	backendsRepo BackendsRepo
	schemaRepo   SchemaRepo
}

func NewModule(
	moduleRepo ModuleRepo,
	backendsRepo BackendsRepo,
	schemaRepo SchemaRepo,
) Module {
	return Module{
		moduleRepo:   moduleRepo,
		backendsRepo: backendsRepo,
		schemaRepo:   schemaRepo,
	}
}

func (s Module) Status(ctx context.Context) ([]domain.ModuleInfo, error) {
	var (
		modules  []entity.Module
		backends []entity.Backend
		schemas  []entity.ConfigSchema
	)
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		modules, err = s.moduleRepo.All(ctx)
		return errors.WithMessage(err, "get all modules")
	})
	group.Go(func() error {
		var err error
		backends, err = s.backendsRepo.All(ctx)
		return errors.WithMessage(err, "get all backends")
	})
	group.Go(func() error {
		var err error
		schemas, err = s.schemaRepo.All(ctx)
		return errors.WithMessage(err, "get all schemas")
	})
	err := group.Wait()
	if err != nil {
		return nil, errors.WithMessage(err, "wait")
	}

	connections := make(map[string][]domain.Connection, len(modules))
	for _, backend := range backends {
		conn, err := s.backendToDto(backend)
		if err != nil {
			return nil, errors.WithMessage(err, "backend to dto")
		}
		connections[backend.ModuleId] = append(connections[backend.ModuleId], conn)
	}

	schemasMap := make(map[string]json.RawMessage)
	for _, schema := range schemas {
		schemasMap[schema.ModuleId] = schema.Data
	}

	moduleInfos := make([]domain.ModuleInfo, 0, len(modules))
	for _, module := range modules {
		info := domain.ModuleInfo{
			Id:                 module.Id,
			Name:               module.Name,
			Active:             len(connections[module.Id]) > 0,
			LastConnectedAt:    (*time.Time)(module.LastConnectedAt),
			LastDisconnectedAt: (*time.Time)(module.LastDisconnectedAt),
			ConfigSchema:       schemasMap[module.Id],
			Status:             connections[module.Id],
			CreatedAt:          time.Time(module.CreatedAt),
		}
		moduleInfos = append(moduleInfos, info)
	}

	return moduleInfos, nil
}

func (s Module) Delete(ctx context.Context, id string) error {
	err := s.moduleRepo.Delete(ctx, id)
	if err != nil {
		return errors.WithMessage(err, "delete modules")
	}
	return nil
}

func (s Module) Connections(ctx context.Context) ([]domain.Connection, error) {
	backends, err := s.backendsRepo.All(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get all backends")
	}

	connections := make([]domain.Connection, 0, len(backends))
	for _, backend := range backends {
		conn, err := s.backendToDto(backend)
		if err != nil {
			return nil, errors.WithMessage(err, "backend to dto")
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

func (s Module) backendToDto(backend entity.Backend) (domain.Connection, error) {
	addr, err := helpers.SplitAddress(backend)
	if err != nil {
		return domain.Connection{}, errors.WithMessage(err, "split address")
	}

	endpointsDescriptors := make([]domain.EndpointDescriptor, 0)
	for _, desc := range backend.Endpoints.Value {
		endpointsDescriptors = append(endpointsDescriptors, domain.EndpointDescriptor{
			Path:  desc.Path,
			Inner: desc.Inner,
		})
	}
	slices.SortFunc(endpointsDescriptors, func(a, b domain.EndpointDescriptor) int {
		return cmp.Compare(a.Path, b.Path)
	})

	conn := domain.Connection{
		Id:         backend.WsConnectionId,
		ModuleName: backend.ModuleName,
		LibVersion: backend.LibVersion,
		Version:    backend.Version,
		Address: domain.Address{
			Ip:   addr.IP,
			Port: addr.Port,
		},
		Endpoints:     endpointsDescriptors,
		EstablishedAt: time.Time(backend.CreatedAt),
	}

	return conn, nil
}
