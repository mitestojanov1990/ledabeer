# Frontend TODOs - Implementation Roadmap

**Date:** 2025-01-27  
**Status:** Frontend Development - Phase 3 & 4 Features

---

## Overview

This document lists all remaining TODOs and implementation needs for the React Native frontend. The core messaging functionality is working with mock data, and we need to implement real backend integration, media handling, calling features, and enhanced UI components.

---

## Current Status

### ✅ Completed Features
- **Basic Messaging UI**: Chat list and conversation screens
- **Mock Backend Integration**: Full mock data system working
- **gRPC-Web Client**: Real gRPC-Web calls implemented (Phase 2 complete)
- **State Management**: Zustand store with conversations, messages, groups
- **Group Messaging Foundation**: Group types, mock data, basic UI
- **Navigation**: React Navigation setup with stack and tab navigators
- **TypeScript**: Full type safety with interfaces

### 🔄 In Progress
- **Real Backend Integration**: gRPC-Web calls working, but some features still use mock
- **Group Conversation UI**: Basic group support, needs conversation screen updates

---

## High Priority TODOs

### 1. Backend Integration Completion

#### ❌ Critical - Real Backend Services
- **TODO: Implement real peer discovery via backend** (`src/services/backend.ts:76`)
  - Backend has peer discovery, need to integrate with gRPC
  - Replace mock peer list with real backend calls
  - Add peer status (online/offline) from backend

- **TODO: Implement real group fetching via backend** (`src/services/backend.ts:88`)
  - Backend has group management, need gRPC integration
  - Replace mock group list with real backend calls
  - Add group member management

- **TODO: Implement real group creation via backend** (`src/services/backend.ts:103`)
  - Backend has group creation, need gRPC integration
  - Replace mock group creation with real backend calls

#### ❌ HTTP Backend Services
- **TODO: Backend doesn't have a "get peers" endpoint yet** (`src/services/backendService.ts:31`)
  - Need to implement HTTP endpoint for peer discovery
  - Or use gRPC instead of HTTP

- **TODO: Implement HTTP endpoint for message history** (`src/services/backendService.ts:50`)
  - Need HTTP endpoint for message history
  - Or use gRPC instead of HTTP

- **TODO: Implement HTTP endpoint for sending messages** (`src/services/backendService.ts:70`)
  - Need HTTP endpoint for sending messages
  - Or use gRPC instead of HTTP

- **TODO: Implement WebSocket connection for real-time messages** (`src/services/backendService.ts:90`)
  - Need WebSocket for real-time message streaming
  - Or use gRPC streaming instead

### 2. gRPC-Web Streaming Implementation

#### ❌ Critical - Server Streaming
- **TODO: Implement proper streaming with ReadableStream** (`src/services/grpcWeb/realClient.ts:259`)
  - Current implementation is placeholder
  - Need to implement ReadableStream handling for message streaming
  - Handle gRPC-Web streaming protocol properly

#### ❌ gRPC-Web Client Completion
- **TODO: Make actual gRPC-Web call once Envoy is set up** (`src/services/grpcWeb/grpcWebClient.ts:83`)
  - Current implementation falls back to mock
  - Need to implement real gRPC-Web calls

- **TODO: Make actual gRPC-Web call once Envoy is set up** (`src/services/grpcWeb/grpcWebClient.ts:104`)
  - Current implementation falls back to mock
  - Need to implement real gRPC-Web calls

---

## Medium Priority TODOs

### 3. Group Conversation UI Updates

#### ❌ UI Implementation
- **Update ChatConversationScreen for Groups** (`src/screens/ChatConversationScreen.tsx`)
  - [ ] Update to accept `conversationId` instead of `peerId`
  - [ ] Display group name and member count in header
  - [ ] Show sender name for group messages
  - [ ] Different header styling for groups
  - [ ] Option to view group members
  - [ ] Option to add/remove members (if admin)

### 4. Media Message Support

#### ❌ Media Components (Need to Create)
- **MessageBubble Component** (`src/components/MessageBubble.tsx`)
  - [ ] Render different message types (text, image, video, audio, file)
  - [ ] Handle message status indicators
  - [ ] Support message reactions
  - [ ] Handle message timestamps

- **ImageMessage Component** (`src/components/ImageMessage.tsx`)
  - [ ] Display image messages with thumbnail
  - [ ] Full-size view on tap
  - [ ] Loading states
  - [ ] Error states for failed images

- **VideoMessage Component** (`src/components/VideoMessage.tsx`)
  - [ ] Display video messages with thumbnail
  - [ ] Play button overlay
  - [ ] Video player integration
  - [ ] Duration display

