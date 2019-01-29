package model

import (
	"isp-config-service/entity"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type InstanceRepository struct {
	DB orm.DB
}

func (r *InstanceRepository) GetInstances(identities []int32) ([]entity.Instance, error) {
	var instances []entity.Instance
	var err error
	if len(identities) > 0 {
		err = r.DB.Model(&instances).
			Where("id IN (?)", pg.In(identities)).
			Order("created_at DESC").
			Limit(500).
			Select()
	} else {
		err = r.DB.Model(&instances).
			Order("created_at DESC").
			Limit(500).
			Select()
	}
	return instances, err
}

func (r *InstanceRepository) CreateInstance(instance entity.Instance) (entity.Instance, error) {
	_, err := r.DB.Model(&instance).
		Returning("id").
		Returning("created_at").
		Insert()
	return instance, err
}

func (r *InstanceRepository) UpdateInstance(instance entity.Instance) (entity.Instance, error) {
	_, err := r.DB.Model(&instance).WherePK().
		Returning("id").
		Returning("created_at").
		Update()
	return instance, err
}

func (r *InstanceRepository) DeleteInstance(identities []int32) (int, error) {
	result, err := r.DB.Exec("DELETE FROM "+schema+".instances WHERE id IN (?)", pg.In(identities))
	return result.RowsAffected(), err
}

func (r *InstanceRepository) GetInstanceById(identity int32) (entity.Instance, error) {
	var module entity.Instance
	err := r.DB.Model(&module).
		Where("id = ?", identity).
		First()
	if err != nil && err == pg.ErrNoRows {
		return module, nil
	}
	return module, err
}

func (r *InstanceRepository) GetInstanceByUuid(instanceUuid string) (entity.Instance, error) {
	var instance entity.Instance
	err := r.DB.Model(&instance).
		Where("uuid = ?", instanceUuid).
		First()
	if err != nil && err == pg.ErrNoRows {
		return instance, nil
	}
	return instance, err
}
