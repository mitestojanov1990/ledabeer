# Ledabeer Mobile

End-to-End Encrypted P2P Chat Application for iOS and Android.

## Features

- 🔐 **End-to-End Encryption** - Signal Protocol (Double Ratchet + X3DH)
- 🌐 **Peer-to-Peer Messaging** - Waku protocol for decentralized communication
- 📞 **Voice & Video Calls** - WebRTC with TURN fallback
- 📁 **Media Sharing** - Encrypted image/video sharing via IPFS
- 📱 **Cross-Platform** - iOS and Android support
- ⚡ **High Performance** - GPU-accelerated UI, 60fps animations
- 🧪 **Test-Driven** - 90%+ test coverage

## Tech Stack

- **Framework**: React Native with Expo
- **Language**: TypeScript
- **UI**: NativeWind (Tailwind CSS for React Native)
- **Animations**: React Native Reanimated
- **State**: Zustand
- **Storage**: MMKV (ultra-fast key-value storage)
- **Lists**: FlashList (high-performance lists)
- **Navigation**: React Navigation
- **Testing**: Jest + Detox

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn
- Expo CLI
- iOS Simulator (macOS only) or Android Emulator

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm start

# Run on iOS
npm run ios

# Run on Android
npm run android

# Run on web
npm run web
```

## Project Structure

```
src/
├── api/              # gRPC & WebSocket clients
├── components/       # React Native components
├── crypto/           # E2EE implementation
├── hooks/            # Custom React hooks
├── store/            # Zustand stores
├── utils/            # Helper functions
├── waku/             # Waku node management
├── types/            # TypeScript types
├── constants/        # App configuration
└── navigation/       # Navigation setup

__tests__/
├── unit/             # Unit tests
├── integration/      # Integration tests
└── e2e/              # End-to-end tests
```

## Development

### Running Tests

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run E2E tests
npm run test:e2e
```

### Building for Production

```bash
# Build for iOS
eas build --platform ios

# Build for Android
eas build --platform android
```

## Architecture

The app follows a **Test-Driven Development (TDD)** approach with comprehensive edge case coverage.

### Key Components

1. **Crypto Layer** - Signal Protocol implementation for E2EE
2. **Waku Layer** - P2P messaging protocol
3. **API Layer** - gRPC and WebSocket clients
4. **UI Layer** - React Native components with GPU optimization
5. **Storage Layer** - MMKV for fast, secure storage

### Security

- All messages encrypted client-side with Signal Protocol
- Keys stored securely in iOS Keychain / Android KeyStore
- Backend cannot decrypt messages (opaque routing only)
- Forward secrecy and post-compromise security

## Contributing

This project follows strict TDD methodology. All new features must:
1. Have failing tests written first
2. Have minimal implementation to pass tests
3. Be refactored while keeping tests green

## License

MIT

## Related

- Backend: [../backend](../../backend)
- Implementation Plan: [../../plans/claude-plan/frontend-react-native-implementation.md](../../plans/claude-plan/frontend-react-native-implementation.md)
