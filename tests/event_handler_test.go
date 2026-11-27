package tests_test

import (
	"context"
	"fmt"
	"testing"

	"isp-config-service/entity"
	"isp-config-service/service/event"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/log"
)

type subscriptionSpy struct {
	backendsChanged []string
	routingCalls    int
}

func (s *subscriptionSpy) NotifyConfigChanged(ctx context.Context, moduleId string) error {
	return nil
}

func (s *subscriptionSpy) NotifyBackendsChanged(ctx context.Context, moduleId string) error {
	s.backendsChanged = append(s.backendsChanged, moduleId)
	return nil
}

func (s *subscriptionSpy) NotifyRoutingChanged(ctx context.Context) error {
	s.routingCalls++
	return nil
}

type eventLoggerSpy struct {
	errors []string
}

func (l *eventLoggerSpy) Error(ctx context.Context, message any, fields ...log.Field) {
	l.errors = append(l.errors, fmt.Sprint(message))
}

func (l *eventLoggerSpy) Warn(ctx context.Context, message any, fields ...log.Field) {}

func (l *eventLoggerSpy) Info(ctx context.Context, message any, fields ...log.Field) {}

func (l *eventLoggerSpy) Debug(ctx context.Context, message any, fields ...log.Field) {}

func moduleReadyEvent(id int) entity.Event {
	event := entity.NewEvent(entity.EventPayload{
		ModuleReady: &entity.PayloadModuleReady{ModuleId: eventTestModuleId},
	})
	event.Id = id
	return event
}

func moduleDisconnectedEvent(id int) entity.Event {
	event := entity.NewEvent(entity.EventPayload{
		ModuleDisconnected: &entity.PayloadModuleDisconnected{ModuleId: eventTestModuleId},
	})
	event.Id = id
	return event
}

func TestHandleLifecycleBatchFansRoutingOnce(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	subscription := &subscriptionSpy{}
	logger := &eventLoggerSpy{}
	handler := event.NewHandler(subscription, logger)

	batch := event.NewCompactor().Compact([]entity.Event{
		moduleReadyEvent(1),
		moduleDisconnectedEvent(2),
	})
	handler.Handle(t.Context(), batch)

	require.Equal([]string{eventTestModuleId}, subscription.backendsChanged,
		"ready and disconnected collapse into the newest lifecycle event")
	require.Equal(1, subscription.routingCalls, "the routing map is broadcast once per batch")
	require.Empty(logger.errors)
}

func TestHandleUnknownEventDoesNotAbortBatch(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	subscription := &subscriptionSpy{}
	logger := &eventLoggerSpy{}
	handler := event.NewHandler(subscription, logger)

	unknown := entity.NewEvent(entity.EventPayload{})

	handler.Handle(t.Context(), []entity.Event{unknown, moduleReadyEvent(1)})

	require.Len(logger.errors, 1)
	require.Contains(logger.errors[0], "unknown event")
	require.Equal([]string{eventTestModuleId}, subscription.backendsChanged,
		"an unknown payload kind must not stop the batch")
	require.Equal(1, subscription.routingCalls)
}
