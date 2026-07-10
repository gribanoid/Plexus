package handler

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/plexus/backend/internal/jobs"
	"github.com/plexus/backend/internal/repository"
)

func (h *Handler) createNotification(ctx context.Context, in repository.CreateNotificationInput) {
	if err := h.Repo.CreateNotification(ctx, in); err != nil {
		return
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
	_, _ = h.JobClient.Enqueue(asynq.NewTask(jobs.TaskSendNotificationEmail, payload))
}
