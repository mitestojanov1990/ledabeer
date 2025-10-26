/**
 * Call Service
 *
 * gRPC wrapper for CallService operations (WebRTC signaling)
 */

import { getBackendClient } from './backendClient';

export type CallStateEnum = 'UNKNOWN' | 'INITIATING' | 'RINGING' | 'CONNECTED' | 'ENDED';

export interface CallState {
  call_id: string;
  state: CallStateEnum;
  participants: string[];
}

export interface InitiateCallResponse {
  call_id: string;
  state: CallState;
}

export interface AnswerCallResponse {
  state: CallState;
}

export interface SignalingMessage {
  call_id: string;
  type: 'offer' | 'answer' | 'ice_candidate';
  sdp?: string;
  candidate?: string;
}

export class CallService {
  /**
   * Initiate a call to a peer
   */
  async initiateCall(
    toPeerId: string,
    audioEnabled: boolean,
    videoEnabled: boolean
  ): Promise<InitiateCallResponse> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getCallClient();

      const request = {
        to_peer_id: toPeerId,
        audio_enabled: audioEnabled,
        video_enabled: videoEnabled,
      };

      (client as any).InitiateCall(request, (error: any, response: any) => {
        if (error) {
          console.error('[CallService] InitiateCall error:', error);
          reject(error);
        } else {
          resolve({
            call_id: response.call_id,
            state: {
              call_id: response.state.call_id,
              state: this.mapCallState(response.state.state),
              participants: response.state.participants || [],
            },
          });
        }
      });
    });
  }

  /**
   * Answer an incoming call
   */
  async answerCall(callId: string, accept: boolean): Promise<AnswerCallResponse> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getCallClient();

      const request = {
        call_id: callId,
        accept: accept,
      };

      (client as any).AnswerCall(request, (error: any, response: any) => {
        if (error) {
          console.error('[CallService] AnswerCall error:', error);
          reject(error);
        } else {
          resolve({
            state: {
              call_id: response.state.call_id,
              state: this.mapCallState(response.state.state),
              participants: response.state.participants || [],
            },
          });
        }
      });
    });
  }

  /**
   * End a call
   */
  async endCall(callId: string): Promise<boolean> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getCallClient();

      const request = { call_id: callId };

      (client as any).EndCall(request, (error: any, response: any) => {
        if (error) {
          console.error('[CallService] EndCall error:', error);
          reject(error);
        } else {
          resolve(response.success);
        }
      });
    });
  }

  /**
   * Get current call state
   */
  async getCallState(callId: string): Promise<CallState> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getCallClient();

      const request = { call_id: callId };

      (client as any).GetCallState(request, (error: any, response: any) => {
        if (error) {
          console.error('[CallService] GetCallState error:', error);
          reject(error);
        } else {
          resolve({
            call_id: response.call_id,
            state: this.mapCallState(response.state),
            participants: response.participants || [],
          });
        }
      });
    });
  }

  /**
   * Stream signaling messages (bidirectional)
   */
  streamSignaling(
    onMessage: (message: SignalingMessage) => void,
    onError?: (error: Error) => void
  ): (message: SignalingMessage) => void {
    const client = getBackendClient().getCallClient();
    const stream = (client as any).StreamSignaling();

    // Handle incoming messages
    stream.on('data', (message: any) => {
      onMessage({
        call_id: message.call_id,
        type: message.type as 'offer' | 'answer' | 'ice_candidate',
        sdp: message.sdp,
        candidate: message.candidate,
      });
    });

    stream.on('error', (error: any) => {
      console.error('[CallService] StreamSignaling error:', error);
      if (onError) {
        onError(error);
      }
    });

    stream.on('end', () => {
      console.log('[CallService] StreamSignaling ended');
    });

    // Return function to send messages
    return (message: SignalingMessage) => {
      stream.write({
        call_id: message.call_id,
        type: message.type,
        sdp: message.sdp || '',
        candidate: message.candidate || '',
      });
    };
  }

  /**
   * Map backend call state enum to frontend
   */
  private mapCallState(backendState: number): CallStateEnum {
    const stateMap: { [key: number]: CallStateEnum } = {
      0: 'UNKNOWN',
      1: 'INITIATING',
      2: 'RINGING',
      3: 'CONNECTED',
      4: 'ENDED',
    };
    return stateMap[backendState] || 'UNKNOWN';
  }
}

// Export singleton instance
export const callService = new CallService();
