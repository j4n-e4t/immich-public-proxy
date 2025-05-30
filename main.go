package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"tailscale.com/tsnet"

)

type ShareResponse struct {
	AllowDownload bool        `json:"allowDownload"`
	AllowUpload   bool        `json:"allowUpload"`
	Assets        []Asset     `json:"assets"`
	CreatedAt     time.Time   `json:"createdAt"`
	Description   interface{} `json:"description"` // Can be null
	ExpiresAt     interface{} `json:"expiresAt"`   // Can be null
	ID            string      `json:"id"`
	Key           string      `json:"key"`
	Password      string      `json:"password"`
	ShowMetadata  bool        `json:"showMetadata"`
	Type          string      `json:"type"`
	UserID        string      `json:"userId"`
}

type Asset struct {
	Checksum         string      `json:"checksum"`
	DeviceAssetID    string      `json:"deviceAssetId"`
	DeviceID         string      `json:"deviceId"`
	DuplicateID      interface{} `json:"duplicateId"` // Can be null
	Duration         string      `json:"duration"`
	ExifInfo         ExifInfo    `json:"exifInfo"`
	FileCreatedAt    time.Time   `json:"fileCreatedAt"`
	FileModifiedAt   time.Time   `json:"fileModifiedAt"`
	HasMetadata      bool        `json:"hasMetadata"`
	ID               string      `json:"id"`
	IsArchived       bool        `json:"isArchived"`
	IsFavorite       bool        `json:"isFavorite"`
	IsOffline        bool        `json:"isOffline"`
	IsTrashed        bool        `json:"isTrashed"`
	LibraryID        interface{} `json:"libraryId"`        // Can be null
	LivePhotoVideoID interface{} `json:"livePhotoVideoId"` // Can be null
	LocalDateTime    time.Time   `json:"localDateTime"`
	OriginalFileName string      `json:"originalFileName"`
	OriginalMimeType string      `json:"originalMimeType"`
	OriginalPath     string      `json:"originalPath"`
	OwnerID          string      `json:"ownerId"`
	People           []interface{} `json:"people"` // Empty array in example
	Resized          bool        `json:"resized"`
	Thumbhash        string      `json:"thumbhash"`
	Type             string      `json:"type"`
	UpdatedAt        time.Time   `json:"updatedAt"`
	Visibility       string      `json:"visibility"`
}

type ExifInfo struct {
	City             string      `json:"city"`
	Country          string      `json:"country"`
	DateTimeOriginal time.Time   `json:"dateTimeOriginal"`
	Description      string      `json:"description"`
	ExifImageHeight  int         `json:"exifImageHeight"`
	ExifImageWidth   int         `json:"exifImageWidth"`
	ExposureTime     string      `json:"exposureTime"`
	FNumber          float64     `json:"fNumber"`
	FileSizeInByte   int         `json:"fileSizeInByte"`
	FocalLength      float64     `json:"focalLength"`
	Iso              int         `json:"iso"`
	Latitude         float64     `json:"latitude"`
	LensModel        string      `json:"lensModel"`
	Longitude        float64     `json:"longitude"`
	Make             string      `json:"make"`
	Model            string      `json:"model"`
	ModifyDate       time.Time   `json:"modifyDate"`
	Orientation      string      `json:"orientation"`
	ProjectionType   interface{} `json:"projectionType"` // Can be null
	Rating           interface{} `json:"rating"`         // Can be null
	State            string      `json:"state"`
	TimeZone         string      `json:"timeZone"`
}

// Template data structures
type Image struct {
	URL   string
	Alt   string
	Title string
}

type PageData struct {
	Title    string
	PreviewURLs []string
	ThumbnailURLs []string
}

var tmpl *template.Template

func init() {
	// Parse the template with custom functions
	funcMap := template.FuncMap{
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
		"title": strings.Title,
	}

	var err error
	tmpl, err = template.New("gallery.html").Funcs(funcMap).ParseFiles("gallery.html")
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}
}

func handleAsset(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/asset/")
	key := r.URL.Query().Get("key")
	thumbnail := r.URL.Query().Get("thumbnail")

	immichURL := ""
	if thumbnail == "true" {
		immichURL = getEnv("IMMICH_BASE_URL", "http://immich:2283/") +
			"api/assets/" + id + "/thumbnail?size=thumbnail&key=" + key
	} else {
		immichURL = getEnv("IMMICH_BASE_URL", "http://immich:2283/") +
			"api/assets/" + id + "/thumbnail?size=preview&key=" + key
	}

	resp, err := http.Get(immichURL)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	w.Header().Set("Content-Type", contentType)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func handleShare(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/share/")

	resp, err := http.Get(getEnv("IMMICH_BASE_URL", "http://immich:2283/") +
		"api/shared-links/me?key=" + id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Parse the JSON response
	var shareData ShareResponse
	if err := json.NewDecoder(resp.Body).Decode(&shareData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Convert assets to template images
	previewURLs := make([]string, 0, len(shareData.Assets))
	thumbnailURLs := make([]string, 0, len(shareData.Assets))
	for _, asset := range shareData.Assets {
		// Skip non-image assets
		if asset.Type != "IMAGE" {
			continue
		}

		previewURLs = append(previewURLs, "/asset/" + asset.ID + "?key=" + id)
		thumbnailURLs = append(thumbnailURLs, "/asset/" + asset.ID + "?key=" + id + "&thumbnail=true")
	}

	// Prepare template data
	data := PageData{
		Title:    "Immich Gallery",
		PreviewURLs:   previewURLs,
		ThumbnailURLs: thumbnailURLs,
	}

	// Set content type and render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func main() {
	srv := &tsnet.Server{
		Hostname: "immich-share",
		Dir: "./ts",
	}
	ln, err := srv.ListenFunnel("tcp", ":443")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Listening on the Tailscale Funnel Hostname")

	http.HandleFunc("/", notFound)
	http.HandleFunc("/share/", handleShare)
	http.HandleFunc("/asset/", handleAsset)

	// Start the server
	// start the server	
	err = http.Serve(ln, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
