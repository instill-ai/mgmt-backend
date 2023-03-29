let proto
let host
let publicPort
let privatePort

if (__ENV.MODE == "api-gateway") {
  // api-gateway mode
  proto = "http"
  host = "api-gateway"
  privatePort = 3084
  publicPort = 8080
} else if (__ENV.MODE == "localhost") {
  // localhost mode for GitHub Actions
  proto = "http"
  host = "localhost"
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

export const defaultUser = {
  name: "users/local-user",
  id: "local-user",
  type: "OWNER_TYPE_USER",
  email: "local-user@instill.tech",
  customer_id: "",
  first_name: "",
  last_name: "",
  org_name: "",
  role: "",
  newsletter_subscription: false,
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