- **AudioMessage Component** (`src/components/AudioMessage.tsx`)
  - [ ] Audio player with controls
  - [ ] Waveform visualization
  - [ ] Play/pause functionality
  - [ ] Duration and progress

- **FileMessage Component** (`src/components/FileMessage.tsx`)
  - [ ] File attachment display
  - [ ] File size and type indicators
  - [ ] Download functionality
  - [ ] File preview

#### ❌ Media Picker (Need to Create)
- **MediaPicker Component** (`src/components/MediaPicker.tsx`)
  - [ ] Bottom sheet with media options
  - [ ] Camera integration
  - [ ] Gallery/photo picker
  - [ ] Video picker
  - [ ] Document picker

- **Media Service** (`src/services/mediaService.ts`)
  - [ ] Handle media upload/download
  - [ ] Image compression
  - [ ] Progress tracking
  - [ ] Error handling

- **Media Picker Hook** (`src/hooks/useMediaPicker.ts`)
  - [ ] Media picker state management
  - [ ] Permission handling
  - [ ] File processing

#### ❌ Dependencies Needed
```bash
npx expo install expo-image-picker expo-media-library
npx expo install expo-document-picker
npx expo install expo-av
```

### 5. Voice/Video Calling UI

#### ❌ Call Screens (Need to Create)
- **CallScreen** (`src/screens/CallScreen.tsx`)
  - [ ] Active call UI with video/audio
  - [ ] Mute/speaker/video toggle buttons
  - [ ] Call duration timer
  - [ ] End call button
  - [ ] Participant list for group calls

- **IncomingCallScreen** (`src/screens/IncomingCallScreen.tsx`)
  - [ ] Incoming call popup/notification
  - [ ] Accept/decline buttons
  - [ ] Caller information display
  - [ ] Ringtone/vibration

- **CallButton Component** (`src/components/CallButton.tsx`)
  - [ ] Call initiation button
  - [ ] Audio/video call options
  - [ ] Call state indicators

#### ❌ Call Service Integration
- **Call Service** (`src/services/callService.ts`)
  - [ ] WebRTC signaling integration
  - [ ] Call state management
  - [ ] Media stream handling
  - [ ] Call history

#### ❌ Dependencies Needed
```bash
npx expo install expo-av
# Later: react-native-webrtc for actual calls
```

---

## Low Priority TODOs

### 6. Enhanced Messaging Features

#### ❌ Advanced Features
- **Typing Indicators**
  - [ ] Show when peer is typing
  - [ ] Typing animation
  - [ ] Typing timeout handling

- **Message Status Indicators**
  - [ ] Sent, delivered, read indicators
  - [ ] Message status icons
  - [ ] Status timestamps

- **Message Actions**
  - [ ] Reply/Quote functionality
  - [ ] Message reactions (emojis)
  - [ ] Message deletion
  - [ ] Message forwarding

- **Search and Navigation**
  - [ ] Search messages within conversation
  - [ ] Search across all conversations
  - [ ] Message jump to date
  - [ ] Message jump to specific message

- **Link Preview**
  - [ ] URL detection in messages
  - [ ] Link preview cards
  - [ ] Image previews for links

### 7. Notification System

#### ❌ Notifications
- **Push Notifications**
  - [ ] Expo notifications setup
  - [ ] Message notifications
  - [ ] Call notifications
  - [ ] Notification permissions

- **In-App Notifications**
  - [ ] Toast notifications
  - [ ] Badge counts
  - [ ] Sound/vibration settings

### 8. Settings and Configuration

#### ❌ Settings Screens
- **Settings Screen** (`src/screens/SettingsScreen.tsx`)
  - [ ] User profile settings
  - [ ] Notification preferences
  - [ ] Privacy settings
  - [ ] About information

- **Profile Screen** (`src/screens/ProfileScreen.tsx`)
  - [ ] User profile editing
  - [ ] Avatar selection
  - [ ] Status message
  - [ ] Contact information

---

## Technical Debt and Improvements

### 9. Code Quality

#### ❌ Refactoring Needed
- **Service Layer Consolidation**
  - [ ] Consolidate multiple backend services
  - [ ] Remove duplicate code
  - [ ] Standardize error handling

- **Type Safety Improvements**
  - [ ] Add stricter TypeScript types
  - [ ] Remove any types
  - [ ] Add runtime type validation

- **State Management**
  - [ ] Optimize Zustand store structure
  - [ ] Add state persistence
  - [ ] Implement state migrations

