import http from "k6/http";
import { check, group } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import * as constant from "./const.js";
import * as helper from "./helper.js";


export function CheckHealth() {
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

export function CheckPublicGetUser(header) {
  group(`Management Public API: Get authenticated user`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/user`, null, header
      ),
      {
        [`GET /${constant.mgmtVersion}/user response status is 200`]:
          (r) => r.status === 200,
        // AIP standard fields (1-6)
        [`GET /${constant.mgmtVersion}/user response name`]:
          (r) => r.json().user.name !== undefined,
        [`GET /${constant.mgmtVersion}/user response id`]:
          (r) => r.json().user.id !== undefined,
        [`GET /${constant.mgmtVersion}/user response id matches`]:
          (r) => r.json().user.id === constant.defaultUser.id,
        [`GET /${constant.mgmtVersion}/user response displayName`]:
          (r) => r.json().user.displayName !== undefined,
        [`GET /${constant.mgmtVersion}/user response slug`]:
          (r) => r.json().user.slug !== undefined,
        [`GET /${constant.mgmtVersion}/user response aliases`]:
          (r) => Array.isArray(r.json().user.aliases),
        // User-specific fields
        [`GET /${constant.mgmtVersion}/user response email`]:
          (r) => r.json().user.email !== undefined,
        [`GET /${constant.mgmtVersion}/user response profile.displayName`]:
          (r) => r.json().user.profile.displayName !== undefined,
        [`GET /${constant.mgmtVersion}/user response profile.companyName`]:
          (r) => r.json().user.profile.companyName !== undefined,
        [`GET /${constant.mgmtVersion}/user response role`]:
          (r) => r.json().user.role !== undefined,
        [`GET /${constant.mgmtVersion}/user response newsletterSubscription`]:
          (r) => r.json().user.newsletterSubscription !== undefined,
        [`GET /${constant.mgmtVersion}/user response createTime`]:
          (r) => r.json().user.createTime !== undefined,
        [`GET /${constant.mgmtVersion}/user response updateTime`]:
          (r) => r.json().user.updateTime !== undefined,
      }
    )
  })
}

export function CheckPublicPatchAuthenticatedUser(header) {
  group(`Management Public API: Update authenticated user`, () => {
    var userUpdate = {
      email: "test@foo.bar",
      profile: {
        displayName: "test",
        companyName: "company",
      },
      role: "ai-engineer",
      newsletterSubscription: true,
    };

    // Merge auth header with Content-Type for PATCH requests
    const patchHeader = {
      headers: {
        ...header.headers,
        "Content-Type": "application/json",
      },
    };

    var getRes = http.request(
      "GET",
      `${constant.mgmtPublicHost}/user`,
      null,
      header,
    );

    // Store original user for comparison
    const originalUser = getRes.json().user;

    var patchRes = http.request(
      "PATCH",
      `${constant.mgmtPublicHost}/user`,
      JSON.stringify(userUpdate),
      patchHeader,
    );

    // Debug: log error response if PATCH fails
    if (patchRes.status !== 200) {
      console.log(`PATCH /user failed with status ${patchRes.status}: ${patchRes.body}`);
    }

    check(patchRes, {
      [`PATCH /${constant.mgmtVersion}/user response 200`]:
        (r) => r.status === 200,
    });

    // Only check response fields if PATCH succeeded
    if (patchRes.status === 200) {
      const updatedUser = patchRes.json().user;
      check(patchRes, {
        // AIP immutable fields unchanged
        [`PATCH /${constant.mgmtVersion}/user response name unchanged`]:
          () => updatedUser.name === originalUser.name,
        [`PATCH /${constant.mgmtVersion}/user response id unchanged`]:
          () => updatedUser.id === originalUser.id,
        // Note: slug is derived from displayName, so it WILL change when profile.displayName changes
        // Updated fields
        [`PATCH /${constant.mgmtVersion}/user response email updated`]:
          () => updatedUser.email === userUpdate.email,
        [`PATCH /${constant.mgmtVersion}/user response profile.displayName updated`]:
          () => updatedUser.profile.displayName === userUpdate.profile.displayName,
        [`PATCH /${constant.mgmtVersion}/user response profile.companyName updated`]:
          () => updatedUser.profile.companyName === userUpdate.profile.companyName,
        [`PATCH /${constant.mgmtVersion}/user response role updated`]:
          () => updatedUser.role === userUpdate.role,
        [`PATCH /${constant.mgmtVersion}/user response newsletterSubscription updated`]:
          () => updatedUser.newsletterSubscription === userUpdate.newsletterSubscription,
        [`PATCH /${constant.mgmtVersion}/user response createTime unchanged`]:
          () => updatedUser.createTime === originalUser.createTime,
        [`PATCH /${constant.mgmtVersion}/user response updateTime updated`]:
          () => updatedUser.updateTime !== originalUser.updateTime,
      });
    }

    // Restore to default user (use defaultUserUpdate which excludes immutable fields)
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/user`,
        JSON.stringify(constant.defaultUserUpdate), patchHeader),

      {
        [`PATCH /${constant.mgmtVersion}/user response status 200`]:
          (r) => r.status === 200,
      }
    );
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/user`,
        null,
        header,
      ),
      {
        [`GET /${constant.mgmtVersion}/user response status 200`]:
          (r) => r.status === 200,
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
        JSON.stringify(userUpdate), header),
      {
        [`PATCH /${constant.mgmtVersion}/user response status 400`]:
          (r) => r.status === 400,
      }
    );
  });

  group(`Management Public API: Update authenticated user UID [not allowed]`, () => {
    var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
    var userUpdate = {
      uid: nonExistUID,
    };
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/user`,
        JSON.stringify(userUpdate), header),
      {
        [`PATCH /${constant.mgmtVersion}/user response status 400`]:
          (r) => r.status === 400,
      }
    );
  });
}

