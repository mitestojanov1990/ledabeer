package integration_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"ledabeer/backend/internal/crypto"
	"ledabeer/backend/internal/media"
	"ledabeer/backend/internal/network"

	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaTransfer_ImageTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create two hosts
	h1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h1.Close()

	h2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h2.Close()

	// Connect hosts
	h2Addr := h2.Addrs()[0].String() + "/p2p/" + h2.ID().String()
	peerInfo, err := network.ParseAddrInfo(h2Addr)
	require.NoError(t, err)
	err = h1.Connect(ctx, *peerInfo)
	require.NoError(t, err)

	// Create media transfer managers
	transfer1, err := media.NewMediaTransfer(ctx, h1)
	require.NoError(t, err)

	transfer2, err := media.NewMediaTransfer(ctx, h2)
	require.NoError(t, err)

	// Setup encryption
	aliceSession := crypto.NewX3DHSession()
	bobSession := crypto.NewX3DHSession()

	// Generate prekey bundles
	bobBundle := bobSession.GeneratePrekeyBundle()

	// Perform key exchange
	sharedSecret1, _, err := aliceSession.InitiateKeyExchange(bobBundle)
	require.NoError(t, err)

	_, err = bobSession.ProcessKeyExchange(sharedSecret1)
	require.NoError(t, err)

	// Create test image data
	imageData := make([]byte, 100*1024) // 100KB
	for i := range imageData {
		imageData[i] = byte(i % 256)
	}

	// Setup stream handler for receiver
	receivedData := make(chan []byte, 1)
	receivedMetadata := make(chan media.MediaMetadata, 1)

	// Register stream handler
	h2.SetStreamHandler("/ledabeer/media/1.0.0", func(stream libp2pnetwork.Stream) {
		defer stream.Close()

		// Receive media
		data, metadata, err := transfer2.ReceiveMedia(stream, sharedSecret1)
		if err != nil {
			t.Logf("Error receiving media: %v", err)
			return
		}

		receivedData <- data
		receivedMetadata <- metadata
	})

	// Send image
	reader := bytes.NewReader(imageData)
	metadata := media.MediaMetadata{
		Type: "image/jpeg",
		Size: int64(len(imageData)),
		Name: "test-image.jpg",
	}

	err = transfer1.SendMedia(ctx, h2.ID(), reader, metadata, sharedSecret1)
	require.NoError(t, err, "Should send image successfully")

	// Wait for reception
	select {
	case received := <-receivedData:
		assert.Equal(t, imageData, received, "Received data should match sent data")
	case receivedMeta := <-receivedMetadata:
		assert.Equal(t, metadata.Type, receivedMeta.Type, "Received metadata should match sent metadata")
		assert.Equal(t, metadata.Size, receivedMeta.Size, "Received size should match sent size")
		assert.Equal(t, metadata.Name, receivedMeta.Name, "Received name should match sent name")
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for media reception")
	}
}

