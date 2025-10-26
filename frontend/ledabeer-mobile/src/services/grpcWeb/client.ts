/**
 * gRPC-Web Client Implementation
 *
 * Implements gRPC-Web client for communicating with the backend
 * through Envoy proxy at http://localhost:8080
 */

// Backend gRPC-Web endpoint (Envoy proxy)
const GRPC_WEB_HOST = 'http://localhost:8080';

// Message types matching proto definitions
export interface Message {
  message_id: string;
  from_peer_id: string;
  content: Uint8Array;
  timestamp: number;
}

export interface SendMessageRequest {
  to_peer_id: string;
  content: Uint8Array;
}

export interface SendMessageResponse {
  message_id: string;
  timestamp: number;
}

export interface SendGroupMessageRequest {
  group_id: string;
  content: Uint8Array;
}

export interface GetMessageHistoryRequest {
  peer_id: string;
  limit: number;
}

export interface MessageHistoryResponse {
  messages: Message[];
}

export interface InitiateCallRequest {
  to_peer_id: string;
  audio_enabled: boolean;
  video_enabled: boolean;
}

export interface InitiateCallResponse {
  call_id: string;
  state: CallState;
}

export interface CallState {
  call_id: string;
  state: 'UNKNOWN' | 'INITIATING' | 'RINGING' | 'CONNECTED' | 'ENDED';
  participants: string[];
}

export interface UploadMediaResponse {
  media_id: string;
  cid: string;
  size: number;
}

/**
 * gRPC-Web Client for Ledabeer Backend
 */
export class GrpcWebBackendClient {
  private host: string;
  private connected: boolean = false;
  private messageListeners: Array<(message: Message) => void> = [];

  constructor(host: string = GRPC_WEB_HOST) {
    this.host = host;
  }

  /**
   * Connect to backend
   */
  async connect(): Promise<void> {
    try {
      console.log('[GrpcWebBackendClient] Connecting to', this.host);

      // Test connection with a simple health check
      const response = await fetch(`${this.host}/`);

      if (response.ok || response.status === 400) {
        // 400 is expected for root path without proper gRPC headers
        this.connected = true;
        console.log('[GrpcWebBackendClient] Connected successfully');
      } else {
        throw new Error(`Unexpected response: ${response.status}`);
      }
    } catch (error) {
      console.error('[GrpcWebBackendClient] Connection failed:', error);
      throw error;
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.connected;
  }

  /**
   * Send a text message to a peer
   */
  async sendMessage(toPeerId: string, content: string): Promise<SendMessageResponse> {
    try {
      const encoder = new TextEncoder();
      const contentBytes = encoder.encode(content);

      // For now, create a mock response since we need proper gRPC-Web encoding
      // This will be replaced with actual gRPC-Web call
      const response: SendMessageResponse = {
        message_id: `msg_${Date.now()}`,
        timestamp: Date.now(),
      };

      console.log('[GrpcWebBackendClient] Message sent (mock):', response);
      return response;
    } catch (error) {
      console.error('[GrpcWebBackendClient] sendMessage error:', error);
      throw error;
    }
  }

  /**
   * Send a message to a group
   */
  async sendGroupMessage(groupId: string, content: string): Promise<SendMessageResponse> {
    try {
      const encoder = new TextEncoder();
      const contentBytes = encoder.encode(content);

      // Mock response for now
      const response: SendMessageResponse = {
        message_id: `msg_${Date.now()}`,
        timestamp: Date.now(),
      };

      console.log('[GrpcWebBackendClient] Group message sent (mock):', response);
      return response;
    } catch (error) {
      console.error('[GrpcWebBackendClient] sendGroupMessage error:', error);
      throw error;
    }
  }

  /**
   * Get message history with a peer
   */
  async getMessageHistory(peerId: string, limit: number = 50): Promise<Message[]> {
    try {
      // For now, return empty array
      // Will implement proper gRPC-Web streaming
      console.log('[GrpcWebBackendClient] Getting message history (not implemented yet)');
      return [];
    } catch (error) {
      console.error('[GrpcWebBackendClient] getMessageHistory error:', error);
      throw error;
    }
  }

  /**
   * Subscribe to incoming messages (server streaming)
   */
  subscribeToMessages(callback: (message: Message) => void): () => void {
    console.log('[GrpcWebBackendClient] Subscribing to messages');
    this.messageListeners.push(callback);

    // Return unsubscribe function
    return () => {
      const index = this.messageListeners.indexOf(callback);
      if (index > -1) {
        this.messageListeners.splice(index, 1);
      }
    };
  }

  /**
   * Initiate a call with a peer
   */
  async initiateCall(
    toPeerId: string,
    audioEnabled: boolean,
    videoEnabled: boolean
  ): Promise<InitiateCallResponse> {
    try {
      // Mock response for now
      const response: InitiateCallResponse = {
        call_id: `call_${Date.now()}`,
        state: {
          call_id: `call_${Date.now()}`,
          state: 'INITIATING',
          participants: [toPeerId],
        },
      };

      console.log('[GrpcWebBackendClient] Call initiated (mock):', response);
      return response;
    } catch (error) {
      console.error('[GrpcWebBackendClient] initiateCall error:', error);
      throw error;
    }
  }

  /**
   * Upload media file
   */
  async uploadMedia(
    file: Blob | File,
    mimeType: string
  ): Promise<UploadMediaResponse> {
    try {
      // Mock response for now
      const response: UploadMediaResponse = {
        media_id: `media_${Date.now()}`,
        cid: `Qm${Math.random().toString(36).substring(2, 15)}`,
        size: file.size,
      };

      console.log('[GrpcWebBackendClient] Media uploaded (mock):', response);
      return response;
    } catch (error) {
      console.error('[GrpcWebBackendClient] uploadMedia error:', error);
      throw error;
    }
  }

  /**
   * Disconnect from backend
   */
  disconnect(): void {
    this.connected = false;
    this.messageListeners = [];
    console.log('[GrpcWebBackendClient] Disconnected');
  }
}

// Singleton instance
let clientInstance: GrpcWebBackendClient | null = null;

/**
 * Get singleton client instance
 */
export function getGrpcWebClient(): GrpcWebBackendClient {
  if (!clientInstance) {
    clientInstance = new GrpcWebBackendClient();
  }
  return clientInstance;
}
