package handler

import (
	"github.com/hibiken/asynq"
	"github.com/plexus/backend/internal/auth"
	"github.com/plexus/backend/internal/repository"
	"github.com/plexus/backend/internal/search"
	"github.com/plexus/backend/internal/service/authz"
	"github.com/plexus/backend/internal/service/workflow"
	"github.com/plexus/backend/internal/websocket"
	"github.com/redis/go-redis/v9"
)

// S3Config holds S3/MinIO configuration.
type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
}

// Deps contains all dependencies needed by handlers.
type Deps struct {
	Repo         *repository.Repo
	Redis        *redis.Client
	Search       *search.Client
	Hub          *websocket.Hub
	JobClient    *asynq.Client
	Authz        *authz.Service
	Workflow     *workflow.Service
	JWTSecret    string
	FrontendURL  string
	S3Config     S3Config
	OIDC         *auth.Provider
}

// Handler embeds all dependencies for use in handler methods.
type Handler struct {
	Deps
}

func New(deps Deps) *Handler {
	return &Handler{Deps: deps}
}