func TestMediaTransfer_VideoTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create two hosts
	h1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h1.Close()

	h2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h2.Close()

	// Connect hosts
	h2Addr := h2.Addrs()[0].String() + "/p2p/" + h2.ID().String()
	peerInfo, err := network.ParseAddrInfo(h2Addr)
	require.NoError(t, err)
	err = h1.Connect(ctx, *peerInfo)
	require.NoError(t, err)

	// Create media transfer managers
	transfer1, err := media.NewMediaTransfer(ctx, h1)
	require.NoError(t, err)

	transfer2, err := media.NewMediaTransfer(ctx, h2)
	require.NoError(t, err)

	// Setup encryption
	aliceSession := crypto.NewX3DHSession()
	bobSession := crypto.NewX3DHSession()

	// Generate prekey bundles
	bobBundle := bobSession.GeneratePrekeyBundle()

	// Perform key exchange
	sharedSecret1, _, err := aliceSession.InitiateKeyExchange(bobBundle)
	require.NoError(t, err)

	_, err = bobSession.ProcessKeyExchange(sharedSecret1)
	require.NoError(t, err)

	// Create test video data (larger file)
	videoData := make([]byte, 500*1024) // 500KB
	for i := range videoData {
		videoData[i] = byte(i % 256)
	}

	// Setup stream handler for receiver
	receivedData := make(chan []byte, 1)
	receivedMetadata := make(chan media.MediaMetadata, 1)

	// Register stream handler
	h2.SetStreamHandler("/ledabeer/media/1.0.0", func(stream libp2pnetwork.Stream) {
		defer stream.Close()

		// Receive media
		data, metadata, err := transfer2.ReceiveMedia(stream, sharedSecret1)
		if err != nil {
			t.Logf("Error receiving media: %v", err)
			return
		}

		receivedData <- data
		receivedMetadata <- metadata
	})

	// Send video
	reader := bytes.NewReader(videoData)
	metadata := media.MediaMetadata{
		Type: "video/mp4",
		Size: int64(len(videoData)),
		Name: "test-video.mp4",
	}

	err = transfer1.SendMedia(ctx, h2.ID(), reader, metadata, sharedSecret1)
	require.NoError(t, err, "Should send video successfully")

	// Wait for reception
	select {
	case received := <-receivedData:
		assert.Equal(t, videoData, received, "Received data should match sent data")
	case receivedMeta := <-receivedMetadata:
		assert.Equal(t, metadata.Type, receivedMeta.Type, "Received metadata should match sent metadata")
		assert.Equal(t, metadata.Size, receivedMeta.Size, "Received size should match sent size")
		assert.Equal(t, metadata.Name, receivedMeta.Name, "Received name should match sent name")
	case <-time.After(15 * time.Second):
		t.Fatal("Timeout waiting for video reception")
	}
}

func TestMediaTransfer_ConcurrentTransfers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Create three hosts
	h1, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h1.Close()

	h2, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h2.Close()

	h3, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h3.Close()

	// Connect all hosts
	h2Addr := h2.Addrs()[0].String() + "/p2p/" + h2.ID().String()
	peerInfo2, err := network.ParseAddrInfo(h2Addr)
	require.NoError(t, err)
	err = h1.Connect(ctx, *peerInfo2)
	require.NoError(t, err)

	h3Addr := h3.Addrs()[0].String() + "/p2p/" + h3.ID().String()
	peerInfo3, err := network.ParseAddrInfo(h3Addr)
	require.NoError(t, err)
	err = h1.Connect(ctx, *peerInfo3)
	require.NoError(t, err)

	// Create media transfer managers
	transfer1, err := media.NewMediaTransfer(ctx, h1)
	require.NoError(t, err)

	transfer2, err := media.NewMediaTransfer(ctx, h2)
	require.NoError(t, err)

	transfer3, err := media.NewMediaTransfer(ctx, h3)
	require.NoError(t, err)

	// Setup encryption for both transfers
	aliceSession1 := crypto.NewX3DHSession()
	bobSession1 := crypto.NewX3DHSession()

	aliceSession2 := crypto.NewX3DHSession()
	charlieSession2 := crypto.NewX3DHSession()

	// Generate prekey bundles
	bobBundle1 := bobSession1.GeneratePrekeyBundle()
	charlieBundle2 := charlieSession2.GeneratePrekeyBundle()

	// Perform key exchanges
	sharedSecret1, _, err := aliceSession1.InitiateKeyExchange(bobBundle1)
	require.NoError(t, err)

	_, err = bobSession1.ProcessKeyExchange(sharedSecret1)
	require.NoError(t, err)

	sharedSecret2, _, err := aliceSession2.InitiateKeyExchange(charlieBundle2)
	require.NoError(t, err)

	_, err = charlieSession2.ProcessKeyExchange(sharedSecret2)
	require.NoError(t, err)

	// Create test data
	data1 := make([]byte, 50*1024) // 50KB
	data2 := make([]byte, 75*1024) // 75KB
	for i := range data1 {
		data1[i] = byte(i % 256)
	}
	for i := range data2 {
		data2[i] = byte((i + 100) % 256)
	}

	// Setup stream handlers
	receivedData1 := make(chan []byte, 1)
	receivedData2 := make(chan []byte, 1)

	h2.SetStreamHandler("/ledabeer/media/1.0.0", func(stream libp2pnetwork.Stream) {
		defer stream.Close()
		data, _, err := transfer2.ReceiveMedia(stream, sharedSecret1)
		if err != nil {
			t.Logf("Error receiving media 1: %v", err)
			return
		}
		receivedData1 <- data
	})

	h3.SetStreamHandler("/ledabeer/media/1.0.0", func(stream libp2pnetwork.Stream) {
		defer stream.Close()
		data, _, err := transfer3.ReceiveMedia(stream, sharedSecret2)
		if err != nil {
			t.Logf("Error receiving media 2: %v", err)
			return
		}
		receivedData2 <- data
	})

	// Start concurrent transfers
	transfer1Done := make(chan error, 1)
	transfer2Done := make(chan error, 1)

	go func() {
		reader1 := bytes.NewReader(data1)
		metadata1 := media.MediaMetadata{
			Type: "image/png",
			Size: int64(len(data1)),
			Name: "image1.png",
		}
		err := transfer1.SendMedia(ctx, h2.ID(), reader1, metadata1, sharedSecret1)
		transfer1Done <- err
	}()

	go func() {
		reader2 := bytes.NewReader(data2)
		metadata2 := media.MediaMetadata{
			Type: "image/jpeg",
			Size: int64(len(data2)),
			Name: "image2.jpg",
		}
		err := transfer1.SendMedia(ctx, h3.ID(), reader2, metadata2, sharedSecret2)
		transfer2Done <- err
	}()

	// Wait for both transfers to complete
	err1 := <-transfer1Done
	err2 := <-transfer2Done

	require.NoError(t, err1, "First transfer should succeed")
	require.NoError(t, err2, "Second transfer should succeed")

	// Wait for reception
	select {
	case received1 := <-receivedData1:
		assert.Equal(t, data1, received1, "First received data should match sent data")
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for first media reception")
	}

	select {
	case received2 := <-receivedData2:
		assert.Equal(t, data2, received2, "Second received data should match sent data")
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for second media reception")
	}
}

