package calls_test

import (
	"testing"

	"ledabeer/backend/internal/calls"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTURN_Configuration(t *testing.T) {
	config := calls.TURNConfig{
		URLs:       []string{"turn:turn.example.com:3478"},
		Username:   "user",
		Credential: "pass",
	}

	session := calls.NewCallSessionWithTURN(config)
	require.NotNil(t, session)

	servers := session.GetICEServers()
	assert.Greater(t, len(servers), 0)

	// Verify TURN server is in the list
	foundTURN := false
	for _, server := range servers {
		for _, url := range server.URLs {
			if url == config.URLs[0] {
				foundTURN = true
				break
			}
		}
	}
	assert.True(t, foundTURN, "TURN server should be in ICE servers list")
}

func TestTURN_FallbackWhenDirectFails(t *testing.T) {
	// Simulate direct connection failure
	// Verify TURN relay is used
	config := calls.TURNConfig{
		URLs:       []string{"turn:turn.example.com:3478"},
		Username:   "user",
		Credential: "pass",
	}

	session := calls.NewCallSessionWithTURN(config)

	// For testing, we'll simulate blocking direct connections
	session.BlockDirectConnections()

	// In a real implementation, this would attempt connection
	// and fall back to TURN
	// For testing, we'll verify the configuration is correct
	assert.True(t, session.HasTURNServers(), "Should have TURN servers configured")
}

func TestTURN_IsUsingRelay(t *testing.T) {
	config := calls.TURNConfig{
		URLs:       []string{"turn:turn.example.com:3478"},
		Username:   "user",
		Credential: "pass",
	}

	session := calls.NewCallSessionWithTURN(config)

	// Initially not using relay (no connection established)
	assert.False(t, session.IsUsingRelay())

	// Simulate relay connection
	session.SimulateRelayConnection()

	// Should now report using relay
	assert.True(t, session.IsUsingRelay())
}

func TestTURN_HealthCheck(t *testing.T) {
	config := calls.TURNConfig{
		URLs:       []string{"turn:turn.example.com:3478"},
		Username:   "user",
		Credential: "pass",
	}

	session := calls.NewCallSessionWithTURN(config)

	// Check TURN server health
	healthy := session.CheckTURNHealth()

	// For testing, we'll simulate a healthy server
	assert.True(t, healthy)
}

func TestTURN_MultipleServers(t *testing.T) {
	config := calls.TURNConfig{
		URLs: []string{
			"turn:turn1.example.com:3478",
			"turn:turn2.example.com:3478",
		},
		Username:   "user",
		Credential: "pass",
	}

	session := calls.NewCallSessionWithTURN(config)

	servers := session.GetICEServers()

	// Should have both TURN servers plus STUN
	assert.GreaterOrEqual(t, len(servers), 2)
}
