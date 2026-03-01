package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"displayName"`
	Gender       *string   `json:"gender"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type BodyPart string

const (
	BodyPartTop       BodyPart = "top"
	BodyPartBottom    BodyPart = "bottom"
	BodyPartOuter     BodyPart = "outer"
	BodyPartShoes     BodyPart = "shoes"
	BodyPartAccessory BodyPart = "accessory"
)

type Item struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Part         BodyPart  `json:"part"`
	ImageURL     string    `json:"imageUrl"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	SourceURL    *string   `json:"sourceUrl,omitempty"`
	UploadedAt   time.Time `json:"uploadedAt"`
}

type Look struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Items        []Item    `json:"items"`
	ModelGender  string    `json:"modelGender"`
	ModelPhotoID string    `json:"modelPhotoId"`
	RenderedURL  string    `json:"renderedUrl"`
	Score        float64   `json:"score"`
	IsFavorite   bool      `json:"isFavorite"`
	IsSaved      bool      `json:"isSaved"`
	CreatedAt    time.Time `json:"createdAt"`
	Meta         LookMeta  `json:"meta"`
}

type LookMeta struct {
	Colors []string `json:"colors"`
	Style  string   `json:"style"`
	Season string   `json:"season"`
}

type ModelPhoto struct {
	ID       string `json:"id"`
	Gender   string `json:"gender"`
	ImageURL string `json:"imageUrl"`
	Pose     string `json:"pose"`
}

type Position struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type UploadItemRequest struct {
	Part      BodyPart `json:"part"`
	SourceURL *string  `json:"sourceUrl,omitempty"`
}

type GenerateLooksRequest struct {
	ItemIDs     []string         `json:"itemIds"`
	Preferences *LookPreferences `json:"preferences,omitempty"`
}

type LookPreferences struct {
	Style    string `json:"style,omitempty"`
	Season   string `json:"season,omitempty"`
	Occasion string `json:"occasion,omitempty"`
}

type GenerateLooksResponse struct {
	Looks []Look `json:"looks"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
