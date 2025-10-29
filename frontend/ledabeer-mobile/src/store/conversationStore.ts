import { create } from 'zustand';
import { Conversation, MessagePreview } from '../services/conversationService';
import { realtimeService, RealtimeMessage } from '../services/realtimeService';

interface ConversationState {
  // Conversations
  conversations: Conversation[];
  currentConversation: Conversation | null;
  
  // Loading states
  loading: boolean;
  loadingMessages: boolean;
  
  // Error states
  error: string | null;
  
  // Actions
  setConversations: (conversations: Conversation[]) => void;
  addConversation: (conversation: Conversation) => void;
  updateConversation: (conversation: Conversation) => void;
  setCurrentConversation: (conversation: Conversation | null) => void;
  setLoading: (loading: boolean) => void;
  setLoadingMessages: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // Message actions
  addMessage: (conversationId: string, message: MessagePreview) => void;
  updateLastMessage: (conversationId: string, message: MessagePreview) => void;
  markAsRead: (conversationId: string) => void;
  
  // Realtime actions
  handleRealtimeMessage: (message: RealtimeMessage) => void;
  initializeRealtime: () => void;
  cleanupRealtime: () => void;
}

export const useConversationStore = create<ConversationState>((set, get) => ({
  // Initial state
  conversations: [],
  currentConversation: null,
  loading: false,
  loadingMessages: false,
  error: null,

  // Conversation actions
  setConversations: (conversations) => set({ conversations }),
  
  addConversation: (conversation) => set((state) => {
    // Check if conversation already exists
    const existingIndex = state.conversations.findIndex(c => c.id === conversation.id);
    if (existingIndex >= 0) {
      // Update existing conversation
      const updated = [...state.conversations];
      updated[existingIndex] = conversation;
      return { conversations: updated };
    } else {
      // Add new conversation at the beginning
      return { conversations: [conversation, ...state.conversations] };
    }
  }),
  
  updateConversation: (conversation) => set((state) => {
    const updated = state.conversations.map(c => 
      c.id === conversation.id ? conversation : c
    );
    return { conversations: updated };
  }),
  
  setCurrentConversation: (conversation) => set({ currentConversation: conversation }),
  
  setLoading: (loading) => set({ loading }),
  
  setLoadingMessages: (loading) => set({ loadingMessages: loading }),
  
  setError: (error) => set({ error }),

  // Message actions
  addMessage: (conversationId, message) => set((state) => {
    const updated = state.conversations.map(conv => {
      if (conv.id === conversationId) {
        return {
          ...conv,
          last_message: message,
          updated_at: message.timestamp,
        };
      }
      return conv;
    });
    return { conversations: updated };
  }),
  
  updateLastMessage: (conversationId, message) => set((state) => {
    const updated = state.conversations.map(conv => {
      if (conv.id === conversationId) {
        return {
          ...conv,
          last_message: message,
          updated_at: message.timestamp,
        };
      }
      return conv;
    });
    return { conversations: updated };
  }),
  
  markAsRead: (conversationId) => set((state) => {
    const updated = state.conversations.map(conv => {
      if (conv.id === conversationId) {
        const newUnreadCount = { ...conv.unread_count };
        // Reset unread count for current user (you'd get this from auth context)
        newUnreadCount['current_user'] = 0;
        
        return {
          ...conv,
          unread_count: newUnreadCount,
        };
      }
      return conv;
    });
    return { conversations: updated };
  }),

  // Realtime actions
  handleRealtimeMessage: (message) => {
    console.log('[ConversationStore] Handling realtime message:', message);
    
    switch (message.type) {
      case 'new_message':
        // Add message to conversation
        get().addMessage(message.conversation_id, message.message);
        break;
        
      case 'conversation_update':
        // Update conversation
        if (message.conversation) {
          get().updateConversation(message.conversation);
        }
        break;
        
      case 'user_online':
      case 'user_offline':
        // Handle user status updates
        console.log('[ConversationStore] User status update:', message);
        break;
    }
  },
  
  initializeRealtime: () => {
    console.log('[ConversationStore] Initializing realtime service');
    
    // Add message handler
    realtimeService.onMessage(get().handleRealtimeMessage);
    
    // Connect to WebSocket
    realtimeService.connect().catch(error => {
      console.error('[ConversationStore] Failed to connect to realtime service:', error);
      get().setError('Failed to connect to realtime service');
    });
  },
  
  cleanupRealtime: () => {
    console.log('[ConversationStore] Cleaning up realtime service');
    
    // Remove message handler
    realtimeService.removeMessageHandler(get().handleRealtimeMessage);
    
    // Disconnect
    realtimeService.disconnect();
  },
}));

// Helper hooks for specific functionality
export const useConversations = () => {
  const conversations = useConversationStore(state => state.conversations);
  const loading = useConversationStore(state => state.loading);
  const error = useConversationStore(state => state.error);
  
  return { conversations, loading, error };
};

export const useCurrentConversation = () => {
  const currentConversation = useConversationStore(state => state.currentConversation);
  const setCurrentConversation = useConversationStore(state => state.setCurrentConversation);
  
  return { currentConversation, setCurrentConversation };
};

export const useRealtimeConnection = () => {
  const initializeRealtime = useConversationStore(state => state.initializeRealtime);
  const cleanupRealtime = useConversationStore(state => state.cleanupRealtime);
  
  return { initializeRealtime, cleanupRealtime };
};
