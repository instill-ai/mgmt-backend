import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

let proto
let host
let publicPort
let privatePort

if (__ENV.API_GATEWAY_HOST && !__ENV.API_GATEWAY_PORT || !__ENV.API_GATEWAY_HOST && __ENV.API_GATEWAY_PORT) {
  fail("both API_GATEWAY_HOST and API_GATEWAY_PORT should be properly configured.")
}

export const apiGatewayMode = (__ENV.API_GATEWAY_HOST && __ENV.API_GATEWAY_PORT);

if (__ENV.API_GATEWAY_PROTOCOL) {
  if (__ENV.API_GATEWAY_PROTOCOL !== "http" && __ENV.API_GATEWAY_PROTOCOL != "https") {
    fail("only allow `http` or `https` for API_GATEWAY_PROTOCOL")
  }
  proto = __ENV.API_GATEWAY_PROTOCOL
} else {
  proto = "http"
}

if (apiGatewayMode) {
  // api gateway mode
  host = __ENV.API_GATEWAY_HOST
  privatePort = 3084
  publicPort = __ENV.API_GATEWAY_PORT
} else {
  // direct microservice mode
  host = "mgmt-backend"
  privatePort = 3084
  publicPort = 8084
}

export const mgmtVersion = "v1alpha";
export const mgmtPrivateHost = `${proto}://${host}:${privatePort}/${mgmtVersion}`
export const mgmtPublicHost = `${proto}://${host}:${publicPort}/${mgmtVersion}`

export const mgmtPrivateGRPCHost = `${host}:${privatePort}`
export const mgmtPublicGRPCHost = `${host}:${publicPort}`

export const restParams = {
  headers: {
    "Content-Type": "application/json",
  },
};

const randomUUID = uuidv4();
export const restParamsWithJwtSub = {
  headers: {
    "Content-Type": "application/json",
    "Jwt-Sub": randomUUID,
  },
}

export const grpcParamsWithJwtSub = {
  metadata: {
    "Jwt-Sub": randomUUID,
  },
}

export const defaultUser = {
  name: "users/instill-ai",
  id: "instill-ai",
  type: "OWNER_TYPE_USER",
  email: "hello@instill.tech",
  customer_id: "",
  first_name: "Instill",
  last_name: "AI",
  org_name: "Instill AI",
  role: "hobbyist",
  newsletter_subscription: true,
  cookie_token: ""
};

export const testToken = {
  name: "tokens/test-token",
  id: "test-token",
  access_token: "at_123456",
  state: "STATE_ACTIVE",
  token_type: "Bearer",
  lifetime: 86400
};
