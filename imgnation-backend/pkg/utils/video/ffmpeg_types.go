package videoUtils

// FFProbe output structures
type FFProbeOutput struct {
	Streams []FFProbeStream `json:"streams"`
	Format  FFProbeFormat   `json:"format"`
}

type FFProbeStream struct {
	Index       int    `json:"index"`
	CodecName   string `json:"codec_name"`
	CodecType   string `json:"codec_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Duration    string `json:"duration"`
	BitRate     string `json:"bit_rate"`
	PixelFormat string `json:"pix_fmt"`
}

type FFProbeFormat struct {
	Filename   string `json:"filename"`
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
}

type CompressionOptions struct {
	OutputPath   string
	Resolution   string // e.g., "720p", "1080p"
	CRF          int    // Constant Rate Factor
	MaxBitrate   int    // in kbps
	OutputFormat string // e.g., "mp4", "mov"
}
