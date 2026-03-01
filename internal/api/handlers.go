package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"fashion-look-generator/internal/auth"
	"fashion-look-generator/internal/database"
	"fashion-look-generator/internal/fashion"
	"fashion-look-generator/internal/models"
	"fashion-look-generator/internal/storage"
	"fashion-look-generator/internal/vision"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handlers struct {
	db             *database.DB
	authService    *auth.Service
	storage        *storage.FileStorage
	fashionDB      *fashion.FashionDB
	genderDetector *vision.GenderDetector
	imageComposer  *vision.ImageComposer
}

func NewHandlers(
	db *database.DB,
	authService *auth.Service,
	fileStorage *storage.FileStorage,
	fashionDB *fashion.FashionDB,
) *Handlers {
	return &Handlers{
		db:             db,
		authService:    authService,
		storage:        fileStorage,
		fashionDB:      fashionDB,
		genderDetector: vision.NewGenderDetector(),
		imageComposer:  vision.NewImageComposer(fileStorage.GetBasePath()),
	}

}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, token, err := h.authService.Register(req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, models.AuthResponse{
		User:  *user,
		Token: token,
	})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, token, err := h.authService.Login(req)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, models.AuthResponse{
		User:  *user,
		Token: token,
	})
}

func (h *Handlers) UploadItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file required")
		return
	}
	defer file.Close()

	part := r.FormValue("part")
	if part == "" {
		part = "top"
	}

	originalURL, thumbnailURL, err := h.storage.SaveUpload(file, header.Filename)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	item := &models.Item{
		ID:           uuid.New().String(),
		UserID:       userID,
		Part:         models.BodyPart(part),
		ImageURL:     originalURL,
		ThumbnailURL: thumbnailURL,
		UploadedAt:   time.Now(),
	}

	if err := h.db.CreateItem(item); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create item")
		return
	}

	go h.updateUserGender(userID)

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"itemId":       item.ID,
		"url":          originalURL,
		"thumbnailUrl": thumbnailURL,
	})
}

