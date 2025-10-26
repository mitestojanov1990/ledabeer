<!-- 56f22704-125a-4b30-a10f-f2ef4b26ab6c 94e112c2-e6d0-406c-9780-0e2f80408de5 -->
# TDD Implementation Plan: Go Libp2p Backend with Signal E2EE

## Overview

Build backend-only (Go + libp2p) implementing Signal Protocol E2EE with text messaging. Follow Red-Green-Refactor cycles strictly. Start with local development, add Docker for CI.

## Project Structure

```
ledabeer/
├── backend/
│   ├── cmd/
│   │   └── node/          # Main entry point
│   ├── internal/
│   │   ├── crypto/        # Signal Protocol E2EE layer
│   │   ├── network/       # Libp2p host, discovery, streams
│   │   ├── messaging/     # Message handling logic
│   │   └── storage/       # Peer keys, message queue
│   ├── pkg/
│   │   └── protocol/      # Public API types
│   ├── test/
│   │   ├── mocks/         # Mock implementations
│   │   └── integration/   # Multi-node tests
│   ├── go.mod
│   └── Makefile
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
└── .github/workflows/ci.yml
```

## Phase 1: Infrastructure Setup (Day 1)

### Step 1.1: Initialize Go Module & Dependencies

**Test**: None (setup step)

- Create `backend/go.mod` with Go 1.21+
- Add core dependencies:
                - `github.com/libp2p/go-libp2p` (v0.32+)
                - `github.com/stretchr/testify` (mocking/assertions)
                - `github.com/RadicalApp/libsignal-protocol-go` (Signal implementation)
- Create `Makefile` with targets: `test`, `build`, `lint`, `cover`

### Step 1.2: Setup Testing Infrastructure

**Test**: Verify test runner works

```go
// backend/internal/crypto/crypto_test.go
package crypto_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestTestingInfrastructure(t *testing.T) {
    assert.True(t, true, "Testing framework operational")
}
```

- Run `go test ./...` - should pass
- Configure coverage: `go test -cover -coverprofile=coverage.out ./...`
- Target: 80%+ coverage

## Phase 2: E2EE Layer - Signal Protocol (Days 2-4)

### Step 2.1: Identity Key Generation (TDD Cycle 1)

**RED** - Write failing test:

```go
// backend/internal/crypto/identity_test.go
func TestGenerateIdentityKeyPair(t *testing.T) {
    // Should generate unique Ed25519 identity keys
    key1, err := crypto.GenerateIdentityKeyPair()
    assert.NoError(t, err)
    assert.NotNil(t, key1.PublicKey)
    assert.NotNil(t, key1.PrivateKey)
    
    key2, err := crypto.GenerateIdentityKeyPair()
    assert.NoError(t, err)
    assert.NotEqual(t, key1.PublicKey, key2.PublicKey, "Keys must be unique")
}

func TestSerializeIdentityKeys(t *testing.T) {
    // Should serialize/deserialize without loss
    original, _ := crypto.GenerateIdentityKeyPair()
    bytes := original.Serialize()
    restored, err := crypto.DeserializeIdentityKeyPair(bytes)
    
    assert.NoError(t, err)
    assert.Equal(t, original.PublicKey, restored.PublicKey)
}
```

**GREEN** - Minimal implementation:

```go
// backend/internal/crypto/identity.go
package crypto

import (
    "crypto/rand"
    "github.com/RadicalApp/libsignal-protocol-go/ecc"
)

type IdentityKeyPair struct {
    PublicKey  *ecc.ECPublicKey
    PrivateKey *ecc.ECPrivateKey
}

func GenerateIdentityKeyPair() (*IdentityKeyPair, error) {
    keyPair, err := ecc.GenerateKeyPair()
    if err != nil {
        return nil, err
    }
    return &IdentityKeyPair{
        PublicKey:  keyPair.PublicKey(),
        PrivateKey: keyPair.PrivateKey(),
    }, nil
}

func (k *IdentityKeyPair) Serialize() []byte {
    return k.PrivateKey.Serialize()
}

func DeserializeIdentityKeyPair(data []byte) (*IdentityKeyPair, error) {
    privateKey := ecc.NewDjbECPrivateKey(data)
    return &IdentityKeyPair{
        PublicKey:  privateKey.PublicKey(),
        PrivateKey: privateKey,
    }, nil
}
```

