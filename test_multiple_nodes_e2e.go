package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "ledabeer/backend/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("🧪 Multiple Nodes E2E Test")
	fmt.Println("==========================")
	fmt.Println("This test requires multiple backend nodes to be running.")
	fmt.Println("Make sure you have started the backend with: docker compose up -d")
	fmt.Println()

	// Test both bootstrap and alice nodes
	testNode("Bootstrap Node", "localhost:50051")
	testNode("Alice Node", "localhost:4001") // Assuming alice runs on different port
}

func testNode(nodeName, address string) {
	fmt.Printf("\n🔍 Testing %s at %s\n", nodeName, address)
	fmt.Println(strings.Repeat("-", 50))

	// Connect to the node
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s: %v\n", nodeName, err)
		return
	}
	defer conn.Close()

	messageClient := pb.NewMessageServiceClient(conn)
	peerClient := pb.NewPeerServiceClient(conn)

	// Get peers
	peersResp, err := peerClient.GetPeers(context.Background(), &pb.GetPeersRequest{})
	if err != nil {
		fmt.Printf("❌ Failed to get peers from %s: %v\n", nodeName, err)
		return
	}

	fmt.Printf("✅ %s found %d peers:\n", nodeName, len(peersResp.Peers))
	for i, peer := range peersResp.Peers {
		fmt.Printf("   %d. ID: %s, Name: %s, Online: %v\n", i+1, peer.Id, peer.Name, peer.Online)
	}

	// Test streaming
	fmt.Printf("📡 Testing streaming on %s...\n", nodeName)
	stream, err := messageClient.ReceiveMessages(context.Background(), &pb.ReceiveMessagesRequest{})
	if err != nil {
		fmt.Printf("❌ Failed to start stream on %s: %v\n", nodeName, err)
		return
	}
	fmt.Printf("✅ %s streaming started\n", nodeName)

	// Try to send a message if there are peers
	if len(peersResp.Peers) > 0 {
		targetPeer := peersResp.Peers[0]
		fmt.Printf("📤 Sending message from %s to %s...\n", nodeName, targetPeer.Name)

		_, err = messageClient.SendMessage(context.Background(), &pb.SendMessageRequest{
			ToPeerId: targetPeer.Id,
			Content:  []byte(fmt.Sprintf("Hello from %s!", nodeName)),
		})
		if err != nil {
			fmt.Printf("❌ Failed to send message from %s: %v\n", nodeName, err)
		} else {
			fmt.Printf("✅ Message sent from %s\n", nodeName)
		}
	} else {
		fmt.Printf("⚠️  No peers available for %s\n", nodeName)
	}

	// Listen for a short time
	fmt.Printf("👂 Listening on %s for 3 seconds...\n", nodeName)
	timeout := time.After(3 * time.Second)
	messageCount := 0

	for {
		select {
		case <-timeout:
			fmt.Printf("⏰ %s received %d messages\n", nodeName, messageCount)
			break
		default:
			msg, err := stream.Recv()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			messageCount++
			fmt.Printf("📦 %s received: %s\n", nodeName, string(msg.Content))
		}
	}
}
