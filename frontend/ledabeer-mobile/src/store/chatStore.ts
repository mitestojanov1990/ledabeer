/**
 * Chat Store
 *
 * Zustand store for managing chat state including messages, peers, and conversations.
 */

import { create } from 'zustand';
import { getBackendService, Message, Peer, Group } from '../services/backend';

// Get backend service instance
const backendService = getBackendService();

export type Conversation = Peer | Group;

interface ChatStore {
  conversations: Conversation[];
  groups: Group[];
  messages: Record<string, Message[]>; // Keyed by conversation ID (peer or group)
  selectedConversationId: string | null;
  loading: boolean;
  error: string | null;

  // Actions
  loadConversations: () => Promise<void>;
  loadMessages: (conversationId: string) => Promise<void>;
  sendMessage: (conversationId: string, content: string) => Promise<void>;
  selectConversation: (conversationId: string) => void;
  addIncomingMessage: (message: Message) => void;
  initializeMessageListener: () => void;
  getConversation: (id: string) => Conversation | undefined;
  isGroup: (id: string) => boolean;
}

export const useChatStore = create<ChatStore>((set, get) => ({
  conversations: [],
  groups: [],
  messages: {},
  selectedConversationId: null,
  loading: false,
  error: null,

  loadConversations: async () => {
    set({ loading: true, error: null });
    try {
      const conversations = await backendService.getAllConversations();
      const groups = await backendService.getGroups();
      set({ conversations, groups, loading: false });
    } catch (error) {
      set({ error: 'Failed to load conversations', loading: false });
    }
  },

  loadMessages: async (conversationId: string) => {
    set({ loading: true, error: null });
    try {
      const messages = await backendService.getMessagesWithPeer(conversationId);
      set((state) => ({
        messages: { ...state.messages, [conversationId]: messages },
        loading: false,
      }));
    } catch (error) {
      set({ error: 'Failed to load messages', loading: false });
    }
  },

  sendMessage: async (conversationId: string, content: string) => {
    try {
      const isGroup = backendService.isGroup(conversationId);
      const message = isGroup
        ? await backendService.sendGroupMessage(conversationId, content)
        : await backendService.sendMessage(conversationId, content);

      set((state) => ({
        messages: {
          ...state.messages,
          [conversationId]: [
            ...(state.messages[conversationId] || []),
            message,
          ],
        },
      }));
    } catch (error) {
      set({ error: 'Failed to send message' });
    }
  },

  selectConversation: (conversationId: string) => {
    set({ selectedConversationId: conversationId });
    // Load messages for the selected conversation
    get().loadMessages(conversationId);
  },

  addIncomingMessage: (message: Message) => {
    const conversationId = message.from;
    set((state) => ({
      messages: {
        ...state.messages,
        [conversationId]: [...(state.messages[conversationId] || []), message],
      },
    }));
  },

  initializeMessageListener: () => {
    backendService.onMessage((message) => {
      get().addIncomingMessage(message);
    });
  },

  getConversation: (id: string) => {
    return get().conversations.find((c) => c.id === id);
  },

  isGroup: (id: string) => {
    return backendService.isGroup(id);
  },
}));
