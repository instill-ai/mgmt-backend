import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

let proto

export const apiGatewayMode = (__ENV.API_GATEWAY_URL !== "" && __ENV.API_GATEWAY_URL !== undefined);

if (__ENV.API_GATEWAY_PROTOCOL) {
  if (__ENV.API_GATEWAY_PROTOCOL !== "http" && __ENV.API_GATEWAY_PROTOCOL != "https") {
    fail("only allow `http` or `https` for API_GATEWAY_PROTOCOL")
  }
  proto = __ENV.API_GATEWAY_PROTOCOL
} else {
  proto = "http"
}
export const mgmtVersion = "v1beta";
export const mgmtPrivateHost = apiGatewayMode ? "" : `http://mgmt-backend:3084/${mgmtVersion}`
export const mgmtPublicHost = apiGatewayMode ? `${proto}://${__ENV.API_GATEWAY_URL}/${mgmtVersion}` : `http://mgmt-backend:8084/${mgmtVersion}`

export const mgmtPrivateGRPCHost = `mgmt-backend:3084`
export const mgmtPublicGRPCHost = apiGatewayMode ? `${__ENV.API_GATEWAY_URL}` : `mgmt-backend:8084`

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

export const defaultUser = {
  name: "users/admin",
  id: "admin",
  email: "hello@instill-ai.com",
  customerId: "",
  profile: {
    displayName: "Instill AI",
    companyName: "Instill AI",
  },
  role: "hobbyist",
  newsletter_subscription: true,
};

export const testToken = {
  id: "test-token",
  ttl: -1,
};
