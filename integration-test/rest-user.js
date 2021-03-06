import http from "k6/http";
import {check, group} from "k6";
import {randomString} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import * as constant from "./const.js";
import * as helper from "./helper.js";

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
        (r) => r.json().total_size === "1",
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
  group(`Management API: Get default user`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtHost}/users/${constant.defaultUser.id}`
      ),
      {
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response status 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response name`]:
          (r) => r.json().user.name !== undefined,
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response uid is UUID`]:
          (r) => helper.isUUID(r.json().user.uid),
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response id`]:
          (r) => r.json().user.id !== undefined,
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response email`]:
          (r) => r.json().user.email !== undefined,
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response id`]:
          (r) => r.json().user.id === constant.defaultUser.id,
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response company_name`]:
          (r) => r.json().user.company_name !== undefined,
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response role`]:
          (r) => r.json().user.role !== undefined,
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response newsletter_subscription`]:
          (r) => r.json().user.newsletter_subscription !== undefined,
          [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response cookie_token`]:
          (r) => r.json().user.cookie_token !== undefined,
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response type`]:
          (r) => r.json().user.type === "OWNER_TYPE_USER",
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response create_time`]:
          (r) => r.json().user.create_time !== undefined,
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response update_time`]:
          (r) => r.json().user.update_time !== undefined,
      }
    );
  });

  var nonExistID = "non-exist";
  group(`Management API: Get non-exist user`, () => {
    check(http.request("GET", `${constant.mgmtHost}/users/${nonExistID}`), {
      [`GET /${constant.mgmtVersion}/users/${nonExistID} response status 404`]:
        (r) => r.status === 404,
    });
  });
}

export function CheckLookUp() {
  // Get the uid of the default user
  var res = http.request(
    "GET",
    `${constant.mgmtHost}/users/${constant.defaultUser.id}`
  );
  var defaultUid = res.json().user.uid;

  group(`Management API: Look up default user by permalink`, () => {
    check(
      http.request("GET", `${constant.mgmtHost}/users/${defaultUid}:lookUp`),
      {
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response status 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response name`]: (
          r
        ) => r.json().user.name !== undefined,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response uid is UUID`]:
          (r) => helper.isUUID(r.json().user.uid),
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response id`]: (r) =>
          r.json().user.id !== undefined,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response email`]: (
          r
        ) => r.json().user.email !== undefined,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response id`]: (r) =>
          r.json().user.id === constant.defaultUser.id,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response company_name`]:
          (r) => r.json().user.company_name !== undefined,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response role`]: (
          r
        ) => r.json().user.role !== undefined,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response newsletter_subscription`]:
          (r) => r.json().user.newsletter_subscription !== undefined,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response cookie_token`]:
          (r) => r.json().user.cookie_token !== undefined,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response type`]: (
          r
        ) => r.json().user.type === "OWNER_TYPE_USER",
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response create_time`]:
          (r) => r.json().user.create_time !== undefined,
        [`GET /${constant.mgmtVersion}/users/${defaultUid} response update_time`]:
          (r) => r.json().user.update_time !== undefined,
      }
    );
  });

  var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
  group(`Management API: Look up non-exist user by permalink`, () => {
    check(
      http.request("GET", `${constant.mgmtHost}/users/${nonExistUID}:lookUp`),
      {
        [`GET /${constant.mgmtVersion}/users/${nonExistUID} response status 404`]:
          (r) => r.status === 404,
      }
    );
  });
}

