import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

let proto
let host
let publicPort
let privatePort

if (__ENV.MODE == "api-gateway") {
  // api-gateway mode for accessing api-gateway directly
  proto = "http"
  host = "api-gateway"
  privatePort = 3084
  publicPort = 8080
} else if (__ENV.MODE == "localhost") {
  // localhost mode for accessing api-gateway from localhost
  proto = "http"
  host = "localhost"
  privatePort = 3084
  publicPort = 8080
} else if (__ENV.MODE == "internal") {
  // internal mode for accessing api-gateway from container
  proto = "http"
  host = "host.docker.internal"
  privatePort = 3084
  publicPort = 8080
} else {
  // direct microservice mode
  proto = "http"
  host = "mgmt-backend"
  privatePort = 3084
  publicPort = 8084
}

export const mgmtVersion = "v1alpha";
export const mgmtPrivateHost = `${proto}://${host}:${privatePort}/${mgmtVersion}`
export const mgmtPublicHost = `${proto}://${host}:${publicPort}/${mgmtVersion}`

export const mgmtPrivateGRPCHost = `${host}:${privatePort}`
export const mgmtPublicGRPCHost = `${host}:${publicPort}`

export const params = {
  headers: {
    "Content-Type": "application/json",
  },
};

const randomUUID = uuidv4();
export const paramsWithJwt = {
  headers: {
    "Content-Type": "application/json",
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
