package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/plexus/backend/internal/jobs"
	"github.com/plexus/backend/internal/repository"
)

func (h *Handler) logIssueChanges(ctx context.Context, issueID, actorID uuid.UUID, before repository.IssueDetailDBO, after repository.UpdateIssueInput) {
	if after.Title != nil && *after.Title != before.Title {
		h.insertHistory(ctx, issueID, actorID, "title", before.Title, *after.Title)
	}
	if after.Description != nil {
		oldDesc := nullStringVal(before.Description)
		if *after.Description != oldDesc {
			h.insertHistory(ctx, issueID, actorID, "description", oldDesc, *after.Description)
		}
	}
	if after.StatusID != nil && *after.StatusID != before.StatusID {
		h.insertHistory(ctx, issueID, actorID, "status_id", before.StatusID.String(), after.StatusID.String())
	}
	if after.TypeID != nil && *after.TypeID != before.TypeID {
		h.insertHistory(ctx, issueID, actorID, "type_id", before.TypeID.String(), after.TypeID.String())
	}
	if after.Priority != nil && *after.Priority != before.Priority {
		h.insertHistory(ctx, issueID, actorID, "priority", before.Priority, *after.Priority)
	}
	if after.AssigneeID != nil {
		oldAssignee := nullUUIDVal(before.AssigneeID)
		newAssignee := uuidPtrVal(after.AssigneeID)
		if oldAssignee != newAssignee {
			h.insertHistory(ctx, issueID, actorID, "assignee_id", oldAssignee, newAssignee)
		}
	}
	if after.SprintID != nil {
		oldSprint := nullUUIDVal(before.SprintID)
		newSprint := uuidPtrVal(after.SprintID)
		if oldSprint != newSprint {
			h.insertHistory(ctx, issueID, actorID, "sprint_id", oldSprint, newSprint)
		}
	}
	if after.ParentID != nil {
		oldParent := nullUUIDVal(before.ParentID)
		newParent := uuidPtrVal(after.ParentID)
		if oldParent != newParent {
			h.insertHistory(ctx, issueID, actorID, "parent_id", oldParent, newParent)
		}
	}
	if after.StoryPoints != nil {
		oldPoints := nullFloat32Val(before.StoryPoints)
		newPoints := fmt.Sprintf("%v", *after.StoryPoints)
		if oldPoints != newPoints {
			h.insertHistory(ctx, issueID, actorID, "story_points", oldPoints, newPoints)
		}
	}
	if after.DueDate != nil {
		oldDue := nullTimeVal(before.DueDate)
		newDue := after.DueDate.Format(time.RFC3339)
		if oldDue != newDue {
			h.insertHistory(ctx, issueID, actorID, "due_date", oldDue, newDue)
		}
	}
}

func (h *Handler) logMoveIssueChanges(ctx context.Context, issueID, actorID uuid.UUID, before repository.IssueDetailDBO, after repository.MoveIssueInput) {
	if after.StatusID != nil && *after.StatusID != before.StatusID {
		h.insertHistory(ctx, issueID, actorID, "status_id", before.StatusID.String(), after.StatusID.String())
	}
	if after.SprintID != nil {
		oldSprint := nullUUIDVal(before.SprintID)
		newSprint := uuidPtrVal(after.SprintID)
		if oldSprint != newSprint {
			h.insertHistory(ctx, issueID, actorID, "sprint_id", oldSprint, newSprint)
		}
	}
}

func (h *Handler) insertHistory(ctx context.Context, issueID, actorID uuid.UUID, field, oldValue, newValue string) {
	_ = h.Repo.InsertIssueHistory(ctx, uuid.New(), issueID, actorID, field, oldValue, newValue)
}

func (h *Handler) enqueueIssueIndex(ctx context.Context, issueID, projectID uuid.UUID, title, body string) {
	payload, _ := json.Marshal(jobs.IndexIssuePayload{
		IssueID:   issueID.String(),
		ProjectID: projectID.String(),
		Title:     title,
		Body:      body,
	})
	_, _ = h.JobClient.Enqueue(asynq.NewTask(jobs.TaskIndexIssue, payload))
}

func nullStringVal(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullUUIDVal(u sql.Null[uuid.UUID]) string {
	if u.Valid {
		return u.V.String()
	}
	return ""
}

func uuidPtrVal(u *uuid.UUID) string {
	if u == nil || *u == uuid.Nil {
		return ""
	}
	return u.String()
}

func nullFloat32Val(f sql.NullFloat64) string {
	if f.Valid {
		return fmt.Sprintf("%v", float32(f.Float64))
	}
	return ""
}

func nullTimeVal(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format(time.RFC3339)
	}
	return ""
}
