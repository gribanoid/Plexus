package api

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plexus/backend/internal/api/handler"
	"github.com/plexus/backend/internal/auth"
	"github.com/plexus/backend/internal/config"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/repository"
	"github.com/plexus/backend/internal/search"
	"github.com/plexus/backend/internal/websocket"
	"github.com/redis/go-redis/v9"
)

// Dependencies holds all external dependencies injected into the API layer.
type Dependencies struct {
	Config       *config.Config
	Pool         *pgxpool.Pool
	Redis        *redis.Client
	SearchClient *search.Client
	Hub          *websocket.Hub
	JobClient    *asynq.Client
}

// New creates and configures the Fiber app with all routes.
func New(deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler:          errorHandler,
		DisableStartupMessage: !deps.Config.IsDevelopment(),
	})

	// Global middleware
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	app.Use(recover.New())
	app.Use(middleware.StructuredLogging())
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			for _, o := range deps.Config.CORSOrigins {
				if strings.EqualFold(strings.TrimSpace(o), origin) {
					return true
				}
			}
			return false
		},
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Authorization,X-API-Key",
		AllowCredentials: true,
	}))

	// Health checks
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "plexus"})
	})
	app.Get("/health/detailed", handler.DetailedHealth(handler.HealthDeps{
		Pool:  deps.Pool,
		Redis: deps.Redis,
	}))

	repo := repository.New(deps.Pool)
	orgMember := middleware.RequireOrgMember(repo)
	orgAdmin := middleware.RequireOrgRole("owner", "admin")
	canWrite := middleware.RequireOrgRole("owner", "admin", "member")

	h := handler.New(handler.Deps{
		Repo:        repo,
		Redis:       deps.Redis,
		Search:      deps.SearchClient,
		Hub:         deps.Hub,
		JobClient:   deps.JobClient,
		JWTSecret:   deps.Config.JWTSecret,
		FrontendURL: deps.Config.FrontendURL,
		S3Config: handler.S3Config{
			Endpoint:  deps.Config.S3Endpoint,
			Bucket:    deps.Config.S3Bucket,
			AccessKey: deps.Config.S3AccessKey,
			SecretKey: deps.Config.S3SecretKey,
			Region:    deps.Config.S3Region,
		},
		OIDC: auth.LoadOIDCFromEnv(),
	})

	// Public routes (no auth required)
	auth := app.Group("/api/v1/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", middleware.RateLimitByIP(deps.Redis, "auth:login", 10, time.Minute), h.Login)
	auth.Post("/refresh", h.RefreshToken)
	auth.Post("/logout", h.Logout)
	auth.Get("/oidc/login", h.OIDCLogin)
	auth.Get("/oidc/callback", h.OIDCCallback)

	// Authenticated routes
	api := app.Group("/api/v1", middleware.AuthOrAPIKey(deps.Config.JWTSecret, repo))

	// Current user
	api.Get("/me", h.GetMe)
	api.Patch("/me", h.UpdateMe)

	// Organizations
	orgs := api.Group("/orgs")
	orgs.Post("/", h.CreateOrg)
	orgs.Get("/", h.ListMyOrgs)

	orgScoped := orgs.Group("/:orgSlug", orgMember)
	orgScoped.Get("/", h.GetOrg)
	orgScoped.Patch("/", orgAdmin, h.UpdateOrg)
	orgScoped.Get("/members", h.ListOrgMembers)
	orgScoped.Post("/members/invite", orgAdmin, h.InviteMember)
	orgScoped.Delete("/members/:userID", orgAdmin, h.RemoveMember)
	orgScoped.Get("/search", h.Search)

	// Projects
	projects := orgScoped.Group("/projects")
	projects.Post("/", canWrite, h.CreateProject)
	projects.Get("/", h.ListProjects)
	projects.Get("/:projectKey", h.GetProject)
	projects.Patch("/:projectKey", canWrite, h.UpdateProject)
	projects.Delete("/:projectKey", orgAdmin, h.DeleteProject)

	projects.Get("/:projectKey/members", h.ListProjectMembers)
	projects.Post("/:projectKey/members", canWrite, h.AddProjectMember)
	projects.Delete("/:projectKey/members/:userID", canWrite, h.RemoveProjectMember)

	// Statuses & issue types (board config)
	projects.Get("/:projectKey/statuses", h.ListStatuses)
	projects.Post("/:projectKey/statuses", canWrite, h.CreateStatus)
	projects.Patch("/:projectKey/statuses/:statusID", canWrite, h.UpdateStatus)
	projects.Delete("/:projectKey/statuses/:statusID", canWrite, h.DeleteStatus)

	projects.Get("/:projectKey/issue-types", h.ListIssueTypes)
	projects.Post("/:projectKey/issue-types", canWrite, h.CreateIssueType)

	projects.Get("/:projectKey/labels", h.ListLabels)
	projects.Post("/:projectKey/labels", canWrite, h.CreateLabel)

	projects.Get("/:projectKey/custom-fields", h.ListCustomFields)
	projects.Post("/:projectKey/custom-fields", canWrite, h.CreateCustomField)

	// Sprints
	projects.Get("/:projectKey/sprints", h.ListSprints)
	projects.Post("/:projectKey/sprints", canWrite, h.CreateSprint)
	projects.Patch("/:projectKey/sprints/:sprintID", canWrite, h.UpdateSprint)
	projects.Post("/:projectKey/sprints/:sprintID/start", canWrite, h.StartSprint)
	projects.Post("/:projectKey/sprints/:sprintID/complete", canWrite, h.CompleteSprint)

	// Issues
	issues := projects.Group("/:projectKey/issues")
	issues.Post("/", canWrite, h.CreateIssue)
	issues.Get("/", h.ListIssues)
	issues.Get("/:issueNumber", h.GetIssue)
	issues.Patch("/:issueNumber", canWrite, h.UpdateIssue)
	issues.Delete("/:issueNumber", canWrite, h.DeleteIssue)
	issues.Post("/:issueNumber/move", canWrite, h.MoveIssue)

	// Comments
	issues.Get("/:issueNumber/comments", h.ListComments)
	issues.Post("/:issueNumber/comments", canWrite, h.CreateComment)
	issues.Patch("/:issueNumber/comments/:commentID", canWrite, h.UpdateComment)
	issues.Delete("/:issueNumber/comments/:commentID", canWrite, h.DeleteComment)

	// Attachments
	issues.Get("/:issueNumber/attachments", h.ListAttachments)
	issues.Post("/:issueNumber/attachments/upload-url", canWrite, h.GetUploadURL)
	issues.Post("/:issueNumber/attachments", canWrite, h.CreateAttachment)
	issues.Delete("/:issueNumber/attachments/:attachmentID", canWrite, h.DeleteAttachment)

	// Issue history
	issues.Get("/:issueNumber/history", h.GetIssueHistory)

	// Notifications
	api.Get("/notifications", h.ListNotifications)
	api.Post("/notifications/:notificationID/read", h.MarkNotificationRead)
	api.Post("/notifications/read-all", h.MarkAllNotificationsRead)

	// WebSocket endpoint — JWT validated in handler via ?token=
	app.Get("/ws", h.WebSocketUpgrade)

	return app
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		msg = e.Message
	}

	return c.Status(code).JSON(fiber.Map{"error": msg})
}
