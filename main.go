package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"tailscale.com/tsnet"
)

type PageData struct {
	Title         string
	PreviewURLs   []string
	ThumbnailURLs []string
	IsAlbum       bool
}

var templates *template.Template

func init() {

	var err error
	templates, err = template.ParseFiles(
		"templates/index.html",
		"templates/nav-bar.html",
		"templates/gallery.html",
	)
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
}

func main() {
	srv := &tsnet.Server{
		Hostname: env("TS_HOSTNAME"),
		Dir:      "./ts",
	}
	ln, err := srv.ListenFunnel("tcp", ":443")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening on the Tailscale Funnel Hostname")

	http.HandleFunc("/", notFoundHandler)
	http.HandleFunc("/share/", shareHandler)
	http.HandleFunc("/asset/", assetHandler)

	// Start the server
	err = http.Serve(ln, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// http handlers

func assetHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/asset/")
	key := r.URL.Query().Get("key")

	immichURL := ""
	if r.URL.Query().Get("thumbnail") == "true" {
		immichURL = env("IMMICH_BASE_URL") +
			"api/assets/" + id + "/thumbnail?size=thumbnail&key=" + key
	} else {
		immichURL = env("IMMICH_BASE_URL") +
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

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func shareHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/share/")

	resp, err := http.Get(env("IMMICH_BASE_URL") +
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

	if shareData.Type == "ALBUM" && shareData.Album != nil {
		data.Title = shareData.Album.AlbumName
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

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