**REFACTOR** - Add error handling, extract to interface

### Step 2.2: X3DH Key Exchange (TDD Cycle 2)

**RED** - Write failing test:

```go
// backend/internal/crypto/x3dh_test.go
func TestX3DH_InitiatorReceiverKeyAgreement(t *testing.T) {
    // Alice (initiator) and Bob (receiver) should derive same shared secret
    alice := NewX3DHSession()
    bob := NewX3DHSession()
    
    // Bob publishes prekey bundle
    bobBundle := bob.GeneratePrekeyBundle()
    
    // Alice initiates key exchange
    aliceSharedSecret, initialMsg, err := alice.InitiateKeyExchange(bobBundle)
    assert.NoError(t, err)
    
    // Bob processes initial message
    bobSharedSecret, err := bob.ProcessKeyExchange(initialMsg)
    assert.NoError(t, err)
    
    // Shared secrets must match
    assert.Equal(t, aliceSharedSecret, bobSharedSecret)
}

func TestX3DH_InvalidBundleRejected(t *testing.T) {
    alice := NewX3DHSession()
    invalidBundle := &PrekeyBundle{} // Empty bundle
    
    _, _, err := alice.InitiateKeyExchange(invalidBundle)
    assert.Error(t, err, "Should reject invalid bundle")
}
```

**GREEN** - Implementation using libsignal-protocol-go's X3DH helpers

**REFACTOR** - Extract interfaces for testability

### Step 2.3: Double Ratchet Encryption (TDD Cycle 3)

**RED** - Write failing test:

```go
// backend/internal/crypto/ratchet_test.go
func TestDoubleRatchet_EncryptDecrypt(t *testing.T) {
    // Setup two ratchets with shared secret from X3DH
    sharedSecret := make([]byte, 32)
    rand.Read(sharedSecret)
    
    aliceRatchet := NewDoubleRatchet(sharedSecret, true)  // sender
    bobRatchet := NewDoubleRatchet(sharedSecret, false)   // receiver
    
    plaintext := []byte("Hello Bob!")
    
    // Alice encrypts
    ciphertext, err := aliceRatchet.Encrypt(plaintext)
    assert.NoError(t, err)
    assert.NotEqual(t, plaintext, ciphertext)
    
    // Bob decrypts
    decrypted, err := bobRatchet.Decrypt(ciphertext)
    assert.NoError(t, err)
    assert.Equal(t, plaintext, decrypted)
}

func TestDoubleRatchet_ForwardSecrecy(t *testing.T) {
    // Old keys should not decrypt new messages
    sharedSecret := make([]byte, 32)
    rand.Read(sharedSecret)
    
    alice := NewDoubleRatchet(sharedSecret, true)
    bob := NewDoubleRatchet(sharedSecret, false)
    
    // First message
    msg1 := []byte("Message 1")
    cipher1, _ := alice.Encrypt(msg1)
    bob.Decrypt(cipher1)
    
    // Second message (ratchet advances)
    msg2 := []byte("Message 2")
    cipher2, _ := alice.Encrypt(msg2)
    
    // Try to decrypt cipher2 with old bob state - should fail
    bobOld := NewDoubleRatchet(sharedSecret, false)
    _, err := bobOld.Decrypt(cipher2)
    assert.Error(t, err, "Old keys should not decrypt new messages")
}

func TestDoubleRatchet_OutOfOrderMessages(t *testing.T) {
    // Should handle message reordering
    // (Test message chain with skipped keys)
}
```

**GREEN** - Wrap libsignal's SessionCipher

**REFACTOR** - Add message numbering, key storage

## Phase 3: Libp2p Network Layer (Days 5-7)

