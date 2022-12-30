let proto
let host
let port

if (__ENV.HOST == "localhost") {
  // api-gateway mode (outside container)
  proto = "https"
  host = "localhost"
  port = 8080
} else {
  // container mode (inside container)
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
