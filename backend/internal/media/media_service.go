package media

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"ledabeer/backend/internal/user"
	"github.com/libp2p/go-libp2p/core/peer"
)

// MediaService handles media file operations for authenticated users
type MediaService struct {
	userManager *user.UserManager
	// In a real implementation, you'd have IPFS client, storage backend, etc.
}

// NewMediaService creates a new MediaService
func NewMediaService(userManager *user.UserManager) *MediaService {
	return &MediaService{
		userManager: userManager,
	}
}

// MediaFile represents a media file
type MediaFile struct {
	FileID      string    `json:"file_id"`
	UserID      string    `json:"user_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	FileHash    string    `json:"file_hash"`
	IPFSHash    string    `json:"ipfs_hash"`
	UploadedAt  time.Time `json:"uploaded_at"`
	IsEncrypted bool      `json:"is_encrypted"`
	EncryptionKey []byte  `json:"-"` // Never expose in JSON
}

// MediaUploadRequest represents a media upload request
type MediaUploadRequest struct {
	UserID      string                `json:"user_id"`
	File        multipart.File        `json:"-"`
	FileHeader  *multipart.FileHeader `json:"-"`
	IsEncrypted bool                  `json:"is_encrypted"`
	Recipients  []string              `json:"recipients"` // For encrypted files
}

// MediaUploadResponse represents the response for media upload
type MediaUploadResponse struct {
	FileID      string    `json:"file_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	FileHash    string    `json:"file_hash"`
	IPFSHash    string    `json:"ipfs_hash"`
	DownloadURL string    `json:"download_url"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// MediaDownloadRequest represents a media download request
type MediaDownloadRequest struct {
	FileID string `json:"file_id"`
	UserID string `json:"user_id"`
}

// MediaDownloadResponse represents the response for media download
type MediaDownloadResponse struct {
	FileID      string    `json:"file_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	FileData    []byte    `json:"file_data"`
	DownloadURL string    `json:"download_url"`
}

