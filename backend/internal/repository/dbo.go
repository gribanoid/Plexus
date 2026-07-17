package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type OrgListDBO struct {
	ID        uuid.UUID
	Slug      string
	Name      string
	LogoURL   sql.NullString
	Plan      string
	MyRole    string
	CreatedAt time.Time
}

type OrgDetailDBO struct {
	ID      uuid.UUID
	Slug    string
	Name    string
	LogoURL sql.NullString
	Plan    string
	MyRole  string
}

type OrgMemberDBO struct {
	ID          uuid.UUID
	DisplayName string
	Email       string
	AvatarURL   sql.NullString
	Role        string
	JoinedAt    time.Time
}

type ProjectDBO struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	Key         string
	Name        string
	Description sql.NullString
	IconURL     sql.NullString
	LeadID      sql.Null[uuid.UUID]
	CreatedAt   time.Time
}

type StatusDBO struct {
	ID       uuid.UUID
	Name     string
	Color    string
	Category string
	Position int
}

type IssueTypeDBO struct {
	ID      uuid.UUID
	Name    string
	Color   string
	IconURL sql.NullString
}

type LabelDBO struct {
	ID    uuid.UUID
	Name  string
	Color string
}

type ProjectMemberDBO struct {
	UserID      uuid.UUID
	DisplayName string
	Email       string
	AvatarURL   sql.NullString
	Role        string
	JoinedAt    time.Time
}

type CustomFieldDBO struct {
	ID        uuid.UUID
	Name      string
	Key       string
	FieldType string
	Required  bool
	Options   []byte
	Position  int
	CreatedAt time.Time
}

type IssueListDBO struct {
	ID          uuid.UUID
	Number      int64
	Title       string
	Priority    string
	StoryPoints sql.NullFloat64
	DueDate     sql.NullTime
	Position    float64
	StatusID    uuid.UUID
	TypeID      uuid.UUID
	AssigneeID  sql.Null[uuid.UUID]
	ReporterID  uuid.UUID
	SprintID    sql.Null[uuid.UUID]
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type IssueDetailDBO struct {
	ID            uuid.UUID
	Number        int64
	Title         string
	Description   sql.NullString
	Priority      string
	StoryPoints   sql.NullFloat64
	DueDate       sql.NullTime
	Position      float64
	StatusID      uuid.UUID
	TypeID        uuid.UUID
	AssigneeID    sql.Null[uuid.UUID]
	AssigneeName  sql.NullString
	ReporterID    uuid.UUID
	ReporterName  sql.NullString
	SprintID      sql.Null[uuid.UUID]
	ParentID      sql.Null[uuid.UUID]
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type IssueHistoryDBO struct {
	ID        uuid.UUID
	Field     string
	OldValue  sql.NullString
	NewValue  sql.NullString
	ActorID   uuid.UUID
	CreatedAt time.Time
}

type SprintDBO struct {
	ID        uuid.UUID
	Name      string
	Goal      sql.NullString
	State     string
	StartDate sql.NullTime
	EndDate   sql.NullTime
	CreatedAt time.Time
}

type CommentDBO struct {
	ID        uuid.UUID
	Body      string
	AuthorID  uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AttachmentDBO struct {
	ID         uuid.UUID
	Filename   string
	MimeType   string
	Size       int64
	UploaderID uuid.UUID
	CreatedAt  time.Time
}

type NotificationDBO struct {
	ID        uuid.UUID
	Type      string
	Title     string
	Body      sql.NullString
	Read      bool
	IssueID   sql.Null[uuid.UUID]
	CreatedAt time.Time
}

type UserProfileDBO struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	AvatarURL   sql.NullString
	Role        string
	CreatedAt   time.Time
}

type UserCredentials struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

type RegisterInput struct {
	UserID       uuid.UUID
	Email        string
	PasswordHash string
	DisplayName  string
	OrgID        uuid.UUID
	OrgSlug      string
	OrgName      string
	ProjectID    uuid.UUID
}

type CreateIssueInput struct {
	IssueID     uuid.UUID
	ProjectID   uuid.UUID
	Number      int64
	TypeID      uuid.UUID
	StatusID    uuid.UUID
	Title       string
	Description *string
	Priority    string
	AssigneeID  *uuid.UUID
	ReporterID  uuid.UUID
	ParentID    *uuid.UUID
	SprintID    *uuid.UUID
	StoryPoints *float32
	DueDate     *time.Time
	Position    float64
	LabelIDs    []uuid.UUID
}

type IssueFilters struct {
	StatusID   string
	AssigneeID string
	SprintID   string
	Priority   string
}

type UpdateIssueInput struct {
	Title       *string
	Description *string
	StatusID    *uuid.UUID
	TypeID      *uuid.UUID
	Priority    *string
	AssigneeID  *uuid.UUID
	SprintID    *uuid.UUID
	ClearSprint bool
	ParentID    *uuid.UUID
	StoryPoints *float32
	DueDate     *time.Time
}

type MoveIssueInput struct {
	StatusID *uuid.UUID
	Position *float64
	SprintID *uuid.UUID
}
