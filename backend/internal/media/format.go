package media

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

// MediaMessage represents a complete media message with metadata and data
type MediaMessage struct {
	Metadata  MediaMetadata `json:"metadata"`
	Data      []byte        `json:"data"`
	Thumbnail []byte        `json:"thumbnail,omitempty"`
}

// ExtractMetadata extracts metadata from file data
func ExtractMetadata(data []byte, filename string) (MediaMetadata, error) {
	// Generate file hash
	hash := sha256.Sum256(data)

	// Detect MIME type from filename
	mimeType, err := DetectMimeType(filename)
	if err != nil {
		return MediaMetadata{}, fmt.Errorf("failed to detect MIME type: %w", err)
	}

	return MediaMetadata{
		Type: mimeType,
		Size: int64(len(data)),
		Name: filename,
		Hash: fmt.Sprintf("%x", hash),
	}, nil
}

// GenerateThumbnail generates a thumbnail for image files
func GenerateThumbnail(data []byte, mimeType string) ([]byte, error) {
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("thumbnail generation only supported for images")
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize to thumbnail dimensions (max 200x200, preserving aspect ratio)
	thumbnail := resize.Thumbnail(200, 200, img, resize.Lanczos3)

	// Encode back to original format
	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 80})
	case "png":
		err = png.Encode(&buf, thumbnail)
	default:
		// Default to JPEG for other formats
		err = jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 80})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// ValidateMimeType validates if a MIME type is supported
func ValidateMimeType(mimeType string) error {
	supportedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"video/mp4":       true,
		"video/webm":      true,
		"audio/mpeg":      true,
		"audio/ogg":       true,
		"application/pdf": true,
	}

	if !supportedTypes[mimeType] {
		return fmt.Errorf("unsupported MIME type: %s", mimeType)
	}

	return nil
}

// ValidateFileSize validates file size against limits
func ValidateFileSize(size int64) error {
	maxSize := int64(50 * 1024 * 1024) // 50MB limit

	if size > maxSize {
		return fmt.Errorf("file size %d exceeds limit %d", size, maxSize)
	}

	return nil
}

// DetectMimeType detects MIME type from filename extension
func DetectMimeType(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mp3":  "audio/mpeg",
		".ogg":  "audio/ogg",
		".pdf":  "application/pdf",
	}

	mimeType, exists := mimeTypes[ext]
	if !exists {
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	return mimeType, nil
}

// CompressData compresses data using gzip
func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// DecompressData decompresses data using gzip
func DecompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed data: %w", err)
	}

	return decompressed, nil
}

// CreateMediaMessage creates a complete media message
func CreateMediaMessage(metadata MediaMetadata, data []byte) (*MediaMessage, error) {
	// Validate MIME type
	err := ValidateMimeType(metadata.Type)
	if err != nil {
		return nil, fmt.Errorf("invalid MIME type: %w", err)
	}

	// Validate file size
	err = ValidateFileSize(metadata.Size)
	if err != nil {
		return nil, fmt.Errorf("invalid file size: %w", err)
	}

	// Generate thumbnail for images
	var thumbnail []byte
	if strings.HasPrefix(metadata.Type, "image/") {
		thumbnail, err = GenerateThumbnail(data, metadata.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
		}
	}

	return &MediaMessage{
		Metadata:  metadata,
		Data:      data,
		Thumbnail: thumbnail,
	}, nil
}

// SerializeMediaMessage serializes a media message to JSON
func SerializeMediaMessage(message *MediaMessage) ([]byte, error) {
	return json.Marshal(message)
}

// DeserializeMediaMessage deserializes a media message from JSON
func DeserializeMediaMessage(data []byte) (*MediaMessage, error) {
	var message MediaMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize media message: %w", err)
	}

	return &message, nil
}
