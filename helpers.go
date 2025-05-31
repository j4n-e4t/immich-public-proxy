package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func env(key string) string {
	return os.Getenv(key)
}

func getAssetsFromShare(shareData ShareResponse, shareKey string) ([]string, []string, error) {
	var assets []Asset

	switch shareData.Type {
	case "INDIVIDUAL":
		assets = shareData.Assets
	case "ALBUM":
		if shareData.Album == nil {
			return nil, nil, fmt.Errorf("album data missing for ALBUM type share")
		}
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

	resp, err := http.Get(env("IMMICH_BASE_URL") +
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
