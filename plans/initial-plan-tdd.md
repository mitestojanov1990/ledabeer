### Detailed TDD Plan for the Decentralized Chat App

This plan applies **Test-Driven Development (TDD)** to the Golang backend (using libp2p) and JavaScript frontend (using Waku), based on the architectures outlined previously. TDD follows the cycle: **Red (write failing test) → Green (write minimal code to pass) → Refactor (improve code while keeping tests green)**. We'll break it down by component, starting with unit tests and progressing to integration/E2E tests.

The plan assumes:
- **Tools**: Go's `testing` package for backend; Jest/Vitest for frontend (with jsdom for DOM simulation).
- **Coverage Goal**: Aim for 80%+ coverage; use `go test -cover` and `npm test -- --coverage`.
- **CI Integration**: Run tests in GitHub Actions/Docker for every PR.
- **Scope**: Focus on core features (E2EE messaging, media transfer, calls) while ensuring security invariants (e.g., no plaintext leaks).

Testing is iterative: For each feature, write tests before code, then implement in small increments.

#### Overall TDD Workflow
1. **Setup Project Structure**: Initialize repos with test folders (e.g., `backend/internal/pkg/chat_test.go`, `frontend/src/components/Chat.test.tsx`).
2. **Mock Dependencies**: Use mocks for libp2p/Waku (e.g., `github.com/stretchr/testify/mock` in Go; `vi.mock` in JS) to isolate units.
3. **Run Cycles**: For each test, expect failure first, then code to pass, then refactor (e.g., extract helpers).
4. **Progression**: Unit → Integration → E2E.
5. **Security Focus**: Include fuzzing/property-based tests for crypto (e.g., using `go-fuzz` or `fast-check` in JS).

---

### Backend TDD Plan (Golang with Libp2p)

Focus on the core libp2p host, E2EE integration, messaging streams, media chunking, and WebRTC signaling.

#### 1. **Unit Tests (Isolated Components)**
   - **E2EE Layer (Signal Protocol Integration)**
     | Step | Red (Failing Test) | Green (Minimal Code) | Refactor |
     |------|--------------------|-----------------------|----------|
     | 1 | Test key generation: Assert Ed25519 keys are created and unique. | Implement `GenerateIdentity()` returning keys. | Extract to `crypto` package. |
     | 2 | Test X3DH key exchange: Simulate two peers; assert shared secret matches. | Add `PerformX3DH()` with Curve25519. | Use interfaces for pluggable crypto. |
     | 3 | Test Double Ratchet: Encrypt/decrypt message; assert decryption fails with wrong key. | Integrate `libsignal-go`; handle ratchet steps. | Add error handling for key exhaustion. |
     | 4 | Test group MLS: Add/remove members; assert only members decrypt. | Implement sender keys for groups. | Fuzz test with random inputs. |

   - **Libp2p Host and Discovery**
     | Step | Red (Failing Test) | Green (Minimal Code) | Refactor |
     |------|--------------------|-----------------------|----------|
     | 1 | Test host init: Assert Noise/TLS secures connections. | Use `libp2p.New()` with secure opts. | Config via struct for test overrides. |
     | 2 | Test peer discovery: Mock DHT; assert finds peer by ID. | Enable Kademlia; query mock peers. | Use test swarm for multi-node sim. |
     | 3 | Test NAT traversal: Simulate firewall; assert hole-punching connects. | Add AutoNAT; mock network. | Integrate with libp2p's test utils. |

   - **Messaging (Streams and PubSub)**
     | Step | Red (Failing Test) | Green (Minimal Code) | Refactor |
     |------|--------------------|-----------------------|----------|
     | 1 | Test 1:1 stream: Open stream; send/receive encrypted blob; assert integrity. | Use `host.NewStream()`; wrap in E2EE. | Add timeouts and retries. |
     | 2 | Test pubsub: Subscribe to topic; publish encrypted msg; assert received. | Init GossipSub; handle topics. | Validate topics to prevent spam. |
     | 3 | Test offline queuing: Mock disconnect; store in IPFS; retrieve on reconnect. | Integrate go-ipfs-api; query CID. | Use LRU cache for recent messages. |

   - **Media Handling (Images/Videos)**
     | Step | Red (Failing Test) | Green (Minimal Code) | Refactor |
     |------|--------------------|-----------------------|----------|
     | 1 | Test chunking: Split 1MB file; assert reassembled matches original. | Use `io.Reader` for streams; chunk at 64KB. | Add compression (snappy). |
     | 2 | Test secure transfer: Encrypt chunks; assert decrypt fails without key. | Wrap in Signal; send over stream. | Hash checks for integrity. |
     | 3 | Test limits: Reject >10MB; assert error thrown. | Check size in handler. | Configurable limits via env. |

   - **Calls (WebRTC Signaling)**
     | Step | Red (Failing Test) | Green (Minimal Code) | Refactor |
     |------|--------------------|-----------------------|----------|
     | 1 | Test SDP exchange: Send offer; assert answer generated. | Use pion/webrtc; encrypt SDP. | Mock PeerConnection. |
     | 2 | Test ICE candidates: Gather and exchange; assert connection established. | Handle trickle ICE over stream. | Add STUN/TURN fallback. |
     | 3 | Test group signaling: Broadcast to multiple; assert all connect. | Use pubsub for multi-party. | Handle renegotiation. |

