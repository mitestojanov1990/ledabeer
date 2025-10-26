package media_test

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"ledabeer/backend/internal/media"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThumbnail_RealGeneration_JPEG(t *testing.T) {
	// Should generate real thumbnail for JPEG
	jpegData := createTestJPEG(800, 600)

	thumbnail, err := media.GenerateThumbnail(jpegData, "image/jpeg")
	require.NoError(t, err)

	// Verify thumbnail is smaller
	assert.Less(t, len(thumbnail), len(jpegData))

	// Verify thumbnail is valid image
	img, err := decodeImage(thumbnail)
	require.NoError(t, err)

	// Verify thumbnail dimensions
	bounds := img.Bounds()
	assert.LessOrEqual(t, bounds.Dx(), 200) // Max width
	assert.LessOrEqual(t, bounds.Dy(), 200) // Max height
}

func TestThumbnail_RealGeneration_PNG(t *testing.T) {
	// Should generate real thumbnail for PNG
	pngData := createTestPNG(400, 300)

	thumbnail, err := media.GenerateThumbnail(pngData, "image/png")
	require.NoError(t, err)

	assert.Less(t, len(thumbnail), len(pngData))

	// Verify thumbnail is valid
	img, err := decodeImage(thumbnail)
	require.NoError(t, err)

	bounds := img.Bounds()
	assert.LessOrEqual(t, bounds.Dx(), 200)
	assert.LessOrEqual(t, bounds.Dy(), 200)
}

func TestThumbnail_RealGeneration_AspectRatio(t *testing.T) {
	// Should preserve aspect ratio
	imageData := createTestJPEG(800, 400) // 2:1 ratio

	thumbnail, err := media.GenerateThumbnail(imageData, "image/jpeg")
	require.NoError(t, err)

	img, err := decodeImage(thumbnail)
	require.NoError(t, err)

	bounds := img.Bounds()
	ratio := float64(bounds.Dx()) / float64(bounds.Dy())
	assert.InDelta(t, 2.0, ratio, 0.1) // Approximately 2:1
}

func TestThumbnail_RealGeneration_UnsupportedFormat(t *testing.T) {
	// Should return error for unsupported format
	data := []byte("not an image")

	_, err := media.GenerateThumbnail(data, "text/plain")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "thumbnail generation only supported for images")
}

func TestThumbnail_RealGeneration_InvalidImage(t *testing.T) {
	// Should return error for invalid image data
	invalidData := []byte("invalid image data")

	_, err := media.GenerateThumbnail(invalidData, "image/jpeg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode image")
}

// Helper function for creating test PNG data
func createTestPNG(width, height int) []byte {
	img := createTestImage(width, height)

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}
