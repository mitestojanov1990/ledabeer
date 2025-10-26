package media

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"sort"
)

// Chunk represents a piece of a file
type Chunk struct {
	Index       int    `json:"index"`
	Data        []byte `json:"data"`
	TotalChunks int    `json:"total_chunks"`
	FileHash    string `json:"file_hash"`
	Compressed  bool   `json:"compressed"`
	ChunkHash   string `json:"chunk_hash"`
}

// ChunkingConfig holds configuration for file chunking
type ChunkingConfig struct {
	ChunkSize    int  `json:"chunk_size"`
	Compress     bool `json:"compress"`
	ValidateHash bool `json:"validate_hash"`
}

// ChunkFile splits a file into chunks of the specified size
func ChunkFile(reader io.Reader, chunkSize int) ([]Chunk, error) {
	config := ChunkingConfig{
		ChunkSize:    chunkSize,
		Compress:     false,
		ValidateHash: true,
	}
	return ChunkFileWithConfig(reader, config)
}

// ChunkFileWithConfig splits a file into chunks with the given configuration
func ChunkFileWithConfig(reader io.Reader, config ChunkingConfig) ([]Chunk, error) {
	if config.ChunkSize <= 0 {
		return nil, fmt.Errorf("chunk size must be positive")
	}

	var chunks []Chunk
	buffer := make([]byte, config.ChunkSize)
	index := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}

		// Create chunk with actual data read
		chunkData := make([]byte, n)
		copy(chunkData, buffer[:n])

		// Compress if requested
		var compressed bool
		if config.Compress {
			compressedData, err := compressData(chunkData)
			if err != nil {
				return nil, fmt.Errorf("error compressing chunk %d: %w", index, err)
			}
			// Only use compressed data if it's actually smaller
			if len(compressedData) < len(chunkData) {
				chunkData = compressedData
				compressed = true
			}
		}

		// Calculate chunk hash
		chunkHash := calculateChunkHash(chunkData)

		chunk := Chunk{
			Index:       index,
			Data:        chunkData,
			TotalChunks: 0,  // Will be set after all chunks are created
			FileHash:    "", // Will be calculated later
			Compressed:  compressed,
			ChunkHash:   chunkHash,
		}

		chunks = append(chunks, chunk)
		index++

		// If we read less than chunkSize, we've reached the end
		if n < config.ChunkSize {
			break
		}
	}

	// Set total chunks for all chunks
	totalChunks := len(chunks)
	for i := range chunks {
		chunks[i].TotalChunks = totalChunks
	}

	// Calculate file hash from all data
	fileHash, err := calculateFileHash(chunks)
	if err != nil {
		return nil, fmt.Errorf("error calculating file hash: %w", err)
	}

	// Set file hash for all chunks
	for i := range chunks {
		chunks[i].FileHash = fileHash
	}

	return chunks, nil
}

// ReassembleChunks combines chunks back into the original file
func ReassembleChunks(chunks []Chunk) ([]byte, error) {
	if len(chunks) == 0 {
		return []byte{}, nil
	}

	// Validate chunks
	if err := validateChunks(chunks); err != nil {
		return nil, err
	}

	// Sort chunks by index
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Index < chunks[j].Index
	})

	// Reassemble data, decompressing if necessary
	result := make([]byte, 0)
	for _, chunk := range chunks {
		var chunkData []byte
		var err error

		if chunk.Compressed {
			chunkData, err = decompressData(chunk.Data)
			if err != nil {
				return nil, fmt.Errorf("error decompressing chunk %d: %w", chunk.Index, err)
			}
		} else {
			chunkData = chunk.Data
		}

		result = append(result, chunkData...)
	}

	return result, nil
}

// validateChunks checks that chunks are valid and complete
func validateChunks(chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// Check that all chunks have the same total count and file hash
	expectedTotal := chunks[0].TotalChunks
	expectedHash := chunks[0].FileHash

	for i, chunk := range chunks {
		if chunk.TotalChunks != expectedTotal {
			return fmt.Errorf("chunk %d has mismatched total chunks: expected %d, got %d",
				i, expectedTotal, chunk.TotalChunks)
		}
		if chunk.FileHash != expectedHash {
			return fmt.Errorf("chunk %d has mismatched file hash: expected %s, got %s",
				i, expectedHash, chunk.FileHash)
		}
	}

	// Check that we have all expected chunks
	expectedIndices := make(map[int]bool)
	for i := 0; i < expectedTotal; i++ {
		expectedIndices[i] = true
	}

	actualIndices := make(map[int]bool)
	for _, chunk := range chunks {
		actualIndices[chunk.Index] = true
	}

	// Check for missing chunks
	for expectedIndex := range expectedIndices {
		if !actualIndices[expectedIndex] {
			return fmt.Errorf("missing chunk at index %d", expectedIndex)
		}
	}

	// Check for duplicate chunks
	if len(actualIndices) != len(chunks) {
		return fmt.Errorf("duplicate chunk indices found")
	}

	return nil
}

// calculateFileHash computes SHA-256 hash of all chunk data
func calculateFileHash(chunks []Chunk) (string, error) {
	hasher := sha256.New()

	// Sort chunks by index to ensure consistent hashing
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Index < chunks[j].Index
	})

	for _, chunk := range chunks {
		// Use original data for hashing (decompress if necessary)
		var dataToHash []byte
		if chunk.Compressed {
			decompressed, err := decompressData(chunk.Data)
			if err != nil {
				return "", fmt.Errorf("error decompressing chunk %d for hashing: %w", chunk.Index, err)
			}
			dataToHash = decompressed
		} else {
			dataToHash = chunk.Data
		}

		_, err := hasher.Write(dataToHash)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// compressData compresses data using gzip
func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData decompresses data using gzip
func decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// calculateChunkHash computes SHA-256 hash of chunk data
func calculateChunkHash(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
