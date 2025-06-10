package controller

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/http/apierrors"
	"isp-config-service/domain"
)

type metricsService interface {
	Autodiscovery(ctx context.Context) (domain.AutodiscoveryResponse, error)
}

type Metrics struct {
	svc metricsService
}

func NewMetrics(svc metricsService) Metrics {
	return Metrics{svc: svc}
}

// Autodiscovery
// @Summary Метод получения конфигурации для HTTP-based service discovery механизма Prometheus
// @Tags Метрики
// @Accept json
// @Produce json
// @Success 200 {object} domain.AutodiscoveryResponse
// @Failure 500 {object} apierrors.Error
// @Router /internal/metrics/autodiscovery [POST]
func (c Metrics) Autodiscovery(ctx context.Context) (domain.AutodiscoveryResponse, error) {
	resp, err := c.svc.Autodiscovery(ctx)
	if err != nil {
		return nil, apierrors.NewInternalServiceError(errors.WithMessage(err, "autodiscovery"))
	}
	return resp, nil
}