export function CheckUpdate() {
  group(`Management API: Update default user`, () => {
    var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
    var userUpdate = {
      name: "users/non-exist",
      uid: nonExistUID,
      email: "test@foo.bar",
      company_name: "company",
      role: "Manager",
      newsletter_subscription: false,
      cookie_token: "f5730f62-7026-4e11-917a-d890da315d3b",
      type: "OWNER_TYPE_ORGANIZATION",
      create_time: "2000-01-01T00:00:00.000000Z",
      update_time: "2000-01-01T00:00:00.000000Z",
    };

    var res = http.request(
      "GET",
      `${constant.mgmtHost}/users/${constant.defaultUser.id}`
    );

    check(
      http.request(
        "PATCH",
        `${constant.mgmtHost}/users/${constant.defaultUser.id}`,
        JSON.stringify(userUpdate),
        {headers: {"Content-Type": "application/json"}}
      ),
      {
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response 200`]:
          (r) => r.status === 200,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response name unchanged`]:
          (r) => r.json().user.name === res.json().user.name,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response uid unchanged`]:
          (r) => r.json().user.uid === res.json().user.uid,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response id unchanged`]:
          (r) => r.json().user.id === res.json().user.id,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response email updated`]:
          (r) => r.json().user.email === userUpdate.email,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response company_name updated`]:
          (r) => r.json().user.company_name === userUpdate.company_name,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response role updated`]:
          (r) => r.json().user.role === userUpdate.role,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response newsletter_subscription updated`]:
          (r) =>
            r.json().user.newsletter_subscription ===
            userUpdate.newsletter_subscription,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response cookie_token updated`]:
            (r) => r.json().user.cookie_token === userUpdate.cookie_token,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response type unchanged`]:
          (r) => r.json().user.type === res.json().user.type,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response create_time unchanged`]:
          (r) => r.json().user.create_time === res.json().user.create_time,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response update_time updated`]:
          (r) => r.json().user.update_time !== res.json().user.update_time,
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response update_time not updated with request value`]:
          (r) => r.json().user.update_time !== userUpdate.update_time,
      }
    );

    // Restore to default user
    check(
      http.request(
        "PATCH",
        `${constant.mgmtHost}/users/${constant.defaultUser.id}`,
        JSON.stringify(constant.defaultUser),
        {headers: {"Content-Type": "application/json"}}
      ),
      {
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response status 200`]:
          (r) => r.status === 200,
      }
    );
    check(
      http.request(
        "GET",
        `${constant.mgmtHost}/users/${constant.defaultUser.id}`
      ),
      {
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id} response status 200`]:
          (r) => r.status === 200,
      }
    );
  });

  group(`Management API: Update user with a non-exist role`, () => {
    var nonExistRole = "non-exist-role";
    var userUpdate = {
      role: nonExistRole,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtHost}/users/${constant.defaultUser.id}`,
        JSON.stringify(userUpdate),
        {headers: {"Content-Type": "application/json"}}
      ),
      {
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response status 400`]:
          (r) => r.status === 400,
      }
    );
  });

  group(`Management API: Update user ID [not allowed]`, () => {
    var userUpdate = {
      id: `test_${randomString(10)}`,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtHost}/users/${constant.defaultUser.id}`,
        JSON.stringify(userUpdate),
        {headers: {"Content-Type": "application/json"}}
      ),
      {
        [`PATCH /${constant.mgmtVersion}/users/${constant.defaultUser.id} response status 400`]:
          (r) => r.status === 400,
      }
    );
  });

  var nonExistID = "non-exist";
  group(`Management API: Update non-exist user`, () => {
    check(http.request("PATCH", `${constant.mgmtHost}/users/${nonExistID}`), {
      [`PATCH /${constant.mgmtVersion}/users/${nonExistID} response status 404`]:
        (r) => r.status === 404,
    });
  });
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
  group("Management API: Create user with invalid id", () => {
    check(
      http.request(
        "POST",
        `${constant.mgmtHost}/users`,
        JSON.stringify({id: "local user"}),
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
      http.request(
        "POST",
        `${constant.mgmtHost}/users`,
        JSON.stringify({
          id: "local-user",
        }),
        {
          headers: {"Content-Type": "application/json"},
        }
      ),
      {
        [`POST /${constant.mgmtVersion}/users response status 400`]:
          (r) => r.status === 400,
      }
    );

    check(
      http.request(
        "POST",
        `${constant.mgmtHost}/users`,
        JSON.stringify({
          id: "local-user",
          email: "local-user@instill.tech"
        }),
        {
          headers: {"Content-Type": "application/json"},
        }
      ),
      {
        [`POST /${constant.mgmtVersion}/users response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}

export function CheckDelete() {
  group(`Management API: Delete user`, () => {
    check(
      http.request(
        "DELETE",
        `${constant.mgmtHost}/users/${constant.defaultUser.id}`,
        JSON.stringify({}),
        {
          headers: {"Content-Type": "application/json"},
        }
      ),
      {
        [`DELETE /${constant.mgmtVersion}/users/${constant.defaultUser.id} response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}
