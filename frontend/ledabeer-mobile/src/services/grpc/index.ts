/**
 * gRPC Services Index
 *
 * Exports all gRPC services for easy import
 */

export { BackendClient, getBackendClient } from './backendClient';
export { messageService } from './messageService';
export { mediaService } from './mediaService';
export { callService } from './callService';

export type { Message, SendMessageResponse } from './messageService';
export type { UploadMediaResponse, MediaInfo, SendMediaMessageResponse } from './mediaService';
export type { CallState, CallStateEnum, InitiateCallResponse, AnswerCallResponse, SignalingMessage } from './callService';
