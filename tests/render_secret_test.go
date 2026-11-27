package tests_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"isp-config-service/domain"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/metrics"
)

// renderPolicyConfig has one ui-exposed leaf and one plain leaf.
type renderPolicyConfig struct {
	Exposed string `schemaGen:"uiExposed"`
	Plain   string
}

//nolint:funlen,paralleltest,noctx
func TestRenderSecretSkip(t *testing.T) {
	require := require.New(t)
	logger := setupTest(t)

	apiCli, err := client.Default()
	require.NoError(err)
	apiCli.Upgrade([]string{"127.0.0.1:9002"})

	err = apiCli.Invoke("config/variable/upsert").
		JsonRequestBody([]domain.UpsertVariableRequest{{
			Name:        "rdVar",
			Description: "test",
			Type:        "TEXT",
			Value:       "render-visible",
		}}).
		Do(t.Context())
	require.NoError(err)

	eventHandler := newClusterEventHandler()
	renderCli := newClusterClientWith(
		t,
		"RenderPolicy",
		"10.2.9.41",
		renderPolicyConfig{},
		[]byte(`{"exposed":"_Var(rdVar)","plain":"_Var(rdVar)"}`),
		nil,
		logger,
	)
	go func() {
		handler := cluster.NewEventHandler().RemoteConfigReceiver(eventHandler)
		err := renderCli.Run(t.Context(), handler)
		require.NoError(err) //nolint:testifylint
	}()
	time.Sleep(3 * time.Second)

	active := domain.Config{}
	err = apiCli.Invoke("config/config/get_active_config_by_module_name").
		JsonRequestBody(domain.GetByModuleNameRequest{ModuleName: "RenderPolicy"}).
		JsonResponseBody(&active).
		Do(t.Context())
	require.NoError(err)
	require.NotEmpty(active.ModuleId)

	first := renderPolicyConfig{}
	err = json.Unmarshal(eventHandler.ReceivedConfigs()[0], &first)
	require.NoError(err)
	require.Equal("render-visible", first.Exposed)
	require.Equal("render-visible", first.Plain)

	// Legacy regime: the variable referenced by the stored config becomes SECRET.
	err = apiCli.Invoke("config/variable/upsert").
		JsonRequestBody([]domain.UpsertVariableRequest{{
			Name:        "rdVar",
			Description: "test",
			Type:        "SECRET",
			Value:       "render-secret-value",
		}}).
		Do(t.Context())
	require.NoError(err)
	time.Sleep(3 * time.Second)

	last := renderPolicyConfig{}
	err = json.Unmarshal(eventHandler.ReceivedConfigs()[len(eventHandler.ReceivedConfigs())-1], &last)
	require.NoError(err)
	require.Equal("_Var(rdVar)", last.Exposed, "the secret reference must stay a literal in the ui-exposed leaf")
	require.Equal("render-secret-value", last.Plain)

	// The skip is observable: a positive counter for the module and the variable.
	recorder := httptest.NewRecorder()
	metrics.DefaultRegistry.MetricsHandler().ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/internal/metrics", nil))
	require.Regexp(
		fmt.Sprintf(`render_secret_skipped[^\n]*module=%q[^\n]*variable="rdVar"[^\n]* [1-9]`, active.ModuleId),
		recorder.Body.String(),
	)
}
