package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RunSeeder(ctx context.Context, db *mongo.Database) {
	// --- Users ---
	userCol := db.Collection("users")
	count, err := userCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		hash_password, _ := HashPassword("123123123")
		users := []User{
			{Name: "Alice", Email: "alice@mail.com", Password: hash_password, Role: "owner"},
			{Name: "Bob", Email: "bob@mail.com", Password: hash_password, Role: "owner"},
			{Name: "Charlie", Email: "charlie@mail.com", Password: hash_password, Role: "tenant"},
			{Name: "David", Email: "david@mail.com", Password: hash_password, Role: "tenant"},
			{Name: "Eve", Email: "eve@mail.com", Password: hash_password, Role: "admin"},
		}
		for i := range users {
			users[i].ID = primitive.NewObjectID()
			users[i].CreatedAt = time.Now()
			users[i].UpdatedAt = time.Now()
		}
		var userInterfaces []interface{}
		for _, u := range users {
			userInterfaces = append(userInterfaces, u)
		}
		_, err := userCol.InsertMany(ctx, userInterfaces)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Seeded Users ✅")
	} else {
		fmt.Println("Users already exist, skipping seeder")
	}

	// --- Ruko ---
	rukoCol := db.Collection("ruko")
	count, _ = rukoCol.CountDocuments(ctx, bson.M{})
	if count == 0 {
		var usersList []User
		cursor, _ := userCol.Find(ctx, bson.M{"role": "owner"})
		if err := cursor.All(ctx, &usersList); err != nil {
			log.Fatal(err)
		}

		var rukos []interface{}
		rand.Seed(time.Now().UnixNano())
		for i := 1; i <= 20; i++ {
			ruko := Ruko{
				ID:            primitive.NewObjectID(),
				OwnerID:       usersList[rand.Intn(len(usersList))].ID,
				Name:          fmt.Sprintf("Ruko %02d", i),
				Description:   "Ruko strategis di pusat kota",
				Address:       fmt.Sprintf("%d Jalan Mawar", i),
				City:          "Jakarta",
				Latitude:      -6.2 + rand.Float64()*0.1,
				Longitude:     106.8 + rand.Float64()*0.1,
				Price:         float64(5000000 + rand.Intn(5000000)),
				RentalType:    "monthly",
				IsAvailable:   true,
				RentedOffline: false,
				Image:         fmt.Sprintf("ruko%d.jpg", i),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			rukos = append(rukos, ruko)
		}
		_, err := rukoCol.InsertMany(ctx, rukos)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Seeded Ruko ✅")
	} else {
		fmt.Println("Ruko already exist, skipping seeder")
	}

	// --- Discounts ---
	discountCol := db.Collection("discounts")
	count, _ = discountCol.CountDocuments(ctx, bson.M{})
	if count == 0 {
		var rukoList []Ruko
		cursor, _ := rukoCol.Find(ctx, bson.M{})
		if err := cursor.All(ctx, &rukoList); err != nil {
			log.Fatal(err)
		}

		var discounts []interface{}
		rand.Seed(time.Now().UnixNano())
		for i := 1; i <= 7; i++ {
			discount := Discount{
				ID:        primitive.NewObjectID(),
				RukoID:    rukoList[rand.Intn(len(rukoList))].ID,
				Name:      fmt.Sprintf("Promo %d", i),
				Percent:   float64(5 + rand.Intn(20)),
				StartDate: time.Now(),
				EndDate:   time.Now().AddDate(0, 1, 0),
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			discounts = append(discounts, discount)
		}
		_, err := discountCol.InsertMany(ctx, discounts)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Seeded Discounts ✅")
	} else {
		fmt.Println("Discounts already exist, skipping seeder")
	}
}
