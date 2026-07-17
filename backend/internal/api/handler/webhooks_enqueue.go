package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/plexus/backend/internal/jobs"
	"github.com/plexus/backend/internal/metrics"
)

func (h *Handler) enqueueOrgWebhooks(ctx context.Context, orgID uuid.UUID, event string, data map[string]any) {
	if h.JobClient == nil {
		return
	}
	hooks, err := h.Repo.ListActiveWebhooksForEvent(ctx, orgID, event)
	if err != nil {
		slog.Error("list webhooks", "error", err)
		return
	}
	body, err := json.Marshal(map[string]any{
		"id":         uuid.New().String(),
		"type":       event,
		"org_id":     orgID.String(),
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"data":       data,
	})
	if err != nil {
		return
	}
	for _, wh := range hooks {
		payload, _ := json.Marshal(jobs.WebhookPayload{
			WebhookID: wh.ID.String(),
			Event:     event,
			Body:      body,
		})
		if _, err := h.JobClient.Enqueue(asynq.NewTask(jobs.TaskDeliverWebhook, payload)); err != nil {
			slog.Error("enqueue webhook", "error", err)
			continue
		}
		metrics.JobsEnqueued.WithLabelValues(jobs.TaskDeliverWebhook).Inc()
	}
}
