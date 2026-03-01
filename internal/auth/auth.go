package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"fashion-look-generator/internal/database"
	"fashion-look-generator/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	jwtSecret string
	db        *database.DB
}

type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewService(jwtSecret string, db *database.DB) *Service {
	return &Service{
		jwtSecret: jwtSecret,
		db:        db,
	}
}

func (s *Service) Register(req models.RegisterRequest) (*models.User, string, error) {

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return nil, "", fmt.Errorf("all fields are required")
	}

	if len(req.Password) < 6 {
		return nil, "", fmt.Errorf("password must be at least 6 characters")
	}

	existingUser, _, _ := s.db.GetUserByEmail(req.Email)
	if existingUser != nil {
		return nil, "", fmt.Errorf("user with this email already exists")
	}

	user := &models.User{
		ID:          uuid.New().String(),
		Email:       req.Email,
		DisplayName: req.DisplayName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.CreateUser(user, req.Password); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *Service) Login(req models.LoginRequest) (*models.User, string, error) {

	user, passwordHash, err := s.db.GetUserByEmail(req.Email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *Service) generateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func Middleware(authService *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "unauthorized", "message": "missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error": "unauthorized", "message": "invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"error": "unauthorized", "message": "invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
