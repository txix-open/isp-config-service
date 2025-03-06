package tests_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/test/fake"
	"isp-config-service/domain"
	"strconv"
	"testing"
	"time"
)

const defaultRemoteConfigsWithVariables = `
{
   "varA": "_Var({{ .varA }})",
   "varB": "_ToInt(_Var({{ .varB }}))"
}
`

type configWithVars struct {
	VarA string
	VarB int
}

// nolint:funlen
func TestVariableAcceptance(t *testing.T) {
	require := require.New(t)
	logger := setupTest(t)

	apiCli, err := client.Default()
	require.NoError(err)
	apiCli.Upgrade([]string{"127.0.0.1:9002"})
	varAValue := fake.It[string]()
	err = apiCli.Invoke("config/variable/create").
		JsonRequestBody(domain.CreateVariableRequest{
			Name:        "varA",
			Description: fake.It[string](),
			Type:        "TEXT",
			Value:       varAValue,
		}).Do(context.Background())
	require.NoError(err)

	varBValue := fake.It[int]()
	err = apiCli.Invoke("config/variable/upsert").
		JsonRequestBody([]domain.UpsertVariableRequest{{
			Name:        "varB",
			Description: "test",
			Type:        "SECRET",
			Value:       strconv.Itoa(varBValue),
		}}).
		Do(context.Background())
	require.NoError(err)

	eventHandler := newClusterEventHandler()
	clientA1 := newClusterClientWith(
		"A",
		"10.2.9.1",
		configWithVars{},
		[]byte(defaultRemoteConfigsWithVariables),
		logger,
	)
	go func() {
		handler := cluster.NewEventHandler().RemoteConfigReceiver(eventHandler)
		err := clientA1.Run(context.Background(), handler)
		require.NoError(err) //nolint:testifylint
	}()
	time.Sleep(3 * time.Second)

	firstConfig := unmarshalConfig(require, eventHandler.ReceivedConfigs()[0])
	expected := configWithVars{
		VarA: varAValue,
		VarB: varBValue,
	}
	require.EqualValues(expected, firstConfig)

	err = apiCli.Invoke("config/variable/delete").
		JsonRequestBody(domain.VariableByNameRequest{Name: "varB"}).
		Do(context.Background())
	require.Error(err)

	varBValue = fake.It[int]()
	err = apiCli.Invoke("config/variable/upsert").
		JsonRequestBody([]domain.UpsertVariableRequest{{
			Name:        "varB",
			Description: "test",
			Type:        "SECRET",
			Value:       strconv.Itoa(varBValue),
		}}).
		Do(context.Background())
	require.NoError(err)
	time.Sleep(3 * time.Second)

	secondConfig := unmarshalConfig(require, eventHandler.ReceivedConfigs()[1])
	expected = configWithVars{
		VarA: varAValue,
		VarB: varBValue,
	}
	require.EqualValues(expected, secondConfig)

	allResponse := []domain.Variable{}
	err = apiCli.Invoke("config/variable/all").
		JsonResponseBody(&allResponse).
		Do(context.Background())
	require.NoError(err)

	require.Len(allResponse, 2)
	require.EqualValues(varAValue, allResponse[0].Value)
	require.EqualValues("TEXT", allResponse[0].Type)
	require.Empty(allResponse[1].Value)
}

func unmarshalConfig(require *require.Assertions, configData []byte) configWithVars {
	firstConfig := configWithVars{}
	err := json.Unmarshal(configData, &firstConfig)
	require.NoError(err)
	return firstConfig
}