// UploadMedia uploads a media file for a user
func (ms *MediaService) UploadMedia(req *MediaUploadRequest) (*MediaUploadResponse, error) {
	// Convert user ID to peer.ID
	userPeerID, err := peer.Decode(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	// Validate user exists
	_, err = ms.userManager.GetUserByPeerID(userPeerID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Validate file
	if req.File == nil {
		return nil, errors.New("file is required")
	}

	// Read file data
	fileData, err := io.ReadAll(req.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Generate file hash
	fileHash := sha256.Sum256(fileData)
	fileHashStr := hex.EncodeToString(fileHash[:])

	// Generate file ID
	fileID := fmt.Sprintf("file_%s_%d", req.UserID, time.Now().Unix())

	// Determine MIME type
	mimeType := req.FileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(req.FileHeader.Filename))

	// Validate file type
	if !ms.isAllowedFileType(ext, mimeType) {
		return nil, errors.New("file type not allowed")
	}

	// Check file size (max 100MB)
	maxSize := int64(100 * 1024 * 1024) // 100MB
	if int64(len(fileData)) > maxSize {
		return nil, errors.New("file too large (max 100MB)")
	}

	// If encrypted, encrypt the file data
	var encryptedData []byte
	var encryptionKey []byte
	if req.IsEncrypted {
		encryptedData, encryptionKey, err = ms.encryptFileData(fileData, req.Recipients)
		if err != nil {
			return nil, fmt.Errorf("file encryption failed: %w", err)
		}
	} else {
		encryptedData = fileData
	}

	// Upload to IPFS (simplified - in real implementation, use IPFS client)
	ipfsHash, err := ms.uploadToIPFS(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("IPFS upload failed: %w", err)
	}

	// Create media file record
	mediaFile := &MediaFile{
		FileID:       fileID,
		UserID:       req.UserID,
		FileName:     req.FileHeader.Filename,
		FileSize:     int64(len(fileData)),
		MimeType:     mimeType,
		FileHash:     fileHashStr,
		IPFSHash:     ipfsHash,
		UploadedAt:   time.Now(),
		IsEncrypted:  req.IsEncrypted,
		EncryptionKey: encryptionKey,
	}

	// Store media file record (in real implementation, store in database)
	if err := ms.storeMediaFile(mediaFile); err != nil {
		return nil, fmt.Errorf("failed to store media file: %w", err)
	}

	// Generate download URL
	downloadURL := fmt.Sprintf("/api/media/download/%s", fileID)

	return &MediaUploadResponse{
		FileID:      fileID,
		FileName:    req.FileHeader.Filename,
		FileSize:    int64(len(fileData)),
		MimeType:    mimeType,
		FileHash:    fileHashStr,
		IPFSHash:    ipfsHash,
		DownloadURL: downloadURL,
		UploadedAt:  time.Now(),
	}, nil
}

// DownloadMedia downloads a media file for a user
func (ms *MediaService) DownloadMedia(req *MediaDownloadRequest) (*MediaDownloadResponse, error) {
	// Get media file record
	mediaFile, err := ms.getMediaFile(req.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get media file: %w", err)
	}

	// Validate user has access to this file
	if mediaFile.UserID != req.UserID {
		return nil, errors.New("access denied")
	}

	// Download from IPFS (simplified)
	fileData, err := ms.downloadFromIPFS(mediaFile.IPFSHash)
	if err != nil {
		return nil, fmt.Errorf("IPFS download failed: %w", err)
	}

	// If encrypted, decrypt the file data
	var decryptedData []byte
	if mediaFile.IsEncrypted {
		decryptedData, err = ms.decryptFileData(fileData, mediaFile.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("file decryption failed: %w", err)
		}
	} else {
		decryptedData = fileData
	}

	// Generate download URL
	downloadURL := fmt.Sprintf("/api/media/download/%s", req.FileID)

	return &MediaDownloadResponse{
		FileID:      mediaFile.FileID,
		FileName:    mediaFile.FileName,
		FileSize:    mediaFile.FileSize,
		MimeType:    mediaFile.MimeType,
		FileData:    decryptedData,
		DownloadURL: downloadURL,
	}, nil
}

// GetUserMedia gets all media files for a user
func (ms *MediaService) GetUserMedia(userID string) ([]*MediaFile, error) {
	// Convert user ID to peer.ID
	userPeerID, err := peer.Decode(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	// Validate user exists
	_, err = ms.userManager.GetUserByPeerID(userPeerID)
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Get user's media files (in real implementation, query database)
	return ms.getUserMediaFiles(userID)
}

// isAllowedFileType checks if a file type is allowed
func (ms *MediaService) isAllowedFileType(ext, mimeType string) bool {
	allowedTypes := map[string][]string{
		".jpg":  {"image/jpeg"},
		".jpeg": {"image/jpeg"},
		".png":  {"image/png"},
		".gif":  {"image/gif"},
		".mp4":  {"video/mp4"},
		".avi":  {"video/x-msvideo"},
		".mov":  {"video/quicktime"},
		".mp3":  {"audio/mpeg"},
		".wav":  {"audio/wav"},
		".pdf":  {"application/pdf"},
		".doc":  {"application/msword"},
		".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		".txt":  {"text/plain"},
	}

	allowedMimeTypes, exists := allowedTypes[ext]
	if !exists {
		return false
	}

	for _, allowedMime := range allowedMimeTypes {
		if mimeType == allowedMime {
			return true
		}
	}

	return false
}

// encryptFileData encrypts file data for specific recipients
func (ms *MediaService) encryptFileData(data []byte, recipients []string) ([]byte, []byte, error) {
	// In a real implementation, this would use the E2EE service
	// to encrypt the file for each recipient
	// For now, return the data as-is with a placeholder key
	encryptionKey := []byte("encryption_key_placeholder")
	return data, encryptionKey, nil
}

// decryptFileData decrypts file data
func (ms *MediaService) decryptFileData(data []byte, key []byte) ([]byte, error) {
	// In a real implementation, this would use the E2EE service
	// to decrypt the file
	// For now, return the data as-is
	return data, nil
}

// uploadToIPFS uploads data to IPFS (simplified)
func (ms *MediaService) uploadToIPFS(data []byte) (string, error) {
	// In a real implementation, this would use IPFS client
	// For now, return a placeholder hash
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// downloadFromIPFS downloads data from IPFS (simplified)
func (ms *MediaService) downloadFromIPFS(hash string) ([]byte, error) {
	// In a real implementation, this would use IPFS client
	// For now, return empty data
	return []byte{}, nil
}

// storeMediaFile stores media file record (simplified)
func (ms *MediaService) storeMediaFile(mediaFile *MediaFile) error {
	// In a real implementation, this would store in database
	return nil
}

// getMediaFile gets media file record (simplified)
func (ms *MediaService) getMediaFile(fileID string) (*MediaFile, error) {
	// In a real implementation, this would query database
	return nil, errors.New("not implemented")
}

// getUserMediaFiles gets user's media files (simplified)
func (ms *MediaService) getUserMediaFiles(userID string) ([]*MediaFile, error) {
	// In a real implementation, this would query database
	return []*MediaFile{}, nil
}
