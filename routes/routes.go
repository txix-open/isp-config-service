package routes

import (
	"github.com/txix-open/etp/v4"
	"isp-config-service/conf"
	"net/http"

	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/controller"
	"isp-config-service/controller/api"
	mws "isp-config-service/middlewares"
)

type Controllers struct {
	Module           controller.Module
	ModuleApi        api.Module
	ConfigApi        api.Config
	ConfigHistoryApi api.ConfigHistory
	ConfigSchemaApi  api.ConfigSchema
	VariableApi      api.Variable
}

func EndpointDescriptors() []cluster.EndpointDescriptor {
	return endpointDescriptors(Controllers{})
}

func GrpcHandler(wrapper endpoint.Wrapper, c Controllers) *grpc.Mux {
	muxer := grpc.NewMux()
	for _, descriptor := range endpointDescriptors(c) {
		muxer.Handle(descriptor.Path, wrapper.Endpoint(descriptor.Handler))
	}
	return muxer
}

func BindEtp(etpSrv *etp.Server, c Controllers, logger log.Logger) {
	middlewares := []mws.EtpMiddleware{
		mws.EtpLogger(logger),
	}
	etpSrv.OnConnect(c.Module.OnConnect)
	etpSrv.OnDisconnect(c.Module.OnDisconnect)
	etpSrv.OnError(c.Module.OnError)

	onConfigSchema := mws.EtpChain(etp.HandlerFunc(c.Module.OnModuleConfigSchema), middlewares...)
	etpSrv.On(cluster.ModuleSendConfigSchema, onConfigSchema)

	onRequirements := mws.EtpChain(etp.HandlerFunc(c.Module.OnModuleRequirements), middlewares...)
	etpSrv.On(cluster.ModuleSendRequirements, onRequirements)

	onModuleReady := mws.EtpChain(etp.HandlerFunc(c.Module.OnModuleReady), middlewares...)
	etpSrv.On(cluster.ModuleReady, onModuleReady)
}

func HttpHandler(etpSrv *etp.Server, conf conf.Local, rqliteProxy http.Handler) http.Handler {
	httpMux := http.NewServeMux()
	if conf.MaintenanceMode {
		httpMux.Handle("/", rqliteProxy)
	} else {
		httpMux.Handle("/isp-etp/", etpSrv)
	}
	return httpMux
}

func endpointDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{
		// modules
		{
			Path:    "config/module/get_modules_info",
			Inner:   true,
			Handler: c.ModuleApi.Status,
		}, {
			Path:    "config/module/delete_module",
			Inner:   true,
			Handler: c.ModuleApi.DeleteModule,
		}, {
			Path:    "config/routing/get_routes",
			Inner:   true,
			Handler: c.ModuleApi.Connections,
		},
		// configs
		{
			Path:    "config/config/get_active_config_by_module_name",
			Inner:   true,
			Handler: c.ConfigApi.GetActiveConfigByModuleName,
		}, {
			Path:    "config/config/get_configs_by_module_id",
			Inner:   true,
			Handler: c.ConfigApi.GetConfigsByModuleId,
		}, {
			Path:    "config/config/create_update_config",
			Inner:   true,
			Handler: c.ConfigApi.CreateUpdateConfig,
		}, {
			Path:    "config/config/get_config_by_id",
			Inner:   true,
			Handler: c.ConfigApi.GetConfigById,
		}, {
			Path:    "config/config/mark_config_as_active",
			Inner:   true,
			Handler: c.ConfigApi.MarkConfigAsActive,
		}, {
			Path:    "config/config/delete_config",
			Inner:   true,
			Handler: c.ConfigApi.DeleteConfigs,
		}, {
			Path:    "config/config/get_all_version",
			Inner:   true,
			Handler: c.ConfigHistoryApi.GetAllVersion,
		}, {
			Path:    "config/config/delete_version",
			Inner:   true,
			Handler: c.ConfigHistoryApi.DeleteConfigVersion,
		}, {
			Path:    "config/schema/get_by_module_id",
			Inner:   true,
			Handler: c.ConfigSchemaApi.SchemaByModuleId,
		},
		// variables
		{
			Path:    "variable/all",
			Inner:   true,
			Handler: c.VariableApi.All,
		}, {
			Path:    "variable/get_by_name",
			Inner:   true,
			Handler: c.VariableApi.GetByName,
		}, {
			Path:    "variable/create",
			Inner:   true,
			Handler: c.VariableApi.Create,
		}, {
			Path:    "variable/update",
			Inner:   true,
			Handler: c.VariableApi.Update,
		}, {
			Path:    "variable/upsert",
			Inner:   true,
			Handler: c.VariableApi.Upsert,
		}, {
			Path:    "variable/delete",
			Inner:   true,
			Handler: c.VariableApi.Delete,
		},
	}
}
