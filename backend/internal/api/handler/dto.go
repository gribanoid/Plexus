package handler

import (
	"time"

	"github.com/google/uuid"
)

type OrgDTO struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	LogoURL   *string   `json:"logo_url"`
	Plan      string    `json:"plan"`
	MyRole    string    `json:"my_role,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type OrgMemberDTO struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	AvatarURL   *string   `json:"avatar_url"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type ProjectDTO struct {
	ID          uuid.UUID  `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	IconURL     *string    `json:"icon_url"`
	LeadID      *uuid.UUID `json:"lead_id"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type StatusDTO struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Color    string    `json:"color"`
	Category string    `json:"category"`
	Position int       `json:"position"`
}

type IssueTypeDTO struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Color   string    `json:"color"`
	IconURL *string   `json:"icon_url"`
}

type LabelDTO struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

type ProjectMemberDTO struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	AvatarURL   *string   `json:"avatar_url"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type CustomFieldDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	Type      string    `json:"type"`
	Required  bool      `json:"required"`
	Options   []string  `json:"options,omitempty"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

type IssueDTO struct {
	ID          uuid.UUID  `json:"id"`
	Number      int64      `json:"number"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Priority    string     `json:"priority"`
	StoryPoints *float64   `json:"story_points"`
	DueDate     *time.Time `json:"due_date"`
	Position    float64    `json:"position"`
	StatusID    uuid.UUID  `json:"status_id"`
	TypeID      uuid.UUID  `json:"type_id"`
	AssigneeID    *uuid.UUID `json:"assignee_id"`
	AssigneeName  *string    `json:"assignee_name,omitempty"`
	ReporterID    uuid.UUID  `json:"reporter_id"`
	ReporterName  *string    `json:"reporter_name,omitempty"`
	SprintID    *uuid.UUID `json:"sprint_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	CustomFields map[string]*string `json:"custom_fields,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type IssueHistoryDTO struct {
	ID        uuid.UUID `json:"id"`
	Field     string    `json:"field"`
	OldValue  *string   `json:"old_value"`
	NewValue  *string   `json:"new_value"`
	ActorID   uuid.UUID `json:"actor_id"`
	CreatedAt time.Time `json:"created_at"`
}

type SprintDTO struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Goal      *string    `json:"goal"`
	State     string     `json:"state"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	CreatedAt time.Time  `json:"created_at"`
}

type CommentDTO struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	AuthorID  uuid.UUID `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AttachmentDTO struct {
	ID         uuid.UUID `json:"id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	Size       int64     `json:"size"`
	UploaderID uuid.UUID `json:"uploader_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type NotificationDTO struct {
	ID        uuid.UUID  `json:"id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Body      *string    `json:"body"`
	Read      bool       `json:"read"`
	IssueID   *uuid.UUID `json:"issue_id"`
	CreatedAt time.Time  `json:"created_at"`
}

type UserDTO struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}
