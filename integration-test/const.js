let proto
let public_host
let admin_host
let public_port
let admin_port

if (__ENV.MODE == "api-gateway") {
  // api-gateway mode
  proto = "http"
  admin_host = "mgmt-backend-admin"
  public_host = "api-gateway"
  admin_port = 3084
  public_port = 8080
} else if (__ENV.MODE == "localhost") {
  // localhost mode for GitHub Actions
  proto = "http"
  admin_host = "mgmt-backend-admin"
  public_host = "localhost"
  admin_port = 3084
  public_port = 8080
} else {
  // direct microservice mode
  proto = "http"
  admin_host = "mgmt-backend-admin"
  public_host = "mgmt-backend-public"
  admin_port = 3084
  public_port = 8084
}

export const mgmtVersion = "v1alpha";
export const mgmtAdminHost = `${proto}://${admin_host}:${admin_port}/${mgmtVersion}/admin`
export const mgmtPublicHost = `${proto}://${public_host}:${public_port}/${mgmtVersion}`

export const defaultUser = {
  name: "users/local-user",
  id: "local-user",
  type: "OWNER_TYPE_USER",
  email: "local-user@instill.tech",
  plan: "plans/open-source",
  billing_id: "",
  first_name: "",
  last_name: "",
  org_name: "",
  role: "",
  newsletter_subscription: false,
  cookie_token: ""
};
