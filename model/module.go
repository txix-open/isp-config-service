package model

import (
	"isp-config-service/entity"

	"github.com/go-pg/pg"

	"github.com/go-pg/pg/orm"
	"time"
)

type ModulesRepository struct {
	DB orm.DB
}

func (r *ModulesRepository) GetModules(identities []int32) ([]entity.Module, error) {
	var modules []entity.Module
	var err error
	if len(identities) > 0 {
		err = r.DB.Model(&modules).
			Where("id IN (?)", pg.In(identities)).
			Order("created_at DESC").
			Limit(500).
			Select()
	} else {
		err = r.DB.Model(&modules).
			Order("created_at DESC").
			Limit(500).
			Select()
	}
	return modules, err
}

func (r *ModulesRepository) GetModule(identity int32) (*entity.Module, error) {
	var module entity.Module
	var err = r.DB.Model(&module).Where("id = ?", identity).Select()
	return &module, err
}

func (r *ModulesRepository) GetActiveModules() ([]entity.Module, error) {
	var modules []entity.Module
	err := r.DB.Model(&modules).
		Order("created_at DESC").
		Where("active = true").
		Limit(500).
		Select()
	return modules, err
}

func (r *ModulesRepository) CreateModule(module entity.Module) (entity.Module, error) {
	_, err := r.DB.Model(&module).
		Returning("id").
		Returning("active").
		Returning("created_at").
		Insert()
	return module, err
}

func (r *ModulesRepository) UpdateModule(module entity.Module) (entity.Module, error) {
	_, err := r.DB.Model(&module).WherePK().
		Returning("id").
		Returning("active").
		Returning("created_at").
		Update()
	return module, err
}

func (r *ModulesRepository) DeleteModule(identities []int32) (int, error) {
	result, err := r.DB.Exec("DELETE FROM "+schema+".modules WHERE id IN (?)", pg.In(identities))
	return result.RowsAffected(), err
}

func (r *ModulesRepository) GetModuleById(identity int32) (entity.Module, error) {
	var module entity.Module
	err := r.DB.Model(&module).
		Where("id = ?", identity).
		First()
	if err != nil && err == pg.ErrNoRows {
		return module, nil
	}
	return module, err
}

func (r *ModulesRepository) GetModuleByInstanceIdAndName(instanceId int32, moduleName string) (entity.Module, error) {
	var module entity.Module
	err := r.DB.Model(&module).
		Where("name = ?", moduleName).
		Where("instance_id = ?", instanceId).
		First()
	if err == pg.ErrNoRows {
		return module, nil
	}
	return module, err
}

func (r *ModulesRepository) UpdateModuleDisconnect(instanceUuid string, moduleName string) error {
	var module entity.Module

	err := r.DB.Model(&module).
		Where("name = ?", moduleName).
		Where("instance_id = (SELECT id FROM "+schema+".instances WHERE uuid = ?)", instanceUuid).
		First()
	if err != nil {
		return err
	}
	if module.Id != 0 {
		module.LastDisconnectedAt = time.Now()
		_, err := r.DB.Model(&module).WherePK().Update()
		if err != nil {
			return err
		}
	}

	return err
}

func (r *ModulesRepository) GetModulesByInstanceUuid(instanceUuid string) ([]entity.Module, error) {
	var res []entity.Module
	err := r.DB.Model(&res).
		Where("instance_id = (SELECT id FROM "+schema+".instances WHERE uuid = ?)", instanceUuid).
		Select()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ModulesRepository) GetModulesByInstanceUuidAndName(instanceUuid string, name string) (*entity.Module, error) {
	res := new(entity.Module)
	err := r.DB.Model(res).
		Where("instance_id = (SELECT id FROM "+schema+".instances WHERE uuid = ?)", instanceUuid).
		Where("name = ?", name).
		First()
	if err == pg.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}
