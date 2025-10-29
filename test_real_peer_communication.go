package main

import (
	"context"
	"fmt"
	"time"

	pb "ledabeer/backend/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("🧪 Real Peer-to-Peer Communication Test")
	fmt.Println("=======================================")
	fmt.Println("Testing actual peer-to-peer communication between backend nodes")
	fmt.Println()

	// Test Alice node (port 4002)
	fmt.Println("🔍 Testing Alice Node...")
	testNode("Alice", "localhost:4002", "12D3KooWJbbM9kZn5na5ByqGhu67g1qNeRs9SKoy5ZEVevK43scL")

	// Test Bob node (port 4003)
	fmt.Println("\n🔍 Testing Bob Node...")
	testNode("Bob", "localhost:4003", "12D3KooWFUjo7n8WzsAwP3ws4mtWwmTPyLBnbrTZqwUA5yor5HaV")

	// Test communication between Alice and Bob
	fmt.Println("\n📡 Testing Alice → Bob Communication...")
	testCommunication("Alice", "localhost:4002", "Bob", "localhost:4003", "12D3KooWFUjo7n8WzsAwP3ws4mtWwmTPyLBnbrTZqwUA5yor5HaV")
}

func testNode(nodeName, address, expectedPeerID string) {
	// Connect to the node
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("❌ Failed to connect to %s: %v\n", nodeName, err)
		return
	}
	defer conn.Close()

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

	// Check if this node knows about other nodes
	if len(peersResp.Peers) > 0 {
		fmt.Printf("🎉 %s is connected to the network!\n", nodeName)
	} else {
		fmt.Printf("⚠️  %s is not connected to any peers\n", nodeName)
	}
}

func testCommunication(senderName, senderAddr, receiverName, receiverAddr, targetPeerID string) {
	// Connect to sender
	senderConn, err := grpc.Dial(senderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("❌ Failed to connect to sender %s: %v\n", senderName, err)
		return
	}
	defer senderConn.Close()

	// Connect to receiver
	receiverConn, err := grpc.Dial(receiverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("❌ Failed to connect to receiver %s: %v\n", receiverName, err)
		return
	}
	defer receiverConn.Close()

	senderClient := pb.NewMessageServiceClient(senderConn)
	receiverClient := pb.NewMessageServiceClient(receiverConn)

	// Start listening on receiver
	fmt.Printf("👂 Starting message listener on %s...\n", receiverName)
	stream, err := receiverClient.ReceiveMessages(context.Background(), &pb.ReceiveMessagesRequest{})
	if err != nil {
		fmt.Printf("❌ Failed to start stream on %s: %v\n", receiverName, err)
		return
	}

	// Send message from sender to receiver
	fmt.Printf("📤 Sending message from %s to %s...\n", senderName, receiverName)
	_, err = senderClient.SendMessage(context.Background(), &pb.SendMessageRequest{
		ToPeerId: targetPeerID,
		Content:  []byte(fmt.Sprintf("Hello from %s to %s!", senderName, receiverName)),
	})
	if err != nil {
		fmt.Printf("❌ Failed to send message: %v\n", err)
		return
	}
	fmt.Printf("✅ Message sent successfully\n")

	// Listen for response
	fmt.Printf("👂 Listening for response on %s (5 seconds)...\n", receiverName)
	timeout := time.After(5 * time.Second)
	messageCount := 0

	for {
		select {
		case <-timeout:
			fmt.Printf("⏰ %s received %d messages\n", receiverName, messageCount)
			if messageCount == 0 {
				fmt.Printf("ℹ️  No messages received - this could mean:\n")
				fmt.Printf("   - Peers are not properly connected\n")
				fmt.Printf("   - Message routing is not working\n")
				fmt.Printf("   - Network discovery needs more time\n")
			}
			return
		default:
			msg, err := stream.Recv()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			messageCount++
			fmt.Printf("📦 %s received message %d:\n", receiverName, messageCount)
			fmt.Printf("   From: %s\n", msg.FromPeerId)
			fmt.Printf("   Content: %s\n", string(msg.Content))
			fmt.Printf("   Timestamp: %d\n", msg.Timestamp)
		}
	}
}
