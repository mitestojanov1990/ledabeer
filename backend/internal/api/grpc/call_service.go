package grpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"

	pb "ledabeer/backend/pkg/proto"
)

type CallService struct {
	pb.UnimplementedCallServiceServer
	callHandler interface{} // Will be replaced with actual call handler
}

func NewCallService(handler interface{}) *CallService {
	return &CallService{callHandler: handler}
}

func (s *CallService) InitiateCall(ctx context.Context, req *pb.InitiateCallRequest) (*pb.InitiateCallResponse, error) {
	// Generate call ID
	callID := generateCallID()

	// Initiate call via call handler (mock for now)

	return &pb.InitiateCallResponse{
		CallId: callID,
		State: &pb.CallState{
			CallId:       callID,
			State:        pb.CallStateEnum_INITIATING,
			Participants: []string{req.ToPeerId},
		},
	}, nil
}

func (s *CallService) AnswerCall(ctx context.Context, req *pb.AnswerCallRequest) (*pb.AnswerCallResponse, error) {
	// Answer call via call handler (mock for now)
	state := pb.CallStateEnum_ENDED
	if req.Accept {
		state = pb.CallStateEnum_CONNECTED
	}

	return &pb.AnswerCallResponse{
		State: &pb.CallState{
			CallId:       req.CallId,
			State:        state,
			Participants: []string{"peer1", "peer2"},
		},
	}, nil
}

func (s *CallService) EndCall(ctx context.Context, req *pb.EndCallRequest) (*pb.EndCallResponse, error) {
	// End call via call handler (mock for now)
	return &pb.EndCallResponse{
		Success: true,
	}, nil
}

func (s *CallService) StreamSignaling(stream pb.CallService_StreamSignalingServer) error {
	// Handle bidirectional signaling stream (mock for now)
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Echo back the message (mock implementation)
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
}

func (s *CallService) GetCallState(ctx context.Context, req *pb.GetCallStateRequest) (*pb.CallState, error) {
	// Get call state (mock for now)
	return &pb.CallState{
		CallId:       req.CallId,
		State:        pb.CallStateEnum_CONNECTED,
		Participants: []string{"peer1", "peer2"},
	}, nil
}

func generateCallID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("call_%x", bytes)
}
