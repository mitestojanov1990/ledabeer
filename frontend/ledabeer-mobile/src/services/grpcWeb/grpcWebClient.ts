/**
 * gRPC-Web Client
 *
 * Browser-compatible gRPC client using grpc-web
 */

import { grpc } from 'grpc-web';

// Envoy proxy URL (will proxy to gRPC backend)
const GRPC_WEB_URL = 'http://localhost:8080';

export interface GrpcMessage {
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

export interface GetMessageHistoryRequest {
  peer_id: string;
  limit: number;
}

export interface MessageHistoryResponse {
  messages: GrpcMessage[];
}

/**
 * gRPC-Web Client for MessageService
 */
class GrpcWebClient {
  private serviceUrl: string;
  private connected: boolean = false;

  constructor(url: string = GRPC_WEB_URL) {
    this.serviceUrl = url;
  }

  /**
   * Test connection to backend
   */
  async connect(): Promise<void> {
    try {
      console.log('[GrpcWebClient] Connecting to backend at', this.serviceUrl);

      // For now, just mark as connected
      // Real connection test will happen on first RPC call
      this.connected = true;
      console.log('[GrpcWebClient] Connected successfully');
    } catch (error) {
      console.error('[GrpcWebClient] Connection failed:', error);
      throw new Error(`Failed to connect to backend: ${error}`);
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.connected;
  }

  /**
   * Send a message to a peer
   *
   * Note: This is a manual implementation until we set up Envoy proxy
   * For now, it will fall back to mock backend
   */
  async sendMessage(toPeerId: string, content: string): Promise<SendMessageResponse> {
    try {
      console.log('[GrpcWebClient] sendMessage not fully implemented yet');

      // TODO: Make actual gRPC-Web call once Envoy is set up
      // const client = new MessageServiceClient(this.serviceUrl);
      // const request = new SendMessageRequest();
      // request.setToPeerId(toPeerId);
      // request.setContent(Buffer.from(content));
      // const response = await client.sendMessage(request, {});

      throw new Error('gRPC-Web requires Envoy proxy - not yet configured');
    } catch (error) {
      console.error('[GrpcWebClient] sendMessage error:', error);
      throw error;
    }
  }

  /**
   * Get message history with a peer
   */
  async getMessageHistory(peerId: string, limit: number = 50): Promise<GrpcMessage[]> {
    try {
      console.log('[GrpcWebClient] getMessageHistory not fully implemented yet');

      // TODO: Make actual gRPC-Web call once Envoy is set up

      throw new Error('gRPC-Web requires Envoy proxy - not yet configured');
    } catch (error) {
      console.error('[GrpcWebClient] getMessageHistory error:', error);
      throw error;
    }
  }
}

// Singleton instance
let grpcWebClientInstance: GrpcWebClient | null = null;

/**
 * Get singleton gRPC-Web client instance
 */
export function getGrpcWebClient(): GrpcWebClient {
  if (!grpcWebClientInstance) {
    grpcWebClientInstance = new GrpcWebClient();
  }
  return grpcWebClientInstance;
}
