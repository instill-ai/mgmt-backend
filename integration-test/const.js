export const mgmtPort = 8084
export const mgmtVersion = "v1alpha";
export const mgmtHost = __ENV.HOSTNAME ? `http://${__ENV.HOSTNAME}:${mgmtPort}/${mgmtVersion}` : `http://mgmt-backend:${mgmtPort}/${mgmtVersion}`;

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
