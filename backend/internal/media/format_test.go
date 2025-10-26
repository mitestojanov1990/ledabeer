package media_test

import (
	"testing"

	"ledabeer/backend/internal/media"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaMessage_Metadata(t *testing.T) {
	// Test metadata extraction from different file types
	imageData := []byte("fake image data")
	metadata, err := media.ExtractMetadata(imageData, "test-image.jpg")
	require.NoError(t, err)

	assert.Equal(t, "image/jpeg", metadata.Type)
	assert.Equal(t, int64(len(imageData)), metadata.Size)
	assert.Equal(t, "test-image.jpg", metadata.Name)
	assert.NotEmpty(t, metadata.Hash, "Should generate file hash")
}

func TestMediaMessage_Thumbnail(t *testing.T) {
	// Test thumbnail generation for images
	imageData := []byte("fake image data")
	thumbnail, err := media.GenerateThumbnail(imageData, "image/jpeg")
	require.NoError(t, err)

	assert.NotNil(t, thumbnail, "Should generate thumbnail")
	assert.Greater(t, len(thumbnail), 0, "Thumbnail should not be empty")
	assert.Less(t, len(thumbnail), len(imageData), "Thumbnail should be smaller than original")
}

func TestMediaMessage_Validation(t *testing.T) {
	// Test MIME type validation
	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"video/mp4",
		"video/webm",
		"audio/mpeg",
		"audio/ogg",
		"application/pdf",
	}

	for _, mimeType := range validTypes {
		err := media.ValidateMimeType(mimeType)
		assert.NoError(t, err, "Should accept valid MIME type: %s", mimeType)
	}

	// Test invalid MIME types
	invalidTypes := []string{
		"text/plain",
		"application/octet-stream",
		"image/bmp", // Unsupported format
		"video/avi", // Unsupported format
		"invalid/type",
	}

	for _, mimeType := range invalidTypes {
		err := media.ValidateMimeType(mimeType)
		assert.Error(t, err, "Should reject invalid MIME type: %s", mimeType)
	}
}

func TestMediaMessage_FileSizeValidation(t *testing.T) {
	// Test file size validation
	smallFile := make([]byte, 1024)          // 1KB
	largeFile := make([]byte, 100*1024*1024) // 100MB

	err := media.ValidateFileSize(int64(len(smallFile)))
	assert.NoError(t, err, "Should accept small file")

	err = media.ValidateFileSize(int64(len(largeFile)))
	assert.Error(t, err, "Should reject large file")
	assert.Contains(t, err.Error(), "file size", "Should mention file size")
}

func TestMediaMessage_FormatDetection(t *testing.T) {
	// Test automatic format detection
	testCases := []struct {
		filename string
		expected string
	}{
		{"image.jpg", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.png", "image/png"},
		{"image.gif", "image/gif"},
		{"video.mp4", "video/mp4"},
		{"video.webm", "video/webm"},
		{"audio.mp3", "audio/mpeg"},
		{"audio.ogg", "audio/ogg"},
		{"document.pdf", "application/pdf"},
	}

	for _, tc := range testCases {
		detected, err := media.DetectMimeType(tc.filename)
		require.NoError(t, err, "Should detect MIME type for %s", tc.filename)
		assert.Equal(t, tc.expected, detected, "Should detect correct MIME type for %s", tc.filename)
	}
}

func TestMediaMessage_Compression(t *testing.T) {
	// Test compression for different file types
	// Use larger data that will actually compress
	testData := make([]byte, 1000)
	for i := range testData {
		testData[i] = byte(i % 256) // Repetitive pattern for good compression
	}

	// Test compression
	compressed, err := media.CompressData(testData)
	require.NoError(t, err, "Should compress data")
	assert.Less(t, len(compressed), len(testData), "Compressed data should be smaller")

	// Test decompression
	decompressed, err := media.DecompressData(compressed)
	require.NoError(t, err, "Should decompress data")
	assert.Equal(t, testData, decompressed, "Decompressed data should match original")
}

func TestMediaMessage_MessageFormat(t *testing.T) {
	// Test complete message format
	imageData := []byte("fake image data")
	metadata := media.MediaMetadata{
		Type: "image/jpeg",
		Size: int64(len(imageData)),
		Name: "test-image.jpg",
	}

	// Create media message
	message, err := media.CreateMediaMessage(metadata, imageData)
	require.NoError(t, err, "Should create media message")

	assert.NotNil(t, message, "Message should not be nil")
	assert.Equal(t, metadata.Type, message.Metadata.Type)
	assert.Equal(t, metadata.Size, message.Metadata.Size)
	assert.Equal(t, metadata.Name, message.Metadata.Name)
	assert.NotEmpty(t, message.Data, "Message should contain data")
	assert.NotEmpty(t, message.Thumbnail, "Message should have thumbnail")
}

func TestMediaMessage_Serialization(t *testing.T) {
	// Test message serialization/deserialization
	originalData := []byte("test media data")
	metadata := media.MediaMetadata{
		Type: "image/jpeg",
		Size: int64(len(originalData)),
		Name: "test.jpg",
	}

	message, err := media.CreateMediaMessage(metadata, originalData)
	require.NoError(t, err, "Should create media message")

	// Serialize message
	serialized, err := media.SerializeMediaMessage(message)
	require.NoError(t, err, "Should serialize message")
	assert.NotEmpty(t, serialized, "Serialized message should not be empty")

	// Deserialize message
	deserialized, err := media.DeserializeMediaMessage(serialized)
	require.NoError(t, err, "Should deserialize message")

	assert.Equal(t, message.Metadata.Type, deserialized.Metadata.Type)
	assert.Equal(t, message.Metadata.Size, deserialized.Metadata.Size)
	assert.Equal(t, message.Metadata.Name, deserialized.Metadata.Name)
	assert.Equal(t, message.Data, deserialized.Data)
	assert.Equal(t, message.Thumbnail, deserialized.Thumbnail)
}
