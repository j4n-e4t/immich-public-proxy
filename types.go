package main

import (
	"time"
)

type ShareResponse struct {
	AllowDownload bool        `json:"allowDownload"`
	AllowUpload   bool        `json:"allowUpload"`
	Assets        []Asset     `json:"assets"`
	Album         *Album      `json:"album,omitempty"` // Only present for ALBUM type
	CreatedAt     time.Time   `json:"createdAt"`
	Description   interface{} `json:"description"` // Can be null
	ExpiresAt     interface{} `json:"expiresAt"`   // Can be null
	ID            string      `json:"id"`
	Key           string      `json:"key"`
	Password      string      `json:"password"`
	ShowMetadata  bool        `json:"showMetadata"`
	Type          string      `json:"type"` // "INDIVIDUAL" or "ALBUM"
	UserID        string      `json:"userId"`
}

type Album struct {
	AlbumName               string      `json:"albumName"`
	Description             string      `json:"description"`
	AlbumThumbnailAssetID   string      `json:"albumThumbnailAssetId"`
	CreatedAt               time.Time   `json:"createdAt"`
	UpdatedAt               time.Time   `json:"updatedAt"`
	ID                      string      `json:"id"`
	OwnerID                 string      `json:"ownerId"`
	Owner                   Owner       `json:"owner"`
	AlbumUsers              []interface{} `json:"albumUsers"`
	Shared                  bool        `json:"shared"`
	HasSharedLink           bool        `json:"hasSharedLink"`
	StartDate               time.Time   `json:"startDate"`
	EndDate                 time.Time   `json:"endDate"`
	Assets                  []Asset     `json:"assets"`
	AssetCount              int         `json:"assetCount"`
	IsActivityEnabled       bool        `json:"isActivityEnabled"`
	Order                   string      `json:"order"`
}

type Owner struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	Name               string    `json:"name"`
	ProfileImagePath   string    `json:"profileImagePath"`
	AvatarColor        string    `json:"avatarColor"`
	ProfileChangedAt   time.Time `json:"profileChangedAt"`
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