/**
 * Backend Service
 *
 * Unified service layer that tries real backend first, falls back to mock
 */

import { getHttpClient } from './httpBackend/httpClient';
import { mockBackend, Message as MockMessage, Peer } from './mockBackend';

/**
 * Check if backend is available
 */
export function isBackendAvailable(): boolean {
  try {
    return getHttpClient().isConnected();
  } catch {
    return false;
  }
}

// Remove gRPC-specific code for now
// Will implement HTTP endpoints when backend API is ready

/**
 * Get list of peers
 * Falls back to mock if backend unavailable
 */
export async function getPeers(): Promise<Peer[]> {
  if (isBackendAvailable()) {
    try {
      // TODO: Backend doesn't have a "get peers" endpoint yet
      // For now, return empty array and use mock
      console.log('[BackendService] Using mock peers (backend peer discovery not implemented yet)');
      return mockBackend.getPeers();
    } catch (error) {
      console.error('[BackendService] Failed to get peers from backend:', error);
      return mockBackend.getPeers();
    }
  }

  return mockBackend.getPeers();
}

/**
 * Get message history with a peer
 */
export async function getMessagesWithPeer(peerId: string): Promise<MockMessage[]> {
  if (isBackendAvailable()) {
    try {
      // TODO: Implement HTTP endpoint for message history
      // const client = getHttpClient();
      // const response = await client.get(`/messages/${peerId}`);
      console.log('[BackendService] HTTP message endpoints not implemented yet, using mock');
      return mockBackend.getMessagesWithPeer(peerId);
    } catch (error) {
      console.error('[BackendService] Failed to get messages from backend:', error);
      return mockBackend.getMessagesWithPeer(peerId);
    }
  }

  return mockBackend.getMessagesWithPeer(peerId);
}

/**
 * Send a message to a peer
 */
export async function sendMessage(peerId: string, content: string): Promise<MockMessage> {
  if (isBackendAvailable()) {
    try {
      // TODO: Implement HTTP endpoint for sending messages
      // const client = getHttpClient();
      // const response = await client.post('/messages/send', { to: peerId, content });
      console.log('[BackendService] HTTP send endpoint not implemented yet, using mock');
      return mockBackend.sendMessage(peerId, content);
    } catch (error) {
      console.error('[BackendService] Failed to send message via backend:', error);
      return mockBackend.sendMessage(peerId, content);
    }
  }

  return mockBackend.sendMessage(peerId, content);
}

/**
 * Subscribe to incoming messages
 */
export function subscribeToMessages(callback: (message: MockMessage) => void): () => void {
  if (isBackendAvailable()) {
    try {
      // TODO: Implement WebSocket connection for real-time messages
      // const wsUrl = getHttpClient().getWebSocketUrl();
      // const ws = new WebSocket(wsUrl);
      // ws.onmessage = (event) => { /* handle message */ };
      console.log('[BackendService] WebSocket not implemented yet, using mock');
      mockBackend.onMessage(callback);
      return () => {}; // No-op unsubscribe
    } catch (error) {
      console.error('[BackendService] Failed to subscribe to backend messages:', error);
      mockBackend.onMessage(callback);
      return () => {}; // No-op unsubscribe
    }
  }

  // Use mock backend
  mockBackend.onMessage(callback);
  return () => {}; // No-op unsubscribe
}

/**
 * Get current user ID
 */
export function getCurrentUserId(): string {
  if (isBackendAvailable()) {
    // Get peer ID from backend
    // For now, use mock
    return mockBackend.getCurrentUserId();
  }

  return mockBackend.getCurrentUserId();
}
