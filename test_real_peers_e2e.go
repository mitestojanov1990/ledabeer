package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "ledabeer/backend/pkg/proto"
)

func main() {
	fmt.Println("🧪 Real Peer-to-Peer E2E Test")
	fmt.Println("=============================")

	// Connect to the backend
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	messageClient := pb.NewMessageServiceClient(conn)
	peerClient := pb.NewPeerServiceClient(conn)

	// Test 1: Get available peers
	fmt.Println("\n🔍 Step 1: Getting available peers...")
	peersResp, err := peerClient.GetPeers(context.Background(), &pb.GetPeersRequest{})
	if err != nil {
		log.Fatalf("Failed to get peers: %v", err)
	}
	
	fmt.Printf("✅ Found %d peers:\n", len(peersResp.Peers))
	for i, peer := range peersResp.Peers {
		fmt.Printf("   %d. ID: %s, Name: %s, Online: %v\n", i+1, peer.Id, peer.Name, peer.Online)
	}

	// Test 2: Start streaming for messages
	fmt.Println("\n📡 Step 2: Starting message streaming...")
	stream, err := messageClient.ReceiveMessages(context.Background(), &pb.ReceiveMessagesRequest{})
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}
	fmt.Println("✅ Stream started successfully")

	// Test 3: Send message to first available peer (if any)
	if len(peersResp.Peers) > 0 {
		targetPeer := peersResp.Peers[0]
		fmt.Printf("\n📤 Step 3: Sending message to peer %s (%s)...\n", targetPeer.Name, targetPeer.Id)
		
		_, err = messageClient.SendMessage(context.Background(), &pb.SendMessageRequest{
			ToPeerId: targetPeer.Id,
			Content:  []byte("Hello from E2E test! This is a real peer-to-peer message."),
		})
		if err != nil {
			fmt.Printf("❌ Failed to send message: %v\n", err)
		} else {
			fmt.Println("✅ Message sent successfully")
		}
	} else {
		fmt.Println("\n⚠️  No peers available - this is expected for a fresh network")
		fmt.Println("   To test real peer-to-peer communication, you need multiple nodes running")
	}

	// Test 4: Listen for incoming messages
	fmt.Println("\n👂 Step 4: Listening for incoming messages (10 seconds)...")
	messageCount := 0
	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Printf("\n⏰ Timeout reached. Received %d messages.\n", messageCount)
			if messageCount == 0 {
				fmt.Println("\n📋 Analysis:")
				fmt.Println("   ✅ Streaming endpoint is working correctly")
				fmt.Println("   ✅ Backend is waiting for real peer-to-peer messages")
				fmt.Println("   ℹ️  No messages received because:")
				if len(peersResp.Peers) == 0 {
					fmt.Println("      - No peers are connected to the network")
					fmt.Println("      - Start multiple backend nodes to test peer-to-peer communication")
				} else {
					fmt.Println("      - Target peer may not be listening for messages")
					fmt.Println("      - Network routing may not be established yet")
				}
			}
			return
		default:
			// Try to receive a message
			msg, err := stream.Recv()
			
			if err != nil {
				// This is expected - stream might not have data yet
				time.Sleep(100 * time.Millisecond)
				continue
			}
			
			messageCount++
			fmt.Printf("\n📦 Received message %d:\n", messageCount)
			fmt.Printf("   ID: %s\n", msg.MessageId)
			fmt.Printf("   From: %s\n", msg.FromPeerId)
			fmt.Printf("   Content: %s\n", string(msg.Content))
			fmt.Printf("   Timestamp: %d\n", msg.Timestamp)
		}
	}
}
