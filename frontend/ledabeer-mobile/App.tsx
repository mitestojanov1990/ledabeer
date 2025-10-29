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
import { ConversationListScreen } from './src/screens/ConversationListScreen';
import { ConversationDetailScreen } from './src/screens/ConversationDetailScreen';
import { AuthScreen } from './src/screens/AuthScreen';
import { AuthProvider, useAuth } from './src/contexts/AuthContext';
import { initializeBackend } from './src/services/backend';

function AppContent() {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const [selectedConversation, setSelectedConversation] = useState<any>(null);
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
        setBackendError(
          error instanceof Error ? error.message : 'Unknown error'
        );
        // Continue with mock backend
        setBackendReady(true);
      }
    };
    init();
  }, []);

  // Show loading screen while initializing
  if (!backendReady || authLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size='large' color='#3B82F6' />
        <Text style={styles.loadingText}>
          {authLoading
            ? 'Checking authentication...'
            : 'Connecting to backend...'}
        </Text>
      </View>
    );
  }

  // Show authentication screen if not authenticated
  if (!isAuthenticated) {
    return <AuthScreen onAuthSuccess={() => {}} />;
  }

  // Show main app if authenticated
  return (
    <>
      {backendError && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorText}>
            Using mock backend (real backend unavailable)
          </Text>
        </View>
      )}
      {selectedConversation ? (
        <ConversationDetailScreen
          conversation={selectedConversation}
          onBack={() => setSelectedConversation(null)}
        />
      ) : (
        <ConversationListScreen onSelectConversation={setSelectedConversation} />
      )}
    </>
  );
}

export default function App() {
  return (
    <SafeAreaProvider>
      <GestureHandlerRootView style={styles.container}>
        <AuthProvider>
          <AppContent />
        </AuthProvider>
        <StatusBar style='light' />
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
