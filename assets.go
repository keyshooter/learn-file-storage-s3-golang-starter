package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	fileNameRandomized := make([]byte, 32)
	_, err := rand.Read(fileNameRandomized)
	if err != nil {
		panic("failed to generate random bytes")
	}
	fileName := base64.RawURLEncoding.EncodeToString(fileNameRandomized)
	extension := mediaTypeToExtension(mediaType)
	return fmt.Sprintf("%s%s", fileName, extension)
}

func mediaTypeToExtension(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func (cfg apiConfig) getObjectURL(key string) string {
	return fmt.Sprintf("%s,%s", cfg.s3Bucket, key)
}

func (cfg apiConfig) getAssetDiskPAth(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignedClient := s3.NewPresignClient(s3Client)
	presignedRequest, err := presignedClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %v", err)
	}
	return presignedRequest.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	videoParts := strings.Split(*video.VideoURL, ",")
	log.Println("parts: ", videoParts)
	if len(videoParts) == 0 {
		return video, nil
	}
	videoURL, err := generatePresignedURL(cfg.s3Client, videoParts[0], videoParts[1], 5*time.Minute)
	if err != nil {
		return database.Video{}, fmt.Errorf("failed to get presigned url: %v", err)
	}

	video.VideoURL = &videoURL
	return video, nil
}
