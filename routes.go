package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, h *Handlers) {
	r.Use(cors.Default())
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, "Hello World")
	})
	api := r.Group("/api")
	{
		// auth
		api.POST("/auth/register", h.Register) //finished
		api.POST("/auth/login", h.Login)       //finished

		// public ruko listing
		api.GET("/ruko", h.ListRuko)    //finished
		api.GET("/ruko/:id", h.GetRuko) //finished

		// routes requiring authentication
		authed := api.Group("/")
		authed.Use(AuthMiddleware())
		{
			authed.POST("/bookings", h.CreateBooking) //finished
			authed.GET("/bookings", h.ListBookings)   //finished
			authed.GET("/bookings/:id", h.GetBooking) //finished

			authed.POST("/payments", h.CreatePayment) //finished
			authed.GET("/payments/:id", h.GetPayment) //finished

			// owner-only endpoints
			owner := authed.Group("/")
			owner.Use(RoleMiddleware("owner", "admin"))
			{
				owner.POST("/ruko", h.CreateRuko)                                     //finished
				owner.PATCH("/ruko/:id/rented-offline", h.MarkRukoRentedOffline)      //finished
				owner.PATCH("/bookings/:id/confirm-offline", h.ConfirmBookingOffline) //finished
			}

			// admin or owner may access discounts/rental history
			authed.GET("/discounts", h.ListDiscounts)   //finished
			authed.POST("/discounts", h.CreateDiscount) //finished
			authed.GET("/rental-history", h.ListRentalHistory)
		}
	}
}
