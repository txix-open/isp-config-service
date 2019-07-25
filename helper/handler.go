package helper

import (
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/controllers"
	"isp-config-service/domain"
	"isp-config-service/entity"
)

type Handlers struct {
	// ===== INSTANCE =====
	GetInstances         func(identities []int32) ([]entity.Instance, error)      `method:"get_instances" group:"instance" inner:"true"`
	CreateUpdateInstance func(instance entity.Instance) (*entity.Instance, error) `method:"create_update_instance" group:"instance" inner:"true"`
	DeleteInstance       func(identities []int32) (*domain.DeleteResponse, error) `method:"delete_instance" group:"instance" inner:"true"`

	// ===== MODULE =====
	GetModules          func(identities []int32) ([]entity.Module, error)         `method:"get_modules" group:"module" inner:"true"`
	GetSchemaByModuleId func(moduleId int32) (*entity.ConfigSchema, error)        `method:"get_schema" group:"module" inner:"true"`
	GetActiveModules    func() ([]entity.Module, error)                           `method:"get_active_modules" group:"module" inner:"true"`
	GetConnectedModules func() (map[string]interface{}, error)                    `method:"get_connected_modules" group:"module" inner:"true"`
	CreateUpdateModule  func(module entity.Module) (*entity.Module, error)        `method:"create_update_module" group:"module" inner:"true"`
	DeleteModule        func(identities []int32) (*domain.DeleteResponse, error)  `method:"delete_module" group:"module" inner:"true"`
	GetModulesInfo      func(i structure.Isolation) ([]*domain.ModuleInfo, error) `method:"get_modules_info" group:"module" inner:"true"`

	// ===== CONFIG =====
	GetConfigs                                    func(identities []int64) ([]entity.Config, error)                    `method:"get_configs" group:"config" inner:"true"`
	GetConfigByInstanceUUIDAndModuleName          func(request entity.ModuleInstanceIdentity) (*entity.Config, error)  `method:"get_config_by_instance_uuid_and_module_name" group:"config" inner:"true"`
	CreateUpdateConfig                            func(config entity.Config) (*entity.Config, error)                   `method:"create_update_config" group:"config" inner:"true"`
	UpdateActiveConfigByInstanceUUIDAndModuleName func(config domain.ConfigInstanceModuleName) (*entity.Config, error) `method:"update_active_config_by_instance_uuid_and_module_name" group:"config" inner:"true"`
	MarkConfigAsActive                            func(identity domain.LongIdentitiesRequest) (*entity.Config, error)  `method:"mark_config_as_active" group:"config" inner:"true"`
	DeleteConfig                                  func(identities []int64) (*domain.DeleteResponse, error)             `method:"delete_config" group:"config" inner:"true"`

	// ===== ROUTING =====
	GetRoutes func() (map[string]interface{}, error) `method:"get_routes" group:"routing" inner:"true"`
}

func GetHandlers() *Handlers {
	return &Handlers{
		GetInstances:         controllers.GetInstances,
		CreateUpdateInstance: controllers.CreateUpdateInstance,
		DeleteInstance:       controllers.DeleteInstance,

		GetModules:          controllers.GetModules,
		GetSchemaByModuleId: controllers.GetSchemaByModuleId,
		GetActiveModules:    controllers.GetActiveModules,
		GetConnectedModules: controllers.GetConnectedModules,
		CreateUpdateModule:  controllers.CreateUpdateModule,
		DeleteModule:        controllers.DeleteModule,
		GetModulesInfo:      controllers.GetModulesAggregatedInfo,

		GetConfigs:                                    controllers.GetConfigs,
		GetConfigByInstanceUUIDAndModuleName:          controllers.GetConfigByInstanceUUIDAndModuleName,
		CreateUpdateConfig:                            controllers.CreateUpdateConfig,
		UpdateActiveConfigByInstanceUUIDAndModuleName: controllers.UpdateActiveConfigByInstanceUUIDAndModuleName,
		MarkConfigAsActive:                            controllers.MarkConfigAsActive,
		DeleteConfig:                                  controllers.DeleteConfig,

		GetRoutes: controllers.GetRoutes,
	}
}
