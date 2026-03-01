package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fashion-look-generator/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	*sql.DB
}

func Initialize() (*DB, error) {








	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "fashion.db"
	}

	var db *sql.DB
	var err error

	if len(dbURL) > 8 && dbURL[:8] == "postgres" {





		return nil, fmt.Errorf("PostgreSQL support requires additional setup. Please see TODO_USER_FILL comments")
	} else {

		dir := filepath.Dir(dbURL)
		if dir != "." && dir != "" {
			os.MkdirAll(dir, 0755)
		}
		db, err = sql.Open("sqlite3", dbURL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &DB{db}

	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database initialized successfully")
	return database, nil
}

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			display_name TEXT NOT NULL,
			gender TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			part TEXT NOT NULL,
			image_url TEXT NOT NULL,
			thumbnail_url TEXT NOT NULL,
			source_url TEXT,
			uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS looks (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			model_gender TEXT NOT NULL,
			model_photo_id TEXT NOT NULL,
			rendered_url TEXT,
			score REAL DEFAULT 0,
			is_favorite BOOLEAN DEFAULT 0,
			is_saved BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			meta TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS look_items (
			look_id TEXT NOT NULL,
			item_id TEXT NOT NULL,
			PRIMARY KEY (look_id, item_id),
			FOREIGN KEY (look_id) REFERENCES looks(id) ON DELETE CASCADE,
			FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS model_photos (
			id TEXT PRIMARY KEY,
			gender TEXT NOT NULL,
			image_url TEXT NOT NULL,
			pose TEXT DEFAULT 'standing'
		)`,






	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	if err := db.seedModelPhotos(); err != nil {
		log.Printf("Warning: Failed to seed model photos: %v", err)
	}

	return nil
}

func (db *DB) seedModelPhotos() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM model_photos").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}








	photos := []struct {
		id     string
		gender string
		url    string
		pose   string
	}{
		{"male_1", "male", "/models/male_1.png", "standing"},
		{"male_2", "male", "/models/male_2.png", "standing"},
		{"female_1", "female", "/models/female_1.jpg", "standing"},
		{"female_2", "female", "/models/female_2.jpg", "standing"},
	}

	for _, photo := range photos {
		_, err := db.Exec(
			"INSERT INTO model_photos (id, gender, image_url, pose) VALUES (?, ?, ?, ?)",
			photo.id, photo.gender, photo.url, photo.pose,
		)
		if err != nil {
			log.Printf("Failed to insert model photo %s: %v", photo.id, err)
		}
	}

	log.Println("Model photos seeded")
	return nil
}

func (db *DB) CreateUser(user *models.User, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = db.Exec(
		"INSERT INTO users (id, email, password_hash, display_name, gender, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		user.ID, user.Email, string(hash), user.DisplayName, user.Gender, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (db *DB) GetUserByEmail(email string) (*models.User, string, error) {
	var user models.User
	var passwordHash string

	err := db.QueryRow(
		"SELECT id, email, password_hash, display_name, gender, created_at, updated_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &passwordHash, &user.DisplayName, &user.Gender, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, "", fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	return &user, passwordHash, nil
}

func (db *DB) GetUserByID(id string) (*models.User, error) {
	var user models.User

	err := db.QueryRow(
		"SELECT id, email, display_name, gender, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Gender, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (db *DB) UpdateUserGender(userID string, gender string) error {
	_, err := db.Exec(
		"UPDATE users SET gender = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		gender, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user gender: %w", err)
	}
	return nil
}

func (db *DB) CreateItem(item *models.Item) error {
	sourceURL := sql.NullString{}
	if item.SourceURL != nil {
		sourceURL.String = *item.SourceURL
		sourceURL.Valid = true
	}

	_, err := db.Exec(
		"INSERT INTO items (id, user_id, part, image_url, thumbnail_url, source_url, uploaded_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		item.ID, item.UserID, item.Part, item.ImageURL, item.ThumbnailURL, sourceURL, item.UploadedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}
	return nil
}

func (db *DB) GetUserItems(userID string) ([]models.Item, error) {
	rows, err := db.Query(
		"SELECT id, user_id, part, image_url, thumbnail_url, source_url, uploaded_at FROM items WHERE user_id = ? ORDER BY uploaded_at DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		var sourceURL sql.NullString

		err := rows.Scan(&item.ID, &item.UserID, &item.Part, &item.ImageURL, &item.ThumbnailURL, &sourceURL, &item.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}

		if sourceURL.Valid {
			item.SourceURL = &sourceURL.String
		}

		items = append(items, item)
	}

	return items, nil
}

func (db *DB) DeleteItem(itemID, userID string) error {
	result, err := db.Exec(
		"DELETE FROM items WHERE id = ? AND user_id = ?",
		itemID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("item not found or not owned by user")
	}

	return nil
}

func (db *DB) CreateLook(look *models.Look) error {
	query := `
		INSERT INTO looks (id, user_id, model_gender, model_photo_id, rendered_url, score, is_favorite, is_saved, created_at, meta)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	metaJSON, err := json.Marshal(look.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = db.Exec(
		query,
		look.ID,
		look.UserID,
		look.ModelGender,
		look.ModelPhotoID,
		look.RenderedURL,
		look.Score,
		look.IsFavorite,
		look.IsSaved,
		look.CreatedAt,
		string(metaJSON),
	)

	return err
}

func (db *DB) CreateLookItems(lookID string, items []models.Item) error {
	for _, item := range items {
		_, err := db.Exec(
			"INSERT OR IGNORE INTO look_items (look_id, item_id) VALUES (?, ?)",
			lookID, item.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to create look_item relationship: %w", err)
		}
	}
	return nil
}

func (db *DB) GetUserLooks(userID string) ([]models.Look, error) {
	rows, err := db.Query(
		"SELECT id, user_id, model_gender, model_photo_id, rendered_url, score, is_favorite, is_saved, created_at, meta FROM looks WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query looks: %w", err)
	}
	defer rows.Close()

	var looks []models.Look
	for rows.Next() {
		var look models.Look
		var metaJSON string

		err := rows.Scan(&look.ID, &look.UserID, &look.ModelGender, &look.ModelPhotoID, &look.RenderedURL, &look.Score, &look.IsFavorite, &look.IsSaved, &look.CreatedAt, &metaJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan look: %w", err)
		}

		if metaJSON != "" {
			json.Unmarshal([]byte(metaJSON), &look.Meta)
		}

		items, err := db.GetLookItems(look.ID)
		if err != nil {
			log.Printf("Failed to load items for look %s: %v", look.ID, err)
		} else {
			look.Items = items
		}

		looks = append(looks, look)
	}

	return looks, nil
}

func (db *DB) GetLookItems(lookID string) ([]models.Item, error) {
	query := `
		SELECT i.id, i.user_id, i.part, i.image_url, i.thumbnail_url, i.source_url, i.uploaded_at 
		FROM items i
		JOIN look_items li ON i.id = li.item_id
		WHERE li.look_id = ?
	`

	rows, err := db.Query(query, lookID)
	if err != nil {
		return nil, fmt.Errorf("failed to query look items: %w", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		var sourceURL sql.NullString

		err := rows.Scan(&item.ID, &item.UserID, &item.Part, &item.ImageURL, &item.ThumbnailURL, &sourceURL, &item.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan look item: %w", err)
		}

		if sourceURL.Valid {
			item.SourceURL = &sourceURL.String
		}

		items = append(items, item)
	}

	return items, nil
}

func (db *DB) UpdateLook(lookID, userID string, updates map[string]interface{}) error {
	query := "UPDATE looks SET "
	args := []interface{}{}
	updatesCount := 0

	for field, value := range updates {
		if updatesCount > 0 {
			query += ", "
		}
		query += field + " = ?"
		args = append(args, value)
		updatesCount++
	}

	if updatesCount == 0 {
		return fmt.Errorf("no updates provided")
	}

	query += " WHERE id = ? AND user_id = ?"
	args = append(args, lookID, userID)

	result, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update look: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("look not found or not owned by user")
	}

	return nil
}

func (db *DB) DeleteLook(lookID, userID string) error {
	result, err := db.Exec(
		"DELETE FROM looks WHERE id = ? AND user_id = ?",
		lookID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete look: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("look not found or not owned by user")
	}

	return nil
}

func (db *DB) GetModelPhotos(gender string) ([]models.ModelPhoto, error) {

	var rows *sql.Rows
	var err error

	if gender != "" {
		rows, err = db.Query(
			"SELECT id, gender, image_url, pose FROM model_photos WHERE gender = ?",
			gender,
		)
	} else {
		rows, err = db.Query(
			"SELECT id, gender, image_url, pose FROM model_photos",
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query model photos: %w", err)
	}
	defer rows.Close()

	var photos []models.ModelPhoto
	for rows.Next() {
		var photo models.ModelPhoto
		err := rows.Scan(&photo.ID, &photo.Gender, &photo.ImageURL, &photo.Pose)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model photo: %w", err)
		}
		photos = append(photos, photo)
	}

	return photos, nil
}
