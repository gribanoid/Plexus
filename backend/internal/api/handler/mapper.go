package handler

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/plexus/backend/internal/repository"
)

func nullStringPtr(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	s := n.String
	return &s
}

func nullUUIDPtr(n sql.Null[uuid.UUID]) *uuid.UUID {
	if !n.Valid {
		return nil
	}
	id := n.V
	return &id
}

func nullFloat64Ptr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	f := n.Float64
	return &f
}

func nullTimePtr(n sql.NullTime) *time.Time {
	if !n.Valid {
		return nil
	}
	t := n.Time
	return &t
}

func timePtrIfSet(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func toOrgDTO(d repository.OrgListDBO) OrgDTO {
	return OrgDTO{
		ID: d.ID, Slug: d.Slug, Name: d.Name,
		LogoURL: nullStringPtr(d.LogoURL), Plan: d.Plan, MyRole: d.MyRole, CreatedAt: d.CreatedAt,
	}
}

func toOrgDetailDTO(d repository.OrgDetailDBO) OrgDTO {
	return OrgDTO{
		ID: d.ID, Slug: d.Slug, Name: d.Name,
		LogoURL: nullStringPtr(d.LogoURL), Plan: d.Plan, MyRole: d.MyRole,
	}
}

func toOrgMemberDTO(d repository.OrgMemberDBO) OrgMemberDTO {
	return OrgMemberDTO{
		ID: d.ID, DisplayName: d.DisplayName, Email: d.Email,
		AvatarURL: nullStringPtr(d.AvatarURL), Role: d.Role, JoinedAt: d.JoinedAt,
	}
}

func toProjectDTO(d repository.ProjectDBO) ProjectDTO {
	return ProjectDTO{
		ID: d.ID, OrgID: d.OrgID, Key: d.Key, Name: d.Name,
		Description: nullStringPtr(d.Description),
		IconURL:     nullStringPtr(d.IconURL),
		LeadID:      nullUUIDPtr(d.LeadID),
		CreatedAt:   timePtrIfSet(d.CreatedAt),
	}
}

func toStatusDTO(d repository.StatusDBO) StatusDTO {
	return StatusDTO{ID: d.ID, Name: d.Name, Color: d.Color, Category: d.Category, Position: d.Position}
}

func toIssueTypeDTO(d repository.IssueTypeDBO) IssueTypeDTO {
	return IssueTypeDTO{ID: d.ID, Name: d.Name, Color: d.Color, IconURL: nullStringPtr(d.IconURL)}
}

func toLabelDTO(d repository.LabelDBO) LabelDTO {
	return LabelDTO{ID: d.ID, Name: d.Name, Color: d.Color}
}

func toProjectMemberDTO(d repository.ProjectMemberDBO) ProjectMemberDTO {
	return ProjectMemberDTO{
		UserID: d.UserID, DisplayName: d.DisplayName, Email: d.Email,
		AvatarURL: nullStringPtr(d.AvatarURL), Role: d.Role, JoinedAt: d.JoinedAt,
	}
}

func toCustomFieldDTO(d repository.CustomFieldDBO) CustomFieldDTO {
	var options []string
	if len(d.Options) > 0 {
		_ = json.Unmarshal(d.Options, &options)
	}
	return CustomFieldDTO{
		ID: d.ID, Name: d.Name, Key: d.Key, Type: d.FieldType,
		Required: d.Required, Options: options, Position: d.Position, CreatedAt: d.CreatedAt,
	}
}

func toIssueListDTO(d repository.IssueListDBO) IssueDTO {
	return IssueDTO{
		ID: d.ID, Number: d.Number, Title: d.Title, Priority: d.Priority,
		StoryPoints: nullFloat64Ptr(d.StoryPoints),
		DueDate:     nullTimePtr(d.DueDate),
		Position:    d.Position,
		StatusID:    d.StatusID, TypeID: d.TypeID,
		AssigneeID: nullUUIDPtr(d.AssigneeID),
		ReporterID: d.ReporterID,
		SprintID:   nullUUIDPtr(d.SprintID),
		CreatedAt:  d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

func toIssueDetailDTO(d repository.IssueDetailDBO) IssueDTO {
	return IssueDTO{
		ID: d.ID, Number: d.Number, Title: d.Title,
		Description: nullStringPtr(d.Description),
		Priority:    d.Priority,
		StoryPoints: nullFloat64Ptr(d.StoryPoints),
		DueDate:     nullTimePtr(d.DueDate),
		Position:    d.Position,
		StatusID:    d.StatusID, TypeID: d.TypeID,
		AssigneeID:   nullUUIDPtr(d.AssigneeID),
		AssigneeName: nullStringPtr(d.AssigneeName),
		ReporterID:   d.ReporterID,
		ReporterName: nullStringPtr(d.ReporterName),
		SprintID:     nullUUIDPtr(d.SprintID),
		ParentID:     nullUUIDPtr(d.ParentID),
		CreatedAt:    d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

func toIssueHistoryDTO(d repository.IssueHistoryDBO) IssueHistoryDTO {
	return IssueHistoryDTO{
		ID: d.ID, Field: d.Field,
		OldValue: nullStringPtr(d.OldValue), NewValue: nullStringPtr(d.NewValue),
		ActorID: d.ActorID, CreatedAt: d.CreatedAt,
	}
}

func toSprintDTO(d repository.SprintDBO) SprintDTO {
	return SprintDTO{
		ID: d.ID, Name: d.Name, Goal: nullStringPtr(d.Goal), State: d.State,
		StartDate: nullTimePtr(d.StartDate), EndDate: nullTimePtr(d.EndDate),
		CreatedAt: d.CreatedAt,
	}
}

func toCommentDTO(d repository.CommentDBO) CommentDTO {
	return CommentDTO{
		ID: d.ID, Body: d.Body, AuthorID: d.AuthorID,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

func toAttachmentDTO(d repository.AttachmentDBO) AttachmentDTO {
	return AttachmentDTO{
		ID: d.ID, Filename: d.Filename, MimeType: d.MimeType,
		Size: d.Size, UploaderID: d.UploaderID, CreatedAt: d.CreatedAt,
	}
}

func toNotificationDTO(d repository.NotificationDBO) NotificationDTO {
	return NotificationDTO{
		ID: d.ID, Type: d.Type, Title: d.Title,
		Body: nullStringPtr(d.Body), Read: d.Read,
		IssueID: nullUUIDPtr(d.IssueID), CreatedAt: d.CreatedAt,
	}
}

func toUserDTO(d repository.UserProfileDBO) UserDTO {
	return UserDTO{
		ID: d.ID, Email: d.Email, DisplayName: d.DisplayName,
		AvatarURL: nullStringPtr(d.AvatarURL), Role: d.Role, CreatedAt: d.CreatedAt,
	}
}

func mapSlice[A any, B any](in []A, fn func(A) B) []B {
	if in == nil {
		return []B{}
	}
	out := make([]B, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}
