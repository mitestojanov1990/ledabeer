/**
 * Core TypeScript Type Definitions
 *
 * Central type definitions for the Ledabeer mobile app.
 */

// ============================================================================
// User & Identity Types
// ============================================================================

export interface Identity {
  publicKey: Uint8Array;
  privateKey: Uint8Array;
  createdAt: number;
  expiresAt: number;
}

export interface User {
  id: string;
  username: string;
  publicKey: string;
  avatar?: string;
  status: 'online' | 'offline' | 'away';
  lastSeen: number;
}

// ============================================================================
// Message Types
// ============================================================================

export type MessageStatus = 'sending' | 'sent' | 'delivered' | 'read' | 'failed';

export type MessageType = 'text' | 'image' | 'video' | 'audio' | 'file';

export interface Message {
  id: string;
  peerId: string;
  content: string;
  timestamp: number;
  status: MessageStatus;
  sender: 'me' | string;
  type: MessageType;
  mediaId?: string;
  mediaCid?: string;
  mimeType?: string;
  encryptedContent?: Uint8Array;
}

export interface MessageHistoryRequest {
  peerId: string;
  limit?: number;
  offset?: number;
  beforeTimestamp?: number;
}

// ============================================================================
// Contact Types
// ============================================================================

export interface Contact {
  id: string;
  peerId: string;
  username: string;
  publicKey: string;
  avatar?: string;
  status: 'online' | 'offline' | 'away';
  lastSeen: number;
  unreadCount: number;
  lastMessage?: Message;
  isPinned: boolean;
  isMuted: boolean;
}

// ============================================================================
// Call Types
// ============================================================================

export type CallState = 'idle' | 'initiating' | 'ringing' | 'connecting' | 'connected' | 'ended' | 'failed';

export type CallType = 'audio' | 'video';

export interface Call {
  id: string;
  peerId: string;
  type: CallType;
  state: CallState;
  isIncoming: boolean;
  startTime?: number;
  endTime?: number;
  duration?: number;
  localStream?: any; // MediaStream
  remoteStream?: any; // MediaStream
}

export interface CallOptions {
  audio: boolean;
  video: boolean;
  turnConfig?: TURNConfig;
}

export interface TURNConfig {
  urls: string[];
  username: string;
  credential: string;
}

// ============================================================================
// Media Types
// ============================================================================

export interface MediaFile {
  uri: string;
  type: string;
  size: number;
  name: string;
  mimeType: string;
}

export interface MediaUploadProgress {
  mediaId: string;
  progress: number;
  chunkIndex: number;
  totalChunks: number;
  status: 'uploading' | 'processing' | 'completed' | 'failed';
}

export interface MediaChunk {
  data: Uint8Array;
  index: number;
  total: number;
  mediaId: string;
}

// ============================================================================
// Crypto Types
// ============================================================================

export interface CryptoSession {
  peerId: string;
  chainKey: Uint8Array;
  messageNumber: number;
  createdAt: number;
  lastUsed: number;
}

export interface X3DHBundle {
  identityKey: Uint8Array;
  signedPreKey: Uint8Array;
  preKeySignature: Uint8Array;
  oneTimePreKey?: Uint8Array;
}

export interface EncryptedMessage {
  ciphertext: Uint8Array;
  messageNumber: number;
  sessionId: string;
}

// ============================================================================
// Waku Types
// ============================================================================

export interface WakuMessage {
  payload: Uint8Array;
  contentTopic: string;
  timestamp: number;
}

export interface WakuPeer {
  id: string;
  multiaddr: string;
  protocols: string[];
}

export interface WakuStoreQuery {
  contentTopic: string;
  startTime?: number;
  endTime?: number;
  limit?: number;
}

// ============================================================================
// API Types
// ============================================================================

export interface APIError {
  code: string;
  message: string;
  details?: any;
}

export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: APIError;
}

// ============================================================================
// Store Types
// ============================================================================

export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  identity: Identity | null;
  loading: boolean;
  error: string | null;
}

export interface ContactsState {
  contacts: Contact[];
  selectedContact: Contact | null;
  loading: boolean;
  error: string | null;
}

export interface MessagesState {
  messages: Record<string, Message[]>;
  loading: boolean;
  error: string | null;
}

export interface CallsState {
  activeCall: Call | null;
  incomingCall: Call | null;
  callHistory: Call[];
  loading: boolean;
  error: string | null;
}

export interface UIState {
  theme: 'light' | 'dark' | 'auto';
  selectedChatId: string | null;
  isTyping: Record<string, boolean>;
  notifications: Notification[];
}

export interface WakuState {
  isRunning: boolean;
  peers: WakuPeer[];
  isConnected: boolean;
  error: string | null;
}

// ============================================================================
// Notification Types
// ============================================================================

export interface Notification {
  id: string;
  type: 'message' | 'call' | 'system';
  title: string;
  body: string;
  data?: any;
  timestamp: number;
  read: boolean;
}

// ============================================================================
// Navigation Types
// ============================================================================

export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
  Chat: { peerId: string };
  Call: { callId: string };
  Settings: undefined;
  Profile: { userId: string };
  MediaViewer: { mediaId: string };
};

// ============================================================================
// Event Types
// ============================================================================

export interface AppEvent {
  type: string;
  payload?: any;
  timestamp: number;
}

export type EventListener = (event: AppEvent) => void;

// ============================================================================
// Utility Types
// ============================================================================

export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

export type AsyncState<T> = {
  data: T | null;
  loading: boolean;
  error: string | null;
};

export type Prettify<T> = {
  [K in keyof T]: T[K];
} & {};
