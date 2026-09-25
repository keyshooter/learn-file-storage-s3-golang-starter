package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func getVideoAspectRation(filePath string) (string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-print_format",
		"json",
		"-show_streams",
		filePath,
	)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffprobe error: %v", err)
	}

	var body struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	if err = json.Unmarshal(out.Bytes(), &body); err != nil {
		return "", fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	if len(body.Streams) == 0 {
		return "", errors.New("no video streams found")
	}

	videoData := body.Streams[0]
	ratio := videoData.Width / videoData.Height

	if ratio == 1 {
		return "16:9", nil
	} else if ratio == 0 {
		return "9:16", nil
	}
	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outputFilePath := fmt.Sprintf("%s.processing", filePath)

	cmd := exec.Command(
		"ffmpeg",
		"-i",
		filePath,
		"-c",
		"copy",
		"-movflags",
		"faststart",
		"-f",
		"mp4",
		outputFilePath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to process video file: %s, %v", stderr.String(), err)
	}

	fileInfo, err := os.Stat(outputFilePath)
	if err != nil {
		return "", fmt.Errorf("unable to stat file: %v", err)
	}
	if fileInfo.Size() == 0 {
		return "", errors.New("processed file is empty")
	}
	return outputFilePath, nil
}
