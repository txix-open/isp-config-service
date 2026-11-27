package tests_test

import (
	"encoding/json"
	"testing"
	"time"

	"isp-config-service/domain"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/grpc/isp"
	"google.golang.org/grpc/status"
)

type secretPolicyOauthClient struct {
	ClientId      string
	LoginEndpoint string `schemaGen:"uiExposed"`
}

type secretPolicyOauth struct {
	Clients []secretPolicyOauthClient
}

type secretPolicyConfig struct {
	Secrets string `schemaGen:"uiExposed"`
	Plain   string
	Oauth   secretPolicyOauth
}

type businessError struct {
	ErrorCode    int            `json:"errorCode"`
	ErrorMessage string         `json:"errorMessage"`
	Details      map[string]any `json:"details"`
}

// requireBusinessError decodes the business error carried in the gRPC status
// details and asserts its error code.
func requireBusinessError(require *require.Assertions, err error, code int) businessError {
	require.Error(err)
	st, ok := status.FromError(err)
	require.True(ok, "expected a gRPC status error, got: %v", err)
	for _, detail := range st.Details() {
		if msg, ok := detail.(*isp.Message); ok {
			body, ok := msg.GetBody().(*isp.Message_BytesBody)
			require.True(ok, "expected a bytes body, got: %T", msg.GetBody())
			var be businessError
			require.NoError(json.Unmarshal(body.BytesBody, &be))
			require.Equal(code, be.ErrorCode, "unexpected error code: %s", body.BytesBody)
			return be
		}
	}
	require.Failf("business error not found in status details", "details: %+v", st.Details())
	return businessError{}
}