### Step 3.1: Host Initialization (TDD Cycle 4)

**RED** - Write failing test:

```go
// backend/internal/network/host_test.go
func TestNewHost_InitializesWithSecureTransport(t *testing.T) {
    host, err := network.NewHost(context.Background(), &Config{
        ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
    })
    assert.NoError(t, err)
    defer host.Close()
    
    // Should have Noise security
    assert.Contains(t, host.Mux().Protocols(), "/noise")
    
    // Should have TCP transport
    assert.NotEmpty(t, host.Addrs())
}

func TestNewHost_WithCustomIdentity(t *testing.T) {
    // Should use provided identity key for peer ID
    identity, _ := crypto.GenerateIdentityKeyPair()
    
    host, err := network.NewHost(context.Background(), &Config{
        IdentityKey: identity,
    })
    assert.NoError(t, err)
    defer host.Close()
    
    // PeerID should derive from identity key
    expectedPeerID := crypto.PeerIDFromPublicKey(identity.PublicKey)
    assert.Equal(t, expectedPeerID, host.ID())
}
```

**GREEN** - Implementation:

```go
// backend/internal/network/host.go
package network

import (
    "context"
    "github.com/libp2p/go-libp2p"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/crypto"
)

type Config struct {
    ListenAddrs []string
    IdentityKey *crypto.PrivKey
}

func NewHost(ctx context.Context, cfg *Config) (host.Host, error) {
    opts := []libp2p.Option{
        libp2p.ListenAddrStrings(cfg.ListenAddrs...),
        libp2p.Security("/noise", noise.New),
    }
    
    if cfg.IdentityKey != nil {
        opts = append(opts, libp2p.Identity(*cfg.IdentityKey))
    }
    
    return libp2p.New(opts...)
}
```

**REFACTOR** - Extract config validation

### Step 3.2: Stream-Based Messaging (TDD Cycle 5)

**RED** - Write failing test:

```go
// backend/internal/network/stream_test.go
func TestStream_SendReceiveMessage(t *testing.T) {
    // Setup two hosts
    ctx := context.Background()
    host1, _ := network.NewHost(ctx, &Config{ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"}})
    host2, _ := network.NewHost(ctx, &Config{ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"}})
    defer host1.Close()
    defer host2.Close()
    
    // Connect hosts
    host1.Connect(ctx, peer.AddrInfo{
        ID:    host2.ID(),
        Addrs: host2.Addrs(),
    })
    
    // Setup stream handler on host2
    received := make(chan []byte, 1)
    handler := NewStreamHandler(func(data []byte) {
        received <- data
    })
    host2.SetStreamHandler("/chat/1.0.0", handler.Handle)
    
    // Send from host1
    stream, err := host1.NewStream(ctx, host2.ID(), "/chat/1.0.0")
    assert.NoError(t, err)
    
    testData := []byte("test message")
    err = WriteMessage(stream, testData)
    assert.NoError(t, err)
    
    // Verify receipt
    select {
    case msg := <-received:
        assert.Equal(t, testData, msg)
    case <-time.After(2 * time.Second):
        t.Fatal("Message not received")
    }
}
```

**GREEN** - Implement stream helpers

**REFACTOR** - Add length-prefixed framing, error handling

### Step 3.3: Peer Discovery with DHT (TDD Cycle 6)

**RED** - Write failing test with mock DHT

**GREEN** - Integrate Kademlia DHT

**REFACTOR** - Add bootstrap peer config

## Phase 4: Integration - E2EE + Messaging (Days 8-10)

### Step 4.1: End-to-End Message Flow (Integration Test)

**RED** - Write failing integration test:

