package helper

import (
	"github.com/integration-system/isp-lib/structure"
	"isp-config-service/controller"
	"isp-config-service/domain"
	"isp-config-service/entity"
)

type Handlers struct {
	// ===== MODULE =====
	DeleteModules  func(identities []string) (*domain.DeleteResponse, error) `method:"delete_module" group:"module" inner:"true"`
	GetModulesInfo func() ([]domain.ModuleInfo, error)                       `method:"get_modules_info" group:"module" inner:"true"`

	// ===== CONFIG =====
	GetActiveConfigByModuleName func(request domain.GetByModuleNameRequest) (*entity.Config, error) `method:"get_active_config_by_module_name" group:"config" inner:"true"`
	CreateUpdateConfig          func(config entity.Config) (*entity.Config, error)                  `method:"create_update_config" group:"config" inner:"true"`
	MarkConfigAsActive          func(identity domain.ConfigIdRequest) (*entity.Config, error)       `method:"mark_config_as_active" group:"config" inner:"true"`
	DeleteConfig                func(identities []string) (*domain.DeleteResponse, error)           `method:"delete_config" group:"config" inner:"true"`

	// ===== COMMON CONFIG =====
	GetCommonConfigs         func(identities []string) []entity.CommonConfig                              `method:"get_configs" group:"common_config" inner:"true"`
	CreateUpdateCommonConfig func(config entity.CommonConfig) (*entity.CommonConfig, error)               `method:"create_update_config" group:"common_config" inner:"true"`
	DeleteCommonConfig       func(req domain.ConfigIdRequest) (*domain.DeleteCommonConfigResponse, error) `method:"delete_config" group:"common_config" inner:"true"`
	CompileConfigs           func(req domain.CompileConfigsRequest) controller.CompiledConfig             `method:"compile" group:"common_config" inner:"true"`
	GetLinks                 func(req domain.ConfigIdRequest) domain.CommonConfigLinks                    `method:"get_links" group:"common_config" inner:"true"`

	// ===== ROUTING =====
	GetRoutes func() ([]structure.BackendDeclaration, error) `method:"get_routes" group:"routing" inner:"true"`

	// ===== SCHEMA =====
	GetSchemaByModuleId func(domain.GetByModuleIdRequest) (*entity.ConfigSchema, error) `method:"get_by_module_id" group:"schema" inner:"true"`
}

func GetHandlers() *Handlers {
	return &Handlers{
		DeleteModules:  controller.Module.DeleteModules,
		GetModulesInfo: controller.Module.GetModulesAggregatedInfo,

		GetActiveConfigByModuleName: controller.Config.GetActiveConfigByModuleName,
		CreateUpdateConfig:          controller.Config.CreateUpdateConfig,
		MarkConfigAsActive:          controller.Config.MarkConfigAsActive,
		DeleteConfig:                controller.Config.DeleteConfigs,

		GetCommonConfigs:         controller.CommonConfig.GetConfigs,
		CreateUpdateCommonConfig: controller.CommonConfig.CreateUpdateConfig,
		DeleteCommonConfig:       controller.CommonConfig.DeleteConfigs,
		CompileConfigs:           controller.CommonConfig.CompileConfigs,
		GetLinks:                 controller.CommonConfig.GetLinks,

		GetRoutes: controller.Routes.GetRoutes,

		GetSchemaByModuleId: controller.Schema.GetByModuleId,
	}
}
