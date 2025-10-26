/**
 * Application Configuration
 *
 * Central configuration file for the Ledabeer mobile app.
 * Contains all environment-specific settings and constants.
 */

export const Config = {
  // Backend API Configuration
  api: {
    // gRPC server endpoint
    grpcUrl: __DEV__ ? 'http://localhost:5001' : 'https://api.ledabeer.com:5001',

    // WebSocket server endpoint
    wsUrl: __DEV__ ? 'ws://localhost:8080/ws' : 'wss://api.ledabeer.com/ws',

    // Request timeout (milliseconds)
    timeout: 30000,

    // Maximum retry attempts for failed requests
    maxRetries: 5,

    // Retry delay multiplier (exponential backoff)
    retryDelayMultiplier: 2,

    // Initial retry delay (milliseconds)
    initialRetryDelay: 1000,
  },

  // Waku Protocol Configuration
  waku: {
    // Bootstrap nodes for peer discovery
    bootstrapNodes: __DEV__
      ? ['ws://localhost:8546']
      : [
          '/dns4/node-01.do-ams3.ledabeer.com/tcp/443/wss/p2p/16Uiu2HAm...',
          '/dns4/node-02.ac-cn-hongkong-c.ledabeer.com/tcp/443/wss/p2p/16Uiu2HAm...',
        ],

    // Fallback hardcoded nodes
    fallbackNodes: [
      '/dns4/node-01.do-ams3.wakuv2.prod.statusim.net/tcp/443/wss/p2p/16Uiu2HAkvWiyFsgRhuJEb9JfjYxEkoHLgnUQmr1N5mKWnYjxYRVm',
    ],

    // Chat topic for messages
    chatTopic: '/ledabeer/1/chat/proto',

    // Store protocol query limit
    storeQueryLimit: 100,

    // Message deduplication window (milliseconds)
    deduplicationWindow: 3600000, // 1 hour
  },

  // Cryptography Configuration
  crypto: {
    // Key rotation policy (days)
    keyRotationDays: 30,

    // Message key rotation threshold (number of messages)
    messageKeyRotation: 1000,

    // PBKDF2 iterations for key derivation
    pbkdf2Iterations: 100000,

    // Session timeout (milliseconds)
    sessionTimeout: 86400000, // 24 hours
  },

  // Storage Configuration
  storage: {
    // MMKV encryption key
    encryptionKey: 'ledabeer-secure-storage-key',

    // Message retention period (days)
    messageRetentionDays: 30,

    // Maximum messages to keep in memory per chat
    maxMessagesInMemory: 100,

    // Cache size limit (bytes)
    cacheSize: 50 * 1024 * 1024, // 50MB
  },

  // Media Configuration
  media: {
    // Maximum file size (bytes)
    maxFileSize: 10 * 1024 * 1024, // 10MB

    // Chunk size for uploads (bytes)
    chunkSize: 64 * 1024, // 64KB

    // Supported image formats
    supportedImageFormats: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],

    // Supported video formats
    supportedVideoFormats: ['video/mp4', 'video/webm', 'video/quicktime'],

    // Image compression quality (0-1)
    imageQuality: 0.8,

    // Thumbnail size (pixels)
    thumbnailSize: 200,
  },

  // Call Configuration
  calls: {
    // ICE servers for WebRTC
    iceServers: [
      { urls: 'stun:stun.l.google.com:19302' },
      { urls: 'stun:stun1.l.google.com:19302' },
      // TURN servers (add credentials in production)
      // {
      //   urls: 'turn:turn.ledabeer.com:3478',
      //   username: 'ledabeer',
      //   credential: 'secure-password'
      // },
    ],

    // Signaling timeout (milliseconds)
    signalingTimeout: 30000,

    // Call quality settings
    videoConstraints: {
      width: { ideal: 1280 },
      height: { ideal: 720 },
      frameRate: { ideal: 30 },
    },

    audioConstraints: {
      echoCancellation: true,
      noiseSuppression: true,
      autoGainControl: true,
    },
  },

  // UI Configuration
  ui: {
    // Theme colors
    colors: {
      primary: '#3B82F6', // Blue
      secondary: '#10B981', // Green
      danger: '#EF4444', // Red
      warning: '#F59E0B', // Amber
      background: '#111827', // Gray 900
      surface: '#1F2937', // Gray 800
      text: '#F9FAFB', // Gray 50
    },

    // Animation durations (milliseconds)
    animationDuration: {
      fast: 150,
      normal: 300,
      slow: 500,
    },

    // Debounce delays (milliseconds)
    debounceDelay: {
      search: 300,
      typing: 1000,
      input: 500,
    },
  },

  // Feature Flags
  features: {
    // Enable voice calls
    voiceCalls: true,

    // Enable video calls
    videoCalls: true,

    // Enable group calls
    groupCalls: false, // Not yet implemented

    // Enable media sharing
    mediaSharing: true,

    // Enable offline mode
    offlineMode: true,

    // Enable push notifications
    pushNotifications: true,

    // Enable analytics (disable in production for privacy)
    analytics: __DEV__,
  },

  // App Information
  app: {
    name: 'Ledabeer',
    version: '1.0.0',
    buildNumber: 1,
  },
} as const;

// Type-safe config access
export type AppConfig = typeof Config;

// Environment check helpers
export const isDevelopment = __DEV__;
export const isProduction = !__DEV__;