func (h *Handlers) AddItemLink(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		URL  string `json:"url"`
		Part string `json:"part"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.URL == "" {
		respondError(w, http.StatusBadRequest, "url required")
		return
	}

	originalURL, thumbnailURL, err := h.storage.DownloadFromURL(req.URL)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to download image: "+err.Error())
		return
	}

	item := &models.Item{
		ID:           uuid.New().String(),
		UserID:       userID,
		Part:         models.BodyPart(req.Part),
		ImageURL:     originalURL,
		ThumbnailURL: thumbnailURL,
		SourceURL:    &req.URL,
		UploadedAt:   time.Now(),
	}

	if err := h.db.CreateItem(item); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create item")
		return
	}

	go h.updateUserGender(userID)

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"itemId":       item.ID,
		"url":          originalURL,
		"thumbnailUrl": thumbnailURL,
	})
}

func (h *Handlers) GetUserItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.db.GetUserItems(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get items")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
	})
}

func (h *Handlers) DeleteItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	itemID := vars["id"]

	items, _ := h.db.GetUserItems(userID)
	var itemToDelete *models.Item
	for _, item := range items {
		if item.ID == itemID {
			itemToDelete = &item
			break
		}
	}

	if itemToDelete == nil {
		respondError(w, http.StatusNotFound, "item not found")
		return
	}

	if err := h.db.DeleteItem(itemID, userID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete item")
		return
	}

	h.storage.DeleteFile(itemToDelete.ImageURL)
	h.storage.DeleteFile(itemToDelete.ThumbnailURL)

	respondJSON(w, http.StatusOK, map[string]string{"message": "item deleted"})
}

func (h *Handlers) GenerateLooks(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.GenerateLooksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.ItemIDs) < 2 {
		respondError(w, http.StatusBadRequest, "select at least 2 items")
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	allItems, err := h.db.GetUserItems(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get items")
		return
	}

	selectedItems := make([]models.Item, 0)
	for _, item := range allItems {
		for _, id := range req.ItemIDs {
			if item.ID == id {
				selectedItems = append(selectedItems, item)
				break
			}
		}
	}

	if len(selectedItems) < 2 {
		respondError(w, http.StatusBadRequest, "not enough valid items selected")
		return
	}

	userGender := ""
	if user.Gender != nil && *user.Gender != "" && *user.Gender != "unisex" {
		userGender = *user.Gender
	}
	modelPhotos, err := h.db.GetModelPhotos(userGender)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get model photos")
		return
	}

	oldLooks, _ := h.db.GetUserLooks(userID)
	for _, oldLook := range oldLooks {

		if !oldLook.IsSaved && !oldLook.IsFavorite && oldLook.RenderedURL != "" {
			fmt.Printf("[GenerateLooks] Cleaning up old unsaved look: %s\n", oldLook.ID)
			h.storage.DeleteFile(oldLook.RenderedURL)
			h.db.DeleteLook(oldLook.ID, userID)
		}
	}

	generator := fashion.NewLookGenerator(h.fashionDB)
	looks, err := generator.GenerateLooks(selectedItems, userGender, req.Preferences, modelPhotos)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate looks: "+err.Error())
		return
	}

	for i := range looks {
		look := &looks[i]
		look.UserID = userID

		lookItems := look.Items
		if len(lookItems) == 0 {

			lookItems = selectedItems
		}

		fmt.Printf("[GenerateLooks] Rendering look %s with %d items\n", look.ID, len(lookItems))

		storagePath := h.storage.GetBasePath()
		var renderedURL string

		if len(modelPhotos) > 0 {
			renderedURL, err = h.imageComposer.ComposeLook(
				look,
				modelPhotos[0],
				lookItems,
				storagePath,
			)
			if err != nil {
				fmt.Printf("ComposeLook error: %v\n", err)
				renderedURL, _ = h.imageComposer.CreatePlaceholder(look.ID, storagePath)
			}
		} else {
			renderedURL, _ = h.imageComposer.CreatePlaceholder(look.ID, storagePath)
		}

		look.RenderedURL = renderedURL

		if err := h.db.CreateLook(look); err != nil {
			fmt.Printf("Failed to save look: %v\n", err)
			continue
		}

		if err := h.db.CreateLookItems(look.ID, lookItems); err != nil {
			fmt.Printf("Failed to save look items: %v\n", err)
		}
	}

	respondJSON(w, http.StatusOK, models.GenerateLooksResponse{Looks: looks})
}

func (h *Handlers) GetUserLooks(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	looks, err := h.db.GetUserLooks(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get looks")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"looks": looks,
	})
}

func (h *Handlers) GetLook(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	lookID := vars["id"]

	looks, err := h.db.GetUserLooks(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get look")
		return
	}

	for _, look := range looks {
		if look.ID == lookID {
			respondJSON(w, http.StatusOK, look)
			return
		}
	}

	respondError(w, http.StatusNotFound, "look not found")
}

func (h *Handlers) UpdateLook(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	lookID := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validFields := map[string]bool{"is_favorite": true, "is_saved": true, "isFavorite": true, "isSaved": true}
	filteredUpdates := make(map[string]interface{})
	renderImage := false

	for key, value := range updates {
		if validFields[key] {

			if key == "isFavorite" || key == "isSaved" || key == "is_favorite" || key == "is_saved" {
				if val, ok := value.(bool); ok && val {
					renderImage = true
				}
			}

			dbKey := key
			if key == "isFavorite" {
				dbKey = "is_favorite"
			} else if key == "isSaved" {
				dbKey = "is_saved"
			}
			filteredUpdates[dbKey] = value
		}
	}

	if len(filteredUpdates) == 0 {
		respondError(w, http.StatusBadRequest, "no valid fields to update")
		return
	}

	looks, _ := h.db.GetUserLooks(userID)
	var currentLook *models.Look
	for _, l := range looks {
		if l.ID == lookID {
			currentLook = &l
			break
		}
	}

	isRemovingSaved := false
	isRemovingFavorite := false

	for key, value := range updates {
		if key == "isFavorite" || key == "is_favorite" {
			if val, ok := value.(bool); ok && !val && currentLook != nil && currentLook.IsFavorite {
				isRemovingFavorite = true
			}
		}
		if key == "isSaved" || key == "is_saved" {
			if val, ok := value.(bool); ok && !val && currentLook != nil && currentLook.IsSaved {
				isRemovingSaved = true
			}
		}
	}

	if (isRemovingSaved || isRemovingFavorite) && currentLook != nil && currentLook.RenderedURL != "" {
		fmt.Printf("[UpdateLook] Deleting image for look %s (user removed from saved/favorites)\n", lookID)
		h.storage.DeleteFile(currentLook.RenderedURL)
		filteredUpdates["rendered_url"] = ""
	}

	if renderImage && currentLook != nil && (currentLook.RenderedURL == "" || currentLook.RenderedURL == "/looks/placeholder") {
		fmt.Printf("[UpdateLook] Rendering image for look %s on save\n", lookID)

		lookItems, _ := h.db.GetLookItems(lookID)
		if len(lookItems) > 0 {

			modelPhotos, _ := h.db.GetModelPhotos("")
			storagePath := h.storage.GetBasePath()

			var renderedURL string
			if len(modelPhotos) > 0 {
				renderedURL, _ = h.imageComposer.ComposeLook(
					currentLook,
					modelPhotos[0],
					lookItems,
					storagePath,
				)
			} else {
				renderedURL, _ = h.imageComposer.CreatePlaceholder(lookID, storagePath)
			}

			filteredUpdates["rendered_url"] = renderedURL
		}
	}

	if err := h.db.UpdateLook(lookID, userID, filteredUpdates); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update look")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "look updated"})
}

func (h *Handlers) DeleteLook(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	lookID := vars["id"]

	if err := h.db.DeleteLook(lookID, userID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete look")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "look deleted"})
}

func (h *Handlers) GetModels(w http.ResponseWriter, r *http.Request) {
	gender := r.URL.Query().Get("gender")

	models, err := h.db.GetModelPhotos(gender)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get models")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"models": models,
	})
}

func (h *Handlers) updateUserGender(userID string) {
	items, err := h.db.GetUserItems(userID)
	if err != nil || len(items) == 0 {
		return
	}

	gender, confidence, err := h.genderDetector.DetectGender(items)
	if err != nil {
		return
	}

	if confidence >= 0.7 {
		h.db.UpdateUserGender(userID, gender)
	}
}

func (h *Handlers) analyzeItemStyle(itemID string, imagePath string) {

	fmt.Printf("[ItemAnalysis] Item %s at %s - using default analysis\n", itemID, imagePath)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
