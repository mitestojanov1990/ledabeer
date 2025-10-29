package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "ledabeer/backend/pkg/proto"

	"google.golang.org/grpc"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Test PeerService
	peerClient := pb.NewPeerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	peers, err := peerClient.GetPeers(ctx, &pb.GetPeersRequest{})
	if err != nil {
		log.Printf("PeerService error: %v", err)
	} else {
		fmt.Printf("PeerService working: got %d peers\n", len(peers.Peers))
	}

	// Test GroupService
	groupClient := pb.NewGroupServiceClient(conn)
	groups, err := groupClient.GetGroups(ctx, &pb.GetGroupsRequest{})
	if err != nil {
		log.Printf("GroupService error: %v", err)
	} else {
		fmt.Printf("GroupService working: got %d groups\n", len(groups.Groups))
	}
}
