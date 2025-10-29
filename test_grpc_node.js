const fetch = require('node-fetch');

async function testGrpcWeb() {
    try {
        console.log('Testing gRPC-Web with Node.js...');
        
        const response = await fetch('http://localhost:9001/ledabeer.PeerService/GetPeers', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/grpc-web+proto',
                'grpc-timeout': '10s'
            },
            body: Buffer.from([0x00, 0x00, 0x00, 0x00, 0x00])
        });
        
        console.log('Response status:', response.status);
        console.log('Response headers:', Object.fromEntries(response.headers.entries()));
        
        if (response.ok) {
            const data = await response.arrayBuffer();
            console.log('Response size:', data.byteLength, 'bytes');
            console.log('Response data:', Array.from(new Uint8Array(data)).map(b => b.toString(16).padStart(2, '0')).join(' '));
        } else {
            const text = await response.text();
            console.log('Error response:', text);
        }
        
    } catch (error) {
        console.error('Error:', error);
    }
}

testGrpcWeb();

