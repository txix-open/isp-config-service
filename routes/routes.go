package routes

import (
	"net/http"

	"isp-config-service/conf"

	"github.com/txix-open/etp/v4"

	"isp-config-service/controller"
	"isp-config-service/controller/api"
	mws "isp-config-service/middlewares"

	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/log"
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
	c := Controllers{}
	return concatEndpoints(
		modulesDescriptors(c),
		configsDescriptors(c),
		variablesDescriptors(c),
	)
}

func GrpcHandler(wrapper endpoint.Wrapper, c Controllers) *grpc.Mux {
	muxer := grpc.NewMux()
	for _, descriptor := range modulesDescriptors(c) {
		muxer.Handle(descriptor.Path, wrapper.Endpoint(descriptor.Handler))
	}
	for _, descriptor := range configsDescriptors(c) {
		muxer.Handle(descriptor.Path, wrapper.Endpoint(descriptor.Handler))
	}
	for _, descriptor := range variablesDescriptors(c) {
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

func HttpHandler(etpSrv *etp.Server, conf conf.Local, rqliteProxy, rqliteBackupProxy http.Handler) http.Handler {
	httpMux := http.NewServeMux()
	if conf.MaintenanceMode {
		httpMux.Handle("/", rqliteProxy)
	} else {
		httpMux.Handle("/isp-etp/", etpSrv)
		httpMux.Handle("/backup", rqliteBackupProxy)
	}
	return httpMux
}

func modulesDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{
		{
			Path:    "config/module/get_modules_info",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_view"),
			Handler: c.ModuleApi.Status,
		}, {
			Path:    "config/module/delete_module",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_delete"),
			Handler: c.ModuleApi.DeleteModule,
		}, {
			Path:    "config/routing/get_routes",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_view"),
			Handler: c.ModuleApi.Connections,
		}, {
			Path:    "config/module/get_required_modules",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_view"),
			Handler: c.ModuleApi.RequiredModules,
		},
	}
}

func configsDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{
		{
			Path:    "config/config/get_active_config_by_module_name",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_view"),
			Handler: c.ConfigApi.GetActiveConfigByModuleName,
		}, {
			Path:    "config/config/get_configs_by_module_id",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_view"),
			Handler: c.ConfigApi.GetConfigsByModuleId,
		}, {
			Path:    "config/config/create_update_config",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_configuration_edit"),
			Handler: c.ConfigApi.CreateUpdateConfig,
		}, {
			Path:    "config/config/get_config_by_id",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_view"),
			Handler: c.ConfigApi.GetConfigById,
		}, {
			Path:    "config/config/mark_config_as_active",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_configuration_set_active"),
			Handler: c.ConfigApi.MarkConfigAsActive,
		}, {
			Path:    "config/config/update_config_name",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_configuration_edit"),
			Handler: c.ConfigApi.UpdateConfigName,
		}, {
			Path:    "config/config/delete_config",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_configuration_edit"),
			Handler: c.ConfigApi.DeleteConfigs,
		}, {
			Path:    "config/config/sync",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_configuration_edit"),
			Handler: c.ConfigApi.SyncConfig,
		}, {
			Path:    "config/config/get_all_version",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_view"),
			Handler: c.ConfigHistoryApi.GetAllVersion,
		}, {
			Path:    "config/config/delete_version",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_history_delete_version"),
			Handler: c.ConfigHistoryApi.DeleteConfigVersion,
		}, {
			Path:    "config/config/purge_versions",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_history_delete_version"),
			Handler: c.ConfigHistoryApi.PurgeConfigVersions,
		}, {
			Path:    "config/schema/get_by_module_id",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_view"),
			Handler: c.ConfigSchemaApi.SchemaByModuleId,
		}, {
			Path:    "config/schema/update_by_module_name",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("module_configuration_edit"),
			Handler: c.ConfigSchemaApi.UpdateSchemaByModuleName,
		},
	}
}

func variablesDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{
		{
			Path:    "config/variable/all",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("variable_view"),
			Handler: c.VariableApi.All,
		}, {
			Path:    "config/variable/get_by_name",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("variable_view"),
			Handler: c.VariableApi.GetByName,
		}, {
			Path:    "config/variable/create",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("variable_add"),
			Handler: c.VariableApi.Create,
		}, {
			Path:    "config/variable/update",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("variable_edit"),
			Handler: c.VariableApi.Update,
		}, {
			Path:    "config/variable/upsert",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("variable_edit"),
			Handler: c.VariableApi.Upsert,
		}, {
			Path:    "config/variable/delete",
			Inner:   true,
			Extra:   cluster.RequireAdminPermission("variable_delete"),
			Handler: c.VariableApi.Delete,
		},
	}
}

func concatEndpoints(endpoints ...[]cluster.EndpointDescriptor) []cluster.EndpointDescriptor {
	var result []cluster.EndpointDescriptor
	for _, c := range endpoints {
		result = append(result, c...)
	}
	return result
}
