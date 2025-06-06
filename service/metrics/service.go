package metrics

import (
	"context"

	"github.com/pkg/errors"
	"isp-config-service/domain"
	"isp-config-service/entity"
)

type metricsAutodiscoveryRepo interface {
	AllMetricsAutodiscovery(ctx context.Context) ([]entity.MetricsAdWrapper, error)
}

type service struct {
	repo metricsAutodiscoveryRepo
}

func New(repo metricsAutodiscoveryRepo) service {
	return service{repo: repo}
}

func (s service) Autodiscovery(ctx context.Context) (domain.AutodiscoveryResponse, error) {
	metricsAdList, err := s.repo.AllMetricsAutodiscovery(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "all metrics autodiscovery")
	}

	result := make(domain.AutodiscoveryResponse, len(metricsAdList))
	for i, v := range metricsAdList {
		v := v.MetricsAutodiscovery
		result[i] = domain.PrometheusTargets{
			Targets: []string{v.Value.Address},
			Labels:  v.Value.Labels,
		}
	}

	return result, nil
}
