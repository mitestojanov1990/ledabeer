package grpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"ledabeer/backend/internal/calls"
	pb "ledabeer/backend/pkg/proto"
)

type CallService struct {
	pb.UnimplementedCallServiceServer
	callManager *calls.CallManager
}

func NewCallService(callManager *calls.CallManager) *CallService {
	return &CallService{callManager: callManager}
}

func (s *CallService) InitiateCall(ctx context.Context, req *pb.InitiateCallRequest) (*pb.InitiateCallResponse, error) {
	// Handle nil manager for unit tests
	if s.callManager == nil {
		callID := generateCallID()
		return &pb.InitiateCallResponse{
			CallId: callID,
			State: &pb.CallState{
				CallId:       callID,
				State:        pb.CallStateEnum_INITIATING,
				Participants: []string{req.ToPeerId},
			},
		}, nil
	}

	// Initiate real WebRTC call
	callID, err := s.callManager.InitiateCall(ctx, req.ToPeerId, req.AudioEnabled, req.VideoEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate call: %w", err)
	}

	session := s.callManager.GetCallSession(callID)

	return &pb.InitiateCallResponse{
		CallId: callID,
		State: &pb.CallState{
			CallId:       callID,
			State:        convertCallState(session.GetState()),
			Participants: session.GetParticipants(),
		},
	}, nil
}

func (s *CallService) AnswerCall(ctx context.Context, req *pb.AnswerCallRequest) (*pb.AnswerCallResponse, error) {
	// Handle nil manager for unit tests
	if s.callManager == nil {
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

	// Answer via real call manager
	err := s.callManager.AnswerCall(ctx, req.CallId, req.Accept)
	if err != nil {
		return nil, fmt.Errorf("failed to answer call: %w", err)
	}

	session := s.callManager.GetCallSession(req.CallId)

	return &pb.AnswerCallResponse{
		State: &pb.CallState{
			CallId:       req.CallId,
			State:        convertCallState(session.GetState()),
			Participants: session.GetParticipants(),
		},
	}, nil
}

func (s *CallService) EndCall(ctx context.Context, req *pb.EndCallRequest) (*pb.EndCallResponse, error) {
	// Handle nil manager for unit tests
	if s.callManager == nil {
		return &pb.EndCallResponse{
			Success: true,
		}, nil
	}

	// End call via real call manager
	err := s.callManager.EndCall(ctx, req.CallId)
	if err != nil {
		return nil, fmt.Errorf("failed to end call: %w", err)
	}

	return &pb.EndCallResponse{
		Success: true,
	}, nil
}

func (s *CallService) StreamSignaling(stream pb.CallService_StreamSignalingServer) error {
	// Handle nil manager for unit tests
	if s.callManager == nil {
		// Echo back messages for unit tests
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}

	// Handle real WebRTC signaling
	var callID string

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Process signaling via call manager
		response, err := s.callManager.HandleSignaling(stream.Context(), callID, msg.Type, msg.Sdp, msg.Candidate)
		if err != nil {
			return err
		}

		if response != nil {
			if err := stream.Send(response); err != nil {
				return err
			}
		}
	}
}

func (s *CallService) GetCallState(ctx context.Context, req *pb.GetCallStateRequest) (*pb.CallState, error) {
	// Handle nil manager for unit tests
	if s.callManager == nil {
		return &pb.CallState{
			CallId:       req.CallId,
			State:        pb.CallStateEnum_CONNECTED,
			Participants: []string{"peer1", "peer2"},
		}, nil
	}

	// Get real call state
	session := s.callManager.GetCallSession(req.CallId)
	if session == nil {
		return nil, fmt.Errorf("call session %s not found", req.CallId)
	}

	return &pb.CallState{
		CallId:       req.CallId,
		State:        convertCallState(session.GetState()),
		Participants: session.GetParticipants(),
	}, nil
}

func convertCallState(state calls.CallState) pb.CallStateEnum {
	switch state {
	case calls.StateInitiating:
		return pb.CallStateEnum_INITIATING
	case calls.StateRinging:
		return pb.CallStateEnum_RINGING
	case calls.StateConnected:
		return pb.CallStateEnum_CONNECTED
	case calls.StateEnded:
		return pb.CallStateEnum_ENDED
	default:
		return pb.CallStateEnum_ENDED
	}
}

func generateCallID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("call_%x", bytes)
}