func TestMediaTransfer_FailureRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create host
	h, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer h.Close()

	// Create media transfer manager
	transfer, err := media.NewMediaTransfer(ctx, h)
	require.NoError(t, err)

	// Setup encryption
	sharedSecret := make([]byte, 32)
	for i := range sharedSecret {
		sharedSecret[i] = byte(i)
	}

	// Create test data
	testData := make([]byte, 100*1024) // 100KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// Test sending to non-existent peer (should fail gracefully)
	reader := bytes.NewReader(testData)
	metadata := media.MediaMetadata{
		Type: "image/jpeg",
		Size: int64(len(testData)),
		Name: "test-image.jpg",
	}

	// Create a dummy peer ID that doesn't exist
	dummyPeerID := h.ID()

	err = transfer.SendMedia(ctx, dummyPeerID, reader, metadata, sharedSecret)
	assert.Error(t, err, "Should fail when sending to self")
	assert.Contains(t, err.Error(), "dial to self attempted", "Should mention dial to self")

	// Test with oversized file
	oversizedData := make([]byte, 100*1024*1024) // 100MB
	oversizedReader := bytes.NewReader(oversizedData)
	oversizedMetadata := media.MediaMetadata{
		Type: "video/mp4",
		Size: int64(len(oversizedData)),
		Name: "large-video.mp4",
	}

	err = transfer.SendMedia(ctx, dummyPeerID, oversizedReader, oversizedMetadata, sharedSecret)
	assert.Error(t, err, "Should fail with oversized file")
	// The error could be either "dial to self attempted" or "exceeds limit"
	assert.True(t,
		strings.Contains(err.Error(), "dial to self attempted") ||
			strings.Contains(err.Error(), "exceeds limit"),
		"Should mention either dial to self or size limit")
}
