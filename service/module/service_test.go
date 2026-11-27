package module

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"isp-config-service/domain"
	"isp-config-service/entity"

	kitjson "github.com/txix-open/isp-kit/json"
)

type fakeSchemaRepo struct {
	schema *entity.ConfigSchema
}

func (r fakeSchemaRepo) Upsert(ctx context.Context, schema entity.ConfigSchema) error {
	return nil
}

func (r fakeSchemaRepo) GetByModuleId(ctx context.Context, moduleId string) (*entity.ConfigSchema, error) {
	return r.schema, nil
}

type fakeVariableRepo struct {
	variables []entity.Variable
}

func (r fakeVariableRepo) GetVariablesByName(ctx context.Context, names map[string]struct{}) ([]entity.Variable, error) {
	out := make([]entity.Variable, 0, len(names))
	for _, variable := range r.variables {
		if _, ok := names[variable.Name]; ok {
			out = append(out, variable)
		}
	}
	return out, nil
}

func policyTestSchema(t *testing.T) []byte {
	t.Helper()
	schemaData, err := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"secrets": map[string]any{"type": "string", "x-ui-exposed": true},
			"token":   map[string]any{"type": "string", "x-sensitive": true},
			"arr": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string", "x-sensitive": true},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	return schemaData
}

func TestFilterSecretParams(t *testing.T) {
	const moduleId = "m1"
	varRepo := fakeVariableRepo{variables: []entity.Variable{
		{Name: "sec_var", Type: entity.SecretType},
		{Name: "txt_var", Type: entity.TextType},
	}}

	tests := []struct {
		name        string
		config      string
		wantErr     bool
		wantDetails map[string]string
	}{
		{
			name:    "ui-exposed secret reference",
			config:  `{"secrets": "_Var(sec_var)"}`,
			wantErr: true,
			wantDetails: map[string]string{
				"secrets": "references SECRET variable sec_var",
			},
		},
		{
			name:        "sensitive literal without any references",
			config:      `{"token": "plain-pass"}`,
			wantErr:     true,
			wantDetails: map[string]string{"token": "plain value in sensitive field"},
		},
		{
			name:    "sensitive non-secret reference",
			config:  `{"token": "_Var(txt_var)"}`,
			wantErr: true,
			wantDetails: map[string]string{
				"token": "references non-SECRET variable txt_var",
			},
		},
		{
			name:   "sensitive secret reference",
			config: `{"token": "_Var(sec_var)"}`,
		},
		{
			name:   "sensitive empty value",
			config: `{"token": ""}`,
		},
		{
			name:   "sensitive field absent",
			config: `{"secrets": "_Var(txt_var)"}`,
		},
		{
			name:   "sensitive composite json string is report-only",
			config: `{"token": "{\"a\": 1}"}`,
		},
		{
			name:        "sensitive mixed value",
			config:      `{"token": "prefix _Var(sec_var)"}`,
			wantErr:     true,
			wantDetails: map[string]string{"token": "plain value in sensitive field"},
		},
		{
			name:   "sensitive unknown reference is left to the absent-variable check",
			config: `{"token": "_Var(ghost)"}`,
		},
		{
			name:    "sensitive array item literal",
			config:  `{"arr": ["plain-pass"]}`,
			wantErr: true,
			wantDetails: map[string]string{
				"arr.0": "plain value in sensitive field",
			},
		},
		{
			name:    "both violations at once",
			config:  `{"secrets": "_Var(sec_var)", "token": "plain-pass"}`,
			wantErr: true,
			wantDetails: map[string]string{
				"secrets": "references SECRET variable sec_var",
				"token":   "plain value in sensitive field",
			},
		},
		{
			name:   "compliant config",
			config: `{"secrets": "_Var(txt_var)", "token": "_Var(sec_var)"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := Service{
				configWriteDeny: true,
				configSchemaRepo: fakeSchemaRepo{
					schema: &entity.ConfigSchema{Id: "s1", ModuleId: moduleId, Data: policyTestSchema(t)},
				},
				variableRepo: varRepo,
			}

			err := s.FilterSecretParams(context.Background(), kitjson.RawMessage(test.config), moduleId)
			if !test.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			var validation domain.ConfigValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("error = %v, want ConfigValidationError", err)
			}
			if !reflect.DeepEqual(validation.Details, test.wantDetails) {
				t.Fatalf("details = %v, want %v", validation.Details, test.wantDetails)
			}
		})
	}
}

func TestFilterSecretParamsDisabled(t *testing.T) {
	const moduleId = "m1"
	s := Service{
		configWriteDeny: false,
		configSchemaRepo: fakeSchemaRepo{
			schema: &entity.ConfigSchema{Id: "s1", ModuleId: moduleId, Data: policyTestSchema(t)},
		},
		variableRepo: fakeVariableRepo{},
	}

	err := s.FilterSecretParams(context.Background(), kitjson.RawMessage(`{"secrets": "_Var(sec_var)", "token": "plain-pass"}`), moduleId)
	if err != nil {
		t.Fatalf("expected no error with the policy disabled, got: %v", err)
	}
}

func TestFilterSecretParamsWithoutSchema(t *testing.T) {
	s := Service{
		configWriteDeny:  true,
		configSchemaRepo: fakeSchemaRepo{schema: nil},
		variableRepo:     fakeVariableRepo{},
	}

	err := s.FilterSecretParams(context.Background(), kitjson.RawMessage(`{"token": "plain-pass"}`), "m1")
	if err != nil {
		t.Fatalf("expected no error without a schema, got: %v", err)
	}
}

func TestFilterSecretParamsSchemaWithoutPolicyPaths(t *testing.T) {
	schemaData, err := json.Marshal(map[string]any{
		"type":       "object",
		"properties": map[string]any{"plain": map[string]any{"type": "string"}},
	})
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	s := Service{
		configWriteDeny: true,
		configSchemaRepo: fakeSchemaRepo{
			schema: &entity.ConfigSchema{Id: "s1", ModuleId: "m1", Data: schemaData},
		},
		variableRepo: fakeVariableRepo{},
	}

	err = s.FilterSecretParams(context.Background(), kitjson.RawMessage(`{"plain": "any value"}`), "m1")
	if err != nil {
		t.Fatalf("expected no error for a schema without policy paths, got: %v", err)
	}
}
