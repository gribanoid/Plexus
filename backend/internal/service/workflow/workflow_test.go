package workflow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/plexus/backend/internal/service/workflow"
)

type fakeStore struct {
	count int
	ok    bool
}

func (f fakeStore) CountWorkflowTransitions(ctx context.Context, projectID uuid.UUID) (int, error) {
	return f.count, nil
}

func (f fakeStore) IsTransitionAllowed(ctx context.Context, projectID, issueTypeID, fromStatusID, toStatusID uuid.UUID) (bool, error) {
	return f.ok, nil
}

func TestAllowMove_NoRules(t *testing.T) {
	svc := workflow.New(fakeStore{count: 0})
	if err := svc.AllowMove(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New()); err != nil {
		t.Fatal(err)
	}
}

func TestAllowMove_Enforced(t *testing.T) {
	svc := workflow.New(fakeStore{count: 2, ok: false})
	err := svc.AllowMove(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, workflow.ErrTransitionNotAllowed) {
		t.Fatalf("expected ErrTransitionNotAllowed, got %v", err)
	}
}