```go
// backend/test/integration/e2e_message_test.go
func TestE2E_EncryptedMessageExchange(t *testing.T) {
    ctx := context.Background()
    
    // Start two nodes
    alice := startTestNode(t, ctx, "alice")
    bob := startTestNode(t, ctx, "bob")
    defer alice.Shutdown()
    defer bob.Shutdown()
    
    // Exchange identity keys (simulate out-of-band)
    alice.AddTrustedPeer(bob.PeerID(), bob.IdentityPublicKey())
    bob.AddTrustedPeer(alice.PeerID(), alice.IdentityPublicKey())
    
    // Alice sends encrypted message to Bob
    plaintext := "Secret message"
    err := alice.SendMessage(ctx, bob.PeerID(), plaintext)
    assert.NoError(t, err)
    
    // Bob receives and decrypts
    select {
    case msg := <-bob.IncomingMessages():
        assert.Equal(t, plaintext, msg.Content)
        assert.Equal(t, alice.PeerID(), msg.From)
    case <-time.After(5 * time.Second):
        t.Fatal("Message not received")
    }
    
    // Verify no plaintext leaked
    // (inspect network layer - should only see ciphertext)
}
```

**GREEN** - Wire together crypto + network layers in `messaging` package

**REFACTOR** - Add message queue, retry logic

### Step 4.2: Key Exchange Protocol (Integration Test)

**RED** - Test full X3DH exchange over libp2p:

```go
func TestE2E_X3DHKeyExchange(t *testing.T) {
    // Bob publishes prekey bundle to DHT
    // Alice fetches bundle and initiates
    // Both establish same session
}
```

**GREEN** - Implement prekey service

## Phase 5: Docker & CI Setup (Day 11)

### Step 5.1: Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go test ./...
RUN go build -o /node ./cmd/node

FROM alpine:latest
COPY --from=builder /node /node
EXPOSE 4001
CMD ["/node"]
```

### Step 5.2: Docker Compose for Multi-Node Testing

```yaml
version: '3.8'
services:
  bootstrap:
    build: ./backend
    ports: ["4001:4001"]
    environment:
      - NODE_TYPE=bootstrap
  
  node1:
    build: ./backend
    depends_on: [bootstrap]
    environment:
      - BOOTSTRAP_PEER=/ip4/bootstrap/tcp/4001/p2p/...
  
  node2:
    build: ./backend
    depends_on: [bootstrap]
```

### Step 5.3: GitHub Actions CI

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -race -cover ./...
      - run: docker-compose up --abort-on-container-exit
```

## Testing Strategy Summary

### Unit Tests (Fast, Isolated)

- Crypto primitives (identity, X3DH, ratchet)
- Network helpers (streams, framing)
- Message serialization
- **Coverage goal**: 90%+

### Integration Tests (Multi-Node, In-Process)

- Full message flow (2 nodes in same process)
- Key exchange protocols
- Connection handling
- **Run on**: Every PR

### E2E Tests (Docker, Real Network)

- Multi-container scenarios
- Network partitions
- Discovery/bootstrap
- **Run on**: Nightly + releases

## Success Metrics

- [ ] All tests pass in CI
- [ ] 80%+ code coverage
- [ ] Zero plaintext message leaks (verified by inspection)
- [ ] Sub-second message delivery (2 local nodes)
- [ ] Successful key exchange in <3 seconds

## Next Steps After This Phase

1. Add message persistence (offline queuing)
2. Implement group messaging (MLS)
3. Add media transfer (chunked streams)
4. WebRTC signaling for calls
5. Build frontend (Waku + React)

### To-dos

- [ ] Initialize Go module, dependencies, Makefile, and verify test infrastructure works
- [ ] TDD: Identity key generation and serialization (Ed25519)
- [ ] TDD: X3DH key exchange protocol with prekey bundles
- [ ] TDD: Double Ratchet encryption/decryption with forward secrecy tests
- [ ] TDD: Libp2p host initialization with Noise security
- [ ] TDD: Stream-based message sending/receiving between two hosts
- [ ] TDD: Peer discovery using Kademlia DHT with bootstrap peers
- [ ] Integration test: End-to-end encrypted message exchange between two nodes
- [ ] Integration test: Complete X3DH key exchange over libp2p streams
- [ ] Setup Dockerfile, docker-compose for multi-node testing, and GitHub Actions CI