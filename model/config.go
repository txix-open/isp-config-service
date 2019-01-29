package model

import (
	"isp-config-service/entity"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type ConfigRepository struct {
	DB orm.DB
}

func (r *ConfigRepository) GetConfigs(identities []int64) ([]entity.Config, error) {
	var configs []entity.Config
	var err error
	if len(identities) > 0 {
		err = r.DB.Model(&configs).
			Where("id IN (?)", pg.In(identities)).
			Order("created_at DESC").
			Limit(500).
			Select()
	} else {
		err = r.DB.Model(&configs).
			Order("created_at DESC").
			Limit(500).
			Select()
	}
	return configs, err
}

func (r *ConfigRepository) GetConfigsByModulesId(modulesId []int32) ([]entity.Config, error) {
	var res []entity.Config
	err := r.DB.Model(&res).Where("module_id IN (?)", pg.In(modulesId)).Order("created_at DESC").Select()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ConfigRepository) CreateConfig(config *entity.Config) (*entity.Config, error) {
	_, err := r.DB.Model(config).
		Returning("*").
		Insert()
	return config, err
}

func (r *ConfigRepository) UpdateConfig(config *entity.Config) (*entity.Config, error) {
	_, err := r.DB.Model(config).WherePK().
		Returning("*").
		Update()
	return config, err
}

func (r *ConfigRepository) DeleteConfig(identities []int64) (int, error) {
	result, err := r.DB.Exec("DELETE FROM "+schema+".configs WHERE id IN (?)", pg.In(identities))
	return result.RowsAffected(), err
}

func (r *ConfigRepository) GetConfigById(identity int64) (*entity.Config, error) {
	cfg := new(entity.Config)
	err := r.DB.Model(cfg).
		Where("id = ?", identity).
		First()
	if err == pg.ErrNoRows {
		return nil, nil
	}
	return cfg, err
}

func (r *ConfigRepository) GetConfigByInstanceUUIDAndModuleName(instanceUUID string, moduleName string) (*entity.Config, error) {
	cfg := new(entity.Config)
	err := r.DB.Model(cfg).
		Where(`module_id = (SELECT id FROM `+schema+`.modules WHERE name = ?
			AND instance_id = (SELECT id FROM `+schema+`.instances WHERE uuid = ?))`, moduleName, instanceUUID).
		Where("active = true").
		Order("updated_at DESC").
		First()
	if err == pg.ErrNoRows {
		return nil, nil
	}
	return cfg, err
}

func (r *ConfigRepository) GetModuleNameAndInstanceUUIDByConfigId(configId int64) (entity.ModuleInstanceIdentity, error) {
	var cfg entity.ModuleInstanceIdentity
	_, err := r.DB.Query(&cfg, `SELECT temp.name, uuid
		FROM (SELECT name, instance_id FROM `+schema+`.modules WHERE id =
			(SELECT module_id FROM `+schema+`.configs WHERE id = ? AND active = true)) temp
		LEFT JOIN `+schema+`.instances s ON s.id = temp.instance_id`, configId)

	if err != nil && err == pg.ErrNoRows {
		return cfg, nil
	}
	return cfg, err
}
