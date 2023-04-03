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
        (r) => r.json().total_size === "1",
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
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response type`]:
          (r) => r.json().user.type === "OWNER_TYPE_USER",
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response email`]:
          (r) => r.json().user.email !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response customer_id`]:
          (r) => r.json().user.customer_id !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response first_name`]:
          (r) => r.json().user.first_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response last_name`]:
          (r) => r.json().user.last_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response org_name`]:
          (r) => r.json().user.org_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response role`]:
          (r) => r.json().user.role !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response newsletter_subscription`]:
          (r) => r.json().user.newsletter_subscription !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response cookie_token`]:
          (r) => r.json().user.cookie_token !== undefined,
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
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response type`]:
          (r) => r.json().user.type === "OWNER_TYPE_USER",
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response email`]:
          (r) => r.json().user.email !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response customer_id`]:
          (r) => r.json().user.customer_id !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response first_name`]:
          (r) => r.json().user.first_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response last_name`]:
          (r) => r.json().user.last_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response org_name`]:
          (r) => r.json().user.org_name !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response role`]:
          (r) => r.json().user.role !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response newsletter_subscription`]:
          (r) => r.json().user.newsletter_subscription !== undefined,
        [`GET /${constant.mgmtVersion}/admin/users/${defaultUid} response cookie_token`]:
          (r) => r.json().user.cookie_token !== undefined,
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

export function CheckPrivateUpdateUserAdmin() {
  group(`Management Private API: Update default user`, () => {
    var userUpdate = {
      name: `users/${constant.defaultUser.id}`,
      type: "OWNER_TYPE_ORGANIZATION",
      email: "test@foo.bar",
      customer_id: "new_customer_id",
      first_name: "test",
      last_name: "foo",
      org_name: "company",
      role: "ai-researcher",
      newsletter_subscription: true,
      cookie_token: "f5730f62-7026-4e11-917a-d890da315d3b",
      create_time: "2000-01-01T00:00:00.000000Z",
      update_time: "2000-01-01T00:00:00.000000Z",
    };

    var res = http.request(
      "GET",
      `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`
    );

    check(
      http.request(
        "PATCH",
        `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`,
        JSON.stringify(userUpdate), constant.params),
      {
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response 200`]:
          (r) => r.status === 200,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response name unchanged`]:
          (r) => r.json().user.name === res.json().user.name,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response uid unchanged`]:
          (r) => r.json().user.uid === res.json().user.uid,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response id unchanged`]:
          (r) => r.json().user.id === res.json().user.id,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response type unchanged`]:
          (r) => r.json().user.type === res.json().user.type,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response email updated`]:
          (r) => r.json().user.email === userUpdate.email,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response customer_id unchanged`]:
          (r) => r.json().user.customer_id === res.json().user.customer_id,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response first_name updated`]:
          (r) => r.json().user.first_name === userUpdate.first_name,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response last_name updated`]:
          (r) => r.json().user.last_name === userUpdate.last_name,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response org_name updated`]:
          (r) => r.json().user.org_name === userUpdate.org_name,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response role updated`]:
          (r) => r.json().user.role === userUpdate.role,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response newsletter_subscription updated`]:
          (r) => r.json().user.newsletter_subscription === userUpdate.newsletter_subscription,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response cookie_token updated`]:
          (r) => r.json().user.cookie_token === userUpdate.cookie_token,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response create_time unchanged`]:
          (r) => r.json().user.create_time === res.json().user.create_time,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response update_time updated`]:
          (r) => r.json().user.update_time !== res.json().user.update_time,
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response update_time not updated with request value`]:
          (r) => r.json().user.update_time !== userUpdate.update_time,
      }
    );

    // Restore to default user
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`,
        JSON.stringify(constant.defaultUser), constant.params),
      {
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response status 200`]:
          (r) => r.status === 200,
      }
    );
    check(
      http.request(
        "GET",
        `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`
      ),
      {
        [`GET /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response status 200`]:
          (r) => r.status === 200,
      }
    );
  });

  group(`Management Private API: Update user with a non-exist role`, () => {
    var nonExistRole = "non-exist-role";
    var userUpdate = {
      role: nonExistRole,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`,
        JSON.stringify(userUpdate), constant.params),
      {
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response status 400`]:
          (r) => r.status === 400,
      }
    );
  });

  group(`Management Private API: Update user ID [not allowed]`, () => {
    var userUpdate = {
      id: `test_${randomString(10)}`,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`,
        JSON.stringify(userUpdate), constant.params),
      {
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response status 400`]:
          (r) => r.status === 400,
      }
    );
  });

  group(`Management Private API: Update user UID [not allowed]`, () => {
    var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
    var userUpdate = {
      uid: nonExistUID,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`,
        JSON.stringify(userUpdate), constant.params),
      {
        [`PATCH /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response status 400`]:
          (r) => r.status === 400,
      }
    )
  })

  var nonExistID = "non-exist";
  group(`Management Private API: Update non-exist user`, () => {
    check(http.request("PATCH", `${constant.mgmtPrivateHost}/admin/users/${nonExistID}`), {
      [`PATCH /${constant.mgmtVersion}/admin/users/${nonExistID} response status 404`]:
        (r) => r.status === 404,
    });
  });
}

