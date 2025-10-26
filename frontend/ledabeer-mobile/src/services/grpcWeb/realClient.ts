/**
 * Real gRPC-Web Client Implementation
 *
 * Implements actual gRPC-Web calls to the backend through Envoy proxy
 */

import {
  encodeSendMessageRequest,
  encodeSendGroupMessageRequest,
  encodeGetMessageHistoryRequest,
  encodeReceiveMessagesRequest,
  decodeSendMessageResponse,
  decodeMessage,
  decodeMessageHistoryResponse,
} from './messages';

// Backend gRPC-Web endpoint (Envoy proxy)
const GRPC_WEB_HOST = 'http://localhost:8080';

// Service paths
const MESSAGE_SERVICE = 'ledabeer.MessageService';

// Message types matching proto definitions
export interface Message {
  message_id: string;
  from_peer_id: string;
  content: Uint8Array;
  timestamp: number;
}

export interface SendMessageResponse {
  message_id: string;
  timestamp: number;
}

/**
 * Make a gRPC-Web unary call
 */
async function grpcUnaryCall(
  service: string,
  method: string,
  requestData: Uint8Array
): Promise<Uint8Array> {
  const url = `${GRPC_WEB_HOST}/${service}/${method}`;

  console.log(`[gRPC-Web] Calling ${service}/${method}`);

  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/grpc-web+proto',
      'X-Grpc-Web': '1',
      'X-User-Agent': 'grpc-web-javascript/0.1',
    },
    body: requestData as any, // TypeScript workaround for Uint8Array in body
  });

  if (!response.ok) {
    throw new Error(`gRPC call failed: ${response.status} ${response.statusText}`);
  }

  const contentType = response.headers.get('content-type');
  if (!contentType?.includes('application/grpc-web')) {
    throw new Error(`Unexpected content type: ${contentType}`);
  }

  // Read response as array buffer
  const buffer = await response.arrayBuffer();
  const data = new Uint8Array(buffer);

  // gRPC-Web response format:
  // - First byte: compression flag (0x00 = no compression)
  // - Next 4 bytes: message length (big-endian)
  // - Remaining bytes: protobuf message

  if (data.length < 5) {
    throw new Error('Invalid gRPC-Web response: too short');
  }

  const compressed = data[0];
  if (compressed !== 0) {
    throw new Error('Compressed responses not supported');
  }

  // Read message length (big-endian 32-bit)
  const messageLength =
    (data[1] << 24) | (data[2] << 16) | (data[3] << 8) | data[4];

  // Extract message data
  const messageData = data.slice(5, 5 + messageLength);

  console.log(`[gRPC-Web] Response received: ${messageData.length} bytes`);

  return messageData;
}

/**
 * Real gRPC-Web Backend Client
 */
export class RealGrpcWebClient {
  private host: string;
  private connected: boolean = false;

  constructor(host: string = GRPC_WEB_HOST) {
    this.host = host;
  }

