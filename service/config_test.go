package service

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"isp-config-service/entity"
	"testing"
)

func TestConfigService_validateSchema(t *testing.T) {
	var (
		a = assert.New(t)

		schemaDesc   = []byte(`{"schema":{"properties":{"configuration":{"required":["address","token"],"properties":{"address":{"type":"string","title":"Адрес"},"token":{"type":"string","title":"Токен"}},"additionalProperties":false,"type":"object"}},"additionalProperties":false,"type":"object","title":"Remote config"},"title":"module_name"}`)
		validData    = map[string]interface{}{"configuration": map[string]interface{}{"address": "127.0.0.1", "token": "token"}}
		invalidValue = map[string]interface{}{"configuration": map[string]interface{}{"address": 127001, "token": "token"}}
		invalidName  = map[string]interface{}{"config": map[string]interface{}{"test": "127.0.0.1", "token": "token"}}

		schema entity.ConfigSchema
		valid  bool
		err    error
	)

	a.NoError(json.Unmarshal(schemaDesc, &schema))

	valid, err = ConfigService.validateSchema(schema, validData)
	a.NoError(err)
	a.True(valid)

	valid, err = ConfigService.validateSchema(schema, invalidValue)
	a.Error(err)
	a.False(valid)

	valid, err = ConfigService.validateSchema(schema, invalidName)
	a.Error(err)
	a.False(valid)
}
