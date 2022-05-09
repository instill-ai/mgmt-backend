import http from "k6/http";
import {check, group} from "k6";
import {randomString} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import * as constant from "./const.js";
import * as helper from "./helper.js";

var checks = {
  ID: constant.defaultUser.id,
  UUID: constant.defaultUser.uid,
};
var nonExistChecks = {
  ID: "non-exist", // non exist ID
  UUID: "2a06c2f7-8da9-4046-91ea-240f88a5d000", // non exist UUID
};

export function CheckList() {
  group("Management API: List users", () => {
    check(http.request("GET", `${constant.mgmtHost}/users`), {
      [`GET /${constant.mgmtVersion}/users status 200`]: (r) =>
        r.status === 200,
      [`GET /${constant.mgmtVersion}/users response body has user array`]: (
        r
      ) => Array.isArray(r.json().users),
    });

    var res = http.request("GET", `${constant.mgmtHost}/users`);
    check(http.request("GET", `${constant.mgmtHost}/users?page_size=0`), {
      [`GET /${constant.mgmtVersion}/users?page_size=0 response status 200`]: (
        r
      ) => r.status === 200,
      [`GET /${constant.mgmtVersion}/users?page_size=0 response all records`]: (
        r
      ) => r.json().users.length === res.json().users.length,
      [`GET /${constant.mgmtVersion}/users?page_size=0 response total_size 1`]:
        (r) => r.json().total_size === res.json().total_size,
    });

    check(http.request("GET", `${constant.mgmtHost}/users?page_size=5`), {
      [`GET /${constant.mgmtVersion}/users?page_size=5 response status 200`]: (
        r
      ) => r.status === 200,
      [`GET /${constant.mgmtVersion}/users?page_size=5 response all records size 1`]:
        (r) => r.json().users.length === 1,
      [`GET /${constant.mgmtVersion}/users?page_size=5 response total_size 1`]:
        (r) => r.json().total_size === 1,
      [`GET /${constant.mgmtVersion}/users?page_size=5 response next_page_token is empty`]:
        (r) => r.json().next_page_token === "",
    });

    var invalidNextPageToken = `${randomString(10)}`;
    check(
      http.request(
        "GET",
        `${constant.mgmtHost}/users?page_size=1&page_token=${invalidNextPageToken}`
      ),
      {
        [`GET /${constant.mgmtVersion}/users?page_size=1&page_token=${invalidNextPageToken} response status 400`]:
          (r) => r.status === 400,
      }
    );
  });
}

export function CheckGet() {
  for (const [key, value] of Object.entries(checks)) {
    group(`Management API: Get default user by ${key}`, () => {
      check(http.request("GET", `${constant.mgmtHost}/users/${value}`), {
        [`GET /${constant.mgmtVersion}/users/${value} response status 200`]: (
          r
        ) => r.status === 200,
        [`GET /${constant.mgmtVersion}/users/${value} response name`]: (r) =>
          r.json().user.name !== undefined,
        [`GET /${constant.mgmtVersion}/users/${value} response uid is UUID`]: (
          r
        ) => helper.isUUID(r.json().user.uid),
        [`GET /${constant.mgmtVersion}/users/${value} response id`]: (r) =>
          r.json().user.id !== undefined,
        [`GET /${constant.mgmtVersion}/users/${value} response email`]: (r) =>
          r.json().user.email !== undefined,
        [`GET /${constant.mgmtVersion}/users/${value} response id`]: (r) =>
          r.json().user.id === constant.defaultUser.id,
        [`GET /${constant.mgmtVersion}/users/${value} response company_name`]: (
          r
        ) => r.json().user.company_name !== undefined,
        [`GET /${constant.mgmtVersion}/users/${value} response role`]: (r) =>
          r.json().user.role !== undefined,
        [`GET /${constant.mgmtVersion}/users/${value} response usage_data_collection`]:
          (r) => r.json().user.usage_data_collection !== undefined,
        [`GET /${constant.mgmtVersion}/users/${value} response newsletter_subscription`]:
          (r) => r.json().user.newsletter_subscription !== undefined,
        [`GET /${constant.mgmtVersion}/users/${value} response type`]: (r) =>
          r.json().user.type === "OWNER_TYPE_USER",
        [`GET /${constant.mgmtVersion}/users/${value} response create_time`]: (
          r
        ) => r.json().user.create_time !== undefined,
        [`GET /${constant.mgmtVersion}/users/${value} response update_time`]: (
          r
        ) => r.json().user.update_time !== undefined,
      });
    });
  }

  for (const [key, value] of Object.entries(nonExistChecks)) {
    group(`Management API: Get non-exist user by ${key}`, () => {
      check(http.request("GET", `${constant.mgmtHost}/users/${value}`), {
        [`GET /${constant.mgmtVersion}/users/${value} response status 404`]: (
          r
        ) => r.status === 404,
      });
    });
  }
}

