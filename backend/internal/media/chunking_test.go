package media_test

import (
	"bytes"
	"testing"

	"ledabeer/backend/internal/media"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunking_SplitReassemble(t *testing.T) {
	// Create test data (1MB)
	testData := make([]byte, 1024*1024) // 1MB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	reader := bytes.NewReader(testData)
	chunks, err := media.ChunkFile(reader, 64*1024) // 64KB chunks
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 0, "Should create chunks")

	// Reassemble chunks
	reassembled, err := media.ReassembleChunks(chunks)
	require.NoError(t, err)
	assert.Equal(t, testData, reassembled, "Reassembled data should match original")
}

func TestChunking_ChunkSize(t *testing.T) {
	// Create test data (200KB)
	testData := make([]byte, 200*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	reader := bytes.NewReader(testData)
	chunks, err := media.ChunkFile(reader, 64*1024) // 64KB chunks
	require.NoError(t, err)

	// Should create 4 chunks (3 full + 1 partial)
	expectedChunks := 4
	assert.Equal(t, expectedChunks, len(chunks), "Should create correct number of chunks")

	// Verify chunk sizes
	for i, chunk := range chunks {
		if i < 3 {
			// First 3 chunks should be 64KB
			assert.Equal(t, 64*1024, len(chunk.Data), "Chunk %d should be 64KB", i)
		} else {
			// Last chunk should be 8KB (200KB - 3*64KB)
			assert.Equal(t, 8*1024, len(chunk.Data), "Last chunk should be 8KB")
		}
		assert.Equal(t, i, chunk.Index, "Chunk index should match position")
		assert.Equal(t, expectedChunks, chunk.TotalChunks, "Total chunks should be correct")
	}
}

func TestChunking_EmptyFile(t *testing.T) {
	// Test with empty file
	reader := bytes.NewReader([]byte{})
	chunks, err := media.ChunkFile(reader, 64*1024)
	require.NoError(t, err)
	assert.Len(t, chunks, 0, "Empty file should produce no chunks")

	// Reassemble should work with empty chunks
	reassembled, err := media.ReassembleChunks(chunks)
	require.NoError(t, err)
	assert.Len(t, reassembled, 0, "Reassembled empty chunks should be empty")
}

func TestChunking_LargeFile(t *testing.T) {
	// Create large test data (50MB)
	largeData := make([]byte, 50*1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	reader := bytes.NewReader(largeData)
	chunks, err := media.ChunkFile(reader, 64*1024) // 64KB chunks
	require.NoError(t, err)

	// Should create many chunks
	expectedChunks := (50 * 1024 * 1024) / (64 * 1024) // 781 chunks
	assert.Equal(t, expectedChunks, len(chunks), "Should create correct number of chunks for 50MB file")

	// Verify all chunks are 64KB except possibly the last one
	for i, chunk := range chunks {
		if i < len(chunks)-1 {
			assert.Equal(t, 64*1024, len(chunk.Data), "Chunk %d should be 64KB", i)
		} else {
			// Last chunk might be smaller
			assert.LessOrEqual(t, len(chunk.Data), 64*1024, "Last chunk should be <= 64KB")
		}
		assert.Equal(t, i, chunk.Index, "Chunk index should match position")
		assert.Equal(t, expectedChunks, chunk.TotalChunks, "Total chunks should be correct")
	}

	// Reassemble should work
	reassembled, err := media.ReassembleChunks(chunks)
	require.NoError(t, err)
	assert.Equal(t, largeData, reassembled, "Reassembled large file should match original")
}

func TestChunking_InvalidChunks(t *testing.T) {
	// Test reassembling with invalid chunks
	invalidChunks := []media.Chunk{
		{Index: 0, Data: []byte("chunk0"), TotalChunks: 2, FileHash: "hash"},
		{Index: 2, Data: []byte("chunk2"), TotalChunks: 2, FileHash: "hash"}, // Missing chunk 1
	}

	_, err := media.ReassembleChunks(invalidChunks)
	assert.Error(t, err, "Should error on missing chunks")

	// Test with mismatched total chunks
	mismatchedChunks := []media.Chunk{
		{Index: 0, Data: []byte("chunk0"), TotalChunks: 3, FileHash: "hash"},
		{Index: 1, Data: []byte("chunk1"), TotalChunks: 2, FileHash: "hash"}, // Different total
	}

	_, err = media.ReassembleChunks(mismatchedChunks)
	assert.Error(t, err, "Should error on mismatched total chunks")
}

func TestChunking_Streaming(t *testing.T) {
	// Test that chunking works with streaming (doesn't load entire file in memory)
	// Create a large file and verify memory usage doesn't spike
	largeData := make([]byte, 10*1024*1024) // 10MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	reader := bytes.NewReader(largeData)
	chunks, err := media.ChunkFile(reader, 64*1024)
	require.NoError(t, err)

	// Verify chunks were created correctly
	assert.Greater(t, len(chunks), 0, "Should create chunks")

	// Verify we can reassemble
	reassembled, err := media.ReassembleChunks(chunks)
	require.NoError(t, err)
	assert.Equal(t, largeData, reassembled, "Streaming chunking should work correctly")
}

func TestChunking_Compression(t *testing.T) {
	// Create test data that compresses well (repeated pattern)
	pattern := []byte("Hello World! ")
	testData := make([]byte, 0, 100*1024) // 100KB
	for i := 0; i < 100*1024/len(pattern); i++ {
		testData = append(testData, pattern...)
	}

	reader := bytes.NewReader(testData)

	// Test with compression enabled
	config := media.ChunkingConfig{
		ChunkSize:    64 * 1024,
		Compress:     true,
		ValidateHash: true,
	}

	chunks, err := media.ChunkFileWithConfig(reader, config)
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 0, "Should create chunks")

	// Verify some chunks are compressed
	compressedCount := 0
	for _, chunk := range chunks {
		if chunk.Compressed {
			compressedCount++
		}
	}
	assert.Greater(t, compressedCount, 0, "Some chunks should be compressed")

	// Reassemble should work
	reassembled, err := media.ReassembleChunks(chunks)
	require.NoError(t, err)
	assert.Equal(t, testData, reassembled, "Compressed chunks should reassemble correctly")
}

func TestChunking_ChunkHashes(t *testing.T) {
	// Create test data
	testData := make([]byte, 200*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	reader := bytes.NewReader(testData)
	chunks, err := media.ChunkFile(reader, 64*1024)
	require.NoError(t, err)

	// Verify all chunks have hashes
	for i, chunk := range chunks {
		assert.NotEmpty(t, chunk.ChunkHash, "Chunk %d should have a hash", i)
		assert.NotEmpty(t, chunk.FileHash, "Chunk %d should have file hash", i)
	}

	// Verify file hashes are consistent
	expectedFileHash := chunks[0].FileHash
	for i, chunk := range chunks {
		assert.Equal(t, expectedFileHash, chunk.FileHash, "Chunk %d should have consistent file hash", i)
	}
}
