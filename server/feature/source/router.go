package source

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br      *router.BaseRouter
	service *Service
}

func NewRouter(br *router.BaseRouter, service *Service) *Router {
	return &Router{
		br,
		service,
	}
}

func (r *Router) AddRoutes() {
	source := r.br.Router.Group("/source").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))

	source.GET("", r.GetSources)
	source.POST("", r.CreateSource)
	source.GET(":id", r.GetSource)
	source.PUT(":id", r.UpdateSource)
	source.DELETE(":id", r.DeleteSource)
	source.PUT(":id/cinema", r.UpdateCinemaDetails)
	source.POST(":id/cinema/refresh", r.RefreshCinemaOsm)
	source.POST(":id/cinema/screen", r.CreateScreen)
	source.DELETE(":id/cinema/screen/:screenId", r.DeleteScreen)
	source.GET(":id/watches", r.GetSourceWatches)
	source.GET(":id/ratings", r.GetSourceRatings)

	// Own group: gin cannot mix a static `geocode` route with the
	// `:id` param routes above.
	geocode := r.br.Router.Group("/geocode").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))
	geocode.GET("", r.Geocode)

	// Own group for the same reason.
	stats := r.br.Router.Group("/stats").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))
	stats.GET("cinema", r.GetCinemaStats)
}

// Get our personal cinema visit stats.
func (r *Router) GetCinemaStats(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	stats, err := r.service.GetCinemaStats(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Get one of our watch sources.
func (r *Router) GetSource(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("getSource route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	source, err := r.service.GetSource(userId, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, source)
}

// Get all watches that used one of our sources.
func (r *Router) GetSourceWatches(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("getSourceWatches route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	watches, err := r.service.GetSourceWatches(userId, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, watches)
}

// Get all visit ratings of a source (anonymised unless opted in).
func (r *Router) GetSourceRatings(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("getSourceRatings route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	ratings, err := r.service.GetSourceRatings(userId, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, ratings)
}

// Get all of our watch sources.
func (r *Router) GetSources(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	sources, err := r.service.GetSources(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, sources)
}

// Create a new watch source.
func (r *Router) CreateSource(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var sr domain.WatchSourceAddRequest
	if err := c.ShouldBindJSON(&sr); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	source, err := r.service.AddSource(userId, sr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, source)
}

// Update one of our watch sources.
func (r *Router) UpdateSource(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("updateSource route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	var sr domain.WatchSourceAddRequest
	if err := c.ShouldBindJSON(&sr); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	if err := r.service.UpdateSource(userId, uint(id), sr); err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// Delete one of our watch sources.
func (r *Router) DeleteSource(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("deleteSource route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	if err := r.service.DeleteSource(userId, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// Refresh the cached OSM data of a cinema (explicit user action).
func (r *Router) RefreshCinemaOsm(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("refreshCinemaOsm route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	details, err := r.service.RefreshCinemaOsm(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}

// Update cinema details of one of our cinema sources.
func (r *Router) UpdateCinemaDetails(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("updateCinemaDetails route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	var cr domain.CinemaDetailsUpdateRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	if err := r.service.UpdateCinemaDetails(userId, uint(id), cr); err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// Add a screen to one of our cinema sources.
func (r *Router) CreateScreen(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("createScreen route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	var sr domain.CinemaScreenAddRequest
	if err := c.ShouldBindJSON(&sr); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	screen, err := r.service.AddScreen(userId, uint(id), sr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, screen)
}

// Delete a screen from one of our cinema sources.
func (r *Router) DeleteScreen(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Error("deleteScreen route failed to convert id param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid id"})
		return
	}
	screenId, err := strconv.Atoi(c.Param("screenId"))
	if err != nil {
		slog.Error("deleteScreen route failed to convert screenId param to int", "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid screen id"})
		return
	}
	if err := r.service.DeleteScreen(userId, uint(id), uint(screenId)); err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// Geocode a query via nominatim, for prefilling cinema coordinates.
func (r *Router) Geocode(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "no query provided"})
		return
	}
	results, err := r.service.Geocode(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}
