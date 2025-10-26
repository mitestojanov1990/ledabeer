/**
 * Proto Loader
 *
 * Loads protobuf definitions for gRPC services
 */

import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import * as path from 'path';

// Proto file paths
const PROTO_DIR = path.join(__dirname, '../../proto');
const MESSAGE_PROTO = path.join(PROTO_DIR, 'message.proto');
const CALL_PROTO = path.join(PROTO_DIR, 'call.proto');
const MEDIA_PROTO = path.join(PROTO_DIR, 'media.proto');

// Proto loader options
const PROTO_LOADER_OPTIONS = {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
};

/**
 * Load message service proto definition
 */
export function loadMessageProto(): grpc.GrpcObject {
  const packageDefinition = protoLoader.loadSync(MESSAGE_PROTO, PROTO_LOADER_OPTIONS);
  return grpc.loadPackageDefinition(packageDefinition);
}

/**
 * Load call service proto definition
 */
export function loadCallProto(): grpc.GrpcObject {
  const packageDefinition = protoLoader.loadSync(CALL_PROTO, PROTO_LOADER_OPTIONS);
  return grpc.loadPackageDefinition(packageDefinition);
}

/**
 * Load media service proto definition
 */
export function loadMediaProto(): grpc.GrpcObject {
  const packageDefinition = protoLoader.loadSync(MEDIA_PROTO, PROTO_LOADER_OPTIONS);
  return grpc.loadPackageDefinition(packageDefinition);
}

/**
 * Get service client constructor from package
 */
export function getServiceClient(
  proto: grpc.GrpcObject,
  packageName: string,
  serviceName: string
): grpc.ServiceClientConstructor {
  const pkg = proto[packageName] as any;
  if (!pkg || !pkg[serviceName]) {
    throw new Error(`Service ${packageName}.${serviceName} not found in proto definition`);
  }
  return pkg[serviceName] as grpc.ServiceClientConstructor;
}
