/**
 * Protobuf Message Helpers
 *
 * Manual protobuf encoding/decoding for gRPC-Web messages
 * Based on message.proto, call.proto, media.proto, and peer.proto
 */

/**
 * Helper function to wrap protobuf data in gRPC frame
 */
function wrapInGrpcFrame(protobufData: Uint8Array): Uint8Array {
  const length = protobufData.length;
  const frame = new Uint8Array(5 + length);

  // gRPC frame header: [compression_flag: 1 byte][message_length: 4 bytes]
  frame[0] = 0x00; // No compression
  frame[1] = (length >> 24) & 0xff; // Length (big-endian)
  frame[2] = (length >> 16) & 0xff;
  frame[3] = (length >> 8) & 0xff;
  frame[4] = length & 0xff;

  // Copy protobuf data
  frame.set(protobufData, 5);

  return frame;
}

/**
 * Encode a SendMessageRequest
 */
export function encodeSendMessageRequest(
  toPeerId: string,
  content: string
): Uint8Array {
  const encoder = new TextEncoder();
  const contentBytes = encoder.encode(content);

  // Simple protobuf encoding
  // Field 1 (to_peer_id): tag = 0x0A (field 1, type string)
  // Field 2 (content): tag = 0x12 (field 2, type bytes)

  const toPeerIdBytes = encoder.encode(toPeerId);
  const output: number[] = [];

  // Field 1: to_peer_id (string)
  output.push(0x0a); // Tag: field 1, wire type 2 (length-delimited)
  output.push(toPeerIdBytes.length); // Length
  output.push(...Array.from(toPeerIdBytes)); // Data

  // Field 2: content (bytes)
  output.push(0x12); // Tag: field 2, wire type 2 (length-delimited)
  output.push(contentBytes.length); // Length
  output.push(...Array.from(contentBytes)); // Data

  const protobufData = new Uint8Array(output);
  return wrapInGrpcFrame(protobufData);
}

/**
 * Encode a SendGroupMessageRequest
 */
export function encodeSendGroupMessageRequest(
  groupId: string,
  content: string
): Uint8Array {
  const encoder = new TextEncoder();
  const contentBytes = encoder.encode(content);

  // Field 1 (group_id): tag = 0x0A
  // Field 2 (content): tag = 0x12

  const groupIdBytes = encoder.encode(groupId);
  const output: number[] = [];

  // Field 1: group_id (string)
  output.push(0x0a);
  output.push(groupIdBytes.length);
  output.push(...Array.from(groupIdBytes));

  // Field 2: content (bytes)
  output.push(0x12);
  output.push(contentBytes.length);
  output.push(...Array.from(contentBytes));

  const protobufData = new Uint8Array(output);
  return wrapInGrpcFrame(protobufData);
}

/**
 * Encode GetMessageHistoryRequest
 */
export function encodeGetMessageHistoryRequest(
  peerId: string,
  limit: number
): Uint8Array {
  const encoder = new TextEncoder();
  const peerIdBytes = encoder.encode(peerId);
  const output: number[] = [];

  // Field 1: peer_id (string)
  output.push(0x0a);
  output.push(peerIdBytes.length);
  output.push(...Array.from(peerIdBytes));

  // Field 2: limit (int32)
  if (limit > 0) {
    output.push(0x10); // Tag: field 2, wire type 0 (varint)
    output.push(limit & 0x7f); // Simple varint encoding (works for values < 128)
  }

  const protobufData = new Uint8Array(output);
  return wrapInGrpcFrame(protobufData);
}

/**
 * Encode ReceiveMessagesRequest (empty)
 */
export function encodeReceiveMessagesRequest(): Uint8Array {
  // gRPC-Web requires a proper gRPC frame header even for empty messages
  return new Uint8Array([0x00, 0x00, 0x00, 0x00, 0x00]);
}

/**
 * Decode SendMessageResponse
 */