export function CheckPublicCreateToken(header) {
  group(`Management Public API: Create API token`, () => {
    // Add Content-Type header for POST request
    const postHeader = {
      headers: {
        ...header.headers,
        "Content-Type": "application/json",
      },
    };

    // First, try to delete any existing test token to avoid 409 conflict
    // Note: Backend expects name format users/{user_id}/tokens/{token_id} but
    // proto route uses tokens/{token_id}. This is a backend/proto mismatch.
    http.request(
      "DELETE",
      `${constant.mgmtPublicHost}/tokens/${constant.testToken.id}`,
      null,
      header,
    );

    var res = http.request(
      "POST",
      `${constant.mgmtPublicHost}/tokens`,
      JSON.stringify(constant.testToken),
      postHeader,
    );

    // Debug: log error response if POST fails
    if (res.status !== 201) {
      console.log(`POST /tokens failed with status ${res.status}: ${res.body}`);
    }

    check(res, {
      [`POST /${constant.mgmtVersion}/tokens response status 201`]:
        (r) => r.status === 201,
    });
  });
}

export function CheckPublicListTokens(header) {
  group(`Management Public API: List API tokens`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/tokens`,
        null,
        header,
      ),
      {
        [`GET /${constant.mgmtVersion}/tokens response status 200`]:
          (r) => r.status === 200,
      }
    );
  });
}

export function CheckPublicGetToken(header) {
  group(`Management Public API: Get API token`, () => {
    // Note: Backend expects name format users/{user_id}/tokens/{token_id} but
    // proto route uses tokens/{token_id}. This is a backend/proto mismatch
    // that causes 500 errors until the backend handler is fixed.
    var res = http.request(
      "GET",
      `${constant.mgmtPublicHost}/tokens/${constant.testToken.id}`,
      null,
      header,
    );

    // Debug: log error response if GET fails
    if (res.status !== 200) {
      console.log(`GET /tokens/${constant.testToken.id} failed with status ${res.status}: ${res.body}`);
    }

    check(res, {
      [`GET /${constant.mgmtVersion}/tokens/${constant.testToken.id} response status 200`]:
        (r) => r.status === 200,
    });
  });
}

export function CheckPublicDeleteToken(header) {
  group(`Management Public API: Delete API token`, () => {
    // Note: Backend expects name format users/{user_id}/tokens/{token_id} but
    // proto route uses tokens/{token_id}. This is a backend/proto mismatch
    // that causes 500 errors until the backend handler is fixed.
    var res = http.request(
      "DELETE",
      `${constant.mgmtPublicHost}/tokens/${constant.testToken.id}`,
      null,
      header,
    );

    // Debug: log error response if DELETE fails
    if (res.status !== 204) {
      console.log(`DELETE /tokens/${constant.testToken.id} failed with status ${res.status}: ${res.body}`);
    }

    check(res, {
      [`DELETE /${constant.mgmtVersion}/tokens/${constant.testToken.id} response status 204`]:
        (r) => r.status === 204,
    });
  });
}

export function CheckPublicGetRemainingCredit(header) {
  group(`Management Public API: Get remaining credit`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/users/${constant.defaultUser.id}/credit`,
        JSON.stringify({}),
        header
      ),
      {
        // Although grpc-gateway returns 501 unimplemented, in the public
        // gateway the endpoint shouldn't be defined, yielding a 404.
        [`GET /${constant.mgmtVersion}/users/${constant.defaultUser.id}/credit response status 404`]:
          (r) => r.status === 404,
      }
    );
  });
}

