package videoUtils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"imgnation-backend/pkg/xerr"

	"github.com/safeblock-dev/werr"
)

type VideoInfo struct {
	Width    int           `bson:"width" json:"width"`
	Height   int           `bson:"height" json:"height"`
	Duration time.Duration `bson:"duration" json:"duration"`
	Bitrate  int64         `bson:"bitrate" json:"bitrate"`
	Format   string        `bson:"format" json:"format"`
}

func checkBinaryAvailability(name string) error {
	cmd := exec.Command(name, "-version")
	if err := cmd.Run(); err != nil {
		return werr.Wrapf(err, "%s not found - video processing requires %s to be installed", name, name)
	}
	return nil
}

func CheckFFmpegAvailability() error {
	if err := checkBinaryAvailability("ffmpeg"); err != nil {
		return err
	}
	if err := checkBinaryAvailability("ffprobe"); err != nil {
		return err
	}
	return nil
}

func GetVideoInfo(ctx context.Context, videoPath string) (*VideoInfo, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, werr.Wrapf(err, "failed to run ffprobe")
	}

	var probe FFProbeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, werr.Wrapf(err, "failed to parse ffprobe output")
	}

	// Find video stream
	var videoStream *FFProbeStream
	for _, stream := range probe.Streams {
		if stream.CodecType == "video" {
			videoStream = &stream
			break
		}
	}

	if videoStream == nil {
		return nil, errors.New("no video stream found")
	}

	// Parse duration
	var duration time.Duration
	if videoStream.Duration != "" {
		if durationFloat, err := strconv.ParseFloat(videoStream.Duration, 64); err == nil {
			duration = time.Duration(durationFloat * float64(time.Second))
		}
	} else if probe.Format.Duration != "" {
		if durationFloat, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
			duration = time.Duration(durationFloat * float64(time.Second))
		}
	} else {
		return nil, werr.Wrapf(xerr.ErrInvalidArgument, "no duration information available in video metadata")
	}

	// Parse bitrate
	var bitrate int64
	if videoStream.BitRate != "" {
		if bitrateInt, err := strconv.ParseInt(videoStream.BitRate, 10, 64); err == nil {
			bitrate = bitrateInt
		}
	} else if probe.Format.BitRate != "" {
		if bitrateInt, err := strconv.ParseInt(probe.Format.BitRate, 10, 64); err == nil {
			bitrate = bitrateInt
		}
	} else {
		return nil, werr.Wrapf(xerr.ErrInvalidArgument, "no bitrate information available in video metadata")
	}

	return &VideoInfo{
		Width:    videoStream.Width,
		Height:   videoStream.Height,
		Duration: duration,
		Bitrate:  bitrate,
		Format:   strings.Split(probe.Format.FormatName, ",")[0],
	}, nil
}

// Keeps the ratio
func GenerateThumbnail(ctx context.Context, inputPath, outputPath string, maxWidth int) error {
	// Use scale filter with width:height ratio where height is -1 (auto-calculated)
	// This maintains aspect ratio while limiting width to maxWidth
	scaleFilter := fmt.Sprintf("thumbnail,scale=%d:-1", maxWidth)

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputPath,
		"-vf", scaleFilter,
		"-frames:v", "1",
		"-y", // overwrite output file
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("thumbnail generation failed: %w", err)
	}

	return nil
}

func CompressVideo(ctx context.Context, inputPath string, opts CompressionOptions) error {
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264", // Use H.264 codec
		"-preset", "medium", // Encoding speed vs compression efficiency
		"-crf", strconv.Itoa(opts.CRF),
	}

	// Add resolution scaling if specified
	if opts.Resolution != "" {
		height := parseResolution(opts.Resolution)
		if height > 0 {
			args = append(args, "-vf", fmt.Sprintf("scale=-2:%d", height))
		}
	}

	// Add bitrate limit if specified
	if opts.MaxBitrate > 0 {
		args = append(args,
			"-maxrate", fmt.Sprintf("%dk", opts.MaxBitrate),
			"-bufsize", fmt.Sprintf("%dk", opts.MaxBitrate*2),
		)
	}

	// Add format-specific options
	switch opts.OutputFormat {
	case "mp4":
		args = append(args, "-movflags", "+faststart") // Optimize for streaming
	}

	// Add output path and overwrite flag
	args = append(args, "-y", opts.OutputPath)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("video compression failed: %w", err)
	}

	return nil
}

func parseResolution(resolution string) int {
	switch strings.ToLower(resolution) {
	case "240p":
		return 240
	case "320p":
		return 320
	case "480p":
		return 480
	case "720p":
		return 720
	case "1080p":
		return 1080
	case "1440p":
		return 1440
	case "2160p", "4k":
		return 2160
	default:
		// Try to parse as number
		if num, err := strconv.Atoi(strings.TrimSuffix(resolution, "p")); err == nil {
			return num
		}
		return 0
	}
}
