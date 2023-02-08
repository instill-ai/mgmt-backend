import http from "k6/http";
import { check, group } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import * as constant from "./const.js";
import * as helper from "./helper.js";


export function CheckHealth(){
  // Health check
  group("Management API: Health check", () => {
    check(http.request("GET", `${constant.mgmtPublicHost}/health/mgmt`), {
      [`GET /${constant.mgmtVersion}/health/mgmt response status is 200`]: (
        r
      ) => r.status === 200,
    });

    check(http.request("GET", `${constant.mgmtPublicHost}/ready/mgmt`), {
      [`GET /${constant.mgmtVersion}/ready/mgmt response status is 200`]: (
        r
      ) => r.status === 200,
    });
  });
}

export function CheckPublicGet(){
  group(`Management Public API: Get authenticated user`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/user`
      ),
      {
        [`GET /${constant.mgmtVersion}/user response name`]:
          (r) => r.json().user.name !== undefined,
        [`GET /${constant.mgmtVersion}/user response uid is UUID`]:
          (r) => helper.isUUID(r.json().user.uid),
        [`GET /${constant.mgmtVersion}/user response id`]:
          (r) => r.json().user.id !== undefined,
        [`GET /${constant.mgmtVersion}/user response id`]:
          (r) => r.json().user.id === constant.defaultUser.id,
        [`GET /${constant.mgmtVersion}/user response type`]:
          (r) => r.json().user.type === "OWNER_TYPE_USER",
        [`GET /${constant.mgmtVersion}/user response email`]:
          (r) => r.json().user.email !== undefined,
        [`GET /${constant.mgmtVersion}/user response plan`]:
          (r) => r.json().user.plan !== undefined,
        [`GET /${constant.mgmtVersion}/user response billing_id`]:
          (r) => r.json().user.billing_id !== undefined,
        [`GET /${constant.mgmtVersion}/user response first_name`]:
          (r) => r.json().user.first_name !== undefined,
        [`GET /${constant.mgmtVersion}/user response last_name`]:
          (r) => r.json().user.last_name !== undefined,
        [`GET /${constant.mgmtVersion}/user response org_name`]:
          (r) => r.json().user.org_name !== undefined,
        [`GET /${constant.mgmtVersion}/user response role`]:
          (r) => r.json().user.role !== undefined,
        [`GET /${constant.mgmtVersion}/user response newsletter_subscription`]:
          (r) => r.json().user.newsletter_subscription !== undefined,
        [`GET /${constant.mgmtVersion}/user response cookie_token`]:
          (r) => r.json().user.cookie_token !== undefined,
        [`GET /${constant.mgmtVersion}/user response create_time`]:
          (r) => r.json().user.create_time !== undefined,
        [`GET /${constant.mgmtVersion}/user response update_time`]:
          (r) => r.json().user.update_time !== undefined,
      }
    )
  })
}

export function CheckPublicUpdate(){
  group(`Management Public API: Update authenticated user`, () => {
    var userUpdate = {
      type: "OWNER_TYPE_ORGANIZATION",
      email: "test@foo.bar",
      plan: "plans/new_plan",
      billing_id: "0",
      first_name: "test",
      last_name: "foo",
      org_name: "company",
      role: "Manager",
      newsletter_subscription: true,
      cookie_token: "f5730f62-7026-4e11-917a-d890da315d3b",
      create_time: "2000-01-01T00:00:00.000000Z",
      update_time: "2000-01-01T00:00:00.000000Z",
    };

    var res = http.request(
      "GET",
      `${constant.mgmtPublicHost}/user`
    );

    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/user`,
        JSON.stringify(userUpdate),
        { headers: { "Content-Type": "application/json" } }
      ),
      {
        [`PATCH /${constant.mgmtVersion}/user response 200`]:
          (r) => r.status === 200,
        [`PATCH /${constant.mgmtVersion}/user response name unchanged`]:
          (r) => r.json().user.name === res.json().user.name,
        [`PATCH /${constant.mgmtVersion}/user response uid unchanged`]:
          (r) => r.json().user.uid === res.json().user.uid,
        [`PATCH /${constant.mgmtVersion}/user response id unchanged`]:
          (r) => r.json().user.id === res.json().user.id,
        [`PATCH /${constant.mgmtVersion}/user response type unchanged`]:
          (r) => r.json().user.type === res.json().user.type,
        [`PATCH /${constant.mgmtVersion}/user response email updated`]:
          (r) => r.json().user.email === userUpdate.email,
        [`PATCH /${constant.mgmtVersion}/user response plan updated`]:
          (r) => r.json().user.plan === userUpdate.plan,
        [`PATCH /${constant.mgmtVersion}/user response billing_id updated`]:
          (r) => r.json().user.billing_id === userUpdate.billing_id,
        [`PATCH /${constant.mgmtVersion}/user response first_name updated`]:
          (r) => r.json().user.first_name === userUpdate.first_name,
        [`PATCH /${constant.mgmtVersion}/user response last_name updated`]:
          (r) => r.json().user.last_name === userUpdate.last_name,
        [`PATCH /${constant.mgmtVersion}/user response org_name updated`]:
          (r) => r.json().user.org_name === userUpdate.org_name,
        [`PATCH /${constant.mgmtVersion}/user response role updated`]:
          (r) => r.json().user.role === userUpdate.role,
        [`PATCH /${constant.mgmtVersion}/user response newsletter_subscription updated`]:
          (r) => r.json().user.newsletter_subscription === userUpdate.newsletter_subscription,
        [`PATCH /${constant.mgmtVersion}/user response cookie_token updated`]:
          (r) => r.json().user.cookie_token === userUpdate.cookie_token,
        [`PATCH /${constant.mgmtVersion}/user response create_time unchanged`]:
          (r) => r.json().user.create_time === res.json().user.create_time,
        [`PATCH /${constant.mgmtVersion}/user response update_time updated`]:
          (r) => r.json().user.update_time !== res.json().user.update_time,
        [`PATCH /${constant.mgmtVersion}/user response update_time not updated with request value`]:
          (r) => r.json().user.update_time !== userUpdate.update_time,
      }
    );

    // Restore to default user
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/user`,
        JSON.stringify(constant.defaultUser),
        { headers: { "Content-Type": "application/json" } }
      ),
      {
        [`PATCH /${constant.mgmtVersion}/user response status 200`]:
          (r) => r.status === 200,
      }
    );
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/user`
      ),
      {
        [`GET /${constant.mgmtVersion}/user response status 200`]:
          (r) => r.status === 200,
      }
    );
  });

  group(`Management Public API: Update authenticated user with a non-exist role`, () => {
    var nonExistRole = "non-exist-role";
    var userUpdate = {
      role: nonExistRole,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/user`,
        JSON.stringify(userUpdate),
        { headers: { "Content-Type": "application/json" } }
      ),
      {
        [`PATCH /${constant.mgmtVersion}/user response status 400`]:
          (r) => r.status === 400,
      }
    );
  });

  group(`Management Public API: Update authenticated user ID [not allowed]`, () => {
    var userUpdate = {
      id: `test_${randomString(10)}`,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/user`,
        JSON.stringify(userUpdate),
        { headers: { "Content-Type": "application/json" } }
      ),
      {
        [`PATCH /${constant.mgmtVersion}/user response status 400`]:
          (r) => r.status === 400,
      }
    );
  });

  group(`Management Public API: Update authenticated user UID [not allowed]`, () =>{
    var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
    var userUpdate = {
      uid: nonExistUID,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/user`,
        JSON.stringify(userUpdate),
        { headers: { "Content-Type": "application/json" } }
      ),
      {
        [`PATCH /${constant.mgmtVersion}/user response status 400`]:
          (r) => r.status === 400,
      }
    );
  });
}
