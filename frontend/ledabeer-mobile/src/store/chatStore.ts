/**
 * Chat Store
 *
 * Zustand store for managing chat state including messages, peers, and conversations.
 */

import { create } from 'zustand';
import { Message, Peer, mockBackend } from '../services/mockBackend';

interface ChatStore {
  peers: Peer[];
  messages: Record<string, Message[]>; // Keyed by peer ID
  selectedPeerId: string | null;
  loading: boolean;
  error: string | null;

  // Actions
  loadPeers: () => Promise<void>;
  loadMessages: (peerId: string) => Promise<void>;
  sendMessage: (peerId: string, content: string) => Promise<void>;
  selectPeer: (peerId: string) => void;
  addIncomingMessage: (message: Message) => void;
  initializeMessageListener: () => void;
}

export const useChatStore = create<ChatStore>((set, get) => ({
  peers: [],
  messages: {},
  selectedPeerId: null,
  loading: false,
  error: null,

  loadPeers: async () => {
    set({ loading: true, error: null });
    try {
      const peers = await mockBackend.getPeers();
      set({ peers, loading: false });
    } catch (error) {
      set({ error: 'Failed to load peers', loading: false });
    }
  },

  loadMessages: async (peerId: string) => {
    set({ loading: true, error: null });
    try {
      const messages = await mockBackend.getMessagesWithPeer(peerId);
      set((state) => ({
        messages: { ...state.messages, [peerId]: messages },
        loading: false,
      }));
    } catch (error) {
      set({ error: 'Failed to load messages', loading: false });
    }
  },

  sendMessage: async (peerId: string, content: string) => {
    try {
      const message = await mockBackend.sendMessage(peerId, content);
      set((state) => ({
        messages: {
          ...state.messages,
          [peerId]: [...(state.messages[peerId] || []), message],
        },
      }));
    } catch (error) {
      set({ error: 'Failed to send message' });
    }
  },

  selectPeer: (peerId: string) => {
    set({ selectedPeerId: peerId });
    // Load messages for the selected peer
    get().loadMessages(peerId);
  },

  addIncomingMessage: (message: Message) => {
    const peerId = message.from;
    set((state) => ({
      messages: {
        ...state.messages,
        [peerId]: [...(state.messages[peerId] || []), message],
      },
    }));
  },

  initializeMessageListener: () => {
    mockBackend.onMessage((message) => {
      get().addIncomingMessage(message);
    });
  },
}));
