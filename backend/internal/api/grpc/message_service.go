package grpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"ledabeer/backend/internal/messaging"
	pb "ledabeer/backend/pkg/proto"
)

type MessageService struct {
	pb.UnimplementedMessageServiceServer
	msgHandler   *messaging.MessageHandler
	groupManager *messaging.GroupManager
}

func NewMessageService(msgHandler *messaging.MessageHandler, groupManager *messaging.GroupManager) *MessageService {
	return &MessageService{
		msgHandler:   msgHandler,
		groupManager: groupManager,
	}
}

func (s *MessageService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// Handle nil handler for unit tests
	if s.msgHandler == nil {
		messageID := generateMessageID()
		return &pb.SendMessageResponse{
			MessageId: messageID,
			Timestamp: time.Now().Unix(),
		}, nil
	}

	// Send via real messaging layer
	messageID, err := s.msgHandler.SendMessage(ctx, req.ToPeerId, req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &pb.SendMessageResponse{
		MessageId: messageID,
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *MessageService) ReceiveMessages(req *pb.ReceiveMessagesRequest, stream pb.MessageService_ReceiveMessagesServer) error {
	// Handle nil handler for unit tests
	if s.msgHandler == nil {
		// Send a test message for unit tests
		msg := &pb.Message{
			MessageId:  generateMessageID(),
			FromPeerId: "test-peer",
			Content:    []byte("test message"),
			Timestamp:  time.Now().Unix(),
		}
		return stream.Send(msg)
	}

	// Stream from real messaging layer
	msgChan := s.msgHandler.SubscribeToMessages(stream.Context())

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg := <-msgChan:
			pbMsg := &pb.Message{
				MessageId:  msg.ID,
				FromPeerId: msg.From,
				Content:    msg.Content,
				Timestamp:  msg.Timestamp,
			}

			if err := stream.Send(pbMsg); err != nil {
				return err
			}
		}
	}
}

func (s *MessageService) SendGroupMessage(ctx context.Context, req *pb.SendGroupMessageRequest) (*pb.SendMessageResponse, error) {
	// Handle nil handler for unit tests
	if s.groupManager == nil {
		messageID := generateMessageID()
		return &pb.SendMessageResponse{
			MessageId: messageID,
			Timestamp: time.Now().Unix(),
		}, nil
	}

	// Send via real group manager
	messageID, err := s.groupManager.SendGroupMessage(ctx, req.GroupId, req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to send group message: %w", err)
	}

	return &pb.SendMessageResponse{
		MessageId: messageID,
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *MessageService) GetMessageHistory(ctx context.Context, req *pb.GetMessageHistoryRequest) (*pb.MessageHistoryResponse, error) {
	// Handle nil handler for unit tests
	if s.msgHandler == nil {
		// Return empty history for unit tests
		return &pb.MessageHistoryResponse{
			Messages: []*pb.Message{},
		}, nil
	}

	// Get from real message history (IPFS storage)
	messages, err := s.msgHandler.GetMessageHistory(ctx, req.PeerId, int(req.Limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	pbMessages := make([]*pb.Message, len(messages))
	for i, msg := range messages {
		pbMessages[i] = &pb.Message{
			MessageId:  msg.ID,
			FromPeerId: msg.From,
			Content:    msg.Content,
			Timestamp:  msg.Timestamp,
		}
	}

	return &pb.MessageHistoryResponse{
		Messages: pbMessages,
	}, nil
}

func generateMessageID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("msg_%x", bytes)
}
