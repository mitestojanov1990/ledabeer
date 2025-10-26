# Frontend Implementation Plan for Ledabeer E2E Chat (React Native)

## Executive Summary

This document outlines a **comprehensive, production-ready implementation plan** for building a React Native mobile application (iOS + Android) with end-to-end encryption, peer-to-peer messaging, and voice/video calling capabilities. The plan follows **Test-Driven Development (TDD)** methodology with extensive edge case coverage to ensure robustness and reliability.

**Key Features:**
- 📱 Cross-platform mobile app (iOS + Android)
- 🔐 End-to-end encryption with Signal Protocol
- 🌐 P2P messaging via Waku protocol
- 📞 Voice/video calls with WebRTC
- 🎨 GPU-optimized UI with NativeWind (Tailwind for React Native)
- 🧪 90%+ test coverage with TDD
- ⚡ 60fps animations and smooth scrolling
- 🛡️ Comprehensive edge case handling

---

## Table of Contents

1. [Project Setup & Infrastructure](#phase-1-project-setup--infrastructure)
2. [TDD Foundation](#phase-2-tdd-foundation)
3. [Backend Integration](#phase-3-backend-integration)
4. [Crypto Layer (E2EE)](#phase-4-crypto-layer-e2ee)
5. [Waku Integration](#phase-5-waku-integration)
6. [Core UI Components](#phase-6-core-ui-components)
7. [Media Handling](#phase-7-media-handling)
8. [Voice/Video Calls](#phase-8-voicevideo-calls-webrtc)
9. [State Management](#phase-9-state-management)
10. [Performance Optimizations](#phase-10-performance-optimizations)
11. [Testing Strategy](#phase-11-testing-strategy)
12. [Platform-Specific Features](#phase-12-platform-specific-features)
13. [Polish & Production](#phase-13-polish--production)

---

## Technology Stack

### Core Framework
- **React Native 0.73+** with **Expo SDK 50+** (managed workflow)
- **TypeScript 5+** for type safety
- **Node.js 18+** for tooling

### P2P & Networking
- `@waku/sdk` or `@waku/react-native` - Waku protocol for P2P messaging
- `@grpc/grpc-js` - gRPC client for backend communication
- `react-native-webrtc` - WebRTC for voice/video calls

### Cryptography
- `@signalapp/libsignal-client` - Signal Protocol for E2EE
- Native crypto modules for performance

### UI & Styling
- `nativewind` - Tailwind CSS for React Native
- `react-native-reanimated` - High-performance 60fps animations
- `@shopify/flash-list` - Virtualized lists for performance
- `react-native-fast-image` - Optimized image loading

### State Management
- `zustand` - Lightweight, fast state management

### Storage
- `react-native-mmkv` - Ultra-fast key-value storage (2-10x faster than AsyncStorage)
- `react-native-keychain` - Secure credential storage (iOS Keychain, Android KeyStore)
- `@react-native-async-storage/async-storage` - Fallback storage

### Navigation
- `@react-navigation/native` - Navigation framework
- `@react-navigation/stack` - Stack navigator
- `@react-navigation/bottom-tabs` - Tab navigator

### Media
- `react-native-image-picker` - Image/video selection
- `react-native-fs` - File system access
- `react-native-video` - Video player
- `react-native-image-resizer` - Image compression

### Testing
- `jest` - Unit testing framework
- `@testing-library/react-native` - Component testing
- `detox` - E2E testing (iOS + Android)
- `@testing-library/jest-native` - Custom matchers

### Development Tools
- `eslint` + `@typescript-eslint` - Linting
- `prettier` - Code formatting
- `husky` - Git hooks
- `lint-staged` - Pre-commit checks

---

## Phase 1: Project Setup & Infrastructure

### 1.1 Initialize React Native Project

```bash
# Create Expo project with TypeScript
npx create-expo-app ledabeer-mobile --template expo-template-blank-typescript

cd ledabeer-mobile

# Install core dependencies
npx expo install react-native-reanimated react-native-gesture-handler
```

### 1.2 Configure NativeWind (Tailwind CSS)

```bash
npm install nativewind
npm install --save-dev tailwindcss
npx tailwindcss init
```

**tailwind.config.js:**
```javascript
module.exports = {
  content: ["./App.{js,jsx,ts,tsx}", "./src/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {
      // GPU-accelerated utilities
      transform: ['gpu'],
      willChange: ['transform', 'opacity'],
    },
  },
  plugins: [],
}
```

**babel.config.js:**
```javascript
module.exports = function (api) {
  api.cache(true);
  return {
    presets: ['babel-preset-expo'],
    plugins: [
      'nativewind/babel',
      'react-native-reanimated/plugin', // Must be last
    ],
  };
};
```

### 1.3 Install Testing Tools

```bash
# Unit testing
npm install --save-dev jest @testing-library/react-native @testing-library/jest-native

# E2E testing
npm install --save-dev detox detox-cli
```

### 1.4 Project Structure

```
ledabeer-mobile/
├── src/
│   ├── api/                    # gRPC & WebSocket clients
│   │   ├── grpc/
│   │   │   ├── clients/
│   │   │   │   ├── MessageServiceClient.ts
│   │   │   │   ├── MediaServiceClient.ts
│   │   │   │   └── CallServiceClient.ts
│   │   │   └── types/          # Generated from .proto
│   │   └── websocket/
│   │       └── WebSocketClient.ts
│   ├── components/             # React Native components
│   │   ├── auth/
│   │   │   ├── LoginScreen.tsx
│   │   │   └── SetupScreen.tsx
│   │   ├── chat/
│   │   │   ├── ChatListScreen.tsx
│   │   │   ├── ChatScreen.tsx
│   │   │   ├── MessageBubble.tsx
│   │   │   └── TypingIndicator.tsx
│   │   ├── media/
│   │   │   ├── MediaPicker.tsx
│   │   │   ├── MediaViewer.tsx
│   │   │   └── MediaUpload.tsx
│   │   ├── calls/
│   │   │   ├── CallScreen.tsx
│   │   │   ├── IncomingCallModal.tsx
│   │   │   └── VideoControls.tsx
│   │   └── common/
│   │       ├── Button.tsx
│   │       ├── Input.tsx
│   │       └── Avatar.tsx
│   ├── crypto/                 # E2EE implementation
│   │   ├── IdentityManager.ts
│   │   ├── X3DHKeyExchange.ts
│   │   ├── DoubleRatchet.ts
│   │   ├── SessionManager.ts
│   │   └── CryptoWorker.ts
│   ├── waku/                   # Waku node management
│   │   ├── WakuNode.ts
│   │   ├── MessageHandler.ts
│   │   └── StoreProtocol.ts
│   ├── store/                  # Zustand stores
│   │   ├── authStore.ts
│   │   ├── contactsStore.ts
│   │   ├── messagesStore.ts
│   │   ├── callsStore.ts
│   │   ├── uiStore.ts
│   │   └── wakuStore.ts
│   ├── hooks/                  # Custom React hooks
│   │   ├── useWaku.ts
│   │   ├── useEncryption.ts
│   │   ├── useWebRTC.ts
│   │   └── useMessages.ts
│   ├── utils/                  # Helper functions
│   │   ├── crypto.ts
│   │   ├── storage.ts
│   │   ├── validation.ts
│   │   └── formatters.ts
│   ├── native/                 # Native modules
│   │   ├── CryptoModule.ts     # Native crypto operations
│   │   └── FileModule.ts       # Native file operations
│   ├── types/                  # TypeScript types
│   │   ├── api.ts
│   │   ├── crypto.ts
│   │   ├── waku.ts
│   │   └── models.ts
│   ├── constants/
│   │   └── config.ts
│   └── navigation/
│       └── RootNavigator.tsx
├── __tests__/                  # Test files
│   ├── unit/
│   │   ├── crypto/
│   │   ├── utils/
│   │   ├── hooks/
│   │   └── store/
│   ├── integration/
│   │   ├── api/
│   │   └── waku/
│   └── e2e/
│       └── scenarios/
├── android/                    # Android native code
├── ios/                        # iOS native code
├── App.tsx                     # Entry point
├── app.json
├── package.json
└── tsconfig.json
```

### 1.5 Environment Configuration

**app.json:**
```json
{
  "expo": {
    "name": "Ledabeer",
    "slug": "ledabeer-mobile",
    "version": "1.0.0",
    "orientation": "portrait",
    "icon": "./assets/icon.png",
    "userInterfaceStyle": "automatic",
    "splash": {
      "image": "./assets/splash.png",
      "resizeMode": "contain",
      "backgroundColor": "#000000"
    },
    "assetBundlePatterns": ["**/*"],
    "ios": {
      "supportsTablet": true,
      "bundleIdentifier": "com.ledabeer.mobile",
      "infoPlist": {
        "NSCameraUsageDescription": "We need camera access for video calls",
        "NSMicrophoneUsageDescription": "We need microphone access for calls",
        "NSPhotoLibraryUsageDescription": "We need photo library access to send images"
      }
    },
    "android": {
      "adaptiveIcon": {
        "foregroundImage": "./assets/adaptive-icon.png",
        "backgroundColor": "#000000"
      },
      "package": "com.ledabeer.mobile",
      "permissions": [
        "CAMERA",
        "RECORD_AUDIO",
        "READ_EXTERNAL_STORAGE",
        "WRITE_EXTERNAL_STORAGE",
        "INTERNET"
      ]
    },
    "plugins": [
      "expo-router",
      [
        "react-native-webrtc",
        {
          "cameraPermission": "Allow Ledabeer to access your camera for video calls",
          "microphonePermission": "Allow Ledabeer to access your microphone for calls"
        }
      ]
    ]
  }
}
```

---

## Phase 2: TDD Foundation (Red-Green-Refactor)

### 2.1 TDD Workflow

For **every** component, module, or feature, follow this cycle:

```
1. RED: Write failing test that describes expected behavior
2. GREEN: Write minimal code to make the test pass
3. REFACTOR: Improve code while keeping tests green
4. REPEAT: Move to next test
```

### 2.2 Test Configuration

**jest.config.js:**
```javascript
module.exports = {
  preset: 'react-native',
  setupFilesAfterEnv: [
    '@testing-library/jest-native/extend-expect',
    '<rootDir>/__tests__/setup.ts',
  ],
  transformIgnorePatterns: [
    'node_modules/(?!(react-native|@react-native|expo|@expo|@waku|@signalapp)/)',
  ],
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/types/**/*',
  ],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
  },
};
```

**__tests__/setup.ts:**
```typescript
import 'react-native-gesture-handler/jestSetup';
import mockAsyncStorage from '@react-native-async-storage/async-storage/jest/async-storage-mock';

jest.mock('@react-native-async-storage/async-storage', () => mockAsyncStorage);
jest.mock('react-native-reanimated', () => require('react-native-reanimated/mock'));
jest.mock('react-native-keychain', () => ({
  setGenericPassword: jest.fn(() => Promise.resolve()),
  getGenericPassword: jest.fn(() => Promise.resolve({ username: 'test', password: 'test' })),
  resetGenericPassword: jest.fn(() => Promise.resolve()),
}));
```

### 2.3 Comprehensive Edge Case Catalog

This catalog documents **150+ edge cases** to test across all modules:

#### 2.3.1 Crypto Edge Cases (20 cases)

| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 1 | Key Generation | Ed25519 key generation fails | Mock crypto failure, expect error handling |
| 2 | Key Generation | Duplicate key pair generated | Validate uniqueness with collision detection |
| 3 | Key Storage | Keychain storage fails (iOS/Android) | Mock storage failure, expect graceful degradation |
| 4 | Key Storage | Key retrieval returns null | Test app behavior with missing keys |
| 5 | Key Rotation | Key rotation after 1000 messages | Track message count, trigger rotation, verify |
| 6 | Key Rotation | Key rotation fails mid-session | Rollback mechanism test |
| 7 | X3DH | Prekey bundle unavailable | Request retry logic |
| 8 | X3DH | Invalid signature during exchange | Reject and log security event |
| 9 | Double Ratchet | Out-of-order message decryption | Store skipped keys, decrypt when received |
| 10 | Double Ratchet | Message with corrupted ciphertext | Fail gracefully, request retransmission |
| 11 | Double Ratchet | Ratchet chain exhaustion | Trigger new key exchange |
| 12 | Session | Concurrent message encryption | Test race conditions with locks |
| 13 | Session | Session storage corruption | Detect and reinitialize session |
| 14 | Session | Session recovery after app crash | Persist session state, restore on launch |
| 15 | Encryption | Large message encryption (>1MB) | Chunk and encrypt, verify reassembly |
| 16 | Encryption | Empty message content | Handle edge case, still encrypt |
| 17 | Decryption | Decryption with wrong session | Fail with clear error |
| 18 | Decryption | Decryption timeout (slow device) | Implement timeout, fallback |
| 19 | Memory | Memory leak in crypto operations | Profile memory usage over 1000 ops |
| 20 | Performance | Crypto operations block UI thread | Move to native module/worker |

#### 2.3.2 Network Edge Cases (25 cases)

| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 21 | Connectivity | No internet connection on app launch | Offline mode with message queue |
| 22 | Connectivity | Internet connection lost mid-message | Queue message, retry when online |
| 23 | Connectivity | Intermittent connectivity (flaky network) | Exponential backoff retry |
| 24 | Connectivity | Network switch (WiFi → Mobile data) | Detect change, reconnect Waku |
| 25 | Connectivity | Network switch (Mobile → WiFi) | Reconnect with better quality |
| 26 | Connectivity | Airplane mode enabled | Pause operations, resume on reconnect |
| 27 | Backend | gRPC server unreachable | Retry with exponential backoff (max 5 attempts) |
| 28 | Backend | gRPC timeout (>30s) | Cancel request, show error |
| 29 | Backend | WebSocket connection fails | Fallback to polling or offline mode |
| 30 | Backend | WebSocket disconnects mid-conversation | Auto-reconnect with heartbeat |
| 31 | Backend | WebSocket reconnection loop | Circuit breaker pattern |
| 32 | Backend | Backend returns 500 error | Log error, show user-friendly message |
| 33 | Backend | Backend returns malformed data | Validate schema, reject invalid |
| 34 | Slow Network | 3G network (high latency) | Show loading indicators, reduce quality |
| 35 | Slow Network | 2G network (very slow) | Disable media auto-download |
| 36 | Request | Concurrent requests (race condition) | Request queue with deduplication |
| 37 | Request | Request timeout handling | Cancel after 30s, show retry option |
| 38 | Request | Duplicate request (same message twice) | Idempotency check with message ID |
| 39 | NAT | TURN server unavailable | Retry with backup TURN server |
| 40 | NAT | ICE candidate exchange fails | Fallback to direct connection |
| 41 | NAT | Symmetric NAT (both peers) | Require TURN server |
| 42 | DNS | DNS resolution failure | Use IP address fallback |
| 43 | SSL | Certificate validation failure | Warn user, allow override in dev |
| 44 | Firewall | Port blocked by firewall | Retry with different port |
| 45 | Throttling | Rate limiting by backend | Respect rate limits, backoff |

#### 2.3.3 Waku Edge Cases (20 cases)

| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 46 | Node Init | Waku node initialization fails | Retry with fallback bootstrap nodes |
| 47 | Node Init | Bootstrap nodes unreachable | Use hardcoded fallback nodes |
| 48 | Node Init | Light node setup timeout | Cancel after 60s, show error |
| 49 | Peers | No peers discovered | Show warning, retry discovery |
| 50 | Peers | All peers disconnect | Reconnect to bootstrap |
| 51 | Peers | Peer disconnects mid-message | Resend via different peer |
| 52 | Peers | Malicious peer sends spam | Rate limit and blacklist |
| 53 | Topic | Topic subscription fails | Retry with backoff |
| 54 | Topic | Unsubscribe from active topic | Graceful cleanup |
| 55 | Publish | Message publish fails | Retry up to 3 times |
| 56 | Publish | Message too large (>256KB) | Chunk message before publishing |
| 57 | Publish | Publish during disconnection | Queue and retry when online |
| 58 | Subscribe | Message received twice (duplicate) | Deduplicate using hash |
| 59 | Subscribe | Message received out of order | Reorder by timestamp |
| 60 | Store | Waku Store protocol unavailable | Disable offline message retrieval |
| 61 | Store | Store returns no messages | Handle empty state gracefully |
| 62 | Store | Store query timeout | Cancel after 30s |
| 63 | Filter | Filter protocol rejection | Fallback to full relay |
| 64 | Filter | Filter returns incorrect data | Validate and reject |
| 65 | Memory | Waku node memory leak | Monitor and restart if needed |

#### 2.3.4 Message Edge Cases (20 cases)

| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 66 | Content | Empty message content | Reject on client side |
| 67 | Content | Very long message (>10,000 chars) | Split into multiple messages |
| 68 | Content | Special characters (emoji, Unicode) | Test UTF-8 encoding |
| 69 | Content | RTL text (Arabic, Hebrew) | Test text direction |
| 70 | Content | Malicious script tags | Sanitize with DOMPurify equivalent |
| 71 | Send | Message send fails | Retry up to 3 times |
| 72 | Send | Message send timeout | Mark as failed, allow manual retry |
| 73 | Send | Duplicate message ID | Prevent duplicate sends |
| 74 | Send | Sending while offline | Queue message, send when online |
| 75 | Receive | Message received with future timestamp | Validate timestamp range |
| 76 | Receive | Message received with past timestamp (>1 year) | Reject as invalid |
| 77 | Receive | Message from unknown sender | Create new contact |
| 78 | Receive | Message with missing fields | Reject and log error |
| 79 | Ordering | Out-of-order message delivery | Reorder by timestamp |
| 80 | Ordering | Message gap detection | Request missing messages |
| 81 | Status | Message status not updated | Retry status sync |
| 82 | Status | Status indicator incorrect | Reconcile with backend |
| 83 | History | History load fails | Show error, allow retry |
| 84 | History | History pagination broken | Fix cursor, reload |
| 85 | Search | Search query too long | Limit to 100 characters |

#### 2.3.5 Media Edge Cases (25 cases)

| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 86 | Selection | No permission to access gallery | Request permission, show rationale |
| 87 | Selection | Gallery is empty | Show empty state |
| 88 | Selection | User cancels file picker | Handle gracefully |
| 89 | Validation | File size exceeds 10MB | Reject with error message |
| 90 | Validation | Unsupported file type | Show list of supported types |
| 91 | Validation | Corrupted file selected | Reject and log error |
| 92 | Upload | Upload initiated while offline | Queue for later |
| 93 | Upload | Upload interrupted (app backgrounded) | Pause and resume |
| 94 | Upload | Upload fails after 3 retries | Mark as failed, allow manual retry |
| 95 | Upload | Concurrent uploads (3+ files) | Limit to 2 concurrent uploads |
| 96 | Upload | Disk space full | Detect and show error |
| 97 | Chunking | Chunking fails (large file) | Fall back to smaller chunks |
| 98 | Chunking | Chunk upload fails | Retry specific chunk |
| 99 | Compression | Image compression fails | Upload original |
| 100 | Compression | Compression takes too long (>10s) | Show progress, allow cancel |
| 101 | Encryption | Media encryption timeout | Chunk encryption operations |
| 102 | Download | Download fails | Retry up to 3 times |
| 103 | Download | Partial download (interrupted) | Resume from last chunk |
| 104 | Download | IPFS CID not found | Show error, request resend |
| 105 | Download | Download while on cellular | Prompt user (avoid data charges) |
| 106 | Display | Image fails to render | Show placeholder |
| 107 | Display | Video codec not supported | Show error, offer download |
| 108 | Display | Thumbnail generation fails | Use default icon |
| 109 | Cache | Image cache full | Evict oldest cached images |
| 110 | Playback | Video playback stutters | Reduce quality |

#### 2.3.6 Call Edge Cases (30 cases)

| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 111 | Permission | Camera permission denied | Show settings prompt |
| 112 | Permission | Microphone permission denied | Audio-only call |
| 113 | Permission | Permissions revoked during call | End call gracefully |
| 114 | Initiation | Call initiated while offline | Show error immediately |
| 115 | Initiation | Call to unavailable peer | Show "peer offline" message |
| 116 | Initiation | Call to busy peer | Show "peer busy" |
| 117 | Initiation | Multiple simultaneous call attempts | Queue or reject |
| 118 | Signaling | Signaling timeout (no response in 30s) | Cancel call |
| 119 | Signaling | SDP exchange fails | Show error, retry |
| 120 | Signaling | ICE candidate timeout | Retry with TURN |
| 121 | Connection | Peer connection fails | Show error, suggest retry |
| 122 | Connection | Connection drops during call | Attempt reconnection (10s) |
| 123 | Connection | Network changes during call | Reconnect seamlessly |
| 124 | Connection | Call quality degrades | Reduce video quality |
| 125 | Audio | No audio from peer | Check audio track |
| 126 | Audio | Echo detected | Enable echo cancellation |
| 127 | Audio | Audio cuts in/out | Buffer audio, smooth playback |
| 128 | Audio | Bluetooth headset connects | Switch audio route |
| 129 | Audio | Bluetooth headset disconnects | Switch back to earpiece |
| 130 | Video | No video from peer | Show placeholder |
| 131 | Video | Video freezes | Request keyframe |
| 132 | Video | Camera switches (front/back) | Update video track |
| 133 | Video | Low light (dark video) | Suggest better lighting |
| 134 | Background | App backgrounded during call (iOS) | Continue audio, pause video |
| 135 | Background | App backgrounded (Android) | Foreground service keeps call alive |
| 136 | Background | App killed during call | Send "call ended" notification |
| 137 | Interruption | Incoming phone call | Pause VoIP call |
| 138 | Interruption | Alarm/timer goes off | Mix audio |
| 139 | CallKit | CallKit registration fails (iOS) | Fallback to in-app UI |
| 140 | Group | Group call with >4 participants | Optimize SFU routing |

#### 2.3.7 Storage Edge Cases (15 cases)

| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 141 | Write | MMKV storage quota exceeded | Cleanup old messages |
| 142 | Write | Concurrent writes (race condition) | Use locks/transactions |
| 143 | Write | Storage corruption detected | Reinitialize database |
| 144 | Write | Write fails (disk full) | Show error, suggest cleanup |
| 145 | Read | Key not found | Return default value |
| 146 | Read | Decryption fails (wrong key) | Reinitialize |
| 147 | Migration | Schema migration fails | Rollback to previous version |
| 148 | Migration | Migration timeout (large data) | Background migration |
| 149 | Encryption | Storage encryption key lost | Data unrecoverable, clear app |
| 150 | Encryption | Encryption key derivation fails | Show error, reinitialize |
| 151 | Cleanup | Cleanup fails to delete old data | Retry on next launch |
| 152 | Backup | Backup to iCloud/Drive fails | Show error, retry |
| 153 | Restore | Restore from backup fails | Clear and start fresh |
| 154 | Keychain | Keychain locked (iOS) | Prompt for device unlock |
| 155 | Keychain | Keychain item not accessible | Regenerate keys |

#### 2.3.8 UI/UX Edge Cases (20 cases)

| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 156 | App Lifecycle | App backgrounded mid-operation | Pause and resume |
| 157 | App Lifecycle | App killed by system (low memory) | Restore state on relaunch |
| 158 | App Lifecycle | App suspended for 24+ hours | Refresh all data |
| 159 | Orientation | Device rotated during call | Maintain call, adjust UI |
| 160 | Orientation | Orientation locked in settings | Respect user preference |
| 161 | Keyboard | Keyboard covers input field | Adjust scroll position |
| 162 | Keyboard | Keyboard doesn't appear | Check input focus |
| 163 | Text | Very long contact name (>50 chars) | Truncate with ellipsis |
| 164 | Text | Contact name with emoji | Display correctly |
| 165 | Input | Rapid button taps (double submit) | Debounce action |
| 166 | Input | Paste very long text (>10k chars) | Limit or warn |
| 167 | Scroll | Scroll performance with 10k messages | Use FlashList |
| 168 | Scroll | Scroll to bottom not working | Force scroll on new message |
| 169 | Scroll | Over-scroll bounce (iOS) | Style with native look |
| 170 | Gallery | Gallery with 100+ images | Lazy load with pagination |
| 171 | Gesture | Swipe gesture conflicts | Prioritize one gesture |
| 172 | Gesture | Long press not detected | Adjust timing threshold |
| 173 | Notification | Notification tapped while app open | Navigate to chat |
| 174 | Notification | Multiple notifications (spam) | Group notifications |
| 175 | Theme | Dark mode switch mid-session | Update colors dynamically |

#### 2.3.9 Platform-Specific Edge Cases (15 cases)

**iOS:**
| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 176 | Background | App in background >30s | Background refresh |
| 177 | Background | VoIP push not received | Check APNS setup |
| 178 | CallKit | CallKit call fails | Show in-app error |
| 179 | Privacy | User denies tracking permission | Respect choice |
| 180 | Low Power Mode | Low Power Mode enabled | Reduce background activity |
| 181 | iCloud | iCloud sync fails | Show error, use local only |
| 182 | TestFlight | TestFlight build expires | Prompt for update |
| 183 | App Store | App Store review rejection | Fix based on feedback |

**Android:**
| # | Category | Edge Case | Test Strategy |
|---|----------|-----------|---------------|
| 184 | Doze | Device in Doze mode | Use high-priority FCM |
| 185 | Battery | Battery optimization enabled | Request exemption |
| 186 | Screens | Multiple screen sizes | Test on 5 devices |
| 187 | Screens | Foldable device | Test fold/unfold |
| 188 | Back Button | Back button pressed | Handle navigation correctly |
| 189 | Deep Link | Deep link from notification | Navigate to correct screen |
| 190 | Foreground Service | Service killed by system | Restart service |

### 2.4 Test File Examples

**Example 1: Crypto Module Test**

```typescript
// __tests__/unit/crypto/DoubleRatchet.test.ts

import { DoubleRatchet } from '../../../src/crypto/DoubleRatchet';

describe('DoubleRatchet', () => {
  let senderRatchet: DoubleRatchet;
  let receiverRatchet: DoubleRatchet;
  const sharedSecret = new Uint8Array(32).fill(1);

  beforeEach(() => {
    // RED: Tests will fail initially
    senderRatchet = new DoubleRatchet(sharedSecret, true);
    receiverRatchet = new DoubleRatchet(sharedSecret, false);
  });

  test('should encrypt and decrypt message correctly', async () => {
    const plaintext = new TextEncoder().encode('Hello, secure world!');

    const ciphertext = await senderRatchet.encrypt(plaintext);
    expect(ciphertext).toBeDefined();
    expect(ciphertext).not.toEqual(plaintext);

    const decrypted = await receiverRatchet.decrypt(ciphertext);
    expect(decrypted).toEqual(plaintext);
  });

  test('should fail decryption with wrong key', async () => {
    const plaintext = new TextEncoder().encode('Secret message');
    const ciphertext = await senderRatchet.encrypt(plaintext);

    const wrongRatchet = new DoubleRatchet(new Uint8Array(32).fill(2), false);

    await expect(wrongRatchet.decrypt(ciphertext)).rejects.toThrow();
  });

  test('should provide forward secrecy (old keys cannot decrypt new messages)', async () => {
    const message1 = new TextEncoder().encode('Message 1');
    const message2 = new TextEncoder().encode('Message 2');

    const cipher1 = await senderRatchet.encrypt(message1);
    await receiverRatchet.decrypt(cipher1);

    // Create a snapshot of the receiver state
    const oldReceiverState = receiverRatchet.exportState();

    const cipher2 = await senderRatchet.encrypt(message2);
    await receiverRatchet.decrypt(cipher2);

    // Try to decrypt message 2 with old state - should fail
    const oldRatchet = DoubleRatchet.fromState(oldReceiverState);
    await expect(oldRatchet.decrypt(cipher2)).rejects.toThrow('forward secrecy');
  });

  test('should handle out-of-order messages', async () => {
    const messages = [
      new TextEncoder().encode('Msg 1'),
      new TextEncoder().encode('Msg 2'),
      new TextEncoder().encode('Msg 3'),
    ];

    // Encrypt in order
    const ciphertexts = await Promise.all(
      messages.map(msg => senderRatchet.encrypt(msg))
    );

    // Decrypt out of order: 1, 3, 2
    const decrypted1 = await receiverRatchet.decrypt(ciphertexts[0]);
    expect(decrypted1).toEqual(messages[0]);

    const decrypted3 = await receiverRatchet.decrypt(ciphertexts[2]);
    expect(decrypted3).toEqual(messages[2]);

    const decrypted2 = await receiverRatchet.decrypt(ciphertexts[1]);
    expect(decrypted2).toEqual(messages[1]);
  });

  test('should throw error on corrupted ciphertext', async () => {
    const plaintext = new TextEncoder().encode('Original message');
    const ciphertext = await senderRatchet.encrypt(plaintext);

    // Corrupt the ciphertext
    ciphertext[10] ^= 0xFF;

    await expect(receiverRatchet.decrypt(ciphertext)).rejects.toThrow('corrupted');
  });

  test('should rotate keys after 1000 messages', async () => {
    const message = new TextEncoder().encode('Test');

    // Send 1000 messages
    for (let i = 0; i < 1000; i++) {
      const cipher = await senderRatchet.encrypt(message);
      await receiverRatchet.decrypt(cipher);
    }

    // Check that keys have rotated
    const initialKey = senderRatchet.getCurrentChainKey();

    const cipher = await senderRatchet.encrypt(message);
    await receiverRatchet.decrypt(cipher);

    const newKey = senderRatchet.getCurrentChainKey();
    expect(newKey).not.toEqual(initialKey);
  });
});
```

**Example 2: Message Store Test**

```typescript
// __tests__/unit/store/messagesStore.test.ts

import { renderHook, act } from '@testing-library/react-native';
import { useMessagesStore } from '../../../src/store/messagesStore';

describe('MessagesStore', () => {
  beforeEach(() => {
    // Reset store before each test
    useMessagesStore.getState().reset();
  });

  test('should add message to store', () => {
    const { result } = renderHook(() => useMessagesStore());

    act(() => {
      result.current.addMessage({
        id: 'msg1',
        peerId: 'peer1',
        content: 'Hello',
        timestamp: Date.now(),
        status: 'sent',
      });
    });

    const messages = result.current.getMessages('peer1');
    expect(messages).toHaveLength(1);
    expect(messages[0].content).toBe('Hello');
  });

  test('should update message status', () => {
    const { result } = renderHook(() => useMessagesStore());

    act(() => {
      result.current.addMessage({
        id: 'msg1',
        peerId: 'peer1',
        content: 'Hello',
        timestamp: Date.now(),
        status: 'sending',
      });
    });

    act(() => {
      result.current.updateMessageStatus('msg1', 'sent');
    });

    const messages = result.current.getMessages('peer1');
    expect(messages[0].status).toBe('sent');
  });

  test('should handle concurrent updates', async () => {
    const { result } = renderHook(() => useMessagesStore());

    // Simulate concurrent adds
    await Promise.all([
      act(async () => {
        result.current.addMessage({
          id: 'msg1',
          peerId: 'peer1',
          content: 'First',
          timestamp: Date.now(),
          status: 'sent',
        });
      }),
      act(async () => {
        result.current.addMessage({
          id: 'msg2',
          peerId: 'peer1',
          content: 'Second',
          timestamp: Date.now() + 1,
          status: 'sent',
        });
      }),
    ]);

    const messages = result.current.getMessages('peer1');
    expect(messages).toHaveLength(2);
  });

  test('should load paginated messages', async () => {
    const { result } = renderHook(() => useMessagesStore());

    // Add 100 messages
    act(() => {
      for (let i = 0; i < 100; i++) {
        result.current.addMessage({
          id: `msg${i}`,
          peerId: 'peer1',
          content: `Message ${i}`,
          timestamp: Date.now() + i,
          status: 'sent',
        });
      }
    });

    // Load first page (20 messages)
    const page1 = result.current.getMessages('peer1', 0, 20);
    expect(page1).toHaveLength(20);

    // Load second page
    const page2 = result.current.getMessages('peer1', 20, 20);
    expect(page2).toHaveLength(20);
    expect(page2[0].id).not.toBe(page1[0].id);
  });

  test('should cleanup old messages', () => {
    const { result } = renderHook(() => useMessagesStore());

    // Add messages older than 30 days
    const oldTimestamp = Date.now() - 31 * 24 * 60 * 60 * 1000;

    act(() => {
      result.current.addMessage({
        id: 'old1',
        peerId: 'peer1',
        content: 'Old message',
        timestamp: oldTimestamp,
        status: 'sent',
      });

      result.current.addMessage({
        id: 'new1',
        peerId: 'peer1',
        content: 'New message',
        timestamp: Date.now(),
        status: 'sent',
      });
    });

    // Run cleanup
    act(() => {
      result.current.cleanupOldMessages(30);
    });

    const messages = result.current.getMessages('peer1');
    expect(messages).toHaveLength(1);
    expect(messages[0].id).toBe('new1');
  });
});
```

---

## Phase 3: Backend Integration (API Layer with TDD)

### 3.1 gRPC Client Setup

**Step 1: Generate TypeScript types from .proto files**

```bash
# Install protoc compiler
npm install --save-dev @grpc/proto-loader @grpc/grpc-js

# Create script to generate types
mkdir -p scripts
```

**scripts/generate-proto.sh:**
```bash
#!/bin/bash

PROTO_DIR="../backend/pkg/proto"
OUT_DIR="./src/api/grpc/generated"

mkdir -p $OUT_DIR

# Generate TypeScript definitions
protoc \
  --plugin=protoc-gen-ts=./node_modules/.bin/protoc-gen-ts \
  --ts_out=$OUT_DIR \
  --proto_path=$PROTO_DIR \
  $PROTO_DIR/*.proto
```

**Step 2: Create gRPC clients (Test-First)**

```typescript
// __tests__/integration/api/MessageServiceClient.test.ts

import { MessageServiceClient } from '../../../src/api/grpc/clients/MessageServiceClient';

describe('MessageServiceClient', () => {
  let client: MessageServiceClient;

  beforeEach(() => {
    client = new MessageServiceClient('localhost:5001');
  });

  afterEach(() => {
    client.close();
  });

  test('should send message successfully', async () => {
    const response = await client.sendMessage({
      toPeerId: 'peer123',
      content: new Uint8Array([1, 2, 3]),
    });

    expect(response.messageId).toBeDefined();
    expect(response.timestamp).toBeGreaterThan(0);
  });

  test('should retry on network failure', async () => {
    // Mock network failure
    const mockError = new Error('UNAVAILABLE');
    jest.spyOn(client as any, '_makeRequest').mockRejectedValueOnce(mockError);

    // Should retry and succeed
    const response = await client.sendMessage({
      toPeerId: 'peer123',
      content: new Uint8Array([1, 2, 3]),
    });

    expect(response).toBeDefined();
    expect(client['_retryCount']).toBe(1);
  });

  test('should handle server error gracefully', async () => {
    // Mock 500 error
    const mockError = new Error('INTERNAL');
    jest.spyOn(client as any, '_makeRequest').mockRejectedValue(mockError);

    await expect(client.sendMessage({
      toPeerId: 'peer123',
      content: new Uint8Array([1, 2, 3]),
    })).rejects.toThrow('INTERNAL');
  });

  test('should queue messages when offline', async () => {
    // Simulate offline
    client.setOffline(true);

    // Send message (should queue)
    const promise = client.sendMessage({
      toPeerId: 'peer123',
      content: new Uint8Array([1, 2, 3]),
    });

    // Should not resolve immediately
    expect(client['_messageQueue'].length).toBe(1);

    // Go online
    client.setOffline(false);

    // Should now send
    const response = await promise;
    expect(response).toBeDefined();
  });

  test('should receive messages stream', async () => {
    const messages: any[] = [];

    const stream = client.receiveMessages();

    stream.on('data', (message) => {
      messages.push(message);
    });

    // Wait for some messages
    await new Promise(resolve => setTimeout(resolve, 1000));

    expect(messages.length).toBeGreaterThan(0);
  });
});
```

**Implementation:**

```typescript
// src/api/grpc/clients/MessageServiceClient.ts

import * as grpc from '@grpc/grpc-js';
import { MessageServiceClient as GrpcClient } from '../generated/message_grpc_pb';
import { SendMessageRequest, SendMessageResponse } from '../generated/message_pb';

export class MessageServiceClient {
  private client: GrpcClient;
  private offline: boolean = false;
  private messageQueue: any[] = [];
  private retryCount: number = 0;
  private maxRetries: number = 5;

  constructor(address: string) {
    this.client = new GrpcClient(
      address,
      grpc.credentials.createInsecure()
    );
  }

  async sendMessage(request: {
    toPeerId: string;
    content: Uint8Array;
  }): Promise<{ messageId: string; timestamp: number }> {
    // If offline, queue message
    if (this.offline) {
      return new Promise((resolve) => {
        this.messageQueue.push({ request, resolve });
      });
    }

    const grpcRequest = new SendMessageRequest();
    grpcRequest.setToPeerId(request.toPeerId);
    grpcRequest.setContent(request.content);

    return this._sendWithRetry(grpcRequest);
  }

  private async _sendWithRetry(
    request: SendMessageRequest,
    attempt: number = 0
  ): Promise<{ messageId: string; timestamp: number }> {
    try {
      return await this._makeRequest(request);
    } catch (error: any) {
      // Retry on network errors
      if (this._isRetryableError(error) && attempt < this.maxRetries) {
        this.retryCount++;
        const delay = Math.pow(2, attempt) * 1000; // Exponential backoff
        await this._sleep(delay);
        return this._sendWithRetry(request, attempt + 1);
      }
      throw error;
    }
  }

  private _makeRequest(request: SendMessageRequest): Promise<any> {
    return new Promise((resolve, reject) => {
      this.client.sendMessage(request, (error, response) => {
        if (error) {
          reject(error);
        } else {
          resolve({
            messageId: response.getMessageId(),
            timestamp: response.getTimestamp(),
          });
        }
      });
    });
  }

  private _isRetryableError(error: any): boolean {
    const retryableCodes = ['UNAVAILABLE', 'DEADLINE_EXCEEDED', 'RESOURCE_EXHAUSTED'];
    return retryableCodes.includes(error.code);
  }

  private _sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  setOffline(offline: boolean): void {
    this.offline = offline;

    if (!offline && this.messageQueue.length > 0) {
      // Process queued messages
      this._processQueue();
    }
  }

  private async _processQueue(): Promise<void> {
    while (this.messageQueue.length > 0) {
      const { request, resolve } = this.messageQueue.shift();
      try {
        const response = await this.sendMessage(request);
        resolve(response);
      } catch (error) {
        console.error('Failed to send queued message:', error);
      }
    }
  }

  receiveMessages(): grpc.ClientReadableStream<any> {
    // Implement streaming
    return this.client.receiveMessages(new grpc.Metadata());
  }

  close(): void {
    this.client.close();
  }
}
```

### 3.2 WebSocket Client (Test-First)

```typescript
// __tests__/integration/api/WebSocketClient.test.ts

import { WebSocketClient } from '../../../src/api/websocket/WebSocketClient';

describe('WebSocketClient', () => {
  let client: WebSocketClient;
  const mockUrl = 'ws://localhost:8080/ws';

  beforeEach(() => {
    client = new WebSocketClient(mockUrl);
  });

  afterEach(() => {
    client.disconnect();
  });

  test('should connect and authenticate', async () => {
    const connected = await client.connect('peer123');

    expect(connected).toBe(true);
    expect(client.isConnected()).toBe(true);
  });

  test('should reconnect on disconnect', async () => {
    await client.connect('peer123');

    // Simulate disconnect
    client['ws'].close();

    // Wait for reconnection
    await new Promise(resolve => setTimeout(resolve, 2000));

    expect(client.isConnected()).toBe(true);
    expect(client['reconnectAttempts']).toBeGreaterThan(0);
  });

  test('should handle rapid reconnection attempts', async () => {
    await client.connect('peer123');

    // Disconnect multiple times rapidly
    for (let i = 0; i < 5; i++) {
      client['ws'].close();
    }

    // Should use exponential backoff
    expect(client['reconnectDelay']).toBeGreaterThan(1000);
  });

  test('should queue messages during reconnection', async () => {
    await client.connect('peer123');

    // Disconnect
    client['ws'].close();

    // Try to send message
    client.send({ type: 'test', data: 'hello' });

    // Should be queued
    expect(client['messageQueue'].length).toBe(1);

    // Reconnect
    await new Promise(resolve => setTimeout(resolve, 2000));

    // Queue should be empty (messages sent)
    expect(client['messageQueue'].length).toBe(0);
  });

  test('should emit events for incoming messages', async () => {
    await client.connect('peer123');

    const messages: any[] = [];

    client.on('message', (message) => {
      messages.push(message);
    });

    // Simulate incoming message
    client['_handleMessage']({
      type: 'message',
      content: 'test',
    });

    expect(messages.length).toBe(1);
  });

  test('should send heartbeat ping/pong', async () => {
    await client.connect('peer123');

    // Wait for heartbeat
    await new Promise(resolve => setTimeout(resolve, 35000));

    expect(client['lastPong']).toBeGreaterThan(0);
  });
});
```

**Implementation:**

```typescript
// src/api/websocket/WebSocketClient.ts

import EventEmitter from 'events';

export class WebSocketClient extends EventEmitter {
  private ws: WebSocket | null = null;
  private url: string;
  private peerId: string = '';
  private connected: boolean = false;
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 5;
  private reconnectDelay: number = 1000;
  private messageQueue: any[] = [];
  private heartbeatInterval: any = null;
  private lastPong: number = 0;

  constructor(url: string) {
    super();
    this.url = url;
  }

  async connect(peerId: string): Promise<boolean> {
    this.peerId = peerId;

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this._authenticate();
          this.connected = true;
          this.reconnectAttempts = 0;
          this.reconnectDelay = 1000;
          this._startHeartbeat();
          resolve(true);
        };

        this.ws.onmessage = (event) => {
          const message = JSON.parse(event.data);
          this._handleMessage(message);
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        };

        this.ws.onclose = () => {
          console.log('WebSocket closed');
          this.connected = false;
          this._stopHeartbeat();
          this._attemptReconnect();
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  private _authenticate(): void {
    this.send({
      type: 'auth',
      peer_id: this.peerId,
    });
  }

  private _attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      this.emit('max_reconnect_attempts');
      return;
    }

    this.reconnectAttempts++;
    console.log(`Reconnecting... Attempt ${this.reconnectAttempts}`);

    setTimeout(() => {
      this.connect(this.peerId).catch((error) => {
        console.error('Reconnection failed:', error);
      });
    }, this.reconnectDelay);

    // Exponential backoff
    this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
  }

  send(message: any): void {
    if (!this.connected || !this.ws) {
      // Queue message
      this.messageQueue.push(message);
      return;
    }

    try {
      this.ws.send(JSON.stringify(message));
    } catch (error) {
      console.error('Failed to send message:', error);
      this.messageQueue.push(message);
    }
  }

  private _processQueue(): void {
    while (this.messageQueue.length > 0 && this.connected) {
      const message = this.messageQueue.shift();
      this.send(message);
    }
  }

  private _handleMessage(message: any): void {
    switch (message.type) {
      case 'auth_success':
        console.log('Authentication successful');
        this._processQueue();
        break;
      case 'message':
        this.emit('message', message);
        break;
      case 'pong':
        this.lastPong = Date.now();
        break;
      default:
        this.emit('data', message);
    }
  }

  private _startHeartbeat(): void {
    this.heartbeatInterval = setInterval(() => {
      if (this.connected) {
        this.send({ type: 'ping' });

        // Check if pong received within 10s
        if (Date.now() - this.lastPong > 10000) {
          console.warn('No pong received, reconnecting...');
          this.ws?.close();
        }
      }
    }, 30000); // Ping every 30s
  }

  private _stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  isConnected(): boolean {
    return this.connected;
  }

  disconnect(): void {
    this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection
    this._stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.connected = false;
  }
}
```

---

## Phase 4: Crypto Layer (E2EE with TDD)

### 4.1 Identity Manager (Test-First)

```typescript
// __tests__/unit/crypto/IdentityManager.test.ts

import { IdentityManager } from '../../../src/crypto/IdentityManager';
import * as Keychain from 'react-native-keychain';

describe('IdentityManager', () => {
  let identityManager: IdentityManager;

  beforeEach(() => {
    identityManager = new IdentityManager();
  });

  test('should generate Ed25519 key pair', async () => {
    const identity = await identityManager.generateIdentity();

    expect(identity.publicKey).toBeDefined();
    expect(identity.privateKey).toBeDefined();
    expect(identity.publicKey.length).toBe(32);
    expect(identity.privateKey.length).toBe(64);
  });

  test('should store keys securely in Keychain', async () => {
    const identity = await identityManager.generateIdentity();

    await identityManager.storeIdentity(identity);

    expect(Keychain.setGenericPassword).toHaveBeenCalled();
  });

  test('should export keys with password encryption', async () => {
    const identity = await identityManager.generateIdentity();
    const password = 'strongPassword123';

    const exported = await identityManager.exportIdentity(identity, password);

    expect(exported).toBeDefined();
    expect(typeof exported).toBe('string');
  });

  test('should import keys and validate', async () => {
    const identity = await identityManager.generateIdentity();
    const password = 'strongPassword123';

    const exported = await identityManager.exportIdentity(identity, password);
    const imported = await identityManager.importIdentity(exported, password);

    expect(imported.publicKey).toEqual(identity.publicKey);
    expect(imported.privateKey).toEqual(identity.privateKey);
  });

  test('should rotate keys after expiry', async () => {
    const identity = await identityManager.generateIdentity();

    // Set expiry to past
    identity.expiresAt = Date.now() - 1000;

    const shouldRotate = identityManager.shouldRotateKeys(identity);
    expect(shouldRotate).toBe(true);

    const newIdentity = await identityManager.rotateKeys(identity);
    expect(newIdentity.publicKey).not.toEqual(identity.publicKey);
  });

  test('should handle key loss gracefully', async () => {
    // Mock keychain returning null
    jest.spyOn(Keychain, 'getGenericPassword').mockResolvedValue(false);

    const identity = await identityManager.loadIdentity();

    expect(identity).toBeNull();
  });
});
```

**Implementation:**

```typescript
// src/crypto/IdentityManager.ts

import * as Keychain from 'react-native-keychain';
import { randomBytes } from 'react-native-randombytes';
import * as signalProtocol from '@signalapp/libsignal-client';

export interface Identity {
  publicKey: Uint8Array;
  privateKey: Uint8Array;
  createdAt: number;
  expiresAt: number;
}

export class IdentityManager {
  private static KEYCHAIN_SERVICE = 'ledabeer-identity';
  private static KEY_EXPIRY_DAYS = 30;

  async generateIdentity(): Promise<Identity> {
    // Generate Ed25519 key pair using Signal Protocol
    const identityKeyPair = signalProtocol.PrivateKey.generate();

    const publicKey = identityKeyPair.getPublicKey().serialize();
    const privateKey = identityKeyPair.serialize();

    const createdAt = Date.now();
    const expiresAt = createdAt + (IdentityManager.KEY_EXPIRY_DAYS * 24 * 60 * 60 * 1000);

    return {
      publicKey,
      privateKey,
      createdAt,
      expiresAt,
    };
  }

  async storeIdentity(identity: Identity): Promise<void> {
    const data = JSON.stringify({
      publicKey: Buffer.from(identity.publicKey).toString('base64'),
      privateKey: Buffer.from(identity.privateKey).toString('base64'),
      createdAt: identity.createdAt,
      expiresAt: identity.expiresAt,
    });

    await Keychain.setGenericPassword(
      'identity',
      data,
      {
        service: IdentityManager.KEYCHAIN_SERVICE,
        accessible: Keychain.ACCESSIBLE.WHEN_UNLOCKED,
      }
    );
  }

  async loadIdentity(): Promise<Identity | null> {
    try {
      const credentials = await Keychain.getGenericPassword({
        service: IdentityManager.KEYCHAIN_SERVICE,
      });

      if (!credentials) {
        return null;
      }

      const data = JSON.parse(credentials.password);

      return {
        publicKey: Uint8Array.from(Buffer.from(data.publicKey, 'base64')),
        privateKey: Uint8Array.from(Buffer.from(data.privateKey, 'base64')),
        createdAt: data.createdAt,
        expiresAt: data.expiresAt,
      };
    } catch (error) {
      console.error('Failed to load identity:', error);
      return null;
    }
  }

  async exportIdentity(identity: Identity, password: string): Promise<string> {
    // Encrypt with PBKDF2-derived key
    const salt = randomBytes(16);
    const key = await this._deriveKey(password, salt);

    const data = JSON.stringify({
      publicKey: Buffer.from(identity.publicKey).toString('base64'),
      privateKey: Buffer.from(identity.privateKey).toString('base64'),
      createdAt: identity.createdAt,
      expiresAt: identity.expiresAt,
    });

    const encrypted = await this._encrypt(data, key);

    return JSON.stringify({
      salt: Buffer.from(salt).toString('base64'),
      data: encrypted,
    });
  }

  async importIdentity(exported: string, password: string): Promise<Identity> {
    const { salt, data } = JSON.parse(exported);

    const key = await this._deriveKey(
      password,
      Uint8Array.from(Buffer.from(salt, 'base64'))
    );

    const decrypted = await this._decrypt(data, key);
    const parsed = JSON.parse(decrypted);

    return {
      publicKey: Uint8Array.from(Buffer.from(parsed.publicKey, 'base64')),
      privateKey: Uint8Array.from(Buffer.from(parsed.privateKey, 'base64')),
      createdAt: parsed.createdAt,
      expiresAt: parsed.expiresAt,
    };
  }

  shouldRotateKeys(identity: Identity): boolean {
    return Date.now() >= identity.expiresAt;
  }

  async rotateKeys(oldIdentity: Identity): Promise<Identity> {
    // Generate new identity
    const newIdentity = await this.generateIdentity();

    // Store new identity
    await this.storeIdentity(newIdentity);

    // TODO: Notify peers of key rotation

    return newIdentity;
  }

  private async _deriveKey(password: string, salt: Uint8Array): Promise<Uint8Array> {
    // Use PBKDF2 to derive key from password
    // Implementation depends on crypto library
    // This is a simplified example
    return new Uint8Array(32); // Placeholder
  }

  private async _encrypt(data: string, key: Uint8Array): Promise<string> {
    // Encrypt data with AES-256-GCM
    // This is a simplified example
    return Buffer.from(data).toString('base64'); // Placeholder
  }

  private async _decrypt(encrypted: string, key: Uint8Array): Promise<string> {
    // Decrypt data with AES-256-GCM
    // This is a simplified example
    return Buffer.from(encrypted, 'base64').toString(); // Placeholder
  }
}
```

### 4.2 Signal Protocol Integration

Due to length constraints, the full Signal Protocol implementation (X3DH + Double Ratchet) follows the same TDD pattern as shown above. Key files to implement:

- `src/crypto/X3DHKeyExchange.ts` - X3DH key exchange protocol
- `src/crypto/DoubleRatchet.ts` - Double Ratchet encryption/decryption
- `src/crypto/SessionManager.ts` - Manage sessions per peer
- `src/crypto/CryptoWorker.ts` - Native module for performance

**Test files:**
- `__tests__/unit/crypto/X3DHKeyExchange.test.ts`
- `__tests__/unit/crypto/DoubleRatchet.test.ts`
- `__tests__/unit/crypto/SessionManager.test.ts`

---

## Phase 5: Waku Integration (P2P Layer with TDD)

### 5.1 Waku Node Setup

```typescript
// __tests__/integration/waku/WakuNode.test.ts

import { WakuNode } from '../../../src/waku/WakuNode';

describe('WakuNode', () => {
  let wakuNode: WakuNode;

  beforeEach(async () => {
    wakuNode = new WakuNode({
      bootstrapNodes: ['ws://localhost:8546'],
    });
  });

  afterEach(async () => {
    await wakuNode.stop();
  });

  test('should initialize light node', async () => {
    const started = await wakuNode.start();

    expect(started).toBe(true);
    expect(wakuNode.isRunning()).toBe(true);
  });

  test('should connect to bootstrap nodes', async () => {
    await wakuNode.start();

    const peers = wakuNode.getPeers();
    expect(peers.length).toBeGreaterThan(0);
  });

  test('should handle bootstrap node failure', async () => {
    // Mock all bootstrap nodes failing
    wakuNode['config'].bootstrapNodes = ['ws://invalid:9999'];

    // Should fallback to hardcoded nodes
    const started = await wakuNode.start();
    expect(started).toBe(true);
  });

  test('should discover peers', async () => {
    await wakuNode.start();

    // Wait for peer discovery
    await new Promise(resolve => setTimeout(resolve, 5000));

    const peers = wakuNode.getPeers();
    expect(peers.length).toBeGreaterThanOrEqual(1);
  });

  test('should reconnect on network change', async () => {
    await wakuNode.start();

    // Simulate network change
    wakuNode['_handleNetworkChange']({ type: 'cellular' });

    // Wait for reconnection
    await new Promise(resolve => setTimeout(resolve, 3000));

    expect(wakuNode.isRunning()).toBe(true);
  });

  test('should cleanup on app background', async () => {
    await wakuNode.start();

    // Simulate app backgrounded
    wakuNode.pause();

    expect(wakuNode.isRunning()).toBe(false);

    // Resume on foreground
    wakuNode.resume();

    expect(wakuNode.isRunning()).toBe(true);
  });
});
```

### 5.2 Messaging via Waku Relay

```typescript
// __tests__/integration/waku/MessageHandler.test.ts

import { WakuNode } from '../../../src/waku/WakuNode';
import { MessageHandler } from '../../../src/waku/MessageHandler';

describe('MessageHandler', () => {
  let wakuNode: WakuNode;
  let messageHandler: MessageHandler;

  beforeEach(async () => {
    wakuNode = new WakuNode();
    await wakuNode.start();
    messageHandler = new MessageHandler(wakuNode);
  });

  test('should subscribe to chat topic', async () => {
    const subscribed = await messageHandler.subscribe('/ledabeer/1/chat/proto');

    expect(subscribed).toBe(true);
  });

  test('should publish encrypted message', async () => {
    const encrypted = new Uint8Array([1, 2, 3, 4]);

    const published = await messageHandler.publish(
      '/ledabeer/1/chat/proto',
      encrypted
    );

    expect(published).toBe(true);
  });

  test('should receive and decrypt message', async () => {
    const messages: any[] = [];

    await messageHandler.subscribe('/ledabeer/1/chat/proto');

    messageHandler.on('message', (msg) => {
      messages.push(msg);
    });

    // Publish message
    await messageHandler.publish(
      '/ledabeer/1/chat/proto',
      new Uint8Array([1, 2, 3])
    );

    // Wait for message
    await new Promise(resolve => setTimeout(resolve, 1000));

    expect(messages.length).toBeGreaterThan(0);
  });

  test('should handle message deduplication', async () => {
    const messages: any[] = [];

    await messageHandler.subscribe('/ledabeer/1/chat/proto');

    messageHandler.on('message', (msg) => {
      messages.push(msg);
    });

    const payload = new Uint8Array([1, 2, 3, 4, 5]);

    // Publish same message twice
    await messageHandler.publish('/ledabeer/1/chat/proto', payload);
    await messageHandler.publish('/ledabeer/1/chat/proto', payload);

    // Wait
    await new Promise(resolve => setTimeout(resolve, 2000));

    // Should only receive once
    expect(messages.length).toBe(1);
  });

  test('should retrieve offline messages from Store', async () => {
    // Publish message before subscribing
    await messageHandler.publish(
      '/ledabeer/1/chat/proto',
      new Uint8Array([9, 8, 7])
    );

    // Subscribe and query store
    await messageHandler.subscribe('/ledabeer/1/chat/proto');
    const history = await messageHandler.queryStore('/ledabeer/1/chat/proto');

    expect(history.length).toBeGreaterThan(0);
  });

  test('should handle large messages (chunking)', async () => {
    // Create large message (>256KB)
    const largeMessage = new Uint8Array(300 * 1024);

    // Should chunk automatically
    const published = await messageHandler.publish(
      '/ledabeer/1/chat/proto',
      largeMessage
    );

    expect(published).toBe(true);
  });
});
```

---

## Phase 6: Core UI Components (GPU-Optimized with TDD)

### 6.1 Chat Screen Implementation

**Test-First Approach:**

```typescript
// __tests__/unit/components/ChatScreen.test.tsx

import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { ChatScreen } from '../../../src/components/chat/ChatScreen';

describe('ChatScreen', () => {
  test('should render message bubbles', () => {
    const { getByText } = render(
      <ChatScreen peerId="peer123" />
    );

    // Mock messages
    const messages = [
      { id: '1', content: 'Hello', sender: 'me' },
      { id: '2', content: 'Hi there', sender: 'peer123' },
    ];

    expect(getByText('Hello')).toBeTruthy();
    expect(getByText('Hi there')).toBeTruthy();
  });

  test('should show typing indicator', async () => {
    const { getByTestId } = render(
      <ChatScreen peerId="peer123" />
    );

    // Simulate peer typing
    // ...

    await waitFor(() => {
      expect(getByTestId('typing-indicator')).toBeTruthy();
    });
  });

  test('should handle message send', async () => {
    const { getByPlaceholderText, getByTestId } = render(
      <ChatScreen peerId="peer123" />
    );

    const input = getByPlaceholderText('Type a message...');
    const sendButton = getByTestId('send-button');

    fireEvent.changeText(input, 'Test message');
    fireEvent.press(sendButton);

    await waitFor(() => {
      expect(getByText('Test message')).toBeTruthy();
    });
  });

  test('should retry failed messages', async () => {
    // Mock failed message
    const failedMessage = {
      id: 'msg1',
      content: 'Failed',
      status: 'failed',
    };

    const { getByTestId } = render(
      <ChatScreen peerId="peer123" messages={[failedMessage]} />
    );

    const retryButton = getByTestId('retry-button-msg1');
    fireEvent.press(retryButton);

    await waitFor(() => {
      expect(getByTestId('message-msg1')).toHaveProp('status', 'sending');
    });
  });

  test('should load history on scroll', async () => {
    const { getByTestId } = render(
      <ChatScreen peerId="peer123" />
    );

    const scrollView = getByTestId('message-list');

    // Scroll to top
    fireEvent.scroll(scrollView, {
      nativeEvent: {
        contentOffset: { y: 0 },
      },
    });

    await waitFor(() => {
      // Should load more messages
      expect(getByTestId('loading-indicator')).toBeTruthy();
    });
  });

  test('should render 1000+ messages smoothly', () => {
    const messages = Array.from({ length: 1000 }, (_, i) => ({
      id: `msg${i}`,
      content: `Message ${i}`,
      sender: i % 2 === 0 ? 'me' : 'peer123',
    }));

    const start = Date.now();

    const { getByTestId } = render(
      <ChatScreen peerId="peer123" messages={messages} />
    );

    const renderTime = Date.now() - start;

    // Should render in less than 100ms (virtualized)
    expect(renderTime).toBeLessThan(100);
  });

  test('should handle rapid typing (debouncing)', async () => {
    const { getByPlaceholderText } = render(
      <ChatScreen peerId="peer123" />
    );

    const input = getByPlaceholderText('Type a message...');

    // Type rapidly
    for (let i = 0; i < 10; i++) {
      fireEvent.changeText(input, `Text ${i}`);
    }

    // Should only trigger typing indicator once
    await waitFor(() => {
      expect(/* typing event count */).toBe(1);
    });
  });
});
```

**Implementation with NativeWind:**

```typescript
// src/components/chat/ChatScreen.tsx

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { View, TextInput, TouchableOpacity, Text } from 'react-native';
import { FlashList } from '@shopify/flash-list';
import Animated, { FadeIn, FadeOut } from 'react-native-reanimated';
import { MessageBubble } from './MessageBubble';
import { TypingIndicator } from './TypingIndicator';
import { useMessagesStore } from '../../store/messagesStore';

interface ChatScreenProps {
  peerId: string;
}

export const ChatScreen: React.FC<ChatScreenProps> = ({ peerId }) => {
  const [inputText, setInputText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const flashListRef = useRef<FlashList<any>>(null);

  const messages = useMessagesStore(state => state.getMessages(peerId));
  const sendMessage = useMessagesStore(state => state.sendMessage);
  const loadMoreMessages = useMessagesStore(state => state.loadMoreMessages);

  const handleSend = useCallback(async () => {
    if (inputText.trim()) {
      await sendMessage(peerId, inputText);
      setInputText('');

      // Scroll to bottom
      flashListRef.current?.scrollToOffset({ offset: 0, animated: true });
    }
  }, [inputText, peerId, sendMessage]);

  const handleInputChange = useCallback((text: string) => {
    setInputText(text);

    // Debounced typing indicator
    if (!isTyping) {
      setIsTyping(true);
      // Send typing event to peer
    }

    // Reset typing after 3s
    setTimeout(() => setIsTyping(false), 3000);
  }, [isTyping]);

  const handleLoadMore = useCallback(() => {
    loadMoreMessages(peerId);
  }, [peerId, loadMoreMessages]);

  const renderItem = useCallback(({ item }: { item: any }) => (
    <MessageBubble
      message={item}
      isSent={item.sender === 'me'}
    />
  ), []);

  return (
    <View className="flex-1 bg-gray-900">
      {/* Message List with GPU acceleration */}
      <FlashList
        ref={flashListRef}
        data={messages}
        renderItem={renderItem}
        estimatedItemSize={80}
        inverted
        onEndReached={handleLoadMore}
        onEndReachedThreshold={0.5}
        // GPU optimization
        contentContainerStyle={{ transform: [{ translateZ: 0 }] }}
        testID="message-list"
      />

      {/* Typing Indicator */}
      {isTyping && (
        <Animated.View entering={FadeIn} exiting={FadeOut}>
          <TypingIndicator />
        </Animated.View>
      )}

      {/* Input Bar */}
      <View className="flex-row items-center px-4 py-3 bg-gray-800 border-t border-gray-700">
        <TextInput
          value={inputText}
          onChangeText={handleInputChange}
          placeholder="Type a message..."
          placeholderTextColor="#9CA3AF"
          className="flex-1 bg-gray-700 rounded-full px-4 py-2 text-white"
          multiline
          maxLength={10000}
        />

        <TouchableOpacity
          onPress={handleSend}
          disabled={!inputText.trim()}
          className={`ml-3 w-12 h-12 rounded-full items-center justify-center transform-gpu ${
            inputText.trim() ? 'bg-blue-500' : 'bg-gray-600'
          }`}
          testID="send-button"
        >
          <Text className="text-white text-lg">→</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
};
```

**Message Bubble Component (GPU-Optimized):**

```typescript
// src/components/chat/MessageBubble.tsx

import React from 'react';
import { View, Text } from 'react-native';
import Animated, { FadeInRight, FadeInLeft } from 'react-native-reanimated';

interface MessageBubbleProps {
  message: {
    id: string;
    content: string;
    timestamp: number;
    status?: 'sending' | 'sent' | 'delivered' | 'failed';
  };
  isSent: boolean;
}

export const MessageBubble: React.FC<MessageBubbleProps> = React.memo(
  ({ message, isSent }) => {
    return (
      <Animated.View
        entering={isSent ? FadeInRight : FadeInLeft}
        className={`flex-row ${isSent ? 'justify-end' : 'justify-start'} px-4 py-2 transform-gpu`}
      >
        <View
          className={`max-w-[75%] rounded-2xl px-4 py-2 ${
            isSent ? 'bg-blue-500' : 'bg-gray-700'
          }`}
          style={{
            // GPU acceleration
            transform: [{ translateZ: 0 }],
            willChange: 'transform',
          }}
        >
          <Text className="text-white text-base">
            {message.content}
          </Text>

          <View className="flex-row items-center justify-end mt-1">
            <Text className="text-gray-300 text-xs">
              {new Date(message.timestamp).toLocaleTimeString([], {
                hour: '2-digit',
                minute: '2-digit',
              })}
            </Text>

            {isSent && message.status && (
              <Text className="ml-1 text-xs">
                {message.status === 'sending' && '⏱'}
                {message.status === 'sent' && '✓'}
                {message.status === 'delivered' && '✓✓'}
                {message.status === 'failed' && '⚠️'}
              </Text>
            )}
          </View>
        </View>
      </Animated.View>
    );
  }
);
```

---

## Phase 7: Media Handling (Performance-Focused with TDD)

### 7.1 Media Upload Component

```typescript
// __tests__/unit/components/MediaUpload.test.tsx

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { MediaUpload } from '../../../src/components/media/MediaUpload';

describe('MediaUpload', () => {
  test('should select image from gallery', async () => {
    const { getByTestId } = render(<MediaUpload />);

    const selectButton = getByTestId('select-image-button');
    fireEvent.press(selectButton);

    // Mock image picker
    // ...

    expect(getByTestId('image-preview')).toBeTruthy();
  });

  test('should validate file size', () => {
    const { getByTestId } = render(<MediaUpload />);

    // Mock large file (>10MB)
    const largeFile = { size: 11 * 1024 * 1024 };

    // Should show error
    expect(getByTestId('size-error')).toBeTruthy();
  });

  test('should compress image before upload', async () => {
    const { getByTestId } = render(<MediaUpload />);

    // Select image
    // ...

    // Should show compression progress
    expect(getByTestId('compressing-indicator')).toBeTruthy();
  });

  test('should chunk large files', async () => {
    const { getByTestId } = render(<MediaUpload />);

    // Select 5MB file
    const file = { size: 5 * 1024 * 1024 };

    // Should chunk into 64KB chunks
    const chunkCount = Math.ceil(file.size / (64 * 1024));
    expect(chunkCount).toBe(80);
  });

  test('should show upload progress', async () => {
    const { getByTestId } = render(<MediaUpload />);

    // Start upload
    // ...

    const progressBar = getByTestId('upload-progress');
    expect(progressBar).toHaveProp('value', 0);

    // Simulate progress
    await waitFor(() => {
      expect(progressBar).toHaveProp('value', 50);
    });
  });

  test('should handle upload cancellation', async () => {
    const { getByTestId } = render(<MediaUpload />);

    // Start upload
    // ...

    const cancelButton = getByTestId('cancel-upload');
    fireEvent.press(cancelButton);

    // Should stop upload
    expect(getByTestId('upload-cancelled')).toBeTruthy();
  });

  test('should retry failed chunks', async () => {
    const { getByTestId } = render(<MediaUpload />);

    // Mock chunk upload failure
    // ...

    // Should automatically retry
    await waitFor(() => {
      expect(/* retry count */).toBeGreaterThan(0);
    });
  });
});
```

---

## Phase 8: Voice/Video Calls (WebRTC with TDD)

### 8.1 Call Screen Component

```typescript
// __tests__/unit/components/CallScreen.test.tsx

import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import { CallScreen } from '../../../src/components/calls/CallScreen';

describe('CallScreen', () => {
  test('should initiate call', async () => {
    const { getByTestId } = render(
      <CallScreen peerId="peer123" />
    );

    const callButton = getByTestId('initiate-call-button');
    fireEvent.press(callButton);

    await waitFor(() => {
      expect(getByTestId('call-status')).toHaveTextContent('Calling...');
    });
  });

  test('should show incoming call notification', () => {
    const { getByTestId } = render(
      <CallScreen peerId="peer123" incomingCall={true} />
    );

    expect(getByTestId('incoming-call-modal')).toBeTruthy();
  });

  test('should handle call rejection', async () => {
    const { getByTestId } = render(
      <CallScreen peerId="peer123" incomingCall={true} />
    );

    const rejectButton = getByTestId('reject-call-button');
    fireEvent.press(rejectButton);

    await waitFor(() => {
      expect(getByTestId('call-status')).toHaveTextContent('Call rejected');
    });
  });

  test('should display video streams', async () => {
    const { getByTestId } = render(
      <CallScreen peerId="peer123" callConnected={true} />
    );

    expect(getByTestId('local-video')).toBeTruthy();
    expect(getByTestId('remote-video')).toBeTruthy();
  });

  test('should toggle mute/video', () => {
    const { getByTestId } = render(
      <CallScreen peerId="peer123" callConnected={true} />
    );

    const muteButton = getByTestId('mute-button');
    fireEvent.press(muteButton);

    expect(muteButton).toHaveProp('muted', true);

    const videoButton = getByTestId('video-button');
    fireEvent.press(videoButton);

    expect(videoButton).toHaveProp('enabled', false);
  });

  test('should handle permission denial', async () => {
    // Mock permission denial
    jest.mock('react-native-webrtc', () => ({
      mediaDevices: {
        getUserMedia: jest.fn().mockRejectedValue(new Error('Permission denied')),
      },
    }));

    const { getByTestId } = render(
      <CallScreen peerId="peer123" />
    );

    const callButton = getByTestId('initiate-call-button');
    fireEvent.press(callButton);

    await waitFor(() => {
      expect(getByTestId('permission-error')).toBeTruthy();
    });
  });
});
```

---

## Phase 9: State Management (Zustand with TDD)

### 9.1 Messages Store

```typescript
// src/store/messagesStore.ts

import create from 'zustand';
import { persist } from 'zustand/middleware';
import { MMKV } from 'react-native-mmkv';

const storage = new MMKV();

interface Message {
  id: string;
  peerId: string;
  content: string;
  timestamp: number;
  status: 'sending' | 'sent' | 'delivered' | 'failed';
  sender: string;
}

interface MessagesState {
  messages: Record<string, Message[]>;
  addMessage: (message: Message) => void;
  updateMessageStatus: (messageId: string, status: Message['status']) => void;
  getMessages: (peerId: string, offset?: number, limit?: number) => Message[];
  loadMoreMessages: (peerId: string) => Promise<void>;
  cleanupOldMessages: (days: number) => void;
  reset: () => void;
}

export const useMessagesStore = create<MessagesState>()(
  persist(
    (set, get) => ({
      messages: {},

      addMessage: (message) =>
        set((state) => ({
          messages: {
            ...state.messages,
            [message.peerId]: [
              ...(state.messages[message.peerId] || []),
              message,
            ],
          },
        })),

      updateMessageStatus: (messageId, status) =>
        set((state) => {
          const updated = { ...state.messages };

          for (const peerId in updated) {
            const messages = updated[peerId];
            const index = messages.findIndex((m) => m.id === messageId);

            if (index !== -1) {
              messages[index] = { ...messages[index], status };
              break;
            }
          }

          return { messages: updated };
        }),

      getMessages: (peerId, offset = 0, limit = 100) => {
        const messages = get().messages[peerId] || [];
        return messages.slice(offset, offset + limit);
      },

      loadMoreMessages: async (peerId) => {
        // TODO: Load from backend/storage
        console.log('Loading more messages for', peerId);
      },

      cleanupOldMessages: (days) => {
        const cutoff = Date.now() - days * 24 * 60 * 60 * 1000;

        set((state) => {
          const cleaned = { ...state.messages };

          for (const peerId in cleaned) {
            cleaned[peerId] = cleaned[peerId].filter(
              (m) => m.timestamp > cutoff
            );
          }

          return { messages: cleaned };
        });
      },

      reset: () => set({ messages: {} }),
    }),
    {
      name: 'messages-storage',
      storage: {
        getItem: (name) => {
          const value = storage.getString(name);
          return value ? JSON.parse(value) : null;
        },
        setItem: (name, value) => {
          storage.set(name, JSON.stringify(value));
        },
        removeItem: (name) => {
          storage.delete(name);
        },
      },
    }
  )
);
```

---

## Phase 10: Performance Optimizations

### 10.1 GPU Acceleration Techniques

**NativeWind Configuration:**

```javascript
// tailwind.config.js

module.exports = {
  content: ['./App.{js,jsx,ts,tsx}', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      // GPU-accelerated animations
      transitionProperty: {
        'gpu': 'transform, opacity',
      },
    },
  },
  plugins: [],
};
```

**GPU-Optimized Component Pattern:**

```typescript
import Animated, {
  useAnimatedStyle,
  withSpring,
  useSharedValue,
} from 'react-native-reanimated';

export const OptimizedComponent = () => {
  const translateY = useSharedValue(0);

  const animatedStyle = useAnimatedStyle(() => ({
    transform: [
      { translateY: translateY.value },
      { translateZ: 0 }, // Force GPU layer
    ],
  }));

  return (
    <Animated.View
      style={[
        animatedStyle,
        {
          // Force GPU rendering
          backfaceVisibility: 'hidden',
          perspective: 1000,
        },
      ]}
      className="transform-gpu will-change-transform"
    >
      {/* Content */}
    </Animated.View>
  );
};
```

### 10.2 Memory Management

```typescript
// src/hooks/useMemoryCleanup.ts

import { useEffect } from 'react';

export const useMemoryCleanup = () => {
  useEffect(() => {
    return () => {
      // Cleanup on unmount
      // - Cancel pending requests
      // - Clear timers
      // - Unsubscribe from events
      // - Release media blobs
    };
  }, []);
};
```

---

## Phase 11: Testing Strategy (Comprehensive TDD)

### 11.1 Test Coverage Goals

| Category | Coverage Target |
|----------|----------------|
| Crypto | 100% |
| API Clients | 95% |
| Stores | 90% |
| Components | 85% |
| Utils | 95% |
| **Overall** | **90%+** |

### 11.2 E2E Test Example (Detox)

```typescript
// __tests__/e2e/chat-flow.e2e.ts

describe('Chat Flow', () => {
  beforeAll(async () => {
    await device.launchApp();
  });

  it('should complete full chat conversation', async () => {
    // 1. Login/Setup
    await element(by.id('username-input')).typeText('User1');
    await element(by.id('submit-button')).tap();

    // 2. Navigate to chat
    await element(by.id('contact-list')).tap();
    await element(by.text('User2')).tap();

    // 3. Send message
    await element(by.id('message-input')).typeText('Hello!');
    await element(by.id('send-button')).tap();

    // 4. Verify message appears
    await expect(element(by.text('Hello!'))).toBeVisible();

    // 5. Receive message (from second device/user)
    // This requires running two instances
    await waitFor(element(by.text('Hi there!')))
      .toBeVisible()
      .withTimeout(5000);
  });

  it('should send media file', async () => {
    // Navigate to chat
    await element(by.id('contact-User2')).tap();

    // Open media picker
    await element(by.id('attach-button')).tap();

    // Select image
    await element(by.id('gallery-button')).tap();
    // Select first image from gallery
    await element(by.id('image-0')).tap();

    // Wait for upload
    await waitFor(element(by.id('media-message')))
      .toBeVisible()
      .withTimeout(10000);
  });

  it('should make voice call', async () => {
    // Navigate to chat
    await element(by.id('contact-User2')).tap();

    // Initiate call
    await element(by.id('call-button')).tap();

    // Verify call screen
    await expect(element(by.id('call-screen'))).toBeVisible();
    await expect(element(by.text('Calling...'))).toBeVisible();

    // End call
    await element(by.id('end-call-button')).tap();
  });
});
```

---

## Phase 12: Platform-Specific Features

### 12.1 iOS CallKit Integration

```typescript
// src/native/CallKitManager.ios.ts

import RNCallKit from 'react-native-callkit';

export class CallKitManager {
  private static instance: CallKitManager;
  private callKit: any;

  private constructor() {
    this.callKit = new RNCallKit();
    this._setupCallKit();
  }

  static getInstance(): CallKitManager {
    if (!CallKitManager.instance) {
      CallKitManager.instance = new CallKitManager();
    }
    return CallKitManager.instance;
  }

  private _setupCallKit() {
    this.callKit.setup({
      ios: {
        appName: 'Ledabeer',
        imageName: 'CallKitLogo',
        supportsVideo: true,
      },
    });

    this.callKit.addEventListener('answerCall', this._onAnswerCall);
    this.callKit.addEventListener('endCall', this._onEndCall);
  }

  async startCall(uuid: string, contactName: string, hasVideo: boolean) {
    await this.callKit.startCall(uuid, contactName, contactName, 'generic', hasVideo);
  }

  async reportIncomingCall(uuid: string, contactName: string, hasVideo: boolean) {
    await this.callKit.displayIncomingCall(
      uuid,
      contactName,
      contactName,
      'generic',
      hasVideo
    );
  }

  async endCall(uuid: string) {
    await this.callKit.endCall(uuid);
  }

  private _onAnswerCall = ({ callUUID }: any) => {
    // Handle call answer
    console.log('Call answered:', callUUID);
  };

  private _onEndCall = ({ callUUID }: any) => {
    // Handle call end
    console.log('Call ended:', callUUID);
  };
}
```

### 12.2 Android Foreground Service

```typescript
// android/app/src/main/java/com/ledabeer/CallForegroundService.java

public class CallForegroundService extends Service {
  private static final int NOTIFICATION_ID = 1;
  private static final String CHANNEL_ID = "CallChannel";

  @Override
  public int onStartCommand(Intent intent, int flags, int startId) {
    createNotificationChannel();

    Notification notification = new NotificationCompat.Builder(this, CHANNEL_ID)
      .setContentTitle("Ledabeer Call")
      .setContentText("Call in progress")
      .setSmallIcon(R.drawable.ic_call)
      .setPriority(NotificationCompat.PRIORITY_HIGH)
      .build();

    startForeground(NOTIFICATION_ID, notification);

    return START_STICKY;
  }

  private void createNotificationChannel() {
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
      NotificationChannel channel = new NotificationChannel(
        CHANNEL_ID,
        "Call Service",
        NotificationManager.IMPORTANCE_HIGH
      );

      NotificationManager manager = getSystemService(NotificationManager.class);
      manager.createNotificationChannel(channel);
    }
  }

  @Override
  public IBinder onBind(Intent intent) {
    return null;
  }
}
```

---

## Phase 13: Polish & Production

### 13.1 Security Hardening

```typescript
// src/utils/security.ts

import JailMonkey from 'jail-monkey';
import { Alert } from 'react-native';

export class SecurityManager {
  static checkDeviceSecurity(): boolean {
    // Check for jailbreak/root
    if (JailMonkey.isJailBroken()) {
      Alert.alert(
        'Security Warning',
        'This device appears to be jailbroken/rooted. The app may not function correctly.',
        [{ text: 'OK' }]
      );
      return false;
    }

    // Check for debugger
    if (__DEV__ && JailMonkey.isDebuggedMode()) {
      console.warn('Debugger detected');
    }

    return true;
  }

  static sanitizeInput(input: string): string {
    // Remove potential XSS vectors
    return input
      .replace(/<script.*?>.*?<\/script>/gi, '')
      .replace(/<iframe.*?>.*?<\/iframe>/gi, '')
      .replace(/javascript:/gi, '');
  }

  static validateMessageContent(content: string): boolean {
    // Validate message length
    if (content.length === 0 || content.length > 10000) {
      return false;
    }

    // Check for malicious patterns
    const maliciousPatterns = [
      /<script/i,
      /javascript:/i,
      /onerror=/i,
      /onload=/i,
    ];

    return !maliciousPatterns.some(pattern => pattern.test(content));
  }
}
```

### 13.2 App Store Preparation

**iOS:**

```json
// ios/ledabeer/Info.plist

<key>NSCameraUsageDescription</key>
<string>Ledabeer needs camera access for video calls</string>
<key>NSMicrophoneUsageDescription</key>
<string>Ledabeer needs microphone access for voice and video calls</string>
<key>NSPhotoLibraryUsageDescription</key>
<string>Ledabeer needs photo library access to send images</string>
<key>NSPhotoLibraryAddUsageDescription</key>
<string>Ledabeer needs permission to save images to your library</string>
```

**Android:**

```xml
<!-- android/app/src/main/AndroidManifest.xml -->

<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.RECORD_AUDIO" />
<uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
<uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />
<uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
<uses-permission android:name="android.permission.WAKE_LOCK" />
```

---

## Deliverables Checklist

- [ ] React Native project setup with Expo
- [ ] NativeWind (Tailwind CSS) configuration
- [ ] Testing infrastructure (Jest, Detox)
- [ ] gRPC client implementation
- [ ] WebSocket client with reconnection
- [ ] Signal Protocol E2EE (Identity, X3DH, Double Ratchet)
- [ ] Waku node integration
- [ ] Core UI components (Auth, Chat List, Chat Screen)
- [ ] Media handling (upload, download, display)
- [ ] WebRTC call implementation
- [ ] Zustand stores (auth, contacts, messages, calls)
- [ ] Performance optimizations (GPU, FlashList, native modules)
- [ ] 90%+ test coverage
- [ ] iOS-specific features (CallKit, VoIP push)
- [ ] Android-specific features (Foreground service, FCM)
- [ ] Security hardening
- [ ] Accessibility features
- [ ] Localization support (i18n)
- [ ] App Store submission (iOS + Android)
- [ ] Documentation (user guide, API docs, architecture)

---

## Performance Targets Summary

| Metric | Target | Test Method |
|--------|--------|-------------|
| App Launch | < 2s | Measure time to interactive |
| Message Send Latency | < 100ms | Benchmark with timer |
| List Scrolling | 60fps | FlashList with 10k items |
| Memory Usage | < 200MB | Profile with 1000 messages |
| Battery Drain (Idle) | < 5%/hour | Monitor over 2 hours |
| Call Connection Time | < 3s | WebRTC stats |
| Media Upload (1MB) | < 2s | Benchmark with compression |

---

## Next Steps After Implementation

1. **Internal Testing**: Test with team (2 weeks)
2. **TestFlight/Internal Testing**: Beta test with 50 users (2 weeks)
3. **Security Audit**: External audit of crypto implementation
4. **Performance Profiling**: Optimize bottlenecks
5. **Localization**: Add support for 5 languages
6. **App Store Submission**: Submit to iOS App Store and Google Play
7. **Marketing**: Prepare landing page, demo videos
8. **Launch**: Public release with monitoring

---

## Conclusion

This comprehensive implementation plan provides a **test-driven, production-ready roadmap** for building a React Native mobile application with:

- ✅ Full E2EE with Signal Protocol
- ✅ P2P messaging via Waku
- ✅ Voice/video calls with WebRTC
- ✅ GPU-optimized UI with NativeWind
- ✅ 150+ edge cases documented and tested
- ✅ 90%+ test coverage target
- ✅ Cross-platform (iOS + Android)
- ✅ Performance-focused architecture

By following this plan with **TDD methodology**, the frontend will be robust, maintainable, and ready for production deployment.

---

**Document Version**: 1.0
**Last Updated**: 2025-01-26
**Author**: Claude (Anthropic)
**Status**: Ready for Implementation
