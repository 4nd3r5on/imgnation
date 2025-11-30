package utils

import "strings"

// GetExt returns the file extension for a given MIME type
func GetExt(contentType string) string {
	switch contentType {
	// Image formats
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "image/svg+xml":
		return "svg"
	case "image/bmp":
		return "bmp"
	case "image/tiff":
		return "tiff"
	case "image/x-icon":
		return "ico"

	// Video formats
	case "video/mp4":
		return "mp4"
	case "video/quicktime":
		return "mov"
	case "video/x-msvideo":
		return "avi"
	case "video/webm":
		return "webm"
	case "video/x-matroska":
		return "mkv"
	case "video/x-flv":
		return "flv"
	case "video/3gpp":
		return "3gp"

	// Audio formats
	case "audio/mpeg":
		return "mp3"
	case "audio/wav":
		return "wav"
	case "audio/ogg":
		return "ogg"
	case "audio/mp4":
		return "m4a"
	case "audio/aac":
		return "aac"
	case "audio/flac":
		return "flac"
	case "audio/webm":
		return "weba"

	// Document formats
	case "application/pdf":
		return "pdf"
	case "application/msword":
		return "doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	case "application/vnd.ms-excel":
		return "xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return "xlsx"
	case "application/vnd.ms-powerpoint":
		return "ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return "pptx"
	case "text/plain":
		return "txt"
	case "text/rtf":
		return "rtf"

	// Code/Text formats
	case "text/html":
		return "html"
	case "text/css":
		return "css"
	case "text/javascript", "application/javascript":
		return "js"
	case "application/json":
		return "json"
	case "text/xml", "application/xml":
		return "xml"
	case "text/csv":
		return "csv"
	case "text/markdown":
		return "md"
	case "application/x-python":
		return "py"
	case "text/x-go":
		return "go"

	// Archive formats
	case "application/zip":
		return "zip"
	case "application/x-rar-compressed":
		return "rar"
	case "application/x-tar":
		return "tar"
	case "application/gzip":
		return "gz"
	case "application/x-7z-compressed":
		return "7z"

	// Streaming formats
	case "application/dash+xml":
		return "mpd"
	case "application/vnd.apple.mpegurl":
		return "m3u8"
	case "video/mp2t":
		return "ts"
	}
	return ""
}

// GetMimeType returns the MIME type for a given file extension (reverse function)
func GetMimeType(ext string) string {
	// Remove leading dot if present and convert to lowercase
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	switch ext {
	// Image formats
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "webp":
		return "image/webp"
	case "gif":
		return "image/gif"
	case "svg":
		return "image/svg+xml"
	case "bmp":
		return "image/bmp"
	case "tiff", "tif":
		return "image/tiff"
	case "ico":
		return "image/x-icon"

	// Video formats
	case "mp4":
		return "video/mp4"
	case "mov":
		return "video/quicktime"
	case "avi":
		return "video/x-msvideo"
	case "webm":
		return "video/webm"
	case "mkv":
		return "video/x-matroska"
	case "flv":
		return "video/x-flv"
	case "3gp":
		return "video/3gpp"

	// Audio formats
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	case "ogg":
		return "audio/ogg"
	case "m4a":
		return "audio/mp4"
	case "aac":
		return "audio/aac"
	case "flac":
		return "audio/flac"
	case "weba":
		return "audio/webm"

	// Document formats
	case "pdf":
		return "application/pdf"
	case "doc":
		return "application/msword"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xls":
		return "application/vnd.ms-excel"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "ppt":
		return "application/vnd.ms-powerpoint"
	case "pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case "txt":
		return "text/plain"
	case "rtf":
		return "text/rtf"

	// Code/Text formats
	case "html", "htm":
		return "text/html"
	case "css":
		return "text/css"
	case "js":
		return "text/javascript"
	case "json":
		return "application/json"
	case "xml":
		return "text/xml"
	case "csv":
		return "text/csv"
	case "md":
		return "text/markdown"
	case "py":
		return "application/x-python"
	case "go":
		return "text/x-go"

	// Archive formats
	case "zip":
		return "application/zip"
	case "rar":
		return "application/x-rar-compressed"
	case "tar":
		return "application/x-tar"
	case "gz":
		return "application/gzip"
	case "7z":
		return "application/x-7z-compressed"

	// Streaming formats
	case "mpd":
		return "application/dash+xml"
	case "m3u8":
		return "application/vnd.apple.mpegurl"
	case "ts":
		return "video/mp2t"
	}
	return ""
}
