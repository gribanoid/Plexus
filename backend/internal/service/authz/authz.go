package authz

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Project roles ranked by privilege.
const (
	RoleViewer = "viewer"
	RoleMember = "member"
	RoleAdmin  = "admin"
)

var roleRank = map[string]int{
	RoleViewer: 1,
	RoleMember: 2,
	RoleAdmin:  3,
}

// ProjectAccess describes effective access to a project.
type ProjectAccess struct {
	ProjectID   uuid.UUID
	OrgID       uuid.UUID
	OrgRole     string
	ProjectRole string // empty if org admin bypass without membership row
	CanRead     bool
	CanWrite    bool
	CanAdmin    bool
}

// Store is the data access needed for authorization checks.
type Store interface {
	GetOrgRole(ctx context.Context, userID uuid.UUID, orgSlug string) (orgID uuid.UUID, role string, err error)
	GetProjectRole(ctx context.Context, userID, projectID uuid.UUID) (role string, err error)
	ResolveProject(ctx context.Context, orgSlug, projectKey string) (projectID, orgID uuid.UUID, err error)
}

// Service evaluates org + project RBAC.
type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

// ResolveProjectAccess loads effective permissions for a user on a project.
func (s *Service) ResolveProjectAccess(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) (*ProjectAccess, error) {
	orgID, orgRole, err := s.store.GetOrgRole(ctx, userID, orgSlug)
	if err != nil {
		return nil, err
	}

	projectID, resolvedOrgID, err := s.store.ResolveProject(ctx, orgSlug, projectKey)
	if err != nil {
		return nil, err
	}
	if resolvedOrgID != orgID {
		return nil, pgx.ErrNoRows
	}

	access := &ProjectAccess{
		ProjectID: projectID,
		OrgID:     orgID,
		OrgRole:   orgRole,
	}

	// Org owners/admins have full project access.
	if orgRole == "owner" || orgRole == "admin" {
		access.ProjectRole = RoleAdmin
		access.CanRead = true
		access.CanWrite = true
		access.CanAdmin = true
		return access, nil
	}

	if orgRole == "guest" {
		// Guests need explicit project membership; read-only at best.
		projRole, err := s.store.GetProjectRole(ctx, userID, projectID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				access.CanRead = false
				return access, nil
			}
			return nil, err
		}
		access.ProjectRole = projRole
		access.CanRead = true
		access.CanWrite = false
		access.CanAdmin = false
		return access, nil
	}

	projRole, err := s.store.GetProjectRole(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Org members without project membership: no access (strict RBAC).
			return access, nil
		}
		return nil, err
	}

	access.ProjectRole = projRole
	access.CanRead = true
	access.CanWrite = roleRank[projRole] >= roleRank[RoleMember]
	access.CanAdmin = roleRank[projRole] >= roleRank[RoleAdmin]
	return access, nil
}

// RequireRead ensures the user can view the project.
func (s *Service) RequireRead(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) (*ProjectAccess, error) {
	a, err := s.ResolveProjectAccess(ctx, userID, orgSlug, projectKey)
	if err != nil {
		return nil, err
	}
	if !a.CanRead {
		return nil, ErrForbidden
	}
	return a, nil
}

// RequireWrite ensures the user can mutate project data.
func (s *Service) RequireWrite(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) (*ProjectAccess, error) {
	a, err := s.ResolveProjectAccess(ctx, userID, orgSlug, projectKey)
	if err != nil {
		return nil, err
	}
	if !a.CanWrite {
		return nil, ErrForbidden
	}
	return a, nil
}

// RequireAdmin ensures project admin (or org admin).
func (s *Service) RequireAdmin(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) (*ProjectAccess, error) {
	a, err := s.ResolveProjectAccess(ctx, userID, orgSlug, projectKey)
	if err != nil {
		return nil, err
	}
	if !a.CanAdmin {
		return nil, ErrForbidden
	}
	return a, nil
}

var ErrForbidden = errors.New("insufficient project permissions")
