package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handlers container
type Handlers struct {
	db *mongo.Database
}

func NewHandlers(db *mongo.Database) *Handlers {
	return &Handlers{db: db}
}

// middleware/json
func JSONContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

// CreateUser
func (h *Handlers) CreateUser(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Phone    string `json:"phone"`
		Role     string `json:"role" binding:"required"`
		Address  string `json:"address"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	user := User{
		Name:      input.Name,
		Email:     input.Email,
		Password:  input.Password, // TODO: hash password in prod
		Phone:     input.Phone,
		Role:      input.Role,
		Address:   input.Address,
		CreatedAt: now,
		UpdatedAt: now,
	}
	res, err := h.db.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed create user: " + err.Error()})
		return
	}
	user.ID = res.InsertedID.(primitive.ObjectID)
	user.Password = "" // hide
	c.JSON(http.StatusCreated, user)
}

// GetUser
func (h *Handlers) GetUser(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var user User
	if err := h.db.Collection("users").FindOne(context.Background(), bson.M{"_id": oid}).Decode(&user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// CreateRuko
func (h *Handlers) CreateRuko(c *gin.Context) {
	collections, _ := h.db.ListCollectionNames(context.Background(), bson.D{})
	log.Println("Collections:", collections)

	var in struct {
		OwnerID         string  `json:"owner_id" binding:"required"`
		Name            string  `json:"name" binding:"required"`
		Description     string  `json:"description"`
		Address         string  `json:"address"`
		City            string  `json:"city"`
		Latitude        float64 `json:"latitude"`
		Longitude       float64 `json:"longitude"`
		Price           float64 `json:"price" binding:"required"`
		DiscountPercent float64 `json:"discount_percent"`
		RentalType      string  `json:"rental_type" binding:"required"`
		Image           string  `json:"image"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	oid, err := primitive.ObjectIDFromHex(in.OwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
		return
	}
	now := time.Now()
	r := Ruko{
		OwnerID:         oid,
		Name:            in.Name,
		Description:     in.Description,
		Address:         in.Address,
		City:            in.City,
		Latitude:        in.Latitude,
		Longitude:       in.Longitude,
		Price:           in.Price,
		DiscountPercent: in.DiscountPercent,
		RentalType:      in.RentalType,
		IsAvailable:     true,
		RentedOffline:   false,
		Image:           in.Image,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	res, err := h.db.Collection("ruko").InsertOne(context.Background(), r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed create ruko"})
		return
	}
	fmt.Println("Inserted Ruko ID:", res.InsertedID)
	r.ID = res.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, r)
}

// ListRuko
func (h *Handlers) ListRuko(c *gin.Context) {
	cur, err := h.db.Collection("ruko").Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list ruko"})
		return
	}
	defer cur.Close(context.Background())
	var results []Ruko
	if err := cur.All(context.Background(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read cursor error"})
		return
	}
	c.JSON(http.StatusOK, results)
}

// GetRuko
func (h *Handlers) GetRuko(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var r Ruko
	if err := h.db.Collection("ruko").FindOne(context.Background(), bson.M{"_id": oid}).Decode(&r); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ruko not found"})
		return
	}
	c.JSON(http.StatusOK, r)
}

// MarkRukoRentedOffline (owner marks ruko as rented offline)
func (h *Handlers) MarkRukoRentedOffline(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	update := bson.M{"$set": bson.M{"rented_offline": true, "is_available": false, "updated_at": time.Now()}}
	_, err = h.db.Collection("ruko").UpdateByID(context.Background(), oid, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed update ruko"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ruko marked as rented offline"})
}

