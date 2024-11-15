package event

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"isp-config-service/entity"
)

type SubscriptionService interface {
	NotifyConfigChanged(ctx context.Context, moduleId string) error
	NotifyBackendsChanged(ctx context.Context, moduleId string) error
	NotifyRoutingChanged(ctx context.Context) error
}

type Handler struct {
	subscriptionService SubscriptionService
	logger              log.Logger
}

func NewHandler(
	subscriptionService SubscriptionService,
	logger log.Logger,
) Handler {
	return Handler{
		subscriptionService: subscriptionService,
		logger:              logger,
	}
}

func (h Handler) Handle(ctx context.Context, events []entity.Event) {
	routingChanged := false
	for _, event := range events {
		payload := event.Payload.Value
		routingChanged = routingChanged || payload.ModuleReady != nil || payload.ModuleDisconnected != nil
		err := h.handleEvent(ctx, event)
		if err != nil {
			h.logger.Error(ctx, errors.WithMessagef(err, "handle event: %s", event.Key()))
		}
	}

	if routingChanged {
		err := h.subscriptionService.NotifyRoutingChanged(ctx)
		if err != nil {
			h.logger.Error(ctx, errors.WithMessage(err, "handle routing changed"))
		}
	}
}

func (h Handler) handleEvent(ctx context.Context, event entity.Event) error {
	payload := event.Payload.Value
	switch {
	case payload.ConfigUpdated != nil:
		err := h.subscriptionService.NotifyConfigChanged(ctx, payload.ConfigUpdated.ModuleId)
		if err != nil {
			return errors.WithMessage(err, "handle config changed event")
		}
	case payload.ModuleReady != nil:
		err := h.subscriptionService.NotifyBackendsChanged(ctx, payload.ModuleReady.ModuleId)
		if err != nil {
			return errors.WithMessage(err, "handle module ready event")
		}
	case payload.ModuleDisconnected != nil:
		err := h.subscriptionService.NotifyBackendsChanged(ctx, payload.ModuleDisconnected.ModuleId)
		if err != nil {
			return errors.WithMessage(err, "handle module disconnected event")
		}
	default:
		return errors.Errorf("unknown event: %v", payload)
	}
	return nil
}
