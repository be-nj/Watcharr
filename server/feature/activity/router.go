package activity

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
	activity := r.br.Router.Group("/activity").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))

	activity.GET(":watchedId", r.GetActivity)
	activity.PUT(":id", r.UpdateActivity)
	activity.PUT(":id/details", r.UpdateActivityDetails)
	activity.POST(":id/companions", r.LogCompanions)
	activity.DELETE(":id", r.DeleteActivity)
}

// Log one of our plays for friends that watched together with us.
func (r *Router) LogCompanions(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "check id route param"})
		return
	}
	var cr domain.CompanionLogRequest
	if err := c.ShouldBindJSON(&cr); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	results, err := r.service.LogCompanions(userId, uint(id), cr.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// Set/replace details (watch source, language, tags, note) on one of our activities.
func (r *Router) UpdateActivityDetails(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "check id route param"})
		return
	}
	var dr domain.ActivityDetailsUpdateRequest
	if err := c.ShouldBindJSON(&dr); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	details, err := r.service.UpdateActivityDetails(userId, uint(id), dr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, details)
}

func (r *Router) GetActivity(c *gin.Context) {
	watchedId, err := strconv.ParseUint(c.Param("watchedId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "check watched id route param"})
		return
	}
	userId := c.MustGet("userId").(uint)
	activity, err := r.service.GetActivity(userId, uint(watchedId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, activity)
}

func (r *Router) UpdateActivity(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Status(400)
		return
	}
	var activityUpdateRequest domain.ActivityUpdateRequest
	err = c.ShouldBindJSON(&activityUpdateRequest)
	if err == nil {
		err = r.service.UpdateActivity(userId, uint(id), activityUpdateRequest)
		if err != nil {
			c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusOK)
		return
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
}

func (r *Router) DeleteActivity(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Status(400)
		slog.Error("Could not process activity id when attempting a deletion", "error", err.Error(), "id", c.Param("id"))
		return
	}
	err = r.service.DeleteActivity(userId, uint(id))
	if err != nil {
		c.JSON(http.StatusForbidden, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
