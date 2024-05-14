package routes

import (
	"net/http"

	"github.com/txix-open/etp/v3"
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

func HttpHandler(etpSrv *etp.Server) http.Handler {
	httpMux := http.NewServeMux()
	httpMux.Handle("/isp-etp/", etpSrv)
	return httpMux
}

func endpointDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{{
		Path:    "config/module/get_modules_info",
		Inner:   true,
		Handler: c.ModuleApi.GetModulesAggregatedInfo,
	}, {
		Path:    "config/module/delete_module",
		Inner:   true,
		Handler: c.ModuleApi.DeleteModule,
	}, {
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
	}}
}