export function CheckPublicMetrics(header) {
  let pipeline_id = randomString(10);

  group(`Management Public API: List Pipeline Trigger Records`, () => {
    let emptyPipelineTriggerRecordResponse = {
      "pipelineTriggerRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    };

    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/triggers`,
        null,
        header
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response has pipelineTriggerRecords`]:
          (r) => r.json().pipelineTriggerRecords !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response has nextPageToken`]:
          (r) => r.json().nextPageToken !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response has totalSize`]:
          (r) => r.json().totalSize !== undefined,
      }
    );

    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/triggers?filter=triggerMode=MODE_SYNC%20AND%20pipelineId=%22${pipeline_id}%22`,
        null,
        header
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers with filter response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers with filter response pipelineTriggerRecords length is 0`]:
          (r) => r.json().pipelineTriggerRecords.length === 0,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers with filter response nextPageToken is empty`]:
          (r) => r.json().nextPageToken === emptyPipelineTriggerRecordResponse.nextPageToken,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers with filter response totalSize is 0`]:
          (r) => r.json().totalSize === emptyPipelineTriggerRecordResponse.totalSize,
      }
    );
  });

  group(`Management Public API: List Pipeline Trigger Table Records`, () => {
    let emptyPipelineTriggerTableRecordResponse = {
      "pipelineTriggerTableRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    };

    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/tables`,
        null,
        header
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response has pipelineTriggerTableRecords`]:
          (r) => r.json().pipelineTriggerTableRecords !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response has nextPageToken`]:
          (r) => r.json().nextPageToken !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response has totalSize`]:
          (r) => r.json().totalSize !== undefined,
      }
    );
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/tables?filter=pipelineId=%22${pipeline_id}%22`,
        null,
        header
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables with filter response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables with filter response pipelineTriggerTableRecords length is 0`]:
          (r) => r.json().pipelineTriggerTableRecords.length === 0,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables with filter response nextPageToken is empty`]:
          (r) => r.json().nextPageToken === emptyPipelineTriggerTableRecordResponse.nextPageToken,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables with filter response totalSize is 0`]:
          (r) => r.json().totalSize === emptyPipelineTriggerTableRecordResponse.totalSize,
      }
    );
  });

  group(`Management Public API: List Pipeline Trigger Chart Records`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/charts`,
        null,
        header
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts response has pipelineTriggerRecords`]:
          (r) => r.json().pipelineTriggerChartRecords !== undefined,
      }
    );
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/charts?filter=triggerMode=MODE_SYNC%20AND%20pipelineId=%22${pipeline_id}%22`,
        null,
        header
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts with filter response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts with filter response pipelineTriggerRecords length is 0`]:
          (r) => r.json().pipelineTriggerChartRecords.length === 0,
      }
    );
  });
}
