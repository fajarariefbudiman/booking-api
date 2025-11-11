package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, h *Handlers) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://ruko-space.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, "Hello World")
	})

	api := r.Group("/api")
	{
		// auth
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", h.Login)

		// public ruko listing
		api.GET("/ruko", h.ListRuko)
		api.GET("/ruko/:id", h.GetRuko)

		// authenticated routes
		authed := api.Group("/")
		authed.Use(AuthMiddleware())
		{
			authed.POST("/bookings", h.CreateBooking)
			authed.GET("/bookings", h.ListBookings)
			authed.GET("/bookings/:id", h.GetBooking)

			authed.POST("/payments", h.CreatePayment)
			authed.GET("/payments/:id", h.GetPayment)

			// owner-only routes
			owner := authed.Group("/")
			owner.Use(RoleMiddleware("owner", "admin"))
			{
				owner.POST("/ruko", h.CreateRuko)
				owner.PATCH("/ruko/:id/rented-offline", h.MarkRukoRentedOffline)
				owner.PATCH("/bookings/:id/confirm-offline", h.ConfirmBookingOffline)

				// owner dashboard endpoints
				owner.GET("/:ownerId/stats", h.GetOwnerStats)
				owner.GET("/:ownerId/rukos", h.GetOwnerRukos)
				owner.GET("/:ownerId/bookings/pending", h.GetPendingBookings)
				owner.GET("/:ownerId/bookings", h.GetAllBookings)
				owner.GET("/:ownerId/income", h.GetIncomeData)
				owner.GET("/:ownerId/activities/recent", h.GetRecentActivities)
			}
			// accept/reject booking
			authed.PUT("/bookings/:id/accept", h.AcceptBooking)
			authed.PUT("/bookings/:id/reject", h.RejectBooking)

			// admin/owner discounts & rental history
			authed.GET("/discounts", h.ListDiscounts)
			authed.POST("/discounts", h.CreateDiscount)
			authed.GET("/rental-history", h.ListRentalHistory)

			admin := authed.Group("/")
			admin.Use(RoleMiddleware("admin"))
			{
				admin.GET("/users", h.ListUsers)
			}

		}
	}
}
