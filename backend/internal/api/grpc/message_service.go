package grpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	pb "ledabeer/backend/pkg/proto"
)

type MessageService struct {
	pb.UnimplementedMessageServiceServer
	msgHandler interface{} // Will be replaced with actual messaging handler
}

func NewMessageService(handler interface{}) *MessageService {
	return &MessageService{msgHandler: handler}
}

func (s *MessageService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// Generate message ID
	messageID := generateMessageID()

	// Send message via messaging layer (mock for now)
	// In real implementation, this would use the messaging handler

	return &pb.SendMessageResponse{
		MessageId: messageID,
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *MessageService) ReceiveMessages(req *pb.ReceiveMessagesRequest, stream pb.MessageService_ReceiveMessagesServer) error {
	// Stream incoming messages (mock for now)
	// In real implementation, this would stream from messaging layer

	// Send a test message
	msg := &pb.Message{
		MessageId:  generateMessageID(),
		FromPeerId: "peer1",
		Content:    []byte("test message"),
		Timestamp:  time.Now().Unix(),
	}

	return stream.Send(msg)
}

func (s *MessageService) SendGroupMessage(ctx context.Context, req *pb.SendGroupMessageRequest) (*pb.SendMessageResponse, error) {
	// Generate message ID
	messageID := generateMessageID()

	// Send group message via messaging layer (mock for now)

	return &pb.SendMessageResponse{
		MessageId: messageID,
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *MessageService) GetMessageHistory(ctx context.Context, req *pb.GetMessageHistoryRequest) (*pb.MessageHistoryResponse, error) {
	// Get message history (mock for now)
	messages := []*pb.Message{
		{
			MessageId:  generateMessageID(),
			FromPeerId: req.PeerId,
			Content:    []byte("historical message"),
			Timestamp:  time.Now().Unix(),
		},
	}

	return &pb.MessageHistoryResponse{
		Messages: messages,
	}, nil
}

func generateMessageID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("msg_%x", bytes)
}