  /**
   * Connect to backend
   */
  async connect(): Promise<void> {
    try {
      console.log('[RealGrpcWebClient] Connecting to', this.host);

      // Test connection - 415 or 400 means Envoy is accessible
      const response = await fetch(`${this.host}/`).catch(() => ({ status: 0, ok: false }));

      if (response.ok || response.status === 400 || response.status === 415) {
        this.connected = true;
        console.log('[RealGrpcWebClient] Connected successfully');
      } else {
        throw new Error(`Unexpected response: ${response.status}`);
      }
    } catch (error) {
      console.error('[RealGrpcWebClient] Connection failed:', error);
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
   * Send a text message to a peer (REAL IMPLEMENTATION)
   */
  async sendMessage(toPeerId: string, content: string): Promise<SendMessageResponse> {
    try {
      console.log('[RealGrpcWebClient] Sending message to', toPeerId);

      // Encode request
      const requestData = encodeSendMessageRequest(toPeerId, content);

      // Make gRPC call
      const responseData = await grpcUnaryCall(
        MESSAGE_SERVICE,
        'SendMessage',
        requestData
      );

      // Decode response
      const response = decodeSendMessageResponse(responseData);

      console.log('[RealGrpcWebClient] Message sent:', response.message_id);

      return response;
    } catch (error) {
      console.error('[RealGrpcWebClient] sendMessage error:', error);
      throw error;
    }
  }

  /**
   * Send a message to a group (REAL IMPLEMENTATION)
   */
  async sendGroupMessage(groupId: string, content: string): Promise<SendMessageResponse> {
    try {
      console.log('[RealGrpcWebClient] Sending group message to', groupId);

      // Encode request
      const requestData = encodeSendGroupMessageRequest(groupId, content);

      // Make gRPC call
      const responseData = await grpcUnaryCall(
        MESSAGE_SERVICE,
        'SendGroupMessage',
        requestData
      );

      // Decode response
      const response = decodeSendMessageResponse(responseData);

      console.log('[RealGrpcWebClient] Group message sent:', response.message_id);

      return response;
    } catch (error) {
      console.error('[RealGrpcWebClient] sendGroupMessage error:', error);
      throw error;
    }
  }

  /**
   * Get message history with a peer (REAL IMPLEMENTATION)
   */
  async getMessageHistory(peerId: string, limit: number = 50): Promise<Message[]> {
    try {
      console.log('[RealGrpcWebClient] Getting message history for', peerId);

      // Encode request
      const requestData = encodeGetMessageHistoryRequest(peerId, limit);

      // Make gRPC call
      const responseData = await grpcUnaryCall(
        MESSAGE_SERVICE,
        'GetMessageHistory',
        requestData
      );

      // Decode response
      const messages = decodeMessageHistoryResponse(responseData);

      console.log('[RealGrpcWebClient] Message history received:', messages.length, 'messages');

      return messages;
    } catch (error) {
      console.error('[RealGrpcWebClient] getMessageHistory error:', error);
      throw error;
    }
  }

  /**
   * Subscribe to incoming messages (server streaming)
   *
   * Note: Server streaming with fetch API requires careful handling
   * This is a simplified implementation
   */
  subscribeToMessages(callback: (message: Message) => void): () => void {
    console.log('[RealGrpcWebClient] Subscribing to messages (streaming)');

    let cancelled = false;

    // Start streaming
    (async () => {
      try {
        const requestData = encodeReceiveMessagesRequest();
        const url = `${GRPC_WEB_HOST}/${MESSAGE_SERVICE}/ReceiveMessages`;

        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/grpc-web+proto',
            'X-Grpc-Web': '1',
            'X-User-Agent': 'grpc-web-javascript/0.1',
          },
          body: requestData as any, // TypeScript workaround for Uint8Array in body
        });

        if (!response.ok) {
          throw new Error(`Streaming failed: ${response.status}`);
        }

        // For now, this is a placeholder
        // Full streaming implementation requires reading the response body as a stream
        console.log('[RealGrpcWebClient] Message streaming connected');

        // TODO: Implement proper streaming with ReadableStream
        // const reader = response.body?.getReader();
        // while (!cancelled && reader) {
        //   const { done, value } = await reader.read();
        //   if (done) break;
        //   // Parse and decode each message frame
        // }

      } catch (error) {
        if (!cancelled) {
          console.error('[RealGrpcWebClient] Streaming error:', error);
        }
      }
    })();

    // Return unsubscribe function
    return () => {
      cancelled = true;
      console.log('[RealGrpcWebClient] Unsubscribed from messages');
    };
  }

  /**
   * Disconnect from backend
   */
  disconnect(): void {
    this.connected = false;
    console.log('[RealGrpcWebClient] Disconnected');
  }
}

// Singleton instance
let clientInstance: RealGrpcWebClient | null = null;

/**
 * Get singleton client instance
 */
export function getRealGrpcWebClient(): RealGrpcWebClient {
  if (!clientInstance) {
    clientInstance = new RealGrpcWebClient();
  }
  return clientInstance;
}
