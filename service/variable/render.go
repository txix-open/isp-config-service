package variable

import (
	"regexp"

	"isp-config-service/entity"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/rc"
)

var (
	variableRefPattern = regexp.MustCompile(`_Var\((.*?)\)`)
	stringToIntPattern = regexp.MustCompile(`"_ToInt\((\d+)\)"`)
)

// skippedRef names a UI-exposed path and a SECRET variable left unrendered there.
type skippedRef struct {
	Path     string
	Variable string
}

// Render substitutes _Var(name) references in the config's string leaves.
// On UI-exposed paths (uiPaths) references to SECRET variables are left as
// literals and reported in skipped; every other reference is substituted.
// Composite JSON string leaves are report-only (see rc.VariableRefReport):
// the secret policy does not apply to them, so every known reference is
// substituted. References to unknown variables are collected in undefined.
// The document is re-assembled into JSON and _ToInt is applied to the
// assembled bytes, as before; object key order is not preserved.
func Render(input []byte, uiPaths []string, vars map[string]entity.Variable) ([]byte, []skippedRef, []string, error) {
	uiSet := make(map[string]struct{}, len(uiPaths))
	for _, path := range uiPaths {
		uiSet[path] = struct{}{}
	}

	skipped := make([]skippedRef, 0)
	skippedSeen := make(map[string]struct{})
	undefined := make([]string, 0)

	substitute := func(path, value string, policyActive bool) string {
		return variableRefPattern.ReplaceAllStringFunc(value, func(match string) string {
			name := variableRefPattern.FindStringSubmatch(match)[1]
			variable, ok := vars[name]
			if !ok {
				undefined = append(undefined, name)
				return match
			}
			if policyActive && variable.Type == entity.SecretType {
				if _, isUI := uiSet[path]; isUI {
					key := path + "\x00" + name
					if _, dup := skippedSeen[key]; !dup {
						skippedSeen[key] = struct{}{}
						skipped = append(skipped, skippedRef{Path: path, Variable: name})
					}
					return match
				}
			}
			return variable.Value
		})
	}

	out, err := rc.TransformStringLeaves(input, func(path, value string) string {
		if rc.IsCompositeJSONString(value) {
			return substitute(path, value, false)
		}
		return substitute(path, value, true)
	})
	if err != nil {
		return nil, nil, nil, errors.WithMessage(err, "transform config leaves")
	}

	return stringToIntPattern.ReplaceAll(out, []byte("$1")), skipped, undefined, nil
}
