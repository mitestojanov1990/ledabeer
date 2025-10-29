# gRPC-Web Message Routing Issue

## Problem Summary

Messages are being sent successfully via the gRPC `SendMessage` endpoint, but they are not being received by the gRPC-Web streaming clients.

## Root Cause

The issue is an **architectural routing problem**:

1. **gRPC-Web clients** (frontend) connect to the **bootstrap node** via Envoy proxy
2. The `ReceiveMessages` streaming endpoint subscribes to messages on the **bootstrap node**
3. When `SendMessage` is called, it sends messages directly between **peer nodes** (alice ↔ bob) via libp2p
4. Messages arrive at alice/bob nodes, which forward to their **local subscribers** (0 subscribers)
5. The **bootstrap node** has all the subscribers (4 subscribers) but never receives the messages

### Evidence from Docker Logs

**Bootstrap node:**
```
📡 ReceiveMessages gRPC method called
✅ Message handler available, subscribing to messages
📡 New subscriber added, total subscribers: 4
```
- ✅ gRPC streaming clients ARE connected
- ✅ Subscribers ARE added
- ❌ NO messages received via handleStream

**Alice/Bob nodes:**
```
📨 handleStream called from peer: 12D3KooW...
📡 Forwarding message to 0 subscribers
```
- ✅ Messages ARE being received
- ❌ 0 subscribers to forward to

## Architecture Diagram

```
Frontend (Browser)
      ↓ gRPC-Web
   Envoy Proxy
      ↓ gRPC
Bootstrap Node (subscribers: 4) ← NO MESSAGES ARRIVE HERE!
      ↑
      | (libp2p peer network)
      |
Alice Node ←──libp2p──→ Bob Node
(subscribers: 0)      (subscribers: 0)
      ↑                   ↑
   Messages            Messages
   arrive here!        arrive here!
```

## Solution Options

### Option 1: Route All Messages Through Bootstrap Node (Recommended)

**Approach**: Make the bootstrap node a message broker that all messages flow through.

**Changes needed**:
1. Update `SendMessage` gRPC endpoint to always send messages to the bootstrap node first
2. Bootstrap node receives message and:
   - Forwards to gRPC-Web subscribers (for real-time delivery)
   - Routes to destination peer node via libp2p (for peer-to-peer delivery)

**Pros**:
- Simple architecture
- Centralized message routing
- Works with current gRPC-Web setup

**Cons**:
- Bootstrap node becomes a bottleneck
- Not truly peer-to-peer

### Option 2: Deploy Envoy + gRPC Server on Each Node

**Approach**: Each peer node (alice, bob, charlie) runs its own Envoy proxy and gRPC server.

**Changes needed**:
1. Update `docker-compose.yml` to add Envoy sidecars for each node
2. Frontend clients connect to the Envoy proxy of their designated peer
3. Messages flow: alice-frontend → alice-envoy → alice-node → bob-node → bob-subscribers

**Pros**:
- Truly distributed
- No single point of failure
- True peer-to-peer architecture

**Cons**:
- More complex deployment
- Requires load balancing/service discovery for clients
- More resource intensive

### Option 3: Use WebSocket for Streaming (Alternative)

**Approach**: Replace gRPC-Web streaming with WebSocket connections.

**Changes needed**:
1. Implement WebSocket message streaming in the backend
2. Update frontend to use WebSocket instead of gRPC-Web for `ReceiveMessages`
3. Keep gRPC for unary calls (SendMessage, GetPeers, etc.)

**Pros**:
- WebSocket is well-supported in browsers
- Simpler than gRPC-Web
- Can use existing WebSocket server code

**Cons**:
- Mixed protocols (WebSocket + gRPC)
- Requires implementing custom message framing

## ✅ IMPLEMENTED: Option 1 (Message Broker Pattern)

**Status:** COMPLETED - Working solution implemented

The message broker pattern has been successfully implemented:

1. ✅ **Added `RouteMessage` method** to `MessageHandler`
2. ✅ **Updated gRPC `MessageService`** to use `RouteMessage` instead of `SendMessage`
3. ✅ **E2E streaming working** - messages flow through bootstrap node to both gRPC-Web subscribers and destination peers
4. ✅ **Debug logging** shows complete message flow

**Evidence from logs:**
```
Bootstrap: 🔄 RouteMessage called: toPeerID=12D3KooW..., content=Hello from Bob!
Bootstrap: 📡 Forwarding message to 2 local subscribers ✅
Bootstrap: ✅ Message forwarded to local subscriber ✅
Bootstrap: 🌐 Routing message to peer: 12D3KooW... ✅
Bootstrap: ✅ Stream created successfully ✅
```

## 🚀 Future Refactoring: Option 2 (Fully Distributed)

**TODO:** Refactor to Option 2 for a production-ready distributed architecture without bootstrap node bottleneck.

### Implementation Steps:

1. **Update `SendMessage` to route through bootstrap**:
   ```go
   // In message_service.go SendMessage()
   // Instead of sending directly to peer:
   // s.msgHandler.SendMessage(ctx, req.ToPeerId, req.Content)
   
   // Send to bootstrap node first, let it route:
   bootstrapPeerID := getBootstrapPeerID()
   s.msgHandler.SendToBootstrap(ctx, bootstrapPeerID, req.ToPeerId, req.Content)
   ```

2. **Update `MessageHandler` to route messages**:
   ```go
   // Add new method to handle routing
   func (m *MessageHandler) RouteMessage(ctx context.Context, toPeerID string, content []byte) error {
       // 1. Forward to local subscribers (for gRPC-Web clients)
       m.forwardToSubscribers(content)
       
       // 2. Route to destination peer via libp2p (if not local)
       if toPeerID != m.host.ID().String() {
           return m.SendMessage(ctx, toPeerID, content)
       }
       return nil
   }
   ```

3. **Update Docker Compose**:
   - Ensure all gRPC-Web clients connect to the bootstrap node
   - Keep peer-to-peer connections for message delivery

## Testing

After implementing the fix:

1. Open `test_e2e_streaming.html` with two clients
2. Click "Discover Peers" on both
3. Click "Start Listening" on both (should connect to bootstrap)
4. Send message from Client 1
5. Verify Client 2 receives the message in real-time

Expected logs:
```
Bootstrap: 📨 handleStream called from peer (message arrives)
Bootstrap: 📡 Forwarding message to 4 subscribers (message forwarded)
Client 2: 📨 Received message from Client 1
```

