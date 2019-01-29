package module

import (
	"github.com/integration-system/isp-lib/config"
	"isp-config-service/conf"
	"isp-config-service/domain"
	st "isp-config-service/entity"
	"isp-config-service/model"
	"isp-config-service/socket"
	"sort"
	"time"
)

var (
	startAt = time.Now()
	Version = "0.1.0"
)

func GetAggregatedModuleInfo(instanceUuid string) ([]*domain.ModuleInfo, error) {
	modules, err := model.ModulesRep.GetModulesByInstanceUuid(instanceUuid)
	if err != nil {
		return nil, err
	}
	config := config.Get().(*conf.Configuration)
	modules = append(modules, st.Module{
		Id:         -1,
		Name:       config.ModuleName,
		Active:     true,
		CreatedAt:  startAt,
		InstanceId: -1,
	})
	l := len(modules)
	res := make([]*domain.ModuleInfo, l)
	m := make(map[int32]*domain.ModuleInfo, l)
	idList := make([]int32, l)
	for i, module := range modules {
		idList[i] = module.Id
		info := &domain.ModuleInfo{
			Id:                 module.Id,
			Name:               module.Name,
			Active:             module.Active,
			CreatedAt:          module.CreatedAt,
			LastConnectedAt:    module.LastConnectedAt,
			LastDisconnectedAt: module.LastDisconnectedAt,
		}
		m[module.Id] = info
		res[i] = info
	}

	configs, err := model.ConfigRep.GetConfigsByModulesId(idList)
	if err != nil {
		return nil, err
	}
	for _, cfg := range configs {
		info := m[cfg.ModuleId]
		info.Configs = append(info.Configs, cfg)
	}

	schemas, err := model.SchemaRep.GetSchemasByModulesId(idList)
	if err != nil {
		return nil, err
	}
	for _, s := range schemas {
		info := m[s.ModuleId]
		schema := s.Schema
		info.ConfigSchema = &schema
	}

	connMap := socket.GetModuleConnectionIdMapByInstanceId(instanceUuid)
	routes := socket.GetRoutes().Routes
	//converters := socket.GetConverters()
	//routers := socket.GetRouters()
	roomsStat := socket.GetRoomsCount()
	for _, info := range res {
		if info.Id == -1 {
			info.Status = append(info.Status, domain.Connection{
				Address:       config.GrpcOuterAddress,
				Version:       Version,
				Endpoints:     socket.ConfigServiceBackendDeclaration.Endpoints,
				LibVersion:    socket.ConfigServiceBackendDeclaration.LibVersion,
				EstablishedAt: startAt,
			})
		} else if conns, connected := connMap[info.Name]; connected {
			for _, conId := range conns {
				hasRoutes := false
				var data *domain.Connection
				if backend, ok := routes[conId]; ok {
					data = &domain.Connection{
						Version:    backend.Version,
						LibVersion: backend.LibVersion,
						Address:    backend.Address,
						Endpoints:  backend.Endpoints,
					}

					hasRoutes = true
				} else if roomsStat[instanceUuid][info.Name] <= 0 {
					continue
				}
				if c, ok := socket.GetConnectionById(instanceUuid, conId); ok {
					if !hasRoutes {
						data = &domain.Connection{
							Version: "unknown",
						}
						/*if router, ok := routers[conId]; ok {
							data.Address = router
						} else if converter, ok := converters[conId]; ok {
							data.Address = converter
						}*/
					}

					data.EstablishedAt = c.EstablishedAt
				}

				info.Status = append(info.Status, *data)
			}
		}
	}

	for _, s := range res {
		sort.Slice(s.Status, func(i, j int) bool {
			return s.Status[i].EstablishedAt.Before(s.Status[j].EstablishedAt)
		})
		for _, con := range s.Status {
			sort.Slice(con.Endpoints, func(i, j int) bool {
				return con.Endpoints[i].Path < con.Endpoints[j].Path
			})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	return res, nil
}
