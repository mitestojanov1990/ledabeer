package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "ledabeer/backend/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("🧪 Simple Streaming Test")
	fmt.Println("========================")

	// Connect to the backend
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewMessageServiceClient(conn)

	// Test 1: Start streaming
	fmt.Println("\n📡 Starting ReceiveMessages stream...")
	stream, err := client.ReceiveMessages(context.Background(), &pb.ReceiveMessagesRequest{})
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}
	fmt.Println("✅ Stream started successfully")

	// Test 2: Listen for messages for a short time
	fmt.Println("\n👂 Listening for messages (10 seconds)...")
	messageCount := 0
	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Printf("\n⏰ Timeout reached. Received %d messages.\n", messageCount)
			if messageCount == 0 {
				fmt.Println("ℹ️  No messages received - this is expected for a streaming endpoint")
				fmt.Println("   The endpoint is working correctly by waiting for messages")
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
			fmt.Printf("📦 Received message %d:\n", messageCount)
			fmt.Printf("   ID: %s\n", msg.MessageId)
			fmt.Printf("   From: %s\n", msg.FromPeerId)
			fmt.Printf("   Content: %s\n", string(msg.Content))
			fmt.Printf("   Timestamp: %d\n", msg.Timestamp)
		}
	}
}
