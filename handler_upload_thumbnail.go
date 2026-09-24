package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20 // 10MB
	r.ParseMultipartForm(int64(maxMemory))
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not read file", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse content type", err)
		return
	}
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusUnsupportedMediaType, "File not supported", nil)
		return
	}

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not access video data", err)
		return
	}
	if videoData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Video is not yours", nil)
		return
	}

	fileFormat := strings.Split(mediaType, "/")[1]
	fileNameRandomized := make([]byte, 32)
	rand.Read(fileNameRandomized)
	fileName := base64.RawURLEncoding.EncodeToString(fileNameRandomized)
	filePath := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s.%s", fileName, fileFormat))
	thumbnailFile, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInsufficientStorage, "Could not store thumbnail", err)
		return
	}
	defer thumbnailFile.Close()
	if _, err := io.Copy(thumbnailFile, file); err != nil {
		respondWithError(w, http.StatusInsufficientStorage, "Could not store thumbnail", err)
		return
	}

	thumbnailUrl := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, fileName, fileFormat)
	videoData.ThumbnailURL = &thumbnailUrl
	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not update video metadata", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoData)
}
