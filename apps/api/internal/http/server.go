package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/3dprint-hub/api/internal/app"
	"github.com/3dprint-hub/api/internal/http/handlers"
	httpmw "github.com/3dprint-hub/api/internal/http/middleware"
)

func New(app *app.Application) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{app.Config.FrontendURL, "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Use(httprate.LimitByIP(100, time.Minute))

	h := handlers.New(app)

	router.Get("/healthz", h.Health)

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)
		r.Post("/auth/refresh", h.Refresh)
		r.Post("/auth/forgot-password", h.ForgotPassword)
		r.Post("/auth/reset-password", h.ResetPassword)
		r.Get("/auth/oauth/{provider}/start", h.OAuthStart)
		r.Get("/auth/oauth/{provider}/callback", h.OAuthCallback)

		r.Post("/pricing/estimate", h.EstimatePrice)

		r.Group(func(protected chi.Router) {
			protected.Use(func(next http.Handler) http.Handler {
				return httpmw.WithAuth(next, app.Tokens)
			})
			protected.Get("/auth/me", h.Me)

			protected.Get("/cart", h.GetCart)
			protected.Post("/cart/items", h.AddCartItem)
			protected.Delete("/cart/items/{itemID}", h.RemoveCartItem)

			protected.Post("/orders/checkout", h.Checkout)
			protected.Get("/orders", h.ListOrders)
			protected.Get("/orders/{orderID}", h.GetOrder)

			protected.Route("/admin", func(admin chi.Router) {
				admin.Use(func(next http.Handler) http.Handler {
					return httpmw.RequireRole("admin", next)
				})
				admin.Get("/orders", h.AdminListOrders)
				admin.Patch("/orders/{orderID}/status", h.AdminUpdateOrderStatus)
			})
		})
	})

	return router
}
