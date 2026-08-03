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
	source.PUT(":id", r.UpdateSource)
	source.DELETE(":id", r.DeleteSource)
	source.PUT(":id/cinema", r.UpdateCinemaDetails)
	source.POST(":id/cinema/screen", r.CreateScreen)
	source.DELETE(":id/cinema/screen/:screenId", r.DeleteScreen)
	source.GET("geocode", r.Geocode)
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
