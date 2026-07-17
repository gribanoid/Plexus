package workflow

import (
	"context"

	"github.com/google/uuid"
)

// TransitionStore loads workflow rules for a project.
type TransitionStore interface {
	CountWorkflowTransitions(ctx context.Context, projectID uuid.UUID) (int, error)
	IsTransitionAllowed(ctx context.Context, projectID uuid.UUID, issueTypeID, fromStatusID, toStatusID uuid.UUID) (bool, error)
}

// Service enforces status transitions when rules exist.
type Service struct {
	store TransitionStore
}

func New(store TransitionStore) *Service {
	return &Service{store: store}
}

// AllowMove returns nil if the move is permitted.
// If the project has no transitions configured, any move is allowed (legacy boards).
func (s *Service) AllowMove(ctx context.Context, projectID, issueTypeID, fromStatusID, toStatusID uuid.UUID) error {
	if fromStatusID == toStatusID {
		return nil
	}
	n, err := s.store.CountWorkflowTransitions(ctx, projectID)
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	ok, err := s.store.IsTransitionAllowed(ctx, projectID, issueTypeID, fromStatusID, toStatusID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrTransitionNotAllowed
	}
	return nil
}

var ErrTransitionNotAllowed = errTransition("transition not allowed by workflow")

type errTransition string

func (e errTransition) Error() string { return string(e) }