export function CheckPrivateCreateUserAdmin() {
  group("Management Private API: Create user with UUID as id", () => {
    check(
      http.request(
        "POST",
        `${constant.mgmtPrivateHost}/admin/users`,
        JSON.stringify({ id: "2a06c2f7-8da9-4046-91ea-240f88a5d000" }),
        {
          headers: { "Content-Type": "application/json" },
        }
      ),
      {
        [`POST /${constant.mgmtVersion}/admin/users response status 400`]: (r) =>
          r.status === 400,
      }
    );
  });
  group("Management Private API: Create user with invalid id", () => {
    check(
      http.request(
        "POST",
        `${constant.mgmtPrivateHost}/admin/users`,
        JSON.stringify({ id: "local user" }),
        {
          headers: { "Content-Type": "application/json" },
        }
      ),
      {
        [`POST /${constant.mgmtVersion}/admin/users response status 400`]: (r) =>
          r.status === 400,
      }
    );
  });
  group("Management Private API: Create user", () => {
    check(
      http.request(
        "POST",
        `${constant.mgmtPrivateHost}/admin/users`,
        JSON.stringify({
          id: "local-user",
        }),
        {
          headers: { "Content-Type": "application/json" },
        }
      ),
      {
        [`POST /${constant.mgmtVersion}/admin/users response status 400`]:
          (r) => r.status === 400,
      }
    );

    check(
      http.request(
        "POST",
        `${constant.mgmtPrivateHost}/admin/users`,
        JSON.stringify({
          id: "local-user-2",
          email: "local-user-2@instill.tech"
        }),
        {
          headers: { "Content-Type": "application/json" },
        }
      ),
      {
        [`POST /${constant.mgmtVersion}/admin/users response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}

export function CheckPrivateDeleteUserAdmin() {
  group(`Management Private API: Delete user`, () => {
    check(
      http.request(
        "DELETE",
        `${constant.mgmtPrivateHost}/admin/users/${constant.defaultUser.id}`,
        JSON.stringify({}),
        {
          headers: { "Content-Type": "application/json" },
        }
      ),
      {
        [`DELETE /${constant.mgmtVersion}/admin/users/${constant.defaultUser.id} response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}

export function CheckPrivateValidateToken() {
  group(`Management Private API: Validate API token`, () => {
    check(
      http.request(
        "POST",
        `${constant.mgmtPrivateHost}/tokens/${constant.testToken.id}/validate`,
        JSON.stringify({}),
        {
          headers: { "Content-Type": "application/json" },
        }
      ),
      {
        [`POST /${constant.mgmtVersion}/tokens/${constant.testToken.id} response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}
