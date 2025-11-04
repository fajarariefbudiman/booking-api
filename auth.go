package main

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// helper: hash password
func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hashed, pw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pw))
	return err == nil
}

// JWT claims
type MyClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		// default for dev — ubah di production
		s = "change_me_please"
	}
	return []byte(s)
}

func jwtExpiry() time.Duration {
	hrs := 24 // default
	if v := os.Getenv("JWT_EXP_HOURS"); v != "" {
		// ignore parse error, keep default
		if parsed, err := strconv.Atoi(v); err == nil {
			hrs = parsed
		}
	}
	return time.Hour * time.Duration(hrs)
}

func GenerateToken(userID primitive.ObjectID, role string) (string, time.Time, error) {
	exp := time.Now().Add(jwtExpiry())
	claims := MyClaims{
		UserID: userID.Hex(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret())
	return signed, exp, err
}

// Auth middleware: parse token, set user into context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}
		tokenStr := parts[1]
		token, err := jwt.ParseWithClaims(tokenStr, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret(), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			return
		}
		claims, ok := token.Claims.(*MyClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}
		// pass user id & role in context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// Role guard middleware: only allow if role in allowedRoles
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool)
	for _, r := range allowedRoles {
		roleSet[r] = true
	}
	return func(c *gin.Context) {
		roleIf, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role not found"})
			return
		}
		role := roleIf.(string)
		if !roleSet[role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient role"})
			return
		}
		c.Next()
	}
}

// helper untuk ambil user id dari context (hex string)
func GetUserIDFromContext(c *gin.Context) (primitive.ObjectID, error) {
	uidIf, ok := c.Get("user_id")
	if !ok {
		return primitive.NilObjectID, errors.New("user_id not set")
	}
	uidHex := uidIf.(string)
	return primitive.ObjectIDFromHex(uidHex)
}
