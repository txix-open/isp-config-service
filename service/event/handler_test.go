package event_test

import (
	"context"
	"fmt"
	"testing"

	"isp-config-service/entity"
	"isp-config-service/service/event"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/log"
)

type recomputeCall struct {
	moduleId       string
	reason         string
	previousLeader string
}

type subscriptionSpy struct {
	backendsChanged []string
	recomputes      []recomputeCall
	routingCalls    int
}

func (s *subscriptionSpy) NotifyConfigChanged(ctx context.Context, moduleId string) error {
	return nil
}

func (s *subscriptionSpy) NotifyBackendsChanged(ctx context.Context, moduleId string) error {
	s.backendsChanged = append(s.backendsChanged, moduleId)
	return nil
}

func (s *subscriptionSpy) NotifyModuleLeaderRecompute(
	ctx context.Context,
	recompute entity.PayloadModuleLeaderRecompute,
) {
	s.recomputes = append(s.recomputes, recomputeCall{
		moduleId:       recompute.ModuleId,
		reason:         recompute.Reason,
		previousLeader: recompute.PreviousLeader,
	})
}

func (s *subscriptionSpy) NotifyRoutingChanged(ctx context.Context) error {
	s.routingCalls++
	return nil
}

type loggerSpy struct {
	errors []string
}

func (l *loggerSpy) Error(ctx context.Context, message any, fields ...log.Field) {
	l.errors = append(l.errors, fmt.Sprint(message))
}

func (l *loggerSpy) Warn(ctx context.Context, message any, fields ...log.Field) {}

func (l *loggerSpy) Info(ctx context.Context, message any, fields ...log.Field) {}

func (l *loggerSpy) Debug(ctx context.Context, message any, fields ...log.Field) {}

func TestHandleModuleLeaderRecomputeEvent(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	subscription := &subscriptionSpy{}
	logger := &loggerSpy{}
	handler := event.NewHandler(subscription, logger)

	handler.Handle(t.Context(), []entity.Event{
		moduleLeaderRecomputeEvent(entity.ModuleLeaderReasonDisconnect, "10.0.0.1:9000"),
	})

	require.Equal([]recomputeCall{{
		moduleId:       testModuleId,
		reason:         entity.ModuleLeaderReasonDisconnect,
		previousLeader: "10.0.0.1:9000",
	}}, subscription.recomputes)
	require.Empty(subscription.backendsChanged, "the recompute event maps the address list on its own")
	require.Equal(1, subscription.routingCalls, "the recompute event fans the routing map out with the batch")
	require.Empty(logger.errors)
}

func TestHandleBatchKeepsModuleLeaderRecompute(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	subscription := &subscriptionSpy{}
	logger := &loggerSpy{}
	handler := event.NewHandler(subscription, logger)

	ready := entity.NewEvent(entity.EventPayload{
		ModuleReady: &entity.PayloadModuleReady{ModuleId: testModuleId},
	})
	ready.Id = 1

	disconnected := entity.NewEvent(entity.EventPayload{
		ModuleDisconnected: &entity.PayloadModuleDisconnected{ModuleId: testModuleId},
	})
	disconnected.Id = 2

	recompute := moduleLeaderRecomputeEvent(entity.ModuleLeaderReasonDisconnect, "10.0.0.1:9000")
	recompute.Id = 3

	batch := event.NewCompactor().Compact([]entity.Event{ready, disconnected, recompute})
	handler.Handle(t.Context(), batch)

	require.Equal([]string{testModuleId}, subscription.backendsChanged,
		"ready and disconnected collapse into the newest lifecycle event")
	require.Len(subscription.recomputes, 1, "the leader recompute must survive compaction")
	require.Equal(1, subscription.routingCalls, "the routing map is broadcast once per batch")
	require.Empty(logger.errors)
}

func TestHandleUnknownEventDoesNotAbortBatch(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	subscription := &subscriptionSpy{}
	logger := &loggerSpy{}
	handler := event.NewHandler(subscription, logger)

	unknown := entity.NewEvent(entity.EventPayload{})

	handler.Handle(t.Context(), []entity.Event{unknown, moduleLeaderRecomputeEvent(
		entity.ModuleLeaderReasonReady,
		"",
	)})

	require.Len(logger.errors, 1)
	require.Contains(logger.errors[0], "unknown event")
	require.Len(subscription.recomputes, 1, "an unknown payload kind must not stop the batch")
	require.Equal(1, subscription.routingCalls)
}
