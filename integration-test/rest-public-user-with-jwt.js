import http from "k6/http";
import { check, group } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import * as constant from "./const.js";
import * as helper from "./helper.js";


export function CheckPublicGetUser() {
  group(`Management Public API: Get authenticated user [with random "instill-user-uid" header]`, () => {

    check(http.request("GET", `${constant.mgmtPublicHost}/user`, {}, constant.restParamsWithInstillUserUid),
      {
        [`[with random "instill-user-uid" header] GET /${constant.mgmtVersion}/user response status 401`]:
          (r) => r.status === 401,
      }
    )
  })
}

export function CheckPublicPatchAuthenticatedUser() {
  group(`Management Public API: Update authenticated user [with random "instill-user-uid" header]`, () => {
    var userUpdate = {
      email: "test@foo.bar",
      customer_id: "new_customer_id",
      profile: {
        display_name: "test",
        company_name: "company"
      },
      role: "ai-engineer",
      newsletter_subscription: true,
      createTime: "2000-01-01T00:00:00.000000Z",
      updateTime: "2000-01-01T00:00:00.000000Z",
    };

    check(http.request("PATCH", `${constant.mgmtPublicHost}/user`, JSON.stringify(userUpdate), constant.restParamsWithInstillUserUid),
      {
        [`[with random "instill-user-uid" header] PATCH /${constant.mgmtVersion}/user response status 401`]: (r) => r.status === 401,
      }
    );
  });
}
