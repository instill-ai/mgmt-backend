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
        [`GET /${constant.mgmtVersion}/user response name`]:
          (r) => r.json().user.name !== undefined,
        [`GET /${constant.mgmtVersion}/user response uid is UUID`]:
          (r) => helper.isUUID(r.json().user.uid),
        [`GET /${constant.mgmtVersion}/user response id`]:
          (r) => r.json().user.id !== undefined,
        [`GET /${constant.mgmtVersion}/user response id`]:
          (r) => r.json().user.id === constant.defaultUser.id,
        [`GET /${constant.mgmtVersion}/user response email`]:
          (r) => r.json().user.email !== undefined,
        [`GET /${constant.mgmtVersion}/user response customerId`]:
          (r) => r.json().user.customerId !== undefined,
        [`GET /${constant.mgmtVersion}/user response displayName`]:
          (r) => r.json().user.profile.displayName !== undefined,
        [`GET /${constant.mgmtVersion}/user response companyName`]:
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
      customerId: "new_customer_id",
      profile: {
        displayName: "test",
        companyName: "company",
      },
      role: "ai-engineer",
      newsletterSubscription: true,
      createTime: "2000-01-01T00:00:00.000000Z",
      updateTime: "2000-01-01T00:00:00.000000Z",
    };

    var res = http.request(
      "GET",
      `${constant.mgmtPublicHost}/user`,
      null,
      header,
    );

    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/user`,
        JSON.stringify(userUpdate), header),
      {
        [`PATCH /${constant.mgmtVersion}/user response 200`]:
          (r) => r.status === 200,
        [`PATCH /${constant.mgmtVersion}/user response name unchanged`]:
          (r) => r.json().user.name === res.json().user.name,
        [`PATCH /${constant.mgmtVersion}/user response uid unchanged`]:
          (r) => r.json().user.uid === res.json().user.uid,
        [`PATCH /${constant.mgmtVersion}/user response id unchanged`]:
          (r) => r.json().user.id === res.json().user.id,
        [`PATCH /${constant.mgmtVersion}/user response email updated`]:
          (r) => r.json().user.email === userUpdate.email,
        [`PATCH /${constant.mgmtVersion}/user response customerId unchanged`]:
          (r) => r.json().user.customerId === res.json().user.customerId,
        [`PATCH /${constant.mgmtVersion}/user response displayName updated`]:
          (r) => r.json().user.profile.displayName === userUpdate.profile.displayName,
        [`PATCH /${constant.mgmtVersion}/user response companyName updated`]:
          (r) => r.json().user.profile.companyName === userUpdate.profile.companyName,
        [`PATCH /${constant.mgmtVersion}/user response role updated`]:
          (r) => r.json().user.role === userUpdate.role,
        [`PATCH /${constant.mgmtVersion}/user response newsletterSubscription updated`]:
          (r) => r.json().user.newsletterSubscription === userUpdate.newsletterSubscription,
        [`PATCH /${constant.mgmtVersion}/user response createTime unchanged`]:
          (r) => r.json().user.createTime === res.json().user.createTime,
        [`PATCH /${constant.mgmtVersion}/user response updateTime updated`]:
          (r) => r.json().user.updateTime !== res.json().user.updateTime,
        [`PATCH /${constant.mgmtVersion}/user response updateTime not updated with request value`]:
          (r) => r.json().user.updateTime !== userUpdate.updateTime,
      }
    );

    // Restore to default user
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/user`,
        JSON.stringify(constant.defaultUser), header),

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
    check(
      http.request(
        "POST",
        `${constant.mgmtPublicHost}/tokens`,
        JSON.stringify(constant.testToken),
        header,
      ),
      {
        [`POST /${constant.mgmtVersion}/tokens response status 201`]:
          (r) => r.status === 201,
      }
    );
  });
}

export function CheckPublicListTokens(header) {
  group(`Management Public API: List API tokens`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/tokens`,
        JSON.stringify({}),
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
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/tokens/${constant.testToken.id}`,
        JSON.stringify({}),
        header,
      ),
      {
        [`GET /${constant.mgmtVersion}/tokens/${constant.testToken.id} response status 200`]:
          (r) => r.status === 200,
      }
    );
  });
}

export function CheckPublicDeleteToken(header) {
  group(`Management Public API: Delete API token`, () => {
    check(
      http.request(
        "DELETE",
        `${constant.mgmtPublicHost}/tokens/${constant.testToken.id}`,
        JSON.stringify({}),
        header,
      ),
      {
        [`DELETE /${constant.mgmtVersion}/tokens/${constant.testToken.id} response status 204`]:
          (r) => r.status === 204,
      }
    );
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
  group(`Management Public API: List Pipeline Trigger Records`, () => {

    let pipeline_id = randomString(10)

    let emptyPipelineTriggerRecordResponse = {
      "pipelineTriggerRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    }

    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/triggers`,
        null,
        header,
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response has pipelineTriggerRecords`]:
          (r) => r.json().pipelineTriggerRecords !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response has nextPageToken`]:
          (r) => r.json().totalSize !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response has totalSize`]:
          (r) => r.json().nextPageToken !== undefined,
      }
    )
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/triggers?filter=trigger_mode=MODE_SYNC%20AND%20pipeline_id=%22${pipeline_id}%22`,
        null,
        header,
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
    )
  })
  group(`Management Public API: List Pipeline Trigger Table Records`, () => {

    let pipeline_id = randomString(10)

    let emptyPipelineTriggerTableRecordResponse = {
      "pipelineTriggerTableRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    }

    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/tables`,
        null,
        header,
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response has pipelineTriggerTableRecords`]:
          (r) => r.json().pipelineTriggerTableRecords !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response has nextPageToken`]:
          (r) => r.json().totalSize !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response has totalSize`]:
          (r) => r.json().nextPageToken !== undefined,
      }
    )
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/tables?filter=pipeline_id=%22${pipeline_id}%22`,
        null,
        header,
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
    )
  })
  group(`Management Public API: List Pipeline Trigger Chart Records`, () => {

    let pipeline_id = randomString(10)

    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/charts`,
        null,
        header,
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts response has pipelineTriggerRecords`]:
          (r) => r.json().pipelineTriggerChartRecords !== undefined,
      }
    )
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/charts?filter=trigger_mode=MODE_SYNC%20AND%20pipeline_id=%22${pipeline_id}%22`,
        null,
        header,
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts with filter response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts with filter response pipelineTriggerRecords length is 0`]:
          (r) => r.json().pipelineTriggerChartRecords.length === 0,
      }
    )
  })
}