#### 2. **Integration Tests**
   - Simulate full flows: Use libp2p's test network (e.g., `test.NewTestNetwork`) to run two nodes.
   - Examples:
     - End-to-end message: Node A encrypts/sends; Node B decrypts; assert no leaks.
     - Media transfer: Send video chunk-by-chunk; verify reassembly.
     - Call setup: Establish WebRTC peer; send dummy RTP packet.
   - Security: Use property-based testing (e.g., quickcheck) for crypto invariants.

#### 3. **E2E Tests**
   - Run in Docker: Spin up multiple containers (nodes + mock clients).
   - Tools: Use `github.com/ory/dockertest` for orchestration.
   - Cover: Full chat session with disconnects, media, and calls.

---

### Frontend TDD Plan (JavaScript with Waku)

Focus on React components, Waku client, E2EE, UI interactions, and WebRTC.

#### 1. **Unit Tests (Isolated Components)**
   - **Waku Client and E2EE**
     | Step | Red (Failing Test) | Green (Minimal Code) | Refactor |
     |------|--------------------|-----------------------|----------|
     | 1 | Test node init: Assert light node connects to bootstrap. | Use `createLightNode()`; mock peers. | Env-based bootstrap config. |
     | 2 | Test key exchange: Simulate X3DH; assert shared key. | Integrate libsignal-client; handle promises. | Async hooks for testing. |
     | 3 | Test ratchet: Encrypt/decrypt; assert forward secrecy (old keys fail). | Advance ratchet; test decryption. | Fuzz with random messages. |

   - **Messaging UI**
     | Step | Red (Failing Test) | Green (Minimal Code) | Refactor |
     |------|--------------------|-----------------------|----------|
     | 1 | Test chat component: Render messages; assert displayed. | Use React Testing Library; mock state. | Use custom render hooks. |
     | 2 | Test send: Input text; encrypt/publish; assert Waku called. | Mock `@waku/sdk`; spy on publish. | Debounce input for perf. |
     | 3 | Test offline: Mock disconnect; query store on reconnect. | Use Waku Store API; assert history loads. | Error boundaries for failures. |

   - **Media Handling**
     | Step | Red (Failing Test) | Green (Minimal Code) | Refactor |
     |------|--------------------|-----------------------|----------|
     | 1 | Test file picker: Select image; assert preview URL. | Mock File API; render <img>. | Compress before encrypt. |
     | 2 | Test upload: Chunk/encrypt; send via Waku; assert progress. | Use Web Workers for chunking. | Cancelable uploads. |
     | 3 | Test display: Receive encrypted blob; decrypt/render video. | Mock media element; assert src set. | Lazy load for performance. |

   - **Calls UI**
     | Step | Red (Failing Test) | Green (Minimal Code) | Refactor |
     |------|--------------------|-----------------------|----------|
     | 1 | Test call button: Click to initiate; assert signaling sent. | Mock WebRTC; encrypt offer. | State machine for call states. |
     | 2 | Test video stream: Get user media; assert <video> plays. | Use `navigator.mediaDevices`; mock streams. | Handle permissions errors. |
     | 3 | Test group: Add participants; assert multi-streams. | Broadcast signaling via pubsub. | Mute/unmute controls. |

#### 2. **Integration Tests**
   - Combine components: Use `@testing-library/react` with mocked Waku.
   - Examples:
     - Full chat flow: Type message → Encrypt → Publish → Receive → Decrypt → Render.
     - Media: Upload image → Chunk → Send → Download → Display.
     - Call: Initiate → Exchange SDP → Establish connection → Teardown.

#### 3. **E2E Tests**
   - Tools: Cypress/Playwright for browser; Appium for mobile (if React Native).
   - Simulate: Open two tabs; send message/media/call; assert end-to-end.
   - Mobile-Specific: Use emulators first, then real devices for NAT/push.

---

### Cross-Project TDD Considerations
- **Backend-Frontend Integration**: Use contract tests (e.g., Pact) for API schemas (e.g., message formats).
- **Security Testing**: 
  - Fuzz crypto inputs.
  - Static analysis: `gosec` for Go; ESLint-security for JS.
  - Penetration: Simulate MITM attacks on libp2p/Waku.
- **Performance**: Benchmark under load (e.g., 100 peers); use `go test -bench`.
- **Edge Cases**: Disconnects, invalid keys, large files, low bandwidth.
- **Iteration Cadence**: Commit after each green phase; run all tests before merge.
- **Tools Setup**:
  - Backend: `go mod init`; add `testify` for assertions.
  - Frontend: `create-react-app` or Vite; add `vitest`, `@testing-library`.

This TDD plan ensures robust, secure code from the start. If a feature changes, update tests first. Want code examples for a specific test (e.g., E2EE unit)? Let me know!