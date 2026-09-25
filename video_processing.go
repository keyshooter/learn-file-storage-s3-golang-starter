package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
)

func getVideoAspectRation(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", nil
	}

	var body FFProbeOutput
	err = json.Unmarshal(out.Bytes(), &body)
	if err != nil {
		return "", err
	}

	if len(body.Streams) == 0 {
		return "", errors.New("failed to parse video metadata")
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
	outputFilePath := filePath + ".processing"
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

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outputFilePath, nil
}
