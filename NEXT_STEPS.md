# Ledabeer - Next Steps

## Current Status

### ✅ Completed
- **Backend (Go + Libp2p)**: Fully implemented with E2EE
  - Identity key generation (Ed25519)
  - X3DH key exchange protocol
  - Double Ratchet encryption/decryption
  - Libp2p host with secure transports
  - Stream-based messaging
  - Peer discovery with Kademlia DHT
  - PubSub for group messaging
  - Test coverage with TDD approach

- **Frontend (React Native/Expo)**: Basic E2EE chat interface
  - Mock backend service for development
  - Zustand state management
  - Chat list screen with peer status
  - Chat conversation screen
  - Real-time messaging simulation
  - E2EE UI indicators
  - Running on web at http://localhost:8083

## Next Steps

### Phase 1: Backend Integration (High Priority)

#### 1.1 Backend API Layer
Create an API layer to connect the frontend to the Go backend:

**Options:**
- **WebSocket API** (Recommended for real-time)
  - Create WebSocket server in Go backend
  - Handle connection lifecycle
  - Stream messages bidirectionally
  - Event-based architecture (connect, disconnect, message, etc.)

- **gRPC API** (Alternative)
  - Define protobuf schemas
  - Generate client/server code
  - Stream support for real-time updates

**Files to create:**
```
backend/
├── internal/
│   └── api/
│       ├── websocket/
│       │   ├── server.go      # WebSocket server
│       │   ├── handler.go     # Message handlers
│       │   └── client.go      # Client connection manager
│       └── models/
│           └── messages.go    # API message types
```

#### 1.2 Frontend Backend Client
Replace the mock backend with a real backend client:

**Files to create/modify:**
```
frontend/ledabeer-mobile/src/services/
├── backendClient.ts           # WebSocket/gRPC client
├── e2eeService.ts            # E2EE operations (placeholder for now)
└── mockBackend.ts            # Keep for testing
```

**Key features:**
- WebSocket connection management
- Automatic reconnection
- Message queue for offline support
- Error handling and retry logic

#### 1.3 Integration Testing
Test the full stack end-to-end:

**Tasks:**
- Start Go backend server
- Connect frontend to backend
- Test peer discovery
- Test message sending/receiving
- Verify E2EE encryption/decryption

### Phase 2: E2EE Client-Side Implementation (Critical Security)

The backend already has Signal Protocol E2EE. Now implement client-side crypto:

#### 2.1 Choose E2EE Library for TypeScript/React Native

**Options:**
1. **@privacyresearch/libsignal-protocol-typescript**
   - Pure TypeScript implementation
   - Works in React Native
   - Compatible with Signal Protocol

2. **libsignal-client** (Official)
   - Rust-based with WASM/Native bindings
   - More secure, audited
   - Requires native module setup

**Recommended:** Start with TypeScript version, migrate to official later

#### 2.2 Implement Client-Side E2EE

**Files to create:**
```
frontend/ledabeer-mobile/src/crypto/
├── identity.ts               # Identity key management
├── x3dh.ts                  # X3DH key exchange
├── ratchet.ts               # Double Ratchet
├── keyStorage.ts            # Secure key storage (MMKV)
└── e2eeManager.ts           # High-level E2EE API
```

**Key tasks:**
- Generate and store identity keys securely
- Implement X3DH handshake with backend peers
- Encrypt messages before sending
- Decrypt messages after receiving
- Key rotation and forward secrecy

#### 2.3 Secure Storage
Use react-native-mmkv (already installed) for secure key storage:

**Implementation:**
- Store private keys encrypted
- Store session states
- Store prekey bundles
- Implement key deletion on logout

### Phase 3: Enhanced Features

#### 3.1 Group Messaging
**Backend:** Already has PubSub implementation
**Frontend:** Build group chat UI
- Group creation screen
- Group member management
- Group encryption (MLS or Sender Keys)

#### 3.2 Media Sharing
**Backend:** Implement chunked file transfer
**Frontend:**
- Image picker integration
- Video picker integration
- File upload progress
- Thumbnail generation
- Media gallery

**Libraries to add:**
```bash
npx expo install expo-image-picker expo-media-library
npx expo install expo-video expo-image
```

#### 3.3 Voice/Video Calling
**Backend:** Already has WebRTC signaling via libp2p
**Frontend:**
- Integrate WebRTC client
- Call UI (incoming/outgoing)
- Audio/video controls
- Screen sharing

