/**
 * Protobuf Message Helpers
 *
 * Manual protobuf encoding/decoding for gRPC-Web messages
 * Based on message.proto, call.proto, and media.proto
 */

/**
 * Encode a SendMessageRequest
 */
export function encodeSendMessageRequest(toPeerId: string, content: string): Uint8Array {
  const encoder = new TextEncoder();
  const contentBytes = encoder.encode(content);

  // Simple protobuf encoding
  // Field 1 (to_peer_id): tag = 0x0A (field 1, type string)
  // Field 2 (content): tag = 0x12 (field 2, type bytes)

  const toPeerIdBytes = encoder.encode(toPeerId);
  const output: number[] = [];

  // Field 1: to_peer_id (string)
  output.push(0x0A); // Tag: field 1, wire type 2 (length-delimited)
  output.push(toPeerIdBytes.length); // Length
  output.push(...Array.from(toPeerIdBytes)); // Data

  // Field 2: content (bytes)
  output.push(0x12); // Tag: field 2, wire type 2 (length-delimited)
  output.push(contentBytes.length); // Length
  output.push(...Array.from(contentBytes)); // Data

  return new Uint8Array(output);
}

/**
 * Encode a SendGroupMessageRequest
 */
export function encodeSendGroupMessageRequest(groupId: string, content: string): Uint8Array {
  const encoder = new TextEncoder();
  const contentBytes = encoder.encode(content);

  // Field 1 (group_id): tag = 0x0A
  // Field 2 (content): tag = 0x12

  const groupIdBytes = encoder.encode(groupId);
  const output: number[] = [];

  // Field 1: group_id (string)
  output.push(0x0A);
  output.push(groupIdBytes.length);
  output.push(...Array.from(groupIdBytes));

  // Field 2: content (bytes)
  output.push(0x12);
  output.push(contentBytes.length);
  output.push(...Array.from(contentBytes));

  return new Uint8Array(output);
}

/**
 * Encode GetMessageHistoryRequest
 */
export function encodeGetMessageHistoryRequest(peerId: string, limit: number): Uint8Array {
  const encoder = new TextEncoder();
  const peerIdBytes = encoder.encode(peerId);
  const output: number[] = [];

  // Field 1: peer_id (string)
  output.push(0x0A);
  output.push(peerIdBytes.length);
  output.push(...Array.from(peerIdBytes));

  // Field 2: limit (int32)
  if (limit > 0) {
    output.push(0x10); // Tag: field 2, wire type 0 (varint)
    output.push(limit & 0x7F); // Simple varint encoding (works for values < 128)
  }

  return new Uint8Array(output);
}

/**
 * Encode ReceiveMessagesRequest (empty)
 */
export function encodeReceiveMessagesRequest(): Uint8Array {
  // Empty message
  return new Uint8Array(0);
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
        value |= (byte & 0x7F) << shift;
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
        value |= (byte & 0x7F) << shift;
        if ((byte & 0x80) === 0) break;
        shift += 7;
      }

      if (fieldNumber === 4) {
        // timestamp
        timestamp = value;
      }
    }
  }

  return { message_id: messageId, from_peer_id: fromPeerId, content, timestamp };
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
