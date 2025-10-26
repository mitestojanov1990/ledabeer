package network

import (
	"context"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Config holds configuration for a libp2p host
type Config struct {
	ListenAddrs    []string
	BootstrapPeers []string
}

// Host wraps libp2p host with additional functionality
type Host interface {
	host.Host
	AddrInfo() peer.AddrInfo
}

// NewHost creates a new libp2p host with secure transports and DHT
func NewHost(ctx context.Context, cfg *Config) (host.Host, error) {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(cfg.ListenAddrs...),
		// Libp2p will automatically use Noise and TLS for secure connections
	}

	// Create the host
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	// Create DHT
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		h.Close()
		return nil, err
	}

	// Bootstrap the DHT
	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		h.Close()
		return nil, err
	}

	// Connect to bootstrap peers if provided
	if len(cfg.BootstrapPeers) > 0 {
		for _, peerAddr := range cfg.BootstrapPeers {
			peerInfo, err := peer.AddrInfoFromString(peerAddr)
			if err != nil {
				continue // Skip invalid addresses
			}
			h.Connect(ctx, *peerInfo)
		}
	}

	return h, nil
}

// ParseAddrInfo parses a multiaddr string into a peer.AddrInfo
func ParseAddrInfo(addr string) (*peer.AddrInfo, error) {
	return peer.AddrInfoFromString(addr)
}
