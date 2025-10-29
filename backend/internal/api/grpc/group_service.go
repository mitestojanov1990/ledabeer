package grpc

import (
	"context"
	"time"

	"ledabeer/backend/internal/messaging"
	pb "ledabeer/backend/pkg/proto"
)

type GroupService struct {
	pb.UnimplementedGroupServiceServer
	groupManager *messaging.GroupManager
}

func NewGroupService(groupManager *messaging.GroupManager) *GroupService {
	return &GroupService{
		groupManager: groupManager,
	}
}

func (s *GroupService) GetGroups(ctx context.Context, req *pb.GetGroupsRequest) (*pb.GetGroupsResponse, error) {
	// Get all groups from group manager
	groups := s.groupManager.GetGroups()
	
	// Convert to protobuf format
	pbGroups := make([]*pb.Group, 0, len(groups))
	for _, group := range groups {
		// Get member IDs
		memberIDs := make([]string, 0, len(group.Members))
		for memberID := range group.Members {
			memberIDs = append(memberIDs, memberID)
		}
		
		pbGroup := &pb.Group{
			Id:          group.ID,
			Name:        group.ID, // Use ID as name for now
			Description: "",       // No description in current Group struct
			MemberIds:   memberIDs,
			CreatedAt:   time.Now().Unix(), // Use current time as fallback
			KeyVersion:  int32(group.KeyVersion),
		}
		pbGroups = append(pbGroups, pbGroup)
	}

	return &pb.GetGroupsResponse{
		Groups: pbGroups,
	}, nil
}

func (s *GroupService) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupResponse, error) {
	group, exists := s.groupManager.GetGroup(req.GroupId)
	
	if !exists {
		return &pb.GetGroupResponse{
			Found: false,
		}, nil
	}

	// Get member IDs
	memberIDs := make([]string, 0, len(group.Members))
	for memberID := range group.Members {
		memberIDs = append(memberIDs, memberID)
	}

	pbGroup := &pb.Group{
		Id:          group.ID,
		Name:        group.ID, // Use ID as name for now
		Description: "",       // No description in current Group struct
		MemberIds:   memberIDs,
		CreatedAt:   time.Now().Unix(), // Use current time as fallback
		KeyVersion:  int32(group.KeyVersion),
	}

	return &pb.GetGroupResponse{
		Group: pbGroup,
		Found: true,
	}, nil
}

func (s *GroupService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error) {
	// Create group using group manager
	err := s.groupManager.CreateGroup(ctx, req.GroupId)
	if err != nil {
		return &pb.CreateGroupResponse{
			Success: false,
		}, err
	}

	// Add members if provided
	for _, memberID := range req.MemberIds {
		err = s.groupManager.AddMember(req.GroupId, memberID)
		if err != nil {
			// If adding member fails, we still consider group creation successful
			// but log the error
			continue
		}
	}

	// Get the created group
	group, exists := s.groupManager.GetGroup(req.GroupId)
	if !exists {
		return &pb.CreateGroupResponse{
			Success: false,
		}, nil
	}

	// Get member IDs
	memberIDs := make([]string, 0, len(group.Members))
	for memberID := range group.Members {
		memberIDs = append(memberIDs, memberID)
	}

	pbGroup := &pb.Group{
		Id:          group.ID,
		Name:        req.Name,
		Description: req.Description,
		MemberIds:   memberIDs,
		CreatedAt:   time.Now().Unix(),
		KeyVersion:  int32(group.KeyVersion),
	}

	return &pb.CreateGroupResponse{
		Group:   pbGroup,
		Success: true,
	}, nil
}

func (s *GroupService) JoinGroup(ctx context.Context, req *pb.JoinGroupRequest) (*pb.JoinGroupResponse, error) {
	err := s.groupManager.JoinGroup(ctx, req.GroupId)
	if err != nil {
		return &pb.JoinGroupResponse{
			Success: false,
		}, err
	}

	return &pb.JoinGroupResponse{
		Success: true,
	}, nil
}

func (s *GroupService) LeaveGroup(ctx context.Context, req *pb.LeaveGroupRequest) (*pb.LeaveGroupResponse, error) {
	err := s.groupManager.LeaveGroup(req.GroupId)
	if err != nil {
		return &pb.LeaveGroupResponse{
			Success: false,
		}, err
	}

	return &pb.LeaveGroupResponse{
		Success: true,
	}, nil
}

func (s *GroupService) AddMember(ctx context.Context, req *pb.AddMemberRequest) (*pb.AddMemberResponse, error) {
	err := s.groupManager.AddMember(req.GroupId, req.MemberId)
	if err != nil {
		return &pb.AddMemberResponse{
			Success: false,
		}, err
	}

	return &pb.AddMemberResponse{
		Success: true,
	}, nil
}

func (s *GroupService) RemoveMember(ctx context.Context, req *pb.RemoveMemberRequest) (*pb.RemoveMemberResponse, error) {
	err := s.groupManager.RemoveMember(req.GroupId, req.MemberId)
	if err != nil {
		return &pb.RemoveMemberResponse{
			Success: false,
		}, err
	}

	return &pb.RemoveMemberResponse{
		Success: true,
	}, nil
}

func (s *GroupService) GetGroupMembers(ctx context.Context, req *pb.GetGroupMembersRequest) (*pb.GetGroupMembersResponse, error) {
	memberIDs, err := s.groupManager.GetGroupMembers(ctx, req.GroupId)
	if err != nil {
		return &pb.GetGroupMembersResponse{
			MemberIds: []string{},
		}, err
	}

	return &pb.GetGroupMembersResponse{
		MemberIds: memberIDs,
	}, nil
}
