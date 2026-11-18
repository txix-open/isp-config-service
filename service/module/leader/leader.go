package leader

import (
	"slices"
	"strings"
	"time"

	"github.com/txix-open/isp-kit/cluster"
	"isp-config-service/entity"
)

type Result struct {
	HasElection bool
	Found       bool
	Leader      entity.Backend
}

func Derive(backends []entity.Backend) Result {
	result := Result{
		HasElection: hasElection(backends),
	}
	if !result.HasElection {
		return result
	}

	candidates := slices.Clone(backends)
	slices.SortFunc(candidates, compare)

	result.Found = true
	result.Leader = candidates[0]

	return result
}

func hasElection(backends []entity.Backend) bool {
	for _, backend := range backends {
		if backend.ModuleLeaderMode == cluster.ModuleLeaderModeRequestElection {
			return true
		}
	}

	return false
}

func compare(a, b entity.Backend) int {
	byCreatedAt := time.Time(a.CreatedAt).Compare(time.Time(b.CreatedAt))
	if byCreatedAt != 0 {
		return byCreatedAt
	}

	return strings.Compare(a.WsConnectionId, b.WsConnectionId)
}