**Libraries to add:**
```bash
npx expo install react-native-webrtc
```

### Phase 4: Production Readiness

#### 4.1 Offline Support
- Message queue for offline sending
- Local message persistence
- Sync on reconnection
- Conflict resolution

#### 4.2 Push Notifications
- Firebase Cloud Messaging integration
- Background message handling
- Notification UI

#### 4.3 Performance Optimization
- Message pagination
- Virtual list optimization (already using FlashList)
- Image caching
- Bundle size optimization

#### 4.4 Security Hardening
- Security audit of E2EE implementation
- Rate limiting
- Anti-spam measures
- Secure key backup/recovery
- Certificate pinning for API calls

#### 4.5 Testing
- Unit tests for crypto operations
- Integration tests for backend communication
- E2E tests with Detox
- Performance testing
- Security penetration testing

### Phase 5: Deployment

#### 5.1 Backend Deployment
- Dockerize backend (Dockerfile already exists)
- Set up Kubernetes/Docker Compose
- Configure bootstrap peers
- Set up monitoring (Prometheus/Grafana)
- Set up logging

#### 5.2 Frontend Deployment
- Build iOS app (TestFlight)
- Build Android app (Google Play)
- Web deployment (optional)
- App Store submissions

### Phase 6: Future Enhancements

- Message search
- Message reactions
- Read receipts (privacy-preserving)
- Typing indicators
- User profiles with avatars
- Status updates
- Message forwarding
- Multi-device support
- Desktop app (Electron)
- Message backups (encrypted)

## Immediate Action Items (Priority Order)

### Week 1: Backend API Integration
1. ✅ Create WebSocket server in Go backend
2. ✅ Implement basic message routing
3. ✅ Create frontend WebSocket client
4. ✅ Test basic message flow (without E2EE)

### Week 2: Client-Side E2EE
1. ✅ Add libsignal library to frontend
2. ✅ Implement key generation and storage
3. ✅ Implement X3DH key exchange
4. ✅ Implement message encryption/decryption
5. ✅ Test E2EE flow end-to-end

### Week 3: Core Features
1. ✅ Implement proper peer discovery UI
2. ✅ Add contact management
3. ✅ Implement group chats
4. ✅ Add message status indicators
5. ✅ Implement offline message queue

### Week 4: Polish & Testing
1. ✅ Bug fixes and UI polish
2. ✅ Performance optimization
3. ✅ Security testing
4. ✅ Write documentation
5. ✅ Prepare for beta release

## Technical Decisions to Make

1. **WebSocket vs gRPC for API?**
   - Recommendation: WebSocket (simpler, better for React Native)

2. **Which Signal library for frontend?**
   - Recommendation: Start with TypeScript version, evaluate official later

3. **Key storage strategy?**
   - Recommendation: react-native-mmkv with encryption

4. **Peer discovery UI?**
   - Options: QR code scanning, manual peer ID entry, DHT browse
   - Recommendation: QR codes + manual entry

5. **Group encryption?**
   - Options: MLS (modern), Sender Keys (simpler)
   - Recommendation: Start with Sender Keys, migrate to MLS

## Resources

### Backend
- Go Libp2p examples: https://github.com/libp2p/go-libp2p/tree/master/examples
- Signal Protocol: https://signal.org/docs/
- WebRTC with Pion: https://github.com/pion/webrtc

### Frontend
- React Native docs: https://reactnative.dev
- Expo docs: https://docs.expo.dev
- libsignal-protocol-typescript: https://github.com/privacyresearch/libsignal-protocol-typescript
- React Native MMKV: https://github.com/mrousavy/react-native-mmkv

### Testing
- Jest: https://jestjs.io
- React Native Testing Library: https://callstack.github.io/react-native-testing-library
- Detox: https://wix.github.io/Detox

## Questions to Address

1. **Bootstrap peers**: How will users find each other initially?
2. **Username system**: Do we need usernames or just peer IDs?
3. **Backup/recovery**: How to handle key loss?
4. **Privacy policy**: What data do we store/log?
5. **Compliance**: GDPR, data retention, etc.

## Getting Help

- Backend issues: Check backend/README.md
- Frontend issues: Check frontend/ledabeer-mobile/README.md
- Crypto/security: Refer to Signal Protocol docs
- Libp2p: Join libp2p Slack/Discord

---

**Current App Status:** ✅ Running at http://localhost:8083 with mock backend

**Next Immediate Step:** Implement WebSocket API in backend to connect frontend to real E2EE backend
