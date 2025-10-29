package network

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Config holds configuration for a libp2p host
type Config struct {
	ListenAddrs    []string
	BootstrapPeers []string
	BootstrapHost  string
	BootstrapPort  string
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

	// Add connection event logging
	h.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, c network.Conn) {
			fmt.Printf("🔗 Connected to peer: %s\n", c.RemotePeer())
		},
		DisconnectedF: func(n network.Network, c network.Conn) {
			fmt.Printf("❌ Disconnected from peer: %s\n", c.RemotePeer())
		},
	})

	// Create DHT
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
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

	// Bootstrap the DHT after connecting to bootstrap peers
	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		h.Close()
		return nil, err
	}

	// Dynamic bootstrap peer discovery
	if cfg.BootstrapHost != "" && cfg.BootstrapPort != "" {
		go func(dht *dht.IpfsDHT) {
			// Wait a bit for the host to be ready
			time.Sleep(3 * time.Second)

			// Query bootstrap node for its peer ID via HTTP
			// Use port 8080 for HTTP, not the libp2p port
			bootstrapURL := fmt.Sprintf("http://%s:8080/peer-id", cfg.BootstrapHost)

			// Try to get bootstrap peer ID
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(bootstrapURL)
			if err != nil {
				fmt.Printf("Failed to query bootstrap peer ID: %v\n", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				fmt.Printf("Bootstrap returned status %d\n", resp.StatusCode)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Failed to read bootstrap response: %v\n", err)
				return
			}

			peerID := strings.TrimSpace(string(body))
			if peerID == "" {
				fmt.Printf("Empty peer ID from bootstrap\n")
				return
			}

			// Construct full multiaddr with discovered peer ID
			bootstrapAddr := fmt.Sprintf("/dns4/%s/tcp/%s/p2p/%s", cfg.BootstrapHost, cfg.BootstrapPort, peerID)
			fmt.Printf("Discovered bootstrap peer: %s\n", bootstrapAddr)

			// Connect to the bootstrap peer
			peerInfo, err := peer.AddrInfoFromString(bootstrapAddr)
			if err != nil {
				fmt.Printf("Failed to parse bootstrap address: %v\n", err)
				return
			}

			if err := h.Connect(ctx, *peerInfo); err != nil {
				fmt.Printf("Failed to connect to bootstrap: %v\n", err)
			} else {
				fmt.Printf("Successfully connected to bootstrap peer!\n")

				// Wait a bit for the connection to be fully established
				time.Sleep(2 * time.Second)

				// Bootstrap DHT after connecting to bootstrap peer
				if err := dht.Bootstrap(ctx); err != nil {
					fmt.Printf("Failed to bootstrap DHT: %v\n", err)
				} else {
					fmt.Printf("DHT bootstrap completed!\n")

					// Start peer connection manager
					go func() {
						for {
							time.Sleep(5 * time.Second)

							// Query bootstrap node for all connected peers
							peersURL := fmt.Sprintf("http://%s:8080/peers", cfg.BootstrapHost)
							resp, err := client.Get(peersURL)
							if err != nil {
								fmt.Printf("Failed to query peers: %v\n", err)
								continue
							}
							defer resp.Body.Close()

							if resp.StatusCode != 200 {
								fmt.Printf("Bootstrap returned status %d for peers\n", resp.StatusCode)
								continue
							}

							body, err := io.ReadAll(resp.Body)
							if err != nil {
								fmt.Printf("Failed to read peers response: %v\n", err)
								continue
							}

							fmt.Printf("Bootstrap peers info:\n%s\n", string(body))

							// Parse peer information and try to connect
							lines := strings.Split(string(body), "\n")
							for _, line := range lines {
								if strings.HasPrefix(line, "Peer: ") {
									peerIDStr := strings.TrimPrefix(line, "Peer: ")
									peerID, err := peer.Decode(peerIDStr)
									if err != nil {
										fmt.Printf("Failed to decode peer ID %s: %v\n", peerIDStr, err)
										continue
									}

									// Skip self
									if peerID == h.ID() {
										continue
									}

									// Check if already connected
									if h.Network().Connectedness(peerID) == network.Connected {
										continue
									}

									// Find the address for this peer
									var peerAddr string
									for i, addrLine := range lines {
										if strings.Contains(addrLine, peerIDStr) && i+1 < len(lines) {
											addrLine = lines[i+1]
											if strings.HasPrefix(addrLine, "  Address: ") {
												peerAddr = strings.TrimPrefix(addrLine, "  Address: ")
												break
											}
										}
									}

									if peerAddr != "" {
										// Parse the address
										addr, err := multiaddr.NewMultiaddr(peerAddr)
										if err != nil {
											fmt.Printf("Failed to parse address %s: %v\n", peerAddr, err)
											continue
										}

										// Create peer info and connect
										peerInfo := peer.AddrInfo{ID: peerID, Addrs: []multiaddr.Multiaddr{addr}}
										if err := h.Connect(ctx, peerInfo); err != nil {
											fmt.Printf("Failed to connect to peer %s: %v\n", peerID, err)
										} else {
											fmt.Printf("🔗 Connected to peer from bootstrap: %s\n", peerID)
										}
									}
								}
							}
						}
					}()
				}
			}
		}(kademliaDHT)
	}

	return h, nil
}

// ParseAddrInfo parses a multiaddr string into a peer.AddrInfo
func ParseAddrInfo(addr string) (*peer.AddrInfo, error) {
	return peer.AddrInfoFromString(addr)
}
