package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"iag-inventory/backend/internal/config"
	"iag-inventory/backend/internal/db"
	"iag-inventory/backend/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RouterDeps struct {
	Cfg          config.Config
	Pool         *pgxpool.Pool
	PlatformAuth *middleware.PlatformAuth
	StrictRBAC   bool
}

func NewRouter(deps RouterDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(otelgin.Middleware(deps.Cfg.ServiceName))
	r.Use(gin.Recovery())
	if deps.PlatformAuth != nil {
		r.Use(deps.PlatformAuth.AttachPrincipal())
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": deps.Cfg.ServiceName})
	})
	r.GET("/ready", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context(), deps.Pool); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "database": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready", "database": true})
	})

	v1 := r.Group("/api/v1")
	if deps.PlatformAuth != nil {
		v1.Use(deps.PlatformAuth.RequireAuth())
	}
	if deps.StrictRBAC {
		v1.Use(middleware.StrictRBAC())
	}
	{
		// Platform status — staff-only liveness/identity probe behind auth.
		v1.GET("/platform/status", middleware.RequireStaff(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"service":     deps.Cfg.ServiceName,
				"environment": deps.Cfg.Environment,
			})
		})

		// Domain endpoints (SKU master, on-hand ledger, stock movements) are
		// added here as the inventory domain is implemented. The skeleton
		// exposes a permission-gated overview so RBAC wiring is exercised.
		v1.GET("/overview", middleware.RequirePermission("inventory.view_overview"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"service": deps.Cfg.ServiceName,
				"status":  "scaffold",
				"message": "inventory domain not yet implemented",
			})
		})
	}
	return r
}
