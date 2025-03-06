package tests_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc/client"
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

func TestVariableAcceptance(t *testing.T) {
	require := require.New(t)
	logger := setupTest(t)

	apiCli, err := client.Default()
	require.NoError(err)
	apiCli.Upgrade([]string{"127.0.0.1:9002"})
	err = apiCli.Invoke("config/variable/create").
		JsonRequestBody(domain.CreateVariableRequest{
			Name:        "varA",
			Description: fake.It[string](),
			Type:        "TEXT",
			Value:       fake.It[string](),
		}).Do(context.Background())
	require.NoError(err)

	err = apiCli.Invoke("config/variable/upsert").
		JsonRequestBody([]domain.UpsertVariableRequest{{
			Name:        "varB",
			Description: "test",
			Type:        "SECRET",
			Value:       strconv.Itoa(fake.It[int]()),
		}}).
		Do(context.Background())
	require.NoError(err)

	clientA1 := newClusterClientWith(
		"A",
		"10.2.9.1",
		configWithVars{},
		[]byte(defaultRemoteConfigsWithVariables),
		logger)
	go func() {
		handler := cluster.NewEventHandler()
		err := clientA1.Run(context.Background(), handler)
		require.NoError(err) //nolint:testifylint
	}()
	time.Sleep(3 * time.Second)

	time.Sleep(5 * time.Second)
}
