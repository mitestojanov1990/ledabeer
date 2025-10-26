/**
 * Ledabeer Mobile App
 *
 * Main entry point for the React Native application.
 * Provides E2E encrypted P2P chat with voice/video calling.
 */

import React, { useState, useEffect } from 'react';
import { StatusBar } from 'expo-status-bar';
import { StyleSheet, View, Text, ActivityIndicator } from 'react-native';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { ChatListScreen } from './src/screens/ChatListScreen';
import { ChatConversationScreen } from './src/screens/ChatConversationScreen';
import { initializeBackend } from './src/services/backend';

export default function App() {
  const [selectedPeerId, setSelectedPeerId] = useState<string | null>(null);
  const [backendReady, setBackendReady] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);

  // Initialize backend connection on app start
  useEffect(() => {
    const init = async () => {
      try {
        await initializeBackend();
        setBackendReady(true);
      } catch (error) {
        console.error('[App] Backend initialization failed:', error);
        setBackendError(error instanceof Error ? error.message : 'Unknown error');
        // Continue with mock backend
        setBackendReady(true);
      }
    };
    init();
  }, []);

  // Show loading screen while initializing
  if (!backendReady) {
    return (
      <SafeAreaProvider>
        <GestureHandlerRootView style={styles.container}>
          <View style={styles.loadingContainer}>
            <ActivityIndicator size="large" color="#3B82F6" />
            <Text style={styles.loadingText}>Connecting to backend...</Text>
          </View>
          <StatusBar style="light" />
        </GestureHandlerRootView>
      </SafeAreaProvider>
    );
  }

  return (
    <SafeAreaProvider>
      <GestureHandlerRootView style={styles.container}>
        {backendError && (
          <View style={styles.errorBanner}>
            <Text style={styles.errorText}>Using mock backend (real backend unavailable)</Text>
          </View>
        )}
        {selectedPeerId ? (
          <ChatConversationScreen
            peerId={selectedPeerId}
            onBack={() => setSelectedPeerId(null)}
          />
        ) : (
          <ChatListScreen onSelectChat={setSelectedPeerId} />
        )}
        <StatusBar style="light" />
      </GestureHandlerRootView>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#111827',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#111827',
  },
  loadingText: {
    color: '#9CA3AF',
    fontSize: 16,
    marginTop: 16,
  },
  errorBanner: {
    backgroundColor: '#FEF2F2',
    paddingVertical: 8,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#FCA5A5',
  },
  errorText: {
    color: '#991B1B',
    fontSize: 12,
    textAlign: 'center',
  },
});
