package http

import (
	"net/http"

	"enterprise-core/api/internal/admin"
	auth "enterprise-core/api/internal/auth"
	authmiddleware "enterprise-core/api/internal/auth/middleware"
	"enterprise-core/api/internal/contact"
	contactrepository "enterprise-core/api/internal/contact/repository"
	"enterprise-core/api/internal/customer"
	customerrepository "enterprise-core/api/internal/customer/repository"
	"enterprise-core/api/internal/health"
	"enterprise-core/api/internal/organization"
	organizationrepository "enterprise-core/api/internal/organization/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(
	healthHandler health.Handler,
	authHandler *auth.Handler,
	adminHandler *admin.Handler,
	tokenValidator authmiddleware.TokenValidator,
	db *pgxpool.Pool,
) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3001",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", healthHandler.All)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)

		r.With(
			authmiddleware.JWTAuth(tokenValidator),
		).Get(
			"/me",
			authHandler.Me,
		)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(
			authmiddleware.JWTAuth(tokenValidator),
			authmiddleware.RequireRole("admin"),
		)

		r.Get("/me", adminHandler.Me)
		r.Get("/users", adminHandler.ListUsers)
		r.Get("/users/{id}", adminHandler.GetUser)
		r.Patch("/users/{id}/active", adminHandler.SetUserActive)
		r.Delete("/users/{id}", adminHandler.DeleteUser)
	})

	organizationRepo := organizationrepository.NewRepository(db)
	organizationService := organization.NewService(organizationRepo)
	organizationHandler := organization.NewHandler(organizationService)

	customerRepo := customerrepository.NewRepository(db)
	customerService := customer.NewService(customerRepo)
	customerHandler := customer.NewHandler(
		customerService,
		organizationService,
	)

	contactRepo := contactrepository.NewRepository(db)
	contactService := contact.NewService(contactRepo)
	contactHandler := contact.NewHandler(
		contactService,
		organizationService,
	)

	r.Route("/organizations", func(r chi.Router) {
		r.Use(
			authmiddleware.JWTAuth(tokenValidator),
		)

		r.Post("/", organizationHandler.Create)
		r.Get("/", organizationHandler.List)
		r.Get("/{id}", organizationHandler.Get)

		r.Route("/{id}/customers", func(r chi.Router) {
			r.Post("/", customerHandler.Create)
			r.Get("/", customerHandler.List)

			r.Get(
				"/{customerID}",
				customerHandler.Get,
			)

			r.Put(
				"/{customerID}",
				customerHandler.Update,
			)

			r.Delete(
				"/{customerID}",
				customerHandler.Delete,
			)

			r.Route(
				"/{customerID}/contacts",
				func(r chi.Router) {
					r.Post("/", contactHandler.Create)
					r.Get("/", contactHandler.List)

					r.Get(
						"/{contactID}",
						contactHandler.Get,
					)

					r.Put(
						"/{contactID}",
						contactHandler.Update,
					)

					r.Delete(
						"/{contactID}",
						contactHandler.Delete,
					)
				},
			)
		})
	})

	return r
}
