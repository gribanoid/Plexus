package handler

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/plexus/backend/internal/jobs"
	"github.com/plexus/backend/internal/repository"
	"github.com/plexus/backend/internal/websocket"
)

func (h *Handler) createNotification(ctx context.Context, in repository.CreateNotificationInput) {
	if err := h.Repo.CreateNotification(ctx, in); err != nil {
		return
	}

	if h.Hub != nil {
		uid := in.UserID
		h.Hub.Publish(&websocket.Event{
			Type:   websocket.EventNotification,
			UserID: &uid,
			Payload: map[string]any{
				"id":       in.ID,
				"type":     in.Type,
				"title":    in.Title,
				"body":     in.Body,
				"issue_id": in.IssueID,
			},
		})
	}

	profile, err := h.Repo.GetUserProfile(ctx, in.UserID)
	if err != nil {
		return
	}

	body := in.Title
	if in.Body != nil {
		body = *in.Body
	}

	payload, _ := json.Marshal(jobs.EmailPayload{
		To:      profile.Email,
		Subject: in.Title,
		Body:    body,
	})
	if h.JobClient != nil {
		_, _ = h.JobClient.Enqueue(asynq.NewTask(jobs.TaskSendNotificationEmail, payload))
	}
}