export function decodeSendMessageResponse(data: Uint8Array): {
  message_id: string;
  timestamp: number;
} {
  const decoder = new TextDecoder();
  let messageId = '';
  let timestamp = 0;
  let i = 0;

  while (i < data.length) {
    const tag = data[i++];
    const fieldNumber = tag >> 3;
    const wireType = tag & 0x07;

    if (wireType === 2) {
      // Length-delimited
      const length = data[i++];
      const fieldData = data.slice(i, i + length);
      i += length;

      if (fieldNumber === 1) {
        // message_id (string)
        messageId = decoder.decode(fieldData);
      }
    } else if (wireType === 0) {
      // Varint
      let value = 0;
      let shift = 0;
      while (i < data.length) {
        const byte = data[i++];
        value |= (byte & 0x7f) << shift;
        if ((byte & 0x80) === 0) break;
        shift += 7;
      }

      if (fieldNumber === 2) {
        // timestamp (int64)
        timestamp = value;
      }
    }
  }

  return { message_id: messageId, timestamp };
}

/**
 * Decode Message
 */
export function decodeMessage(data: Uint8Array): {
  message_id: string;
  from_peer_id: string;
  content: Uint8Array;
  timestamp: number;
} {
  const decoder = new TextDecoder();
  let messageId = '';
  let fromPeerId = '';
  let content = new Uint8Array(0);
  let timestamp = 0;
  let i = 0;

  while (i < data.length) {
    const tag = data[i++];
    const fieldNumber = tag >> 3;
    const wireType = tag & 0x07;

    if (wireType === 2) {
      // Length-delimited
      const length = data[i++];
      const fieldData = data.slice(i, i + length);
      i += length;

      if (fieldNumber === 1) {
        // message_id
        messageId = decoder.decode(fieldData);
      } else if (fieldNumber === 2) {
        // from_peer_id
        fromPeerId = decoder.decode(fieldData);
      } else if (fieldNumber === 3) {
        // content (bytes)
        content = fieldData;
      }
    } else if (wireType === 0) {
      // Varint
      let value = 0;
      let shift = 0;
      while (i < data.length) {
        const byte = data[i++];
        value |= (byte & 0x7f) << shift;
        if ((byte & 0x80) === 0) break;
        shift += 7;
      }

      if (fieldNumber === 4) {
        // timestamp
        timestamp = value;
      }
    }
  }

  return {
    message_id: messageId,
    from_peer_id: fromPeerId,
    content,
    timestamp,
  };
}

/**
 * Decode MessageHistoryResponse
 */
export function decodeMessageHistoryResponse(data: Uint8Array): Array<{
  message_id: string;
  from_peer_id: string;
  content: Uint8Array;
  timestamp: number;
}> {
  const messages: Array<{
    message_id: string;
    from_peer_id: string;
    content: Uint8Array;
    timestamp: number;
  }> = [];

  let i = 0;

  while (i < data.length) {
    const tag = data[i++];
    const fieldNumber = tag >> 3;
    const wireType = tag & 0x07;

    if (wireType === 2 && fieldNumber === 1) {
      // messages field (repeated Message)
      const length = data[i++];
      const messageData = data.slice(i, i + length);
      i += length;

      const message = decodeMessage(messageData);
      messages.push(message);
    }
  }

  return messages;
}

// ============================================================================
// Peer Discovery Messages
// ============================================================================

export interface Peer {
  id: string;
  name: string;
  publicKey: string;
  online: boolean;
  lastSeen: number;
  addresses: string[];
}

export interface GetPeersResponse {
  peers: Peer[];
}

export interface GetPeerResponse {
  peer: Peer | null;
  found: boolean;
}

/**
 * Encode a GetPeersRequest (empty message)
 */
export function encodeGetPeersRequest(): Uint8Array {
  // gRPC-Web requires a proper gRPC frame header even for empty messages
  // Format: [compression_flag: 1 byte][message_length: 4 bytes][message_data: N bytes]
  // For empty message: [0x00][0x00, 0x00, 0x00, 0x00]
  return new Uint8Array([0x00, 0x00, 0x00, 0x00, 0x00]);
}

/**
 * Encode a GetPeerRequest
 */
