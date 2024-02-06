import http from "k6/http";
import { check, group } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import * as constant from "./const.js";
import * as helper from "./helper.js";

export function CheckPrivateListUsersAdmin() {
  group("Management Private API: List users", () => {
    check(http.request("GET", `${constant.mgmtPrivateHost}/admin/users`), {
      [`GET /${constant.mgmtVersion}/admin/users status 200`]: (r) =>
        r.status === 200,
      [`GET /${constant.mgmtVersion}/admin/users response body has user array`]: (
        r
      ) => Array.isArray(r.json().users),
    });

    var res = http.request("GET", `${constant.mgmtPrivateHost}/admin/users`);
    check(http.request("GET", `${constant.mgmtPrivateHost}/admin/users?page_size=0`), {
      [`GET /${constant.mgmtVersion}/admin/users?page_size=0 response status 200`]: (
        r
      ) => r.status === 200,
      [`GET /${constant.mgmtVersion}/admin/users?page_size=0 response all records`]: (
        r
      ) => r.json().users.length === res.json().users.length,
      [`GET /${constant.mgmtVersion}/admin/users?page_size=0 response total_size 1`]:
        (r) => r.json().total_size === res.json().total_size,
    });

    check(http.request("GET", `${constant.mgmtPrivateHost}/admin/users?page_size=5`), {
      [`GET /${constant.mgmtVersion}/admin/users?page_size=5 response status 200`]: (
        r
      ) => r.status === 200,
      [`GET /${constant.mgmtVersion}/admin/users?page_size=5 response all records size 1`]:
        (r) => r.json().users.length === 1,
      [`GET /${constant.mgmtVersion}/admin/users?page_size=5 response total_size 1`]:
        (r) => r.json().total_size === 1,
      [`GET /${constant.mgmtVersion}/admin/users?page_size=5 response next_page_token is empty`]:
        (r) => r.json().next_page_token === "",
    });

    var invalidNextPageToken = `${randomString(10)}`;
    check(
      http.request(
        "GET",
        `${constant.mgmtPrivateHost}/admin/users?page_size=1&page_token=${invalidNextPageToken}`
      ),
      {
        [`GET /${constant.mgmtVersion}/admin/users?page_size=1&page_token=${invalidNextPageToken} response status 400`]:
          (r) => r.status === 400,
      }
    );
  });
}

export function CheckPrivateGetUserAdmin() {
  group(`Management Private API: Get default user`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`
      ),
      {
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response status 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response name`]:
          (r) => r.json().user.name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response uid is UUID`]:
          (r) => helper.isUUID(r.json().user.uid),
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response id`]:
          (r) => r.json().user.id !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response id`]:
          (r) => r.json().user.id === constant.defaultUser.id,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response display_name`]:
          (r) => r.json().user.profile.display_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response company_name`]:
          (r) => r.json().user.profile.company_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response create_time`]:
          (r) => r.json().user.create_time !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response update_time`]:
          (r) => r.json().user.update_time !== undefined,
      }
    );
  });

  var nonExistID = "non-exist";
  group(`Management Private API: Get non-exist user`, () => {
    check(http.request("GET", `${constant.mgmtPrivateHost}/admin/users/${nonExistID}`), {
      [`GET /${constant.mgmtVersion}/admin/users/${nonExistID} response status 404`]:
        (r) => r.status === 404,
    });
  });
}

export function CheckPrivateLookUpUserAdmin() {
  // Get the uid of the default user
  var res = http.request(
    "GET",
    `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`
  );
  var defaultUid = res.json().user.uid;

  group(`Management Private API: Look up default user by permalink`, () => {
    check(
      http.request("GET", `${constant.mgmtPrivateHost}/admin/users/${defaultUid}/lookUp`),
      {
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response status 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response name`]: (
          r
        ) => r.json().user.name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response uid is UUID`]:
          (r) => helper.isUUID(r.json().user.uid),
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response id`]:
          (r) => r.json().user.id !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response id`]:
          (r) => r.json().user.id === constant.defaultUser.id,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response display_name`]:
          (r) => r.json().user.profile.display_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response company_name`]:
          (r) => r.json().user.profile.company_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response create_time`]:
          (r) => r.json().user.create_time !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response update_time`]:
          (r) => r.json().user.update_time !== undefined,
      }
    );
  });

  var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
  group(`Management Private API: Look up non-exist user by permalink`, () => {
    check(
      http.request("GET", `${constant.mgmtPrivateHost}/admin/users/${nonExistUID}/lookUp`),
      {
        [`GET /${constant.mgmtVersion}/admin/users/${nonExistUID} response status 404`]:
          (r) => r.status === 404,
      }
    );
  });
}
