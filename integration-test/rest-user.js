import http from "k6/http";
import {check, group} from "k6";
import {randomString} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import * as constant from "./const.js";
import * as helper from "./helper.js";

export function CheckList() {
  group("Management API: List users", () => {
    check(http.request("GET", `${constant.mgmtHost}/users`), {
      "GET /users status 200": (r) => r.status === 200,
      "GET /users response body has user array": (r) =>
        Array.isArray(r.json().users),
    });

    var allRes = http.request("GET", `${constant.mgmtHost}/users`);
    check(http.request("GET", `${constant.mgmtHost}/users?page_size=0`), {
      "GET /users?page_size=0 response status 200": (r) => r.status === 200,
      "GET /users?page_size=0 response all records": (r) =>
        r.json().users.length === allRes.json().users.length,
      "GET /users?page_size=0 response total_size 1": (r) =>
        r.json().total_size === allRes.json().total_size,
    });

    check(http.request("GET", `${constant.mgmtHost}/users?page_size=0`), {
      "GET /users?page_size=1 response status 200": (r) => r.status === 200,
      "GET /users?page_size=1 response all records size 1": (r) =>
        r.json().users.length === 1,
      "GET /users?page_size=1 response total_size 1": (r) =>
        r.json().total_size === allRes.json().total_size,
    });

    var pageRes = http.request("GET", `${constant.mgmtHost}/users?page_size=1`);
    check(
      http.request(
        "GET",
        `${constant.mgmtHost}/users?page_size=1&page_token=${
          pageRes.json().next_page_token
        }`
      ),
      {
        [`GET /users?page_size=1&page_token=${pageRes.json(
          "next_page_token"
        )} response status 200`]: (r) => r.status === 200,
        [`GET /users?page_size=1&page_token=${pageRes.json(
          "next_page_token"
        )} response all records size 0`]: (r) => r.json().users.length === 0,
        [`GET /users?page_size=1&page_token=${pageRes.json(
          "next_page_token"
        )} response total_size 1`]: (r) => r.json().total_size === 1,
      }
    );

    var invalidNextPageToken = `${randomString(10)}`;
    check(
      http.request(
        "GET",
        `${constant.mgmtHost}/users?page_size=1&page_token=${invalidNextPageToken}`
      ),
      {
        [`GET /users?page_size=1&page_token=${invalidNextPageToken} response status 400`]:
          (r) => r.status === 400,
      }
    );
  });
}

