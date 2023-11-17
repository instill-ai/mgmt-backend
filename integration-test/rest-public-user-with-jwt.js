import http from "k6/http";
import { check, group } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import * as constant from "./const.js";
import * as helper from "./helper.js";


export function CheckPublicGetUser() {
  group(`Management Public API: Get authenticated user [with random "jwt-sub" header]`, () => {

    check(http.request("GET", `${constant.mgmtPublicHost}/users/me`, {}, constant.restParamsWithJwtSub),
      {
        [`[with random "jwt-sub" header] GET /${constant.mgmtVersion}/users/me response status 401`]:
          (r) => r.status === 401,
      }
    )
  })
}

export function CheckPublicPatchAuthenticatedUser() {
  group(`Management Public API: Update authenticated user [with random "jwt-sub" header]`, () => {
    var userUpdate = {
      email: "test@foo.bar",
      customer_id: "new_customer_id",
      first_name: "test",
      last_name: "foo",
      org_name: "company",
      role: "ai-engineer",
      newsletter_subscription: true,
      cookie_token: "f5730f62-7026-4e11-917a-d890da315d3b",
      create_time: "2000-01-01T00:00:00.000000Z",
      update_time: "2000-01-01T00:00:00.000000Z",
    };

    check(http.request("PATCH", `${constant.mgmtPublicHost}/users/me`, JSON.stringify(userUpdate), constant.restParamsWithJwtSub),
      {
        [`[with random "jwt-sub" header] PATCH /${constant.mgmtVersion}/users/me response status 401`]: (r) => r.status === 401,
      }
    );
  });
}
