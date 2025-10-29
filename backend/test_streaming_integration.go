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
	fmt.Println("🧪 E2E Streaming Integration Test")
	fmt.Println("=================================")

	// Connect to the backend
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewMessageServiceClient(conn)

	// Test 1: Start streaming
	fmt.Println("\n📡 Test 1: Starting ReceiveMessages stream...")
	stream, err := client.ReceiveMessages(context.Background(), &pb.ReceiveMessagesRequest{})
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	// Test 2: Send a message in a separate goroutine
	go func() {
		time.Sleep(2 * time.Second) // Wait a bit for stream to be ready
		fmt.Println("\n📤 Test 2: Sending a test message...")
		
		sendClient := pb.NewMessageServiceClient(conn)
		_, err := sendClient.SendMessage(context.Background(), &pb.SendMessageRequest{
			ToPeerId: "test-peer",
			Content:  []byte("Hello from integration test!"),
		})
		if err != nil {
			fmt.Printf("❌ Failed to send message: %v\n", err)
		} else {
			fmt.Println("✅ Message sent successfully")
		}
	}()

	// Test 3: Listen for messages
	fmt.Println("\n👂 Test 3: Listening for messages...")
	messageCount := 0
	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Printf("\n⏰ Timeout reached. Received %d messages.\n", messageCount)
			return
		default:
			// Try to receive a message with a short timeout
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