export function CheckUpdate() {
  for (const [key, value] of Object.entries(checks)) {
    group(`Management API: Update default user by ${key}`, () => {
      var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
      var userUpdate = {
        name: "users/non-exist",
        uid: nonExistUID,
        email: "test@foo.bar",
        company_name: "company",
        role: "Manager",
        usage_data_collection: true,
        newsletter_subscription: false,
        type: "OWNER_TYPE_ORGANIZATION",
        create_time: "2000-01-01T00:00:00.000000Z",
        update_time: "2000-01-01T00:00:00.000000Z",
      };

      var res = http.request("GET", `${constant.mgmtHost}/users/${value}`);

      check(
        http.request(
          "PATCH",
          `${constant.mgmtHost}/users/${value}`,
          JSON.stringify(userUpdate),
          {headers: {"Content-Type": "application/json"}}
        ),
        {
          [`PATCH /${constant.mgmtVersion}/users/${value} response 200`]: (r) =>
            r.status === 200,
          [`PATCH /${constant.mgmtVersion}/users/${value} response name unchanged`]:
            (r) => r.json().user.name === res.json().user.name,
          [`PATCH /${constant.mgmtVersion}/users/${value} response uid unchanged`]:
            (r) => r.json().user.uid === res.json().user.uid,
          [`PATCH /${constant.mgmtVersion}/users/${value} response id unchanged`]:
            (r) => r.json().user.id === res.json().user.id,
          [`PATCH /${constant.mgmtVersion}/users/${value} response email updated`]:
            (r) => r.json().user.email === userUpdate.email,
          [`PATCH /${constant.mgmtVersion}/users/${value} response company_name updated`]:
            (r) => r.json().user.company_name === userUpdate.company_name,
          [`PATCH /${constant.mgmtVersion}/users/${value} response role updated`]:
            (r) => r.json().user.role === userUpdate.role,
          [`PATCH /${constant.mgmtVersion}/users/${value} response usage_data_collection updated`]:
            (r) =>
              r.json().user.usage_data_collection ===
              userUpdate.usage_data_collection,
          [`PATCH /${constant.mgmtVersion}/users/${value} response newsletter_subscription updated`]:
            (r) =>
              r.json().user.newsletter_subscription ===
              userUpdate.newsletter_subscription,
          [`PATCH /${constant.mgmtVersion}/users/${value} response type unchanged`]:
            (r) => r.json().user.type === res.json().user.type,
          [`PATCH /${constant.mgmtVersion}/users/${value} response create_time unchanged`]:
            (r) => r.json().user.create_time === res.json().user.create_time,
          [`PATCH /${constant.mgmtVersion}/users/${value} response update_time updated`]:
            (r) => r.json().user.update_time !== res.json().user.update_time,
          [`PATCH /${constant.mgmtVersion}/users/${value} response update_time not updated with request value`]:
            (r) => r.json().user.update_time !== userUpdate.update_time,
        }
      );

      // Restore to default user
      check(
        http.request(
          "PATCH",
          `${constant.mgmtHost}/users/${value}`,
          JSON.stringify(constant.defaultUser),
          {headers: {"Content-Type": "application/json"}}
        ),
        {
          [`PATCH /${constant.mgmtVersion}/users/${value} response status 200`]:
            (r) => r.status === 200,
        }
      );
      check(http.request("GET", `${constant.mgmtHost}/users/${value}`), {
        [`GET /${constant.mgmtVersion}/users/${value} response status 200`]: (
          r
        ) => r.status === 200,
      });
    });

    group(`Management API: Update user with a non-exist role by ${key}`, () => {
      var nonExistRole = "non-exist-role";
      var userUpdate = {
        role: nonExistRole,
      };
      check(
        http.request(
          "PATCH",
          `${constant.mgmtHost}/users/${value}`,
          JSON.stringify(userUpdate),
          {headers: {"Content-Type": "application/json"}}
        ),
        {
          [`PATCH /${constant.mgmtVersion}/users/${value} response status 400`]:
            (r) => r.status === 400,
        }
      );
    });

    group(
      `Management API: Update user ID by querying ${key} [not allowed]`,
      () => {
        var userUpdate = {
          id: `test_${randomString(10)}`,
        };
        check(
          http.request(
            "PATCH",
            `${constant.mgmtHost}/users/${value}`,
            JSON.stringify(userUpdate),
            {headers: {"Content-Type": "application/json"}}
          ),
          {
            [`PATCH /${constant.mgmtVersion}/users/${value} response status 400`]:
              (r) => r.status === 400,
          }
        );
      }
    );
  }

  for (const [key, value] of Object.entries(nonExistChecks)) {
    group(`Management API: Update non-exist user by ${key}`, () => {
      check(http.request("PATCH", `${constant.mgmtHost}/users/${value}`), {
        [`PATCH /${constant.mgmtVersion}/users/${value} response status 404`]: (
          r
        ) => r.status === 404,
      });
    });
  }
}

export function CheckCreate() {
  group("Management API: Create user with UUID as id", () => {
    check(
      http.request(
        "POST",
        `${constant.mgmtHost}/users`,
        JSON.stringify({id: "2a06c2f7-8da9-4046-91ea-240f88a5d000"}),
        {
          headers: {"Content-Type": "application/json"},
        }
      ),
      {
        [`POST /${constant.mgmtVersion}/users response status 400`]: (r) =>
          r.status === 400,
      }
    );
  });
  group("Management API: Create user", () => {
    check(
      http.request("POST", `${constant.mgmtHost}/users`, JSON.stringify({}), {
        headers: {"Content-Type": "application/json"},
      }),
      {
        [`POST /${constant.mgmtVersion}/users response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}

export function CheckDelete() {
  for (const [key, value] of Object.entries(checks)) {
    group(`Management API: Delete user by ${key}`, () => {
      check(
        http.request(
          "DELETE",
          `${constant.mgmtHost}/users/${value}`,
          JSON.stringify({}),
          {
            headers: {"Content-Type": "application/json"},
          }
        ),
        {
          [`DELETE /${constant.mgmtVersion}/users/${value} response status 501 [not implemented]`]:
            (r) => r.status === 501,
        }
      );
    });
  }
}
