/**
 * Backend Initialization
 *
 * Initialize connection to the Ledabeer backend on app startup
 */

import { getHttpClient } from './httpBackend/httpClient';

let isInitialized = false;
let initializationPromise: Promise<void> | null = null;

/**
 * Initialize backend connection
 */
export async function initializeBackend(): Promise<void> {
  if (isInitialized) {
    console.log('[BackendInit] Backend already initialized');
    return;
  }

  if (initializationPromise) {
    console.log('[BackendInit] Waiting for existing initialization');
    return initializationPromise;
  }

  initializationPromise = (async () => {
    try {
      console.log('[BackendInit] Initializing backend connection...');
      const client = getHttpClient();
      await client.connect();
      isInitialized = true;
      console.log('[BackendInit] Backend connected successfully');
    } catch (error) {
      console.error('[BackendInit] Failed to initialize backend:', error);
      throw error;
    } finally {
      initializationPromise = null;
    }
  })();

  return initializationPromise;
}

/**
 * Check if backend is initialized
 */
export function isBackendInitialized(): boolean {
  return isInitialized;
}
