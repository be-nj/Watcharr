package export

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	export := r.br.Router.Group("/export").Use(authmiddleware.AuthRequired(nil, r.br.Cfg))

	export.GET("letterboxd", r.Letterboxd)
}

// Download the users movies as a letterboxd import compatible csv.
func (r *Router) Letterboxd(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	csvBytes, err := r.service.LetterboxdCsv(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="watcharr-letterboxd.csv"`)
	c.Data(http.StatusOK, "text/csv", csvBytes)
}
