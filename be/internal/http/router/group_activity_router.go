package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/dhruvsaxena1998/splitplus/internal/http/handlers"
	"github.com/dhruvsaxena1998/splitplus/internal/http/middleware"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func WithGroupActivityRoutes(activityService service.GroupActivityService, jwtService service.JWTService, sessionRepo repository.SessionRepository) Option {
	return optionFunc(func(r chi.Router) {
		r.Route("/groups/{group_id}/activity", func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtService, sessionRepo))

			// GET / - List group activities
			r.Get("/", handlers.ListGroupActivitiesHandler(activityService))
		})

		r.Route("/groups/{group_id}/expenses/{expense_id}/history", func(r chi.Router) {
			r.Use(middleware.RequireAuth(jwtService, sessionRepo))

			// GET / - Get expense history
			r.Get("/", handlers.GetExpenseHistoryHandler(activityService))
		})
	})
}
