/**
 * Chat Conversation Screen
 *
 * Displays messages with a specific peer and allows sending new messages.
 */

import React, { useState, useEffect, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TextInput,
  TouchableOpacity,
  FlatList,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useChatStore } from '../store/chatStore';
import { Message, mockBackend } from '../services/mockBackend';

interface ChatConversationScreenProps {
  peerId: string;
  onBack: () => void;
}

export function ChatConversationScreen({ peerId, onBack }: ChatConversationScreenProps) {
  const { conversations, messages, sendMessage, getConversation } = useChatStore();
  const [inputText, setInputText] = useState('');
  const flatListRef = useRef<FlatList>(null);

  const currentUserId = mockBackend.getCurrentUserId();
  const conversation = messages[peerId] || [];
  const peer = getConversation(peerId);

  useEffect(() => {
    // Scroll to bottom when messages change
    if (conversation.length > 0) {
      setTimeout(() => {
        flatListRef.current?.scrollToEnd({ animated: true });
      }, 100);
    }
  }, [conversation.length]);

  const handleSend = async () => {
    if (inputText.trim().length === 0) return;

    const messageContent = inputText.trim();
    setInputText('');

    await sendMessage(peerId, messageContent);

    // Scroll to bottom after sending
    setTimeout(() => {
      flatListRef.current?.scrollToEnd({ animated: true });
    }, 100);
  };

  const formatMessageTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
  };

  const renderMessage = ({ item }: { item: Message }) => {
    const isOwnMessage = item.from === currentUserId;

    return (
      <View
        style={[
          styles.messageContainer,
          isOwnMessage ? styles.ownMessageContainer : styles.otherMessageContainer,
        ]}
      >
        <View
          style={[
            styles.messageBubble,
            isOwnMessage ? styles.ownMessageBubble : styles.otherMessageBubble,
          ]}
        >
          <Text style={styles.messageText}>{item.content}</Text>
          <View style={styles.messageFooter}>
            <Text style={styles.messageTime}>{formatMessageTime(item.timestamp)}</Text>
            {item.encrypted && <Text style={styles.encryptedIcon}>🔒</Text>}
          </View>
        </View>
      </View>
    );
  };

  if (!peer) {
    return (
      <View style={styles.container}>
        <Text style={styles.errorText}>Peer not found</Text>
      </View>
    );
  }

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      keyboardVerticalOffset={Platform.OS === 'ios' ? 90 : 0}
    >
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={onBack} style={styles.backButton}>
          <Text style={styles.backButtonText}>←</Text>
        </TouchableOpacity>

        <View style={styles.headerInfo}>
          <Text style={styles.headerTitle}>{peer.name}</Text>
          <View style={styles.statusContainer}>
            <View style={[styles.statusDot, peer.online && styles.statusDotOnline]} />
            <Text style={styles.statusText}>
              {peer.online ? 'Online' : 'Offline'}
            </Text>
            <Text style={styles.encryptionBadge}>🔒 E2EE</Text>
          </View>
        </View>

        {/* Call Buttons */}
        <View style={styles.headerActions}>
          <TouchableOpacity
            style={styles.callButton}
            onPress={() => console.log('Voice call initiated')}
          >
            <Text style={styles.callButtonText}>📞</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={styles.callButton}
            onPress={() => console.log('Video call initiated')}
          >
            <Text style={styles.callButtonText}>📹</Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Messages List */}
      <FlatList
        ref={flatListRef}
        data={conversation}
        renderItem={renderMessage}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.messagesContainer}
        onContentSizeChange={() => flatListRef.current?.scrollToEnd({ animated: false })}
      />

      {/* Input Area */}
      <View style={styles.inputContainer}>
        <TextInput
          style={styles.input}
          value={inputText}
          onChangeText={setInputText}
          placeholder="Type a message..."
          placeholderTextColor="#6B7280"
          multiline
          maxLength={1000}
        />
        <TouchableOpacity
          style={[styles.sendButton, inputText.trim().length === 0 && styles.sendButtonDisabled]}
          onPress={handleSend}
          disabled={inputText.trim().length === 0}
        >
          <Text style={styles.sendButtonText}>Send</Text>
        </TouchableOpacity>
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#111827',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingTop: 60,
    paddingBottom: 12,
    paddingHorizontal: 16,
    backgroundColor: '#1F2937',
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  backButton: {
    marginRight: 12,
    padding: 4,
  },
  backButtonText: {
    fontSize: 28,
    color: '#3B82F6',
    fontWeight: 'bold',
  },
  headerInfo: {
    flex: 1,
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#F9FAFB',
    marginBottom: 2,
  },
  statusContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statusDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: '#6B7280',
    marginRight: 6,
  },
  statusDotOnline: {
    backgroundColor: '#10B981',
  },
  statusText: {
    fontSize: 12,
    color: '#9CA3AF',
    marginRight: 8,
  },
  encryptionBadge: {
    fontSize: 11,
    color: '#10B981',
  },
  headerActions: {
    flexDirection: 'row',
    alignItems: 'center',
    marginLeft: 8,
  },
  callButton: {
    padding: 8,
    marginLeft: 4,
  },
  callButtonText: {
    fontSize: 22,
  },
  messagesContainer: {
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  messageContainer: {
    marginBottom: 12,
    maxWidth: '75%',
  },
  ownMessageContainer: {
    alignSelf: 'flex-end',
  },
  otherMessageContainer: {
    alignSelf: 'flex-start',
  },
  messageBubble: {
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 16,
  },
  ownMessageBubble: {
    backgroundColor: '#3B82F6',
    borderBottomRightRadius: 4,
  },
  otherMessageBubble: {
    backgroundColor: '#374151',
    borderBottomLeftRadius: 4,
  },
  messageText: {
    fontSize: 15,
    color: '#FFFFFF',
    marginBottom: 4,
  },
  messageFooter: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'flex-end',
  },
  messageTime: {
    fontSize: 11,
    color: 'rgba(255, 255, 255, 0.7)',
    marginRight: 4,
  },
  encryptedIcon: {
    fontSize: 10,
  },
  inputContainer: {
    flexDirection: 'row',
    padding: 12,
    backgroundColor: '#1F2937',
    borderTopWidth: 1,
    borderTopColor: '#374151',
    alignItems: 'flex-end',
  },
  input: {
    flex: 1,
    backgroundColor: '#374151',
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 10,
    marginRight: 8,
    color: '#F9FAFB',
    fontSize: 15,
    maxHeight: 100,
  },
  sendButton: {
    backgroundColor: '#3B82F6',
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 20,
    justifyContent: 'center',
  },
  sendButtonDisabled: {
    backgroundColor: '#374151',
  },
  sendButtonText: {
    color: '#FFFFFF',
    fontSize: 15,
    fontWeight: '600',
  },
  errorText: {
    color: '#EF4444',
    fontSize: 16,
    textAlign: 'center',
    marginTop: 100,
  },
});