export function CheckGet() {
  group("Management API: Get default user", () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtHost}/users/${constant.defaultUser.login}`
      ),
      {
        [`GET /users/${constant.defaultUser.login} response status 200`]: (r) =>
          r.status === 200,
        [`GET /users/${constant.defaultUser.login} response name`]: (r) =>
          r.json().user.name !== undefined,
        [`GET /users/${constant.defaultUser.login} response id is UUID`]: (r) =>
          helper.isUUID(r.json().user.id),
        [`GET /users/${constant.defaultUser.login} response id`]: (r) =>
          r.json().user.id !== undefined,
        [`GET /users/${constant.defaultUser.login} response email`]: (r) =>
          r.json().user.email !== undefined,
        [`GET /users/${constant.defaultUser.login} response login`]: (r) =>
          r.json().user.login === constant.defaultUser.login,
        [`GET /users/${constant.defaultUser.login} response company_name`]: (
          r
        ) => r.json().user.company_name !== undefined,
        [`GET /users/${constant.defaultUser.login} response role`]: (r) =>
          r.json().user.role !== undefined,
        [`GET /users/${constant.defaultUser.login} response usage_data_collection`]:
          (r) => r.json().user.usage_data_collection !== undefined,
        [`GET /users/${constant.defaultUser.login} response newsletter_subscription`]:
          (r) => r.json().user.newsletter_subscription !== undefined,
        [`GET /users/${constant.defaultUser.login} response type`]: (r) =>
          r.json().user.type === "OWNER_TYPE_USER",
        [`GET /users/${constant.defaultUser.login} response create_time`]: (
          r
        ) => r.json().user.create_time !== undefined,
        [`GET /users/${constant.defaultUser.login} response update_time`]: (
          r
        ) => r.json().user.update_time !== undefined,
      }
    );
  });

  group("Management API: Get non-exist user", () => {
    var nonExistUser = "non-exist";
    check(http.request("GET", `${constant.mgmtHost}/users/${nonExistUser}`), {
      [`GET /users/${nonExistUser} response status 404`]: (r) =>
        r.status === 404,
    });
  });
}

export function CheckUpdate() {
  group("Management API: Update default user", () => {
    var nonExistId = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
    var userUpdate = {
      name: "users/non-exist",
      id: nonExistId,
      email: "test@foo.bar",
      company_name: "company",
      role: "Manager",
      usage_data_collection: true,
      newsletter_subscription: false,
      type: "OWNER_TYPE_ORGANIZATION",
      create_time: "2000-01-01T00:00:00.000000Z",
      update_time: "2000-01-01T00:00:00.000000Z",
    };

    var res = http.request(
      "GET",
      `${constant.mgmtHost}/users/${constant.defaultUser.login}`
    );

    check(
      http.request(
        "PATCH",
        `${constant.mgmtHost}/users/${constant.defaultUser.login}`,
        JSON.stringify(userUpdate),
        {headers: {"Content-Type": "application/json"}}
      ),
      {
        [`PATCH /users/${constant.defaultUser.login} response 200`]: (r) =>
          r.status === 200,
        [`PATCH /users/${constant.defaultUser.login} response name unchanged`]:
          (r) => r.json().user.name === res.json().user.name,
        [`PATCH /users/${constant.defaultUser.login} response id unchanged`]: (
          r
        ) => r.json().user.id === res.json().user.id,
        [`PATCH /users/${constant.defaultUser.login} response email updated`]: (
          r
        ) => r.json().user.email === userUpdate.email,
        [`PATCH /users/${constant.defaultUser.login} response company_name updated`]:
          (r) => r.json().user.company_name === userUpdate.company_name,
        [`PATCH /users/${constant.defaultUser.login} response role updated`]: (
          r
        ) => r.json().user.role === userUpdate.role,
        [`PATCH /users/${constant.defaultUser.login} response usage_data_collection updated`]:
          (r) =>
            r.json().user.usage_data_collection ===
            userUpdate.usage_data_collection,
        [`PATCH /users/${constant.defaultUser.login} response newsletter_subscription updated`]:
          (r) =>
            r.json().user.newsletter_subscription ===
            userUpdate.newsletter_subscription,
        [`PATCH /users/${constant.defaultUser.login} response type unchanged`]:
          (r) => r.json().user.type === res.json().user.type,
        [`PATCH /users/${constant.defaultUser.login} response create_time unchanged`]:
          (r) => r.json().user.create_time === res.json().user.create_time,
        [`PATCH /users/${constant.defaultUser.login} response update_time updated`]:
          (r) => r.json().user.update_time !== res.json().user.update_time,
        [`PATCH /users/${constant.defaultUser.login} response update_time not updated with request value`]:
          (r) => r.json().user.update_time !== userUpdate.update_time,
      }
    );

    // Restore to default user
    check(
      http.request(
        "PATCH",
        `${constant.mgmtHost}/users/${constant.defaultUser.login}`,
        JSON.stringify(constant.defaultUser),
        {headers: {"Content-Type": "application/json"}}
      ),
      {
        [`PATCH /users/${constant.defaultUser.login} response status 200`]: (
          r
        ) => r.status === 200,
      }
    );
    check(
      http.request(
        "GET",
        `${constant.mgmtHost}/users/${constant.defaultUser.login}`
      ),
      {
        [`GET /users/${constant.defaultUser.login} response status 200`]: (r) =>
          r.status === 200,
      }
    );
  });

  group("Management API: Update user with a non-exist role", () => {
    var nonExistRole = "non-exist-role";
    var userUpdate = {
      role: nonExistRole,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtHost}/users/${constant.defaultUser.login}`,
        JSON.stringify(userUpdate),
        {headers: {"Content-Type": "application/json"}}
      ),
      {
        [`PATCH /users/${constant.defaultUser.login} response status 400`]: (
          r
        ) => r.status === 400,
      }
    );
  });

  group("Management API: Update user's login", () => {
    var userUpdate = {
      login: `test_${randomString(10)}`,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtHost}/users/${constant.defaultUser.login}`,
        JSON.stringify(userUpdate),
        {headers: {"Content-Type": "application/json"}}
      ),
      {
        [`PATCH /users/${constant.defaultUser.login} response status 200`]: (
          r
        ) => r.status === 200,
        [`PATCH /users/${constant.defaultUser.login} response login updated`]: (
          r
        ) => r.json().user.login === userUpdate.login,
      }
    );

    check(
      http.request(
        "GET",
        `${constant.mgmtHost}/users/${constant.defaultUser.login}`
      ),
      {
        [`GET /users/${constant.defaultUser.login} response status 404`]: (r) =>
          r.status === 404,
      }
    );

    // Restore to default user
    check(
      http.request(
        "PATCH",
        `${constant.mgmtHost}/users/${userUpdate.login}`,
        JSON.stringify(constant.defaultUser),
        {headers: {"Content-Type": "application/json"}}
      ),
      {
        [`PATCH /users/${userUpdate.login} response status 200`]: (r) =>
          r.status === 200,
      }
    );
    check(
      http.request(
        "GET",
        `${constant.mgmtHost}/users/${constant.defaultUser.login}`
      ),
      {
        [`GET /users/${constant.defaultUser.login} response status 200`]: (r) =>
          r.status === 200,
      }
    );
  });

  group("Management API: Update non-exist user", () => {
    var nonExistUser = "non-exist";
    check(http.request("PATCH", `${constant.mgmtHost}/users/${nonExistUser}`), {
      [`PATCH /users/${nonExistUser} response status 404`]: (r) =>
        r.status === 404,
    });
  });
}

export function CheckCreate() {
  group("Management API: Create user", () => {
    check(
      http.request("POST", `${constant.mgmtHost}/users`, JSON.stringify({}), {
        headers: {"Content-Type": "application/json"},
      }),
      {
        "POST /users response status 501 [not implemented]": (r) =>
          r.status === 501,
      }
    );
  });
}

export function CheckDelete() {
  group("Management API: Delete user", () => {
    check(
      http.request(
        "DELETE",
        `${constant.mgmtHost}/users/${constant.defaultUser.login}`,
        JSON.stringify({}),
        {
          headers: {"Content-Type": "application/json"},
        }
      ),
      {
        "DELETE /users response status 501 [not implemented]": (r) =>
          r.status === 501,
      }
    );
  });
}
