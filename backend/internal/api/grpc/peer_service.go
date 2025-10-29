package grpc

import (
	"context"
	"fmt"
	"time"

	pb "ledabeer/backend/pkg/proto"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerService struct {
	pb.UnimplementedPeerServiceServer
	host host.Host
}

func NewPeerService(h host.Host) *PeerService {
	return &PeerService{
		host: h,
	}
}

func (s *PeerService) GetPeers(ctx context.Context, req *pb.GetPeersRequest) (*pb.GetPeersResponse, error) {
	fmt.Printf("📡 GetPeers called from gRPC client\n")
	// Get connected peers from libp2p host
	connectedPeers := s.host.Network().Peers()

	peers := make([]*pb.Peer, 0, len(connectedPeers))
	for _, peerID := range connectedPeers {
		// Skip self
		if peerID == s.host.ID() {
			continue
		}

		// Get peer info from peerstore
		peerInfo := s.host.Peerstore().PeerInfo(peerID)

		// If we're iterating over Network().Peers(), they are all connected
		// network.Connectedness check sometimes gives false negatives
		connected := true

		// Convert addresses to strings
		addresses := make([]string, len(peerInfo.Addrs))
		for i, addr := range peerInfo.Addrs {
			addresses[i] = addr.String()
		}

		// Get last seen time from peerstore (use current time as fallback)
		lastSeen := time.Now()

		peers = append(peers, &pb.Peer{
			Id:        peerID.String(),
			Name:      peerID.ShortString(), // Use short peer ID as name
			Online:    connected,
			LastSeen:  lastSeen.Unix(),
			Addresses: addresses,
		})
	}

	return &pb.GetPeersResponse{
		Peers: peers,
	}, nil
}

func (s *PeerService) GetPeer(ctx context.Context, req *pb.GetPeerRequest) (*pb.GetPeerResponse, error) {
	peerID, err := peer.Decode(req.PeerId)
	if err != nil {
		return &pb.GetPeerResponse{
			Found: false,
		}, nil
	}

	// Check if peer exists in peerstore
	peerInfo := s.host.Peerstore().PeerInfo(peerID)
	if len(peerInfo.Addrs) == 0 {
		return &pb.GetPeerResponse{
			Found: false,
		}, nil
	}

	// Check if peer is connected
	connected := s.host.Network().Connectedness(peerID) == network.Connected

	// Convert addresses to strings
	addresses := make([]string, len(peerInfo.Addrs))
	for i, addr := range peerInfo.Addrs {
		addresses[i] = addr.String()
	}

	// Get last seen time from peerstore (use current time as fallback)
	lastSeen := time.Now()

	pbPeer := &pb.Peer{
		Id:        peerID.String(),
		Name:      peerID.ShortString(),
		Online:    connected,
		LastSeen:  lastSeen.Unix(),
		Addresses: addresses,
	}

	return &pb.GetPeerResponse{
		Peer:  pbPeer,
		Found: true,
	}, nil
}

func (s *PeerService) GetConnectedPeers(ctx context.Context, req *pb.GetConnectedPeersRequest) (*pb.GetConnectedPeersResponse, error) {
	// Get only connected peers
	connectedPeers := s.host.Network().Peers()

	peers := make([]*pb.Peer, 0)
	for _, peerID := range connectedPeers {
		// Skip self
		if peerID == s.host.ID() {
			continue
		}

		// Only include connected peers
		if s.host.Network().Connectedness(peerID) != network.Connected {
			continue
		}

		// Get peer info from peerstore
		peerInfo := s.host.Peerstore().PeerInfo(peerID)

		// Convert addresses to strings
		addresses := make([]string, len(peerInfo.Addrs))
		for i, addr := range peerInfo.Addrs {
			addresses[i] = addr.String()
		}

		// Get last seen time from peerstore (use current time as fallback)
		lastSeen := time.Now()

		peers = append(peers, &pb.Peer{
			Id:        peerID.String(),
			Name:      peerID.ShortString(),
			Online:    true, // We know it's connected
			LastSeen:  lastSeen.Unix(),
			Addresses: addresses,
		})
	}

	return &pb.GetConnectedPeersResponse{
		Peers: peers,
	}, nil
}
