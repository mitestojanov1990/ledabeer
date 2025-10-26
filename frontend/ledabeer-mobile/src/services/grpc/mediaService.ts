/**
 * Media Service
 *
 * gRPC wrapper for MediaService operations
 */

import { getBackendClient } from './backendClient';

export interface UploadMediaResponse {
  media_id: string;
  cid: string;
  size: number;
}

export interface MediaInfo {
  cid: string;
  mime_type: string;
  size: number;
  timestamp: number;
}

export interface SendMediaMessageResponse {
  message_id: string;
  timestamp: number;
}

const CHUNK_SIZE = 64 * 1024; // 64KB chunks

export class MediaService {
  /**
   * Upload media file to IPFS
   */
  async uploadMedia(
    fileUri: string,
    fileData: Uint8Array,
    onProgress?: (progress: number) => void
  ): Promise<UploadMediaResponse> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getMediaClient();
      const stream = (client as any).UploadMedia((error: any, response: any) => {
        if (error) {
          console.error('[MediaService] UploadMedia error:', error);
          reject(error);
        } else {
          resolve({
            media_id: response.media_id,
            cid: response.cid,
            size: parseInt(response.size),
          });
        }
      });

      // Split file into chunks
      const totalChunks = Math.ceil(fileData.length / CHUNK_SIZE);
      let uploadedChunks = 0;

      for (let i = 0; i < totalChunks; i++) {
        const start = i * CHUNK_SIZE;
        const end = Math.min(start + CHUNK_SIZE, fileData.length);
        const chunk = fileData.slice(start, end);

        stream.write({
          data: chunk,
          chunk_index: i,
          total_chunks: totalChunks,
        });

        uploadedChunks++;
        if (onProgress) {
          onProgress((uploadedChunks / totalChunks) * 100);
        }
      }

      stream.end();
    });
  }

  /**
   * Download media file from IPFS
   */
  async downloadMedia(
    cid: string,
    onProgress?: (progress: number) => void
  ): Promise<Uint8Array> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getMediaClient();

      const request = { cid };
      const stream = (client as any).DownloadMedia(request);

      const chunks: Uint8Array[] = [];
      let totalChunks = 0;
      let receivedChunks = 0;

      stream.on('data', (chunk: any) => {
        if (totalChunks === 0) {
          totalChunks = chunk.total_chunks;
        }
        chunks.push(chunk.data);
        receivedChunks++;

        if (onProgress && totalChunks > 0) {
          onProgress((receivedChunks / totalChunks) * 100);
        }
      });

      stream.on('end', () => {
        // Combine all chunks
        const totalSize = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
        const result = new Uint8Array(totalSize);
        let offset = 0;

        for (const chunk of chunks) {
          result.set(chunk, offset);
          offset += chunk.length;
        }

        resolve(result);
      });

      stream.on('error', (error: any) => {
        console.error('[MediaService] DownloadMedia error:', error);
        reject(error);
      });
    });
  }

  /**
   * Get media information
   */
  async getMediaInfo(cid: string): Promise<MediaInfo> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getMediaClient();

      const request = { cid };

      (client as any).GetMediaInfo(request, (error: any, response: any) => {
        if (error) {
          console.error('[MediaService] GetMediaInfo error:', error);
          reject(error);
        } else {
          resolve({
            cid: response.cid,
            mime_type: response.mime_type,
            size: parseInt(response.size),
            timestamp: parseInt(response.timestamp),
          });
        }
      });
    });
  }

  /**
   * Send media message to a peer
   */
  async sendMediaMessage(
    toPeerId: string,
    mediaCid: string,
    mimeType: string
  ): Promise<SendMediaMessageResponse> {
    return new Promise((resolve, reject) => {
      const client = getBackendClient().getMediaClient();

      const request = {
        to_peer_id: toPeerId,
        media_cid: mediaCid,
        mime_type: mimeType,
      };

      (client as any).SendMediaMessage(request, (error: any, response: any) => {
        if (error) {
          console.error('[MediaService] SendMediaMessage error:', error);
          reject(error);
        } else {
          resolve({
            message_id: response.message_id,
            timestamp: parseInt(response.timestamp),
          });
        }
      });
    });
  }
}

// Export singleton instance
export const mediaService = new MediaService();