// CreateBooking (simple)
func (h *Handlers) CreateBooking(c *gin.Context) {
	var in struct {
		RukoID        string `json:"ruko_id" binding:"required"`
		TenantID      string `json:"tenant_id" binding:"required"`
		StartDateStr  string `json:"start_date" binding:"required"` // yyyy-mm-dd
		EndDateStr    string `json:"end_date" binding:"required"`
		PaymentMethod string `json:"payment_method" binding:"required"` // online/offline
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rukoOID, _ := primitive.ObjectIDFromHex(in.RukoID)
	tenantOID, _ := primitive.ObjectIDFromHex(in.TenantID)
	startDate, _ := time.Parse("2006-01-02", in.StartDateStr)
	endDate, _ := time.Parse("2006-01-02", in.EndDateStr)

	// fetch ruko to compute price (apply discount if any)
	var r Ruko
	if err := h.db.Collection("ruko").FindOne(context.Background(), bson.M{"_id": rukoOID}).Decode(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ruko not found"})
		return
	}

	// compute simple total: price * months (for simplicity). You can complexify later.
	months := calculateMonthsBetween(startDate, endDate)
	if months < 1 {
		months = 1
	}
	total := r.Price * float64(months)
	if r.DiscountPercent > 0 {
		total = total * (1 - r.DiscountPercent/100.0)
	}

	now := time.Now()
	booking := Booking{
		RukoID:        rukoOID,
		TenantID:      tenantOID,
		StartDate:     startDate,
		EndDate:       endDate,
		TotalPrice:    total,
		PaymentStatus: "pending",
		BookingStatus: "waiting",
		PaymentMethod: in.PaymentMethod,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	res, err := h.db.Collection("bookings").InsertOne(context.Background(), booking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed create booking"})
		return
	}
	booking.ID = res.InsertedID.(primitive.ObjectID)

	// If paymentMethod == offline, we let owner verify later; if online maybe create payment entry after gateway callback.
	c.JSON(http.StatusCreated, booking)
}

func calculateMonthsBetween(a, b time.Time) int {
	ay, am, _ := a.Date()
	by, bm, _ := b.Date()
	months := (by-ay)*12 + int(bm-am)
	if months <= 0 {
		return 1
	}
	return months
}

// GetBooking
func (h *Handlers) GetBooking(c *gin.Context) {
	id := c.Param("id")
	oid, _ := primitive.ObjectIDFromHex(id)
	var b Booking
	if err := h.db.Collection("bookings").FindOne(context.Background(), bson.M{"_id": oid}).Decode(&b); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	c.JSON(http.StatusOK, b)
}

// ListBookings (simple filter)
func (h *Handlers) ListBookings(c *gin.Context) {
	filter := bson.M{}
	tenant := c.Query("tenant_id")
	if tenant != "" {
		if oid, err := primitive.ObjectIDFromHex(tenant); err == nil {
			filter["tenant_id"] = oid
		}
	}
	cur, err := h.db.Collection("bookings").Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed list bookings"})
		return
	}
	defer cur.Close(context.Background())
	var out []Booking
	if err := cur.All(context.Background(), &out); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read cursor failed"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// ConfirmBookingOffline (owner/admin verifies payment & confirms booking)
func (h *Handlers) ConfirmBookingOffline(c *gin.Context) {
	id := c.Param("id")
	bookingOID, _ := primitive.ObjectIDFromHex(id)
	var in struct {
		VerifierID string `json:"verifier_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	verifierOID, _ := primitive.ObjectIDFromHex(in.VerifierID)

	// update booking status
	_, err := h.db.Collection("bookings").UpdateByID(context.Background(), bookingOID, bson.M{
		"$set": bson.M{"booking_status": "confirmed", "payment_status": "paid", "offline_verified_by": verifierOID, "updated_at": time.Now()},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed update booking"})
		return
	}

	// create rental_history entry
	var booking Booking
	_ = h.db.Collection("bookings").FindOne(context.Background(), bson.M{"_id": bookingOID}).Decode(&booking)
	rHistory := RentalHistory{
		RukoID:        booking.RukoID,
		TenantID:      booking.TenantID,
		StartDate:     booking.StartDate,
		EndDate:       booking.EndDate,
		TotalPaid:     booking.TotalPrice,
		PaymentMethod: "offline",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, _ = h.db.Collection("rental_history").InsertOne(context.Background(), rHistory)

	// mark ruko as not available
	_, _ = h.db.Collection("ruko").UpdateByID(context.Background(), booking.RukoID, bson.M{"$set": bson.M{"is_available": false, "updated_at": time.Now()}})

	c.JSON(http.StatusOK, gin.H{"message": "booking confirmed offline and rental history created"})
}

// CreatePayment
func (h *Handlers) CreatePayment(c *gin.Context) {
	var in struct {
		BookingID     string  `json:"booking_id" binding:"required"`
		PaymentMethod string  `json:"payment_method" binding:"required"` // transfer, cash, gateway
		Amount        float64 `json:"amount" binding:"required"`
		PaymentProof  string  `json:"payment_proof"`
		Status        string  `json:"status"` // pending/confirmed/failed
		ConfirmedBy   string  `json:"confirmed_by"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bid, _ := primitive.ObjectIDFromHex(in.BookingID)
	now := time.Now()
	var confirmedBy *primitive.ObjectID
	if in.ConfirmedBy != "" {
		oid, _ := primitive.ObjectIDFromHex(in.ConfirmedBy)
		confirmedBy = &oid
	}

	p := Payment{
		BookingID:     bid,
		PaymentMethod: in.PaymentMethod,
		Amount:        in.Amount,
		PaymentDate:   now,
		PaymentProof:  in.PaymentProof,
		Status:        in.Status,
		ConfirmedBy:   confirmedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	res, err := h.db.Collection("payments").InsertOne(context.Background(), p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed create payment"})
		return
	}
	p.ID = res.InsertedID.(primitive.ObjectID)

	// If confirmed, update booking payment_status
	if p.Status == "confirmed" {
		_, _ = h.db.Collection("bookings").UpdateByID(context.Background(), bid, bson.M{"$set": bson.M{"payment_status": "paid", "booking_status": "confirmed", "updated_at": time.Now()}})
		// create rental history and mark ruko unavailable
		var booking Booking
		_ = h.db.Collection("bookings").FindOne(context.Background(), bson.M{"_id": bid}).Decode(&booking)
		rHistory := RentalHistory{
			RukoID:        booking.RukoID,
			TenantID:      booking.TenantID,
			StartDate:     booking.StartDate,
			EndDate:       booking.EndDate,
			TotalPaid:     booking.TotalPrice,
			PaymentMethod: in.PaymentMethod,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_, _ = h.db.Collection("rental_history").InsertOne(context.Background(), rHistory)
		_, _ = h.db.Collection("ruko").UpdateByID(context.Background(), booking.RukoID, bson.M{"$set": bson.M{"is_available": false, "updated_at": time.Now()}})
	}

	c.JSON(http.StatusCreated, p)
}

// GetPayment
func (h *Handlers) GetPayment(c *gin.Context) {
	id := c.Param("id")
	oid, _ := primitive.ObjectIDFromHex(id)
	var p Payment
	if err := h.db.Collection("payments").FindOne(context.Background(), bson.M{"_id": oid}).Decode(&p); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// Discounts: Create & List
func (h *Handlers) CreateDiscount(c *gin.Context) {
	var in struct {
		RukoID    string  `json:"ruko_id" binding:"required"`
		Name      string  `json:"name" binding:"required"`
		Percent   float64 `json:"percent" binding:"required"`
		StartDate string  `json:"start_date" binding:"required"`
		EndDate   string  `json:"end_date" binding:"required"`
		Active    bool    `json:"active"`
	}

	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid := c.GetString("user_id") // ini dari JWT
	ownerOID, _ := primitive.ObjectIDFromHex(uid)

	rukoOID, _ := primitive.ObjectIDFromHex(in.RukoID)
	sd, _ := time.Parse("2006-01-02", in.StartDate)
	ed, _ := time.Parse("2006-01-02", in.EndDate)

	d := Discount{
		RukoID:    rukoOID,
		OwnerID:   ownerOID,
		Name:      in.Name,
		Percent:   in.Percent,
		StartDate: sd,
		EndDate:   ed,
		Active:    in.Active,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	res, err := h.db.Collection("discounts").InsertOne(context.Background(), d)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed create discount"})
		return
	}

	d.ID = res.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, d)
}

func (h *Handlers) ListDiscounts(c *gin.Context) {
	uid := c.GetString("user_id")
	ownerOID, _ := primitive.ObjectIDFromHex(uid)

	filter := bson.M{
		"owner_id": ownerOID,
		"active":   true,
	}

	cur, err := h.db.Collection("discounts").Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed list discounts"})
		return
	}
	defer cur.Close(context.Background())

	var out []Discount
	if err := cur.All(context.Background(), &out); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read fail"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// ListRentalHistory
func (h *Handlers) ListRentalHistory(c *gin.Context) {
	filter := bson.M{}
	tenant := c.Query("tenant_id")
	if tenant != "" {
		if oid, err := primitive.ObjectIDFromHex(tenant); err == nil {
			filter["tenant_id"] = oid
		}
	}
	cur, err := h.db.Collection("rental_history").Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed list rental history"})
		return
	}
	defer cur.Close(context.Background())
	var out []RentalHistory
	if err := cur.All(context.Background(), &out); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cursor read fail"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// Register (create user with hashed password + return token)
func (h *Handlers) Register(c *gin.Context) {
	var in struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Phone    string `json:"phone"`
		Role     string `json:"role" binding:"required"` // tenant/owner/admin
		Address  string `json:"address"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashed, err := HashPassword(in.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed hash password"})
		return
	}
	now := time.Now()
	user := User{
		Name:      in.Name,
		Email:     in.Email,
		Password:  hashed,
		Phone:     in.Phone,
		Role:      in.Role,
		Address:   in.Address,
		CreatedAt: now,
		UpdatedAt: now,
	}
	res, err := h.db.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed create user: " + err.Error()})
		return
	}
	uid := res.InsertedID.(primitive.ObjectID)

	// generate token
	token, exp, err := GenerateToken(uid, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed generate token"})
		return
	}
	// hide password
	user.ID = uid
	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{
		"user":         user,
		"access_token": token,
		"expires_at":   exp,
	})
}

// Login
func (h *Handlers) Login(c *gin.Context) {
	var in struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user User
	if err := h.db.Collection("users").FindOne(context.Background(), bson.M{"email": in.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if !CheckPassword(user.Password, in.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, exp, err := GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed generate token"})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"access_token": token,
		"expires_at":   exp,
	})
}
