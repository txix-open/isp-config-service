package state

import (
	"fmt"
	"testing"

	"github.com/integration-system/isp-lib/v2/config"
	"github.com/stretchr/testify/assert"
	"isp-config-service/conf"
	"isp-config-service/entity"
)

func TestVersionConfigStore(t *testing.T) {
	a := assert.New(t)

	type example struct {
		cfg       entity.VersionConfig
		removedId string
	}

	config.UnsafeSet(&conf.Configuration{VersionConfigCount: 5})
	vcs := NewVersionConfigStore()
	for _, e := range []example{
		{cfg: entity.VersionConfig{Id: "1", ConfigId: "1"}, removedId: ""},
		{cfg: entity.VersionConfig{Id: "2", ConfigId: "1"}, removedId: ""},
		{cfg: entity.VersionConfig{Id: "3", ConfigId: "1"}, removedId: ""},
		{cfg: entity.VersionConfig{Id: "4", ConfigId: "1"}, removedId: ""},
		{cfg: entity.VersionConfig{Id: "5", ConfigId: "1"}, removedId: ""},
		{cfg: entity.VersionConfig{Id: "6", ConfigId: "1"}, removedId: "1"},
		{cfg: entity.VersionConfig{Id: "7", ConfigId: "1"}, removedId: "2"},
	} {
		removedId := vcs.Update(e.cfg)
		a.Equal(e.removedId, removedId)
	}
	for i, versionConfig := range vcs.GetByConfigId("1") {
		a.Equal(fmt.Sprintf("%d", 7-i), versionConfig.Id)
	}
}