//nolint:funlen,paralleltest
func TestSecretPolicy(t *testing.T) {
	require := require.New(t)
	logger := setupTest(t)

	apiCli, err := client.Default()
	require.NoError(err)
	apiCli.Upgrade([]string{"127.0.0.1:9002"})

	// Variables: one SECRET, one TEXT.
	err = apiCli.Invoke("config/variable/upsert").
		JsonRequestBody([]domain.UpsertVariableRequest{
			{Name: "polSecret", Description: "test", Type: "SECRET", Value: "top-secret"},
			{Name: "polText", Description: "test", Type: "TEXT", Value: "visible"},
		}).
		Do(t.Context())
	require.NoError(err)

	// Module A: the default config references only the TEXT variable, so the
	// initial config is persisted on the first connection.
	eventHandler := newClusterEventHandler()
	clientA := newClusterClientWith(
		t,
		"SecretPolicyA",
		"10.2.9.31",
		secretPolicyConfig{},
		[]byte(`{"secrets":"_Var(polText)","plain":"p","oauth":{"clients":[{"clientId":"a","loginEndpoint":"_Var(polText)"}]}}`),
		nil,
		logger,
	)
	go func() {
		handler := cluster.NewEventHandler().RemoteConfigReceiver(eventHandler)
		err := clientA.Run(t.Context(), handler)
		require.NoError(err) //nolint:testifylint
	}()
	time.Sleep(3 * time.Second)

	active := domain.Config{}
	err = apiCli.Invoke("config/config/get_active_config_by_module_name").
		JsonRequestBody(domain.GetByModuleNameRequest{ModuleName: "SecretPolicyA"}).
		JsonResponseBody(&active).
		Do(t.Context())
	require.NoError(err)
	require.NotEmpty(active.Id)

	// 1) Update: a SECRET variable in a UI-exposed field is rejected with a
	// per-path detail, and the config is not changed.
	be := requireBusinessError(require,
		apiCli.Invoke("config/config/create_update_config").
			JsonRequestBody(domain.CreateUpdateConfigRequest{
				Id:       active.Id,
				Name:     active.Name,
				ModuleId: active.ModuleId,
				Version:  active.Version,
				Data:     []byte(`{"secrets":"_Var(polSecret)","plain":"p","oauth":{"clients":[{"clientId":"a","loginEndpoint":"_Var(polText)"}]}}`),
			}).
			Do(t.Context()),
		domain.ErrorCodeInvalidConfig)
	require.Equal("references SECRET variable polSecret", be.Details["secrets"])

	err = apiCli.Invoke("config/config/get_active_config_by_module_name").
		JsonRequestBody(domain.GetByModuleNameRequest{ModuleName: "SecretPolicyA"}).
		JsonResponseBody(&active).
		Do(t.Context())
	require.NoError(err)
	require.EqualValues(
		`{"secrets":"_Var(polText)","plain":"p","oauth":{"clients":[{"clientId":"a","loginEndpoint":"_Var(polText)"}]}}`,
		string(active.Data))

	// 2) unsafe=true does not bypass the policy.
	be = requireBusinessError(require,
		apiCli.Invoke("config/config/create_update_config").
			JsonRequestBody(domain.CreateUpdateConfigRequest{
				Id:       active.Id,
				Name:     active.Name,
				ModuleId: active.ModuleId,
				Version:  active.Version,
				Data:     []byte(`{"secrets":"_Var(polSecret)","plain":"p"}`),
				Unsafe:   true,
			}).
			Do(t.Context()),
		domain.ErrorCodeInvalidConfig)
	require.Equal("references SECRET variable polSecret", be.Details["secrets"])

	// 3) Insert path (empty Id): the new config is not created.
	be = requireBusinessError(require,
		apiCli.Invoke("config/config/create_update_config").
			JsonRequestBody(domain.CreateUpdateConfigRequest{
				Name:     "policy-insert",
				ModuleId: active.ModuleId,
				Data:     []byte(`{"secrets":"_Var(polSecret)"}`),
			}).
			Do(t.Context()),
		domain.ErrorCodeInvalidConfig)
	require.Equal("references SECRET variable polSecret", be.Details["secrets"])

	// 4) Nested array path: the detail key matches the gojsonschema field
	// format used by schema validation errors.
	be = requireBusinessError(require,
		apiCli.Invoke("config/config/create_update_config").
			JsonRequestBody(domain.CreateUpdateConfigRequest{
				Id:       active.Id,
				Name:     active.Name,
				ModuleId: active.ModuleId,
				Version:  active.Version,
				Data:     []byte(`{"secrets":"s","plain":"p","oauth":{"clients":[{"clientId":"a","loginEndpoint":"_Var(polSecret)"}]}}`),
			}).
			Do(t.Context()),
		domain.ErrorCodeInvalidConfig)
	require.Equal("references SECRET variable polSecret", be.Details["oauth.clients.0.loginEndpoint"])

	// 5) A SECRET variable outside a UI-exposed field is allowed.
	err = apiCli.Invoke("config/config/create_update_config").
		JsonRequestBody(domain.CreateUpdateConfigRequest{
			Id:       active.Id,
			Name:     active.Name,
			ModuleId: active.ModuleId,
			Version:  active.Version,
			Data:     []byte(`{"secrets":"s","plain":"_Var(polSecret)","oauth":{"clients":[{"clientId":"a","loginEndpoint":"x"}]}}`),
		}).
		Do(t.Context())
	require.NoError(err)

	// Restore the original data for the sync checks below.
	active = domain.Config{}
	err = apiCli.Invoke("config/config/get_active_config_by_module_name").
		JsonRequestBody(domain.GetByModuleNameRequest{ModuleName: "SecretPolicyA"}).
		JsonResponseBody(&active).
		Do(t.Context())
	require.NoError(err)
	err = apiCli.Invoke("config/config/create_update_config").
		JsonRequestBody(domain.CreateUpdateConfigRequest{
			Id:       active.Id,
			Name:     active.Name,
			ModuleId: active.ModuleId,
			Version:  active.Version,
			Data:     []byte(`{"secrets":"_Var(polText)","plain":"p","oauth":{"clients":[{"clientId":"a","loginEndpoint":"_Var(polText)"}]}}`),
		}).
		Do(t.Context())
	require.NoError(err)

	// 6) Sync with a compliant active config succeeds and re-pushes it.
	err = apiCli.Invoke("config/config/sync").
		JsonRequestBody(domain.SyncConfigRequest{ModuleName: "SecretPolicyA"}).
		Do(t.Context())
	require.NoError(err)
	time.Sleep(3 * time.Second)
	require.GreaterOrEqual(len(eventHandler.ReceivedConfigs()), 2)
	pushedBefore := len(eventHandler.ReceivedConfigs())

	// 7) The variable referenced by the stored config becomes SECRET (the
	// legacy regime): sync is rejected with per-path details and nothing is
	// re-pushed.
	err = apiCli.Invoke("config/variable/upsert").
		JsonRequestBody([]domain.UpsertVariableRequest{{
			Name: "polText", Description: "test", Type: "SECRET", Value: "now-secret",
		}}).
		Do(t.Context())
	require.NoError(err)
	time.Sleep(3 * time.Second)
	// The variable change itself triggers a re-render push.
	require.Greater(len(eventHandler.ReceivedConfigs()), pushedBefore)

	be = requireBusinessError(require,
		apiCli.Invoke("config/config/sync").
			JsonRequestBody(domain.SyncConfigRequest{ModuleName: "SecretPolicyA"}).
			Do(t.Context()),
		domain.ErrorCodeInvalidConfig)
	require.Equal("references SECRET variable polText", be.Details["secrets"])
	require.Equal("references SECRET variable polText", be.Details["oauth.clients.0.loginEndpoint"])

	time.Sleep(3 * time.Second)
	// No extra push: the rejected sync did not emit a change event.
	require.Equal(pushedBefore+1, len(eventHandler.ReceivedConfigs()))

	// 8) First connection of module B: the default config references a SECRET
	// variable in a UI-exposed field, so the initial config is not persisted.
	clientB := newClusterClientWith(
		t,
		"SecretPolicyB",
		"10.2.9.32",
		secretPolicyConfig{},
		[]byte(`{"secrets":"_Var(polSecret)","plain":"p","oauth":{"clients":[]}}`),
		nil,
		logger,
	)
	go func() {
		handler := cluster.NewEventHandler()
		err := clientB.Run(t.Context(), handler)
		_ = err // the connection is rejected by the write policy
	}()
	time.Sleep(3 * time.Second)

	requireBusinessError(require,
		apiCli.Invoke("config/config/get_active_config_by_module_name").
			JsonRequestBody(domain.GetByModuleNameRequest{ModuleName: "SecretPolicyB"}).
			Do(t.Context()),
		domain.ErrorCodeConfigNotFound)

	// 9) Unknown module id (no schema row at all): the write policy must not
	// panic on the missing schema — the request fails with the regular
	// schema-not-found error, not an internal one.
	be = requireBusinessError(require,
		apiCli.Invoke("config/config/create_update_config").
			JsonRequestBody(domain.CreateUpdateConfigRequest{
				Name:     "policy-no-schema",
				ModuleId: "00000000-0000-0000-0000-000000000000",
				Data:     []byte(`{"secrets":"_Var(polSecret)"}`),
			}).
			Do(t.Context()),
		domain.ErrorCodeSchemaNotFound)
	require.Empty(be.Details)

	// 10) mark_config_as_active: the active config of module A now violates
	// the policy (its referenced variable became SECRET after the config was
	// saved — the legacy regime). Activating a saved config must pass the
	// policy, so rolling back to the violating config is rejected with
	// per-path details and the active config is not switched.
	var marked domain.CreateUpdateConfigResponse
	err = apiCli.Invoke("config/config/create_update_config").
		JsonRequestBody(domain.CreateUpdateConfigRequest{
			Name:     "policy-mark",
			ModuleId: active.ModuleId,
			Data:     []byte(`{"secrets":"s","plain":"p","oauth":{"clients":[{"clientId":"a","loginEndpoint":"x"}]}}`),
		}).
		JsonResponseBody(&marked).
		Do(t.Context())
	require.NoError(err)
	require.NotEmpty(marked.Id)

	err = apiCli.Invoke("config/config/mark_config_as_active").
		JsonRequestBody(domain.IdRequest{Id: marked.Id}).
		Do(t.Context())
	require.NoError(err)

	be = requireBusinessError(require,
		apiCli.Invoke("config/config/mark_config_as_active").
			JsonRequestBody(domain.IdRequest{Id: active.Id}).
			Do(t.Context()),
		domain.ErrorCodeInvalidConfig)
	require.Equal("references SECRET variable polText", be.Details["secrets"])
	require.Equal("references SECRET variable polText", be.Details["oauth.clients.0.loginEndpoint"])

	fetched := domain.Config{}
	err = apiCli.Invoke("config/config/get_active_config_by_module_name").
		JsonRequestBody(domain.GetByModuleNameRequest{ModuleName: "SecretPolicyA"}).
		JsonResponseBody(&fetched).
		Do(t.Context())
	require.NoError(err)
	require.Equal(marked.Id, fetched.Id)
}
