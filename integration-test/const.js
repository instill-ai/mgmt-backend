import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

let proto;

if (__ENV.API_GATEWAY_PROTOCOL) {
  if (__ENV.API_GATEWAY_PROTOCOL !== "http" && __ENV.API_GATEWAY_PROTOCOL != "https") {
    fail("only allow `http` or `https` for API_GATEWAY_PROTOCOL")
  }
  proto = __ENV.API_GATEWAY_PROTOCOL
} else {
  proto = "http"
}

// API Gateway URL (localhost:8080 from host, api-gateway:8080 from container)
const apiGatewayUrl = __ENV.API_GATEWAY_URL || "localhost:8080";

// Determine if running from host (localhost) or container
export const isHostMode = apiGatewayUrl === "localhost:8080";

export const mgmtVersion = "v1beta";

// Public hosts (via API Gateway)
export const mgmtPublicHost = `${proto}://${apiGatewayUrl}/${mgmtVersion}`;
export const mgmtPublicGRPCHost = apiGatewayUrl;

// Private hosts (direct backend, for internal service calls)
export const mgmtPrivateHost = `http://mgmt-backend:3084/${mgmtVersion}`;
export const mgmtPrivateGRPCHost = `mgmt-backend:3084`;

export const defaultUsername = "admin"
export const defaultPassword = "password"

export const restParams = {
  headers: {
    "Content-Type": "application/json",
  },
};

const randomUUID = uuidv4();
export const restParamsWithInstillUserUid = {
  headers: {
    "Content-Type": "application/json",
    "instill-user-uid": randomUUID,
  },
}

export const grpcParamsWithInstillUserUid = {
  metadata: {
    "instill-user-uid": randomUUID,
  },
}

// Default user update payload (excludes immutable/output-only fields like name, id, uid)
export const defaultUserUpdate = {
  email: "hello@instill-ai.com",
  profile: {
    displayName: "Instill AI",
    companyName: "Instill AI",
  },
  role: "hobbyist",
  newsletterSubscription: true,
};

// Full default user info for reference (includes immutable fields)
export const defaultUser = {
  name: "users/admin",
  id: "admin",
  email: "hello@instill-ai.com",
  profile: {
    displayName: "Instill AI",
    companyName: "Instill AI",
  },
  role: "hobbyist",
  newsletterSubscription: true,
};

export const testToken = {
  id: "test-token",
  ttl: -1,
};
