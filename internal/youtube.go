package internal

import (
	"bytes"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

var ErrInvalidDownloadUrl = errors.New("invalid download youtube url")
var ErrInvalidDirPath = errors.New("invalid directory")
var ErrUnsuccessfulDownload = errors.New("unsuccessful download")

type YouTubeDownloader interface {
	DownloadWav(url string, logger *slog.Logger) (string, error)
}

type ytdlpDownloader struct {
	outputDir string
}

func NewYtDlpDownloader(outputDir string, logger *slog.Logger) (YouTubeDownloader, error) {
	logger = logger.With(slog.String("output_dir", outputDir))

	st, err := os.Stat(outputDir)

	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("Invalid path")
		return nil, ErrInvalidDirPath
	}

	if !st.IsDir() {
		logger.Error("The file is not a dir")
		return nil, ErrInvalidDirPath
	}

	logger.Info("YtDlp downloader is created successfully")
	return &ytdlpDownloader{
		outputDir: outputDir,
	}, nil
}

func validateUrl(rawUrl string) bool {
	u, err := url.ParseRequestURI(rawUrl)

	if err != nil {
		return false
	}

	if u.Host != "youtu.be" {
		return false
	}

	return true
}

func (downloader *ytdlpDownloader) DownloadWav(rawUrl string, logger *slog.Logger) (string, error) {
	logger = logger.With(slog.String("url", rawUrl))

	if !validateUrl(rawUrl) {
		logger.Debug("Invalid download youtube url")
		return "", ErrInvalidDownloadUrl
	}

	cmd := exec.Command(
		"venv/bin/yt-dlp",
		"-x",
		"--audio-format", "wav",
		"-o", filepath.Join(downloader.outputDir, "%(title)s.%(ext)s"),
		"--postprocessor-args", "-ac 1",
		rawUrl,
	)

	cmdOutput, err := cmd.Output()
	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("YtDlp failed")
		return "", ErrUnsuccessfulDownload
	}

	ind1 := bytes.LastIndexByte(cmdOutput[:len(cmdOutput)-1], '\n')
	if ind1 == -1 {
		logger.With(slog.String("ytdlp_output", string(cmdOutput))).Error("No new line found")
		return "", ErrUnsuccessfulDownload
	}

	ind2 := bytes.LastIndexByte(cmdOutput[:ind1], '\n')
	if ind2 == -1 {
		logger.With(slog.String("ytdlp_output", string(cmdOutput))).Error("No new line found")
		return "", ErrUnsuccessfulDownload
	}

	destinationOutput := cmdOutput[ind2+1 : ind1]

	ind3 := bytes.IndexByte(destinationOutput, ':')
	if ind3 == -1 {
		logger.With(slog.String("dest_output", string(destinationOutput))).Error("No colon found")
		return "", ErrUnsuccessfulDownload
	}

	outputPath := string(destinationOutput[ind3+2:])
	logger.With(slog.String("output_path", outputPath)).Debug("Successful audio download")

	return outputPath, nil
}