export function encodeGetPeerRequest(peerId: string): Uint8Array {
  const encoder = new TextEncoder();
  const peerIdBytes = encoder.encode(peerId);
  const output: number[] = [];

  // Field 1: peer_id (string)
  output.push(0x0a); // Tag: field 1, wire type 2 (length-delimited)
  output.push(peerIdBytes.length); // Length
  output.push(...Array.from(peerIdBytes)); // Data

  const protobufData = new Uint8Array(output);
  return wrapInGrpcFrame(protobufData);
}

/**
 * Encode a GetConnectedPeersRequest (empty message)
 */
export function encodeGetConnectedPeersRequest(): Uint8Array {
  // gRPC-Web requires a proper gRPC frame header even for empty messages
  return new Uint8Array([0x00, 0x00, 0x00, 0x00, 0x00]);
}

/**
 * Decode a GetPeersResponse
 */
export function decodeGetPeersResponse(data: Uint8Array): GetPeersResponse {
  const peers: Peer[] = [];
  let i = 0;

  while (i < data.length) {
    const tag = data[i++];
    const fieldNumber = tag >> 3;
    const wireType = tag & 0x07;

    if (wireType === 2 && fieldNumber === 1) {
      // peers field (repeated Peer)
      const length = data[i++];
      const peerData = data.slice(i, i + length);
      i += length;

      const peer = decodePeer(peerData);
      peers.push(peer);
    }
  }

  return { peers };
}

/**
 * Decode a GetPeerResponse
 */
export function decodeGetPeerResponse(data: Uint8Array): GetPeerResponse {
  let peer: Peer | null = null;
  let found = false;
  let i = 0;

  while (i < data.length) {
    const tag = data[i++];
    const fieldNumber = tag >> 3;
    const wireType = tag & 0x07;

    if (wireType === 2 && fieldNumber === 1) {
      // peer field (Peer)
      const length = data[i++];
      const peerData = data.slice(i, i + length);
      i += length;

      peer = decodePeer(peerData);
    } else if (wireType === 0 && fieldNumber === 2) {
      // found field (bool)
      found = data[i++] !== 0;
    }
  }

  return { peer, found };
}

/**
 * Decode a GetConnectedPeersResponse
 */
export function decodeGetConnectedPeersResponse(
  data: Uint8Array
): GetPeersResponse {
  // Same as GetPeersResponse
  return decodeGetPeersResponse(data);
}

/**
 * Decode a Peer message
 */
function decodePeer(data: Uint8Array): Peer {
  let id = '';
  let name = '';
  let online = false;
  let lastSeen = 0;
  const addresses: string[] = [];
  let i = 0;

  while (i < data.length) {
    const tag = data[i++];
    const fieldNumber = tag >> 3;
    const wireType = tag & 0x07;

    if (wireType === 2 && fieldNumber === 1) {
      // id field (string)
      const length = data[i++];
      const idBytes = data.slice(i, i + length);
      i += length;
      id = new TextDecoder().decode(idBytes);
    } else if (wireType === 2 && fieldNumber === 2) {
      // name field (string)
      const length = data[i++];
      const nameBytes = data.slice(i, i + length);
      i += length;
      name = new TextDecoder().decode(nameBytes);
    } else if (wireType === 0 && fieldNumber === 3) {
      // online field (bool)
      online = data[i++] !== 0;
    } else if (wireType === 0 && fieldNumber === 4) {
      // last_seen field (int64)
      let value = 0;
      let shift = 0;
      let byte = data[i++];
      while (byte & 0x80) {
        value |= (byte & 0x7f) << shift;
        shift += 7;
        byte = data[i++];
      }
      value |= byte << shift;
      lastSeen = value;
    } else if (wireType === 2 && fieldNumber === 5) {
      // addresses field (repeated string)
      const length = data[i++];
      const addressBytes = data.slice(i, i + length);
      i += length;
      const address = new TextDecoder().decode(addressBytes);
      addresses.push(address);
    }
  }

  return { id, name, online, lastSeen, addresses, publicKey: '' };
}
