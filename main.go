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

	"tailscale.com/tsnet"
)


type PageData struct {
	Title         string
	PreviewURLs   []string
	ThumbnailURLs []string
	IsAlbum       bool
	AlbumName     string
	Description   string
	AssetCount    int
}

var templates *template.Template

func init() {

	var err error
	templates, err = template.ParseFiles(
		"templates/index.html",
		"templates/nav-bar.html",
		"templates/gallery.html",
		// Add any other partial templates here
	)
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
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

func getAssetsFromShare(shareData ShareResponse, shareKey string) ([]string, []string, error) {
	var assets []Asset
	
	// Determine which assets to use based on share type
	switch shareData.Type {
	case "INDIVIDUAL":
		assets = shareData.Assets
	case "ALBUM":
		if shareData.Album == nil {
			return nil, nil, fmt.Errorf("album data missing for ALBUM type share")
		}
		
		// For albums, we need to fetch the assets separately
		// since they're not included in the initial response
		albumAssets, err := fetchAlbumAssets(shareData.Album.ID, shareKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch album assets: %v", err)
		}
		assets = albumAssets
	default:
		return nil, nil, fmt.Errorf("unknown share type: %s", shareData.Type)
	}

	// Convert assets to URLs
	previewURLs := make([]string, 0, len(assets))
	thumbnailURLs := make([]string, 0, len(assets))
	
	for _, asset := range assets {
		// Skip non-image assets
		if asset.Type != "IMAGE" {
			continue
		}

		previewURLs = append(previewURLs, "/asset/"+asset.ID+"?key="+shareKey)
		thumbnailURLs = append(thumbnailURLs, "/asset/"+asset.ID+"?key="+shareKey+"&thumbnail=true")
	}

	return previewURLs, thumbnailURLs, nil
}

func fetchAlbumAssets(albumID, shareKey string) ([]Asset, error) {
	// Fetch album assets using the album ID and share key
	resp, err := http.Get(getEnv("IMMICH_BASE_URL", "http://immich:2283/") +
		"api/albums/" + albumID + "?key=" + shareKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch album assets, status: %d", resp.StatusCode)
	}

	var albumData Album
	if err := json.NewDecoder(resp.Body).Decode(&albumData); err != nil {
		return nil, err
	}

	return albumData.Assets, nil
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

	// Get assets based on share type
	previewURLs, thumbnailURLs, err := getAssetsFromShare(shareData, id)
	if err != nil {
		log.Printf("Error getting assets from share: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Prepare template data
	data := PageData{
		Title:         "Immich Gallery", 
		PreviewURLs:   previewURLs,
		ThumbnailURLs: thumbnailURLs,
		IsAlbum:       shareData.Type == "ALBUM",
	}

	// Add album-specific data if it's an album
	if shareData.Type == "ALBUM" && shareData.Album != nil {
		data.Title = shareData.Album.AlbumName
		data.AlbumName = shareData.Album.AlbumName
		data.Description = shareData.Album.Description
		data.AssetCount = shareData.Album.AssetCount
	}

	// Set content type and render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := templates.Execute(w, data); err != nil {
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
		Hostname: "immich-share-dev",
		Dir:      "./ts",
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