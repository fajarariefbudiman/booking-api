package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string
type RentalType string

// Users
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"` // store hashed
	Phone     string             `bson:"phone,omitempty" json:"phone"`
	Role      string             `bson:"role" json:"role"` // owner, tenant, admin
	Address   string             `bson:"address,omitempty" json:"address"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}

// Ruko
type Ruko struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID         primitive.ObjectID `bson:"owner_id" json:"owner_id"`
	Name            string             `bson:"name" json:"name"`
	Description     string             `bson:"description,omitempty" json:"description"`
	Address         string             `bson:"address,omitempty" json:"address"`
	City            string             `bson:"city,omitempty" json:"city"`
	Latitude        float64            `bson:"latitude,omitempty" json:"latitude"`
	Longitude       float64            `bson:"longitude,omitempty" json:"longitude"`
	Price           float64            `bson:"price" json:"price"`
	DiscountPercent float64            `bson:"discount_percent,omitempty" json:"discount_percent"`
	RentalType      string             `bson:"rental_type" json:"rental_type"` // monthly, yearly
	IsAvailable     bool               `bson:"is_available" json:"is_available"`
	RentedOffline   bool               `bson:"rented_offline" json:"rented_offline"`
	Image           string             `bson:"image,omitempty" json:"image"`
	CreatedAt       time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}

// Booking
type Booking struct {
	ID                primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	RukoID            primitive.ObjectID  `bson:"ruko_id" json:"ruko_id"`
	TenantID          primitive.ObjectID  `bson:"tenant_id" json:"tenant_id"`
	StartDate         time.Time           `bson:"start_date" json:"start_date"`
	EndDate           time.Time           `bson:"end_date" json:"end_date"`
	TotalPrice        float64             `bson:"total_price" json:"total_price"`
	PaymentStatus     string              `bson:"payment_status" json:"payment_status"` // pending, paid, cancelled
	BookingStatus     string              `bson:"booking_status" json:"booking_status"` // waiting, confirmed, rejected, cancelled
	PaymentMethod     string              `bson:"payment_method" json:"payment_method"` // online, offline
	OfflineVerifiedBy *primitive.ObjectID `bson:"offline_verified_by,omitempty" json:"offline_verified_by,omitempty"`
	CreatedAt         time.Time           `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt         time.Time           `bson:"updated_at,omitempty" json:"updated_at"`
}

// Payment
type Payment struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	BookingID     primitive.ObjectID  `bson:"booking_id" json:"booking_id"`
	PaymentMethod string              `bson:"payment_method" json:"payment_method"` // transfer, cash, gateway
	Amount        float64             `bson:"amount" json:"amount"`
	PaymentDate   time.Time           `bson:"payment_date" json:"payment_date"`
	PaymentProof  string              `bson:"payment_proof,omitempty" json:"payment_proof"`
	Status        string              `bson:"status" json:"status"` // pending, confirmed, failed
	ConfirmedBy   *primitive.ObjectID `bson:"confirmed_by,omitempty" json:"confirmed_by,omitempty"`
	CreatedAt     time.Time           `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt     time.Time           `bson:"updated_at,omitempty" json:"updated_at"`
}

// Discount
type Discount struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RukoID    primitive.ObjectID `bson:"ruko_id" json:"ruko_id"`
	OwnerID   primitive.ObjectID `bson:"owner_id" json:"owner_id"` // <-- baru
	Name      string             `bson:"name" json:"name"`
	Percent   float64            `bson:"percent" json:"percent"`
	StartDate time.Time          `bson:"start_date" json:"start_date"`
	EndDate   time.Time          `bson:"end_date" json:"end_date"`
	Active    bool               `bson:"active" json:"active"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}

// RentalHistory
type RentalHistory struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RukoID        primitive.ObjectID `bson:"ruko_id" json:"ruko_id"`
	TenantID      primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	StartDate     time.Time          `bson:"start_date" json:"start_date"`
	EndDate       time.Time          `bson:"end_date" json:"end_date"`
	TotalPaid     float64            `bson:"total_paid" json:"total_paid"`
	PaymentMethod string             `bson:"payment_method" json:"payment_method"`
	CreatedAt     time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}
