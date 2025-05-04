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
	DownloadWav(url string, logger *slog.Logger) (string, string, error)
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

func ValidateUrl(rawUrl string) bool {
	u, err := url.ParseRequestURI(rawUrl)

	if err != nil {
		return false
	}

	if u.Host != "youtu.be" {
		return false
	}

	return true
}

func StripUrl(raw string) (string, bool) {
	u, err := url.ParseRequestURI(raw)

	if err != nil {
		return "", false
	}

	u.RawQuery = ""
	return u.String(), true
}

func (downloader *ytdlpDownloader) DownloadWav(rawUrl string, logger *slog.Logger) (string, string, error) {
	if !ValidateUrl(rawUrl) {
		logger.Debug("Invalid download youtube url")
		return "", "", ErrInvalidDownloadUrl
	}

	cmd := exec.Command(
		"venv/bin/yt-dlp",
		"--print", `"%(title)s"`,
		"--print", `after_move:"%(filepath)s"`,
		"-x",
		"--audio-format", "wav",
		"-o", filepath.Join(downloader.outputDir, "%(title)s.%(ext)s"),
		"--postprocessor-args", "-ac 1",
		rawUrl,
	)

	cmdOutput, err := cmd.Output()
	if err != nil {
		logger.With(slog.String("err", err.Error())).Error("YtDlp failed")
		return "", "", ErrUnsuccessfulDownload
	}

	// skip last newline with len
	ind := bytes.IndexByte(cmdOutput, '\n')
	if ind == -1 {
		logger.With(slog.String("ytdlp_output", string(cmdOutput))).Error("No new line found")
		return "", "", ErrUnsuccessfulDownload
	}

	title := string(cmdOutput[1 : ind-1])
	// skip last newline with len
	outputPath := string(cmdOutput[ind+2 : len(cmdOutput)-2])

	logger.With(
		slog.String("title", title),
		slog.String("output_path", outputPath),
	).Debug("Successful audio download")

	return title, outputPath, nil
}
