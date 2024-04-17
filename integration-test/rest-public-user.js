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
        [`GET /${constant.mgmtVersion}/user response customer_id`]:
          (r) => r.json().user.customer_id !== undefined,
        [`GET /${constant.mgmtVersion}/user response display_name`]:
          (r) => r.json().user.profile.display_name !== undefined,
        [`GET /${constant.mgmtVersion}/user response company_name`]:
          (r) => r.json().user.profile.company_name !== undefined,
        [`GET /${constant.mgmtVersion}/user response role`]:
          (r) => r.json().user.role !== undefined,
        [`GET /${constant.mgmtVersion}/user response newsletter_subscription`]:
          (r) => r.json().user.newsletter_subscription !== undefined,
        [`GET /${constant.mgmtVersion}/user response create_time`]:
          (r) => r.json().user.create_time !== undefined,
        [`GET /${constant.mgmtVersion}/user response update_time`]:
          (r) => r.json().user.update_time !== undefined,
      }
    )
  })
}

export function CheckPublicPatchAuthenticatedUser(header) {
  group(`Management Public API: Update authenticated user`, () => {
    var userUpdate = {
      email: "test@foo.bar",
      customer_id: "new_customer_id",
      profile: {
        display_name: "test",
        company_name: "company",
      },
      role: "ai-engineer",
      newsletter_subscription: true,
      create_time: "2000-01-01T00:00:00.000000Z",
      update_time: "2000-01-01T00:00:00.000000Z",
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
        [`PATCH /${constant.mgmtVersion}/user response customer_id unchanged`]:
          (r) => r.json().user.customer_id === res.json().user.customer_id,
        [`PATCH /${constant.mgmtVersion}/user response display_name updated`]:
          (r) => r.json().user.profile.display_name === userUpdate.profile.display_name,
        [`PATCH /${constant.mgmtVersion}/user response company_name updated`]:
          (r) => r.json().user.profile.company_name === userUpdate.profile.company_name,
        [`PATCH /${constant.mgmtVersion}/user response role updated`]:
          (r) => r.json().user.role === userUpdate.role,
        [`PATCH /${constant.mgmtVersion}/user response newsletter_subscription updated`]:
          (r) => r.json().user.newsletter_subscription === userUpdate.newsletter_subscription,
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
          (r) => r.json().pipeline_trigger_records !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response has next_page_token`]:
          (r) => r.json().total_size !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers response has total_size`]:
          (r) => r.json().next_page_token !== undefined,
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
          (r) => r.json().pipeline_trigger_records.length === 0,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers with filter response next_page_token is empty`]:
          (r) => r.json().next_page_token === emptyPipelineTriggerRecordResponse.nextPageToken,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/triggers with filter response total_size is 0`]:
          (r) => r.json().total_size === emptyPipelineTriggerRecordResponse.totalSize,
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
          (r) => r.json().pipeline_trigger_table_records !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response has next_page_token`]:
          (r) => r.json().total_size !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables response has total_size`]:
          (r) => r.json().next_page_token !== undefined,
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
          (r) => r.json().pipeline_trigger_table_records.length === 0,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables with filter response next_page_token is empty`]:
          (r) => r.json().next_page_token === emptyPipelineTriggerTableRecordResponse.nextPageToken,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/tables with filter response total_size is 0`]:
          (r) => r.json().total_size === emptyPipelineTriggerTableRecordResponse.totalSize,
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
          (r) => r.json().pipeline_trigger_chart_records !== undefined,
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
          (r) => r.json().pipeline_trigger_chart_records.length === 0,
      }
    )
  })
}
