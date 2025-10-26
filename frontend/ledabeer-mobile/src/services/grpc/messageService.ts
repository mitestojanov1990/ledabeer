/**
 * Message Service
 *
 * gRPC wrapper for MessageService operations
 */

import { getBackendClient } from './backendClient';

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

export class MessageService {
  /**
   * Send a message to a peer
   */
  async sendMessage(toPeerId: string, content: string): Promise<SendMessageResponse> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getMessageClient();

      // Convert string to bytes
      const contentBytes = Buffer.from(content, 'utf-8');

      const request = {
        to_peer_id: toPeerId,
        content: contentBytes,
      };

      (client as any).SendMessage(request, (error: any, response: any) => {
        if (error) {
          console.error('[MessageService] SendMessage error:', error);
          reject(error);
        } else {
          resolve({
            message_id: response.message_id,
            timestamp: parseInt(response.timestamp),
          });
        }
      });
    });
  }

  /**
   * Get message history with a peer
   */
  async getMessageHistory(peerId: string, limit: number = 50): Promise<Message[]> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getMessageClient();

      const request = {
        peer_id: peerId,
        limit: limit,
      };

      (client as any).GetMessageHistory(request, (error: any, response: any) => {
        if (error) {
          console.error('[MessageService] GetMessageHistory error:', error);
          reject(error);
        } else {
          const messages: Message[] = (response.messages || []).map((msg: any) => ({
            message_id: msg.message_id,
            from_peer_id: msg.from_peer_id,
            content: msg.content,
            timestamp: parseInt(msg.timestamp),
          }));
          resolve(messages);
        }
      });
    });
  }

  /**
   * Subscribe to incoming messages (streaming)
   */
  subscribeToMessages(
    onMessage: (message: Message) => void,
    onError?: (error: Error) => void
  ): () => void {
    const client = getBackendClient().getMessageClient();

    const request = {};
    const stream = (client as any).ReceiveMessages(request);

    stream.on('data', (message: any) => {
      onMessage({
        message_id: message.message_id,
        from_peer_id: message.from_peer_id,
        content: message.content,
        timestamp: parseInt(message.timestamp),
      });
    });

    stream.on('error', (error: any) => {
      console.error('[MessageService] ReceiveMessages stream error:', error);
      if (onError) {
        onError(error);
      }
    });

    stream.on('end', () => {
      console.log('[MessageService] ReceiveMessages stream ended');
    });

    // Return unsubscribe function
    return () => {
      stream.cancel();
    };
  }

  /**
   * Send a group message
   */
  async sendGroupMessage(groupId: string, content: string): Promise<SendMessageResponse> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getMessageClient();

      const contentBytes = Buffer.from(content, 'utf-8');

      const request = {
        group_id: groupId,
        content: contentBytes,
      };

      (client as any).SendGroupMessage(request, (error: any, response: any) => {
        if (error) {
          console.error('[MessageService] SendGroupMessage error:', error);
          reject(error);
        } else {
          resolve({
            message_id: response.message_id,
            timestamp: parseInt(response.timestamp),
          });
        }
      });
    });
  }
}

// Export singleton instance
export const messageService = new MessageService();
