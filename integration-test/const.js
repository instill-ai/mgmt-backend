let proto
let host
let port

if (__ENV.MODE == "api-gateway") {
  // api-gateway mode
  proto = "https"
  host = "api-gateway"
  port = 8080
} else if (__ENV.MODE == "localhost") {
  // localhost mode for GitHub Actions
  proto = "https"
  host = "localhost"
  port = 8080
} else {
  // direct microservice mode
  proto = "http"
  host = "mgmt-backend"
  port = 8084
}

export const mgmtVersion = "v1alpha";
export const mgmtHost = `${proto}://${host}:${port}/${mgmtVersion}`

export const defaultUser = {
  name: "users/instill",
  email: "local-user@instill.tech",
  id: "local-user",
  company_name: "",
  role: "",
  newsletter_subscription: false,
  cookie_token: "",
  type: "OWNER_TYPE_USER",
};
