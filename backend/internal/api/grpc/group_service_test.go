package grpc

import (
	"context"
	"testing"

	"ledabeer/backend/internal/messaging"
	"ledabeer/backend/internal/network"
	pb "ledabeer/backend/pkg/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupService_GetGroups(t *testing.T) {
	// Arrange
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := messaging.NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create some test groups
	err = groupManager.CreateGroup(ctx, "test-group-1")
	require.NoError(t, err)
	err = groupManager.CreateGroup(ctx, "test-group-2")
	require.NoError(t, err)

	service := NewGroupService(groupManager)

	// Act
	req := &pb.GetGroupsRequest{}
	resp, err := service.GetGroups(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Groups, 2)
	
	// Check that both groups are present
	groupIDs := make(map[string]bool)
	for _, group := range resp.Groups {
		groupIDs[group.Id] = true
	}
	assert.True(t, groupIDs["test-group-1"])
	assert.True(t, groupIDs["test-group-2"])
}

func TestGroupService_GetGroup(t *testing.T) {
	// Arrange
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := messaging.NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create a test group
	err = groupManager.CreateGroup(ctx, "test-group")
	require.NoError(t, err)

	service := NewGroupService(groupManager)

	// Act - Get existing group
	req := &pb.GetGroupRequest{GroupId: "test-group"}
	resp, err := service.GetGroup(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Found)
	assert.Equal(t, "test-group", resp.Group.Id)

	// Act - Get non-existing group
	req = &pb.GetGroupRequest{GroupId: "non-existing-group"}
	resp, err = service.GetGroup(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Found)
	assert.Nil(t, resp.Group)
}

func TestGroupService_CreateGroup(t *testing.T) {
	// Arrange
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := messaging.NewGroupManager(ctx, host)
	require.NoError(t, err)

	service := NewGroupService(groupManager)

	// Act
	req := &pb.CreateGroupRequest{
		GroupId:     "new-group",
		Name:        "Test Group",
		Description: "A test group",
		MemberIds:   []string{"peer1", "peer2"},
	}
	resp, err := service.CreateGroup(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Group)
	assert.Equal(t, "new-group", resp.Group.Id)
	assert.Equal(t, "Test Group", resp.Group.Name)
	assert.Equal(t, "A test group", resp.Group.Description)
}

func TestGroupService_CreateGroup_Duplicate(t *testing.T) {
	// Arrange
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := messaging.NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create a group first
	err = groupManager.CreateGroup(ctx, "existing-group")
	require.NoError(t, err)

	service := NewGroupService(groupManager)

	// Act - Try to create duplicate group
	req := &pb.CreateGroupRequest{
		GroupId:     "existing-group",
		Name:        "Duplicate Group",
		Description: "This should fail",
		MemberIds:   []string{},
	}
	resp, err := service.CreateGroup(ctx, req)

	// Assert
	require.Error(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestGroupService_JoinGroup(t *testing.T) {
	// Arrange
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := messaging.NewGroupManager(ctx, host)
	require.NoError(t, err)

	service := NewGroupService(groupManager)

	// Act
	req := &pb.JoinGroupRequest{GroupId: "group-to-join"}
	resp, err := service.JoinGroup(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestGroupService_LeaveGroup(t *testing.T) {
	// Arrange
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := messaging.NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create a group first
	err = groupManager.CreateGroup(ctx, "group-to-leave")
	require.NoError(t, err)

	service := NewGroupService(groupManager)

	// Act
	req := &pb.LeaveGroupRequest{GroupId: "group-to-leave"}
	resp, err := service.LeaveGroup(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestGroupService_AddMember(t *testing.T) {
	// Arrange
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := messaging.NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create a group first
	err = groupManager.CreateGroup(ctx, "group-for-member")
	require.NoError(t, err)

	service := NewGroupService(groupManager)

	// Act
	req := &pb.AddMemberRequest{
		GroupId:  "group-for-member",
		MemberId: "new-member",
	}
	resp, err := service.AddMember(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestGroupService_RemoveMember(t *testing.T) {
	// Arrange
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := messaging.NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create a group and add a member first
	err = groupManager.CreateGroup(ctx, "group-for-removal")
	require.NoError(t, err)
	err = groupManager.AddMember("group-for-removal", "member-to-remove")
	require.NoError(t, err)

	service := NewGroupService(groupManager)

	// Act
	req := &pb.RemoveMemberRequest{
		GroupId:  "group-for-removal",
		MemberId: "member-to-remove",
	}
	resp, err := service.RemoveMember(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestGroupService_GetGroupMembers(t *testing.T) {
	// Arrange
	ctx := context.Background()
	host, err := network.NewHost(ctx, &network.Config{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer host.Close()

	groupManager, err := messaging.NewGroupManager(ctx, host)
	require.NoError(t, err)

	// Create a group and add members
	err = groupManager.CreateGroup(ctx, "group-with-members")
	require.NoError(t, err)
	err = groupManager.AddMember("group-with-members", "member1")
	require.NoError(t, err)
	err = groupManager.AddMember("group-with-members", "member2")
	require.NoError(t, err)

	service := NewGroupService(groupManager)

	// Act
	req := &pb.GetGroupMembersRequest{GroupId: "group-with-members"}
	resp, err := service.GetGroupMembers(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.MemberIds, 3) // Current user + 2 added members
}
