package api

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plexus/backend/internal/api/handler"
	"github.com/plexus/backend/internal/auth"
	"github.com/plexus/backend/internal/config"
	"github.com/plexus/backend/internal/crypto"
	"github.com/plexus/backend/internal/metrics"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/repository"
	"github.com/plexus/backend/internal/search"
	"github.com/plexus/backend/internal/service/authz"
	"github.com/plexus/backend/internal/service/workflow"
	"github.com/plexus/backend/internal/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	app.Use(recover.New())
	app.Use(middleware.PrometheusHTTP())
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

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "plexus"})
	})
	metricsAuth := middleware.RequireMetricsAuth(deps.Config.MetricsAuthToken)
	app.Get("/health/detailed", metricsAuth, handler.DetailedHealth(handler.HealthDeps{
		Pool:  deps.Pool,
		Redis: deps.Redis,
	}))
	// Prefer dedicated metrics listen (:9090). Fallback on API port when disabled.
	if !deps.Config.MetricsEnabled {
		promHandler := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{Registry: metrics.Registry})
		app.Get("/metrics", metricsAuth, adaptor.HTTPHandler(promHandler))
	}

	repo := repository.New(deps.Pool)
	authzSvc := authz.New(repo)
	workflowSvc := workflow.New(repo)

	orgMember := middleware.RequireOrgMember(repo)
	orgAdmin := middleware.RequireOrgRole("owner", "admin")
	canWriteOrg := middleware.RequireOrgRole("owner", "admin", "member")

	projRead := middleware.RequireProjectRead(authzSvc)
	projWrite := middleware.RequireProjectWrite(authzSvc)
	projAdmin := middleware.RequireProjectAdmin(authzSvc)

	h := handler.New(handler.Deps{
		Repo:        repo,
		Redis:       deps.Redis,
		Search:      deps.SearchClient,
		Hub:         deps.Hub,
		JobClient:   deps.JobClient,
		Authz:       authzSvc,
		Workflow:    workflowSvc,
		JWTSecret:   deps.Config.JWTSecret,
		FrontendURL: deps.Config.FrontendURL,
		S3Config: handler.S3Config{
			Endpoint:  deps.Config.S3Endpoint,
			Bucket:    deps.Config.S3Bucket,
			AccessKey: deps.Config.S3AccessKey,
			SecretKey: deps.Config.S3SecretKey,
			Region:    deps.Config.S3Region,
		},
		OIDC:              auth.LoadOIDCFromEnv(),
		AllowRegistration: deps.Config.AllowRegistration,
		EncryptionKey:     crypto.KeyFromString(deps.Config.EncryptionKey),
	})

	authGroup := app.Group("/api/v1/auth")
	authGroup.Post("/register", middleware.RateLimitByIP(deps.Redis, "auth:register", 5, time.Minute), h.Register)
	authGroup.Post("/login", middleware.RateLimitByIP(deps.Redis, "auth:login", 10, time.Minute), h.Login)
	authGroup.Post("/refresh", middleware.RateLimitByIP(deps.Redis, "auth:refresh", 30, time.Minute), h.RefreshToken)
	authGroup.Post("/logout", h.Logout)
	authGroup.Get("/oidc/login", h.OIDCLogin)
	authGroup.Get("/oidc/callback", h.OIDCCallback)
	authGroup.Post("/oidc/exchange", middleware.RateLimitByIP(deps.Redis, "auth:oidc-exchange", 20, time.Minute), h.OIDCExchange)
	authGroup.Get("/saml/metadata", h.SAMLMetadata)
	authGroup.Post("/saml/acs", h.SAMLACS)

	apiGroup := app.Group("/api/v1", middleware.AuthOrAPIKey(deps.Config.JWTSecret, repo))

	apiGroup.Get("/me", h.GetMe)
	apiGroup.Patch("/me", h.UpdateMe)

	orgs := apiGroup.Group("/orgs")
	orgs.Post("/", h.CreateOrg)
	orgs.Get("/", h.ListMyOrgs)

	orgScoped := orgs.Group("/:orgSlug", orgMember)
	orgScoped.Get("/", h.GetOrg)
	orgScoped.Patch("/", orgAdmin, h.UpdateOrg)
	orgScoped.Get("/members", h.ListOrgMembers)
	orgScoped.Post("/members/invite", orgAdmin, h.InviteMember)
	orgScoped.Delete("/members/:userID", orgAdmin, h.RemoveMember)
	orgScoped.Get("/search", h.Search)
	orgScoped.Get("/audit", orgAdmin, h.ListAuditEvents)

	orgScoped.Get("/api-keys", orgAdmin, h.ListAPIKeys)
	orgScoped.Post("/api-keys", orgAdmin, h.CreateAPIKey)
	orgScoped.Delete("/api-keys/:keyID", orgAdmin, h.RevokeAPIKey)

	orgScoped.Get("/webhooks", orgAdmin, h.ListWebhooks)
	orgScoped.Post("/webhooks", orgAdmin, h.CreateWebhook)
	orgScoped.Patch("/webhooks/:webhookID", orgAdmin, h.UpdateWebhook)
	orgScoped.Delete("/webhooks/:webhookID", orgAdmin, h.DeleteWebhook)

	orgScoped.Get("/permission-schemes", orgAdmin, h.ListPermissionSchemes)
	orgScoped.Post("/permission-schemes", orgAdmin, h.CreatePermissionScheme)
	orgScoped.Patch("/permission-schemes/:schemeID", orgAdmin, h.UpdatePermissionScheme)

	projects := orgScoped.Group("/projects")
	projects.Post("/", canWriteOrg, h.CreateProject)
	projects.Get("/", h.ListProjects)

	pk := projects.Group("/:projectKey", projRead)
	pk.Get("/", h.GetProject)
	pk.Patch("/", projAdmin, h.UpdateProject)
	pk.Delete("/", orgAdmin, h.DeleteProject)
	pk.Patch("/permission-scheme", projAdmin, h.AssignPermissionScheme)

	pk.Get("/members", h.ListProjectMembers)
	pk.Post("/members", projAdmin, h.AddProjectMember)
	pk.Delete("/members/:userID", projAdmin, h.RemoveProjectMember)

	pk.Get("/statuses", h.ListStatuses)
	pk.Post("/statuses", projAdmin, h.CreateStatus)
	pk.Patch("/statuses/:statusID", projAdmin, h.UpdateStatus)
	pk.Delete("/statuses/:statusID", projAdmin, h.DeleteStatus)

	pk.Get("/issue-types", h.ListIssueTypes)
	pk.Post("/issue-types", projAdmin, h.CreateIssueType)

	pk.Get("/labels", h.ListLabels)
	pk.Post("/labels", projWrite, h.CreateLabel)
	pk.Patch("/labels/:labelID", projWrite, h.UpdateLabel)
	pk.Delete("/labels/:labelID", projWrite, h.DeleteLabel)

	pk.Get("/custom-fields", h.ListCustomFields)
	pk.Post("/custom-fields", projAdmin, h.CreateCustomField)
	pk.Patch("/custom-fields/:fieldID", projAdmin, h.UpdateCustomField)
	pk.Delete("/custom-fields/:fieldID", projAdmin, h.DeleteCustomField)

	pk.Get("/workflow-transitions", h.ListWorkflowTransitions)
	pk.Post("/workflow-transitions", projAdmin, h.CreateWorkflowTransition)
	pk.Delete("/workflow-transitions/:transitionID", projAdmin, h.DeleteWorkflowTransition)

	pk.Get("/saved-filters", h.ListSavedFilters)
	pk.Post("/saved-filters", projWrite, h.CreateSavedFilter)
	pk.Patch("/saved-filters/:filterID", projWrite, h.UpdateSavedFilter)
	pk.Delete("/saved-filters/:filterID", projWrite, h.DeleteSavedFilter)

	pk.Get("/versions", h.ListVersions)
	pk.Post("/versions", projAdmin, h.CreateVersion)
	pk.Patch("/versions/:versionID", projAdmin, h.UpdateVersion)
	pk.Delete("/versions/:versionID", projAdmin, h.DeleteVersion)

	pk.Get("/components", h.ListComponents)
	pk.Post("/components", projAdmin, h.CreateComponent)
	pk.Delete("/components/:componentID", projAdmin, h.DeleteComponent)

	pk.Get("/automation-rules", projAdmin, h.ListAutomationRules)
	pk.Post("/automation-rules", projAdmin, h.CreateAutomationRule)
	pk.Patch("/automation-rules/:ruleID", projAdmin, h.UpdateAutomationRule)
	pk.Delete("/automation-rules/:ruleID", projAdmin, h.DeleteAutomationRule)

	pk.Get("/reports/summary", h.ProjectReportSummary)
	pk.Get("/epics/:issueNumber/children", h.ListEpicChildren)
	pk.Post("/issues/bulk", projWrite, h.BulkUpdateIssues)
	pk.Post("/issues/import", projAdmin, h.ImportIssuesCSV)

	pk.Get("/sprints", h.ListSprints)
	pk.Post("/sprints", projWrite, h.CreateSprint)
	pk.Patch("/sprints/:sprintID", projWrite, h.UpdateSprint)
	pk.Post("/sprints/:sprintID/start", projWrite, h.StartSprint)
	pk.Post("/sprints/:sprintID/complete", projWrite, h.CompleteSprint)

	issues := pk.Group("/issues")
	issues.Post("/", projWrite, h.CreateIssue)
	issues.Get("/", h.ListIssues)
	issues.Get("/:issueNumber", h.GetIssue)
	issues.Patch("/:issueNumber", projWrite, h.UpdateIssue)
	issues.Delete("/:issueNumber", projWrite, h.SoftDeleteIssue)
	issues.Post("/:issueNumber/restore", projAdmin, h.RestoreIssue)
	issues.Post("/:issueNumber/move", projWrite, h.MoveIssue)

	issues.Get("/:issueNumber/comments", h.ListComments)
	issues.Post("/:issueNumber/comments", projWrite, h.CreateComment)
	issues.Patch("/:issueNumber/comments/:commentID", projWrite, h.UpdateComment)
	issues.Delete("/:issueNumber/comments/:commentID", projWrite, h.DeleteComment)

	issues.Get("/:issueNumber/attachments", h.ListAttachments)
	issues.Post("/:issueNumber/attachments/upload-url", projWrite, h.GetUploadURL)
	issues.Post("/:issueNumber/attachments", projWrite, h.CreateAttachment)
	issues.Delete("/:issueNumber/attachments/:attachmentID", projWrite, h.DeleteAttachment)

	issues.Get("/:issueNumber/history", h.GetIssueHistory)
	issues.Get("/:issueNumber/links", h.ListIssueLinks)
	issues.Post("/:issueNumber/links", projWrite, h.CreateIssueLink)
	issues.Delete("/:issueNumber/links/:linkID", projWrite, h.DeleteIssueLink)

	issues.Get("/:issueNumber/watchers", h.ListWatchers)
	issues.Post("/:issueNumber/watchers", projWrite, h.AddWatcher)
	issues.Delete("/:issueNumber/watchers/:userID", projWrite, h.RemoveWatcher)

	issues.Put("/:issueNumber/versions", projWrite, h.SetIssueVersions)
	issues.Put("/:issueNumber/components", projWrite, h.SetIssueComponents)

	apiGroup.Get("/notifications", h.ListNotifications)
	apiGroup.Post("/notifications/:notificationID/read", h.MarkNotificationRead)
	apiGroup.Post("/notifications/read-all", h.MarkAllNotificationsRead)

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