### 10. Performance Optimizations

#### ❌ Performance
- **List Performance**
  - [ ] Implement FlatList optimizations
  - [ ] Add message virtualization
  - [ ] Optimize image loading

- **Memory Management**
  - [ ] Implement proper cleanup
  - [ ] Add memory monitoring
  - [ ] Optimize image caching

- **Network Optimization**
  - [ ] Implement request batching
  - [ ] Add offline support
  - [ ] Implement retry logic

---

## Implementation Priority

### Phase 1: Backend Integration (High Priority)
1. **Complete gRPC-Web streaming** - Essential for real-time messaging
2. **Implement real peer discovery** - Replace mock data
3. **Implement real group management** - Replace mock data
4. **Update conversation screen for groups** - Complete group UI

### Phase 2: Media Support (Medium Priority)
1. **Create media message components** - Image, video, audio, file
2. **Implement media picker** - Camera, gallery, document picker
3. **Add media service** - Upload/download handling
4. **Integrate with backend media service** - Real media storage

### Phase 3: Calling Features (Medium Priority)
1. **Create call screens** - Active call, incoming call
2. **Implement call buttons** - Call initiation
3. **Integrate with backend call service** - WebRTC signaling
4. **Add call state management** - Call lifecycle

### Phase 4: Enhanced Features (Low Priority)
1. **Add typing indicators** - Real-time typing status
2. **Implement message status** - Sent, delivered, read
3. **Add message actions** - Reply, react, forward
4. **Create settings screens** - User preferences

### Phase 5: Polish and Optimization (Low Priority)
1. **Performance optimizations** - List performance, memory
2. **Code quality improvements** - Refactoring, type safety
3. **Notification system** - Push notifications
4. **Advanced features** - Search, link preview

---

## Dependencies to Install

### Media Support
```bash
npx expo install expo-image-picker expo-media-library
npx expo install expo-document-picker
npx expo install expo-av
```

### Calling Features
```bash
npx expo install expo-av
# Later: react-native-webrtc
```

### Notifications
```bash
npx expo install expo-notifications
```

### Additional Utilities
```bash
npx expo install expo-camera
npx expo install expo-file-system
npx expo install expo-sharing
```

---

## File Structure to Create

```
src/
├── components/
│   ├── MessageBubble.tsx
│   ├── ImageMessage.tsx
│   ├── VideoMessage.tsx
│   ├── AudioMessage.tsx
│   ├── FileMessage.tsx
│   ├── MediaPicker.tsx
│   ├── CallButton.tsx
│   └── TypingIndicator.tsx
├── screens/
│   ├── CallScreen.tsx
│   ├── IncomingCallScreen.tsx
│   ├── SettingsScreen.tsx
│   └── ProfileScreen.tsx
├── services/
│   ├── mediaService.ts
│   ├── callService.ts
│   └── notificationService.ts
├── hooks/
│   ├── useMediaPicker.ts
│   ├── useCall.ts
│   └── useNotifications.ts
└── utils/
    ├── mediaUtils.ts
    ├── callUtils.ts
    └── notificationUtils.ts
```

---

## Success Criteria

### Phase 1 Complete
- [ ] Real-time messaging with backend
- [ ] Group conversations working
- [ ] Peer discovery from backend
- [ ] Group management from backend

### Phase 2 Complete
- [ ] Media messages display correctly
- [ ] Can send images, videos, audio, files
- [ ] Media picker working
- [ ] Media upload/download working

### Phase 3 Complete
- [ ] Can make voice calls
- [ ] Can make video calls
- [ ] Call UI working
- [ ] Call state management working

### Phase 4 Complete
- [ ] Typing indicators working
- [ ] Message status working
- [ ] Message actions working
- [ ] Settings screens working

### Phase 5 Complete
- [ ] Performance optimized
- [ ] Code quality high
- [ ] Notifications working
- [ ] All features polished

---

## Notes

- **Current Status**: Basic messaging UI working with mock data
- **Backend Integration**: gRPC-Web calls working, but some features still use mock
- **Priority**: Focus on backend integration first, then media support
- **Dependencies**: Need to install media and calling dependencies
- **Architecture**: Well-structured with TypeScript and Zustand
- **Testing**: Need to add comprehensive testing

---

**Last Updated:** 2025-01-27  
**Total TODOs:** 50+  
**Estimated Completion Time:** 20-30 days

---

**App Running:** http://localhost:8090  
**Current Features:** Mock messaging, basic group support, gRPC-Web integration  
**Next Steps:** Complete backend integration, add media support
