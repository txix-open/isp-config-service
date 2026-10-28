package tests_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/test/fake"
	"isp-config-service/domain"
)

const defaultRemoteConfigsWithVariables = `
{
   "varA": "_Var(varA)",
   "varB": "_ToInt(_Var(varB))",
   "varC": "{{ msp_pgsql_name }}"
}
`

type configWithVars struct {
	VarA string
	VarB int
	VarC string
}

// nolint:funlen,paralleltest
func TestVariableAcceptance(t *testing.T) {
	require := require.New(t)
	logger := setupTest(t)

	apiCli, err := client.Default()
	require.NoError(err)
	apiCli.Upgrade([]string{"127.0.0.1:9002"})
	varAValue := fake.It[string]()
	varA := domain.CreateVariableRequest{
		Name:        "varA",
		Description: fake.It[string](),
		Type:        "TEXT",
		Value:       varAValue,
	}
	err = apiCli.Invoke("config/variable/create").
		JsonRequestBody(varA).
		Do(t.Context())
	require.NoError(err)
	err = apiCli.Invoke("config/variable/create").
		JsonRequestBody(varA).
		Do(t.Context())
	require.Error(err)

	varBValue := fake.It[int]()
	err = apiCli.Invoke("config/variable/upsert").
		JsonRequestBody([]domain.UpsertVariableRequest{{
			Name:        "varB",
			Description: "test",
			Type:        "SECRET",
			Value:       strconv.Itoa(varBValue),
		}}).
		Do(t.Context())
	require.NoError(err)

	eventHandler := newClusterEventHandler()
	clientA1 := newClusterClientWith(
		t,
		"A",
		"10.2.9.1",
		configWithVars{},
		[]byte(defaultRemoteConfigsWithVariables),
		nil,
		logger,
	)
	go func() {
		handler := cluster.NewEventHandler().RemoteConfigReceiver(eventHandler)
		err := clientA1.Run(t.Context(), handler)
		require.NoError(err) //nolint:testifylint
	}()
	time.Sleep(3 * time.Second)

	firstConfig := unmarshalConfig(require, eventHandler.ReceivedConfigs()[0])
	expected := configWithVars{
		VarA: varAValue,
		VarB: varBValue,
		VarC: "{{ msp_pgsql_name }}",
	}
	require.EqualValues(expected, firstConfig)

	err = apiCli.Invoke("config/variable/delete").
		JsonRequestBody(domain.VariableByNameRequest{Name: "varB"}).
		Do(t.Context())
	require.Error(err)

	varBValue = fake.It[int]()
	err = apiCli.Invoke("config/variable/upsert").
		JsonRequestBody([]domain.UpsertVariableRequest{{
			Name:        "varB",
			Description: "test",
			Type:        "SECRET",
			Value:       strconv.Itoa(varBValue),
		}}).
		Do(t.Context())
	require.NoError(err)
	time.Sleep(3 * time.Second)

	secondConfig := unmarshalConfig(require, eventHandler.ReceivedConfigs()[1])
	expected = configWithVars{
		VarA: varAValue,
		VarB: varBValue,
		VarC: "{{ msp_pgsql_name }}",
	}
	require.EqualValues(expected, secondConfig)

	allResponse := []domain.Variable{}
	err = apiCli.Invoke("config/variable/all").
		JsonResponseBody(&allResponse).
		Do(t.Context())
	require.NoError(err)

	require.Len(allResponse, 2)
	require.EqualValues(varAValue, allResponse[0].Value)
	require.EqualValues("TEXT", allResponse[0].Type)
	require.Empty(allResponse[1].Value) // secret
}

//nolint:funlen,paralleltest
func TestVariableLinksAfterConfigUpdate(t *testing.T) {
	require := require.New(t)
	logger := setupTest(t)

	apiCli, err := client.Default()
	require.NoError(err)
	apiCli.Upgrade([]string{"127.0.0.1:9002"})

	err = apiCli.Invoke("config/variable/create").
		JsonRequestBody(domain.CreateVariableRequest{
			Name:        "linkVar",
			Description: "test",
			Type:        "TEXT",
			Value:       fake.It[string](),
		}).
		Do(t.Context())
	require.NoError(err)

	type linkTestConfig struct {
		Addr  string
		Alias string
	}
	clientC := newClusterClientWith(
		t,
		"LinkTestModule",
		"10.2.9.77",
		linkTestConfig{},
		[]byte(`{"addr":"init","alias":"init"}`),
		nil,
		logger,
	)
	go func() {
		handler := cluster.NewEventHandler()
		err := clientC.Run(t.Context(), handler)
		require.NoError(err) //nolint:testifylint
	}()
	time.Sleep(2 * time.Second)

	activeConfig := domain.Config{}
	err = apiCli.Invoke("config/config/get_active_config_by_module_name").
		JsonRequestBody(domain.GetByModuleNameRequest{ModuleName: "LinkTestModule"}).
		JsonResponseBody(&activeConfig).
		Do(t.Context())
	require.NoError(err)

	// The same variable is referenced twice in the config data.
	err = apiCli.Invoke("config/config/create_update_config").
		JsonRequestBody(domain.CreateUpdateConfigRequest{
			Id:       activeConfig.Id,
			Name:     activeConfig.Name,
			ModuleId: activeConfig.ModuleId,
			Version:  activeConfig.Version,
			Data:     []byte(`{"addr":"_Var(linkVar)","alias":"prefix/_Var(linkVar)"}`),
		}).
		Do(t.Context())
	require.NoError(err)

	varResponse := []domain.Variable{}
	err = apiCli.Invoke("config/variable/all").
		JsonResponseBody(&varResponse).
		Do(t.Context())
	require.NoError(err)
	require.Len(varResponse, 1)
	require.Equal("linkVar", varResponse[0].Name)
	require.Len(varResponse[0].ContainsInConfigs, 1)
	require.Equal(activeConfig.Id, varResponse[0].ContainsInConfigs[0].Id)

	// Re-saving identical config data must keep the links.
	err = apiCli.Invoke("config/config/create_update_config").
		JsonRequestBody(domain.CreateUpdateConfigRequest{
			Id:       activeConfig.Id,
			Name:     activeConfig.Name,
			ModuleId: activeConfig.ModuleId,
			Version:  activeConfig.Version,
			Data:     []byte(`{"alias":"prefix/_Var(linkVar)","addr":"_Var(linkVar)"}`),
		}).
		Do(t.Context())
	require.NoError(err)

	varResponse = []domain.Variable{}
	err = apiCli.Invoke("config/variable/all").
		JsonResponseBody(&varResponse).
		Do(t.Context())
	require.NoError(err)
	require.Len(varResponse[0].ContainsInConfigs, 1)
	require.Equal(activeConfig.Id, varResponse[0].ContainsInConfigs[0].Id)
}

func unmarshalConfig(require *require.Assertions, configData []byte) configWithVars {
	firstConfig := configWithVars{}
	err := json.Unmarshal(configData, &firstConfig)
	require.NoError(err)
	return firstConfig
}
