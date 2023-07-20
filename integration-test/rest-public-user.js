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

export function CheckPublicQueryAuthenticatedUser() {
  group(`Management Public API: Get authenticated user`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/users/me`
      ),
      {
        [`GET /${constant.mgmtVersion}/users/me response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/users/me response name`]:
          (r) => r.json().user.name !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response uid is UUID`]:
          (r) => helper.isUUID(r.json().user.uid),
        [`GET /${constant.mgmtVersion}/users/me response id`]:
          (r) => r.json().user.id !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response id`]:
          (r) => r.json().user.id === constant.defaultUser.id,
        [`GET /${constant.mgmtVersion}/users/me response type`]:
          (r) => r.json().user.type === "OWNER_TYPE_USER",
        [`GET /${constant.mgmtVersion}/users/me response email`]:
          (r) => r.json().user.email !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response customer_id`]:
          (r) => r.json().user.customer_id !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response first_name`]:
          (r) => r.json().user.first_name !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response last_name`]:
          (r) => r.json().user.last_name !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response org_name`]:
          (r) => r.json().user.org_name !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response role`]:
          (r) => r.json().user.role !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response newsletter_subscription`]:
          (r) => r.json().user.newsletter_subscription !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response cookie_token`]:
          (r) => r.json().user.cookie_token !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response create_time`]:
          (r) => r.json().user.create_time !== undefined,
        [`GET /${constant.mgmtVersion}/users/me response update_time`]:
          (r) => r.json().user.update_time !== undefined,
      }
    )
  })
}

export function CheckPublicPatchAuthenticatedUser() {
  group(`Management Public API: Update authenticated user`, () => {
    var userUpdate = {
      type: "OWNER_TYPE_ORGANIZATION",
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

    var res = http.request(
      "GET",
      `${constant.mgmtPublicHost}/users/me`
    );

    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/users/me`,
        JSON.stringify(userUpdate), constant.restParams),
      {
        [`PATCH /${constant.mgmtVersion}/users/me response 200`]:
          (r) => r.status === 200,
        [`PATCH /${constant.mgmtVersion}/users/me response name unchanged`]:
          (r) => r.json().user.name === res.json().user.name,
        [`PATCH /${constant.mgmtVersion}/users/me response uid unchanged`]:
          (r) => r.json().user.uid === res.json().user.uid,
        [`PATCH /${constant.mgmtVersion}/users/me response id unchanged`]:
          (r) => r.json().user.id === res.json().user.id,
        [`PATCH /${constant.mgmtVersion}/users/me response type unchanged`]:
          (r) => r.json().user.type === res.json().user.type,
        [`PATCH /${constant.mgmtVersion}/users/me response email updated`]:
          (r) => r.json().user.email === userUpdate.email,
        [`PATCH /${constant.mgmtVersion}/users/me response customer_id unchanged`]:
          (r) => r.json().user.customer_id === res.json().user.customer_id,
        [`PATCH /${constant.mgmtVersion}/users/me response first_name updated`]:
          (r) => r.json().user.first_name === userUpdate.first_name,
        [`PATCH /${constant.mgmtVersion}/users/me response last_name updated`]:
          (r) => r.json().user.last_name === userUpdate.last_name,
        [`PATCH /${constant.mgmtVersion}/users/me response org_name updated`]:
          (r) => r.json().user.org_name === userUpdate.org_name,
        [`PATCH /${constant.mgmtVersion}/users/me response role updated`]:
          (r) => r.json().user.role === userUpdate.role,
        [`PATCH /${constant.mgmtVersion}/users/me response newsletter_subscription updated`]:
          (r) => r.json().user.newsletter_subscription === userUpdate.newsletter_subscription,
        [`PATCH /${constant.mgmtVersion}/users/me response cookie_token updated`]:
          (r) => r.json().user.cookie_token === userUpdate.cookie_token,
        [`PATCH /${constant.mgmtVersion}/users/me response create_time unchanged`]:
          (r) => r.json().user.create_time === res.json().user.create_time,
        [`PATCH /${constant.mgmtVersion}/users/me response update_time updated`]:
          (r) => r.json().user.update_time !== res.json().user.update_time,
        [`PATCH /${constant.mgmtVersion}/users/me response update_time not updated with request value`]:
          (r) => r.json().user.update_time !== userUpdate.update_time,
      }
    );

    // Restore to default user
    check(
      http.request(
        "PATCH",
        `${constant.mgmtPublicHost}/users/me`,
        JSON.stringify(constant.defaultUser), constant.restParams),
      {
        [`PATCH /${constant.mgmtVersion}/users/me response status 200`]:
          (r) => r.status === 200,
      }
    );
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/users/me`
      ),
      {
        [`GET /${constant.mgmtVersion}/users/me response status 200`]:
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
        `${constant.mgmtPublicHost}/users/me`,
        JSON.stringify(userUpdate), constant.restParams),
      {
        [`PATCH /${constant.mgmtVersion}/users/me response status 400`]:
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
        `${constant.mgmtPublicHost}/users/me`,
        JSON.stringify(userUpdate), constant.restParams),
      {
        [`PATCH /${constant.mgmtVersion}/users/me response status 400`]:
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
        `${constant.mgmtPublicHost}/users/me`,
        JSON.stringify(userUpdate), constant.restParams),
      {
        [`PATCH /${constant.mgmtVersion}/users/me response status 400`]:
          (r) => r.status === 400,
      }
    );
  });
}

export function CheckPublicCreateToken() {
  group(`Management Public API: Create API token`, () => {
    check(
      http.request(
        "POST",
        `${constant.mgmtPublicHost}/tokens`,
        JSON.stringify(constant.testToken),
        constant.restParams
      ),
      {
        [`POST /${constant.mgmtVersion}/tokens response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}

export function CheckPublicListTokens() {
  group(`Management Public API: List API tokens`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/tokens`,
        JSON.stringify({}),
        constant.restParams
      ),
      {
        [`GET /${constant.mgmtVersion}/tokens response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}

export function CheckPublicGetToken() {
  group(`Management Public API: Get API token`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/tokens/${constant.testToken.id}`,
        JSON.stringify({}),
        constant.restParams
      ),
      {
        [`GET /${constant.mgmtVersion}/tokens/${constant.testToken.id} response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}

export function CheckPublicDeleteToken() {
  group(`Management Public API: Delete API token`, () => {
    check(
      http.request(
        "DELETE",
        `${constant.mgmtPublicHost}/tokens/${constant.testToken.id}`,
        JSON.stringify({}),
        constant.restParams
      ),
      {
        [`DELETE /${constant.mgmtVersion}/tokens/${constant.testToken.id} response status 501 [not implemented]`]:
          (r) => r.status === 501,
      }
    );
  });
}

export function CheckPublicMetrics() {
  group(`Management Public API: List Pipeline Trigger Records`, () => {

    let emptyPipelineTriggerRecordResponse = {
      "pipelineTriggerRecords": [],
      "nextPageToken": "",
      "totalSize": "0"
    }

    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/triggers`
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
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/triggers?filter=trigger_mode=MODE_SYNC%20AND%20pipeline_id=%22a%22`
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
  group(`Management Public API: List Pipeline Trigger Chart Records`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/charts`
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
        `${constant.mgmtPublicHost}/metrics/vdp/pipeline/charts?filter=trigger_mode=MODE_SYNC%20AND%20pipeline_id=%22a%22`
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts with filter response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/pipeline/charts with filter response pipelineTriggerRecords length is 0`]:
          (r) => r.json().pipeline_trigger_chart_records.length === 0,
      }
    )
  })
  group(`Management Public API: List Connector Execute Records`, () => {

    let emptyConnectorExecuteRecordResponse = {
      "connectorExecuteRecords": [],
      "nextPageToken": "",
      "totalSize": "0"
    }

    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/connector/executes`
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/executes response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/executes response has connectorExecuteRecords`]:
          (r) => r.json().connector_execute_records !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/executes response has next_page_token`]:
          (r) => r.json().total_size !== undefined,
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/executes response has total_size`]:
          (r) => r.json().next_page_token !== undefined,
      }
    )
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/connector/executes?filter=status=STATUS_COMPLETED%20AND%20connector_id=%22a%22`
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/executes with filter response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/executes with filter response connectorExecuteRecords length is 0`]:
          (r) => r.json().connector_execute_records.length === 0,
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/executes with filter response next_page_token is empty`]:
          (r) => r.json().next_page_token === emptyConnectorExecuteRecordResponse.nextPageToken,
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/executes with filter response total_size is 0`]:
          (r) => r.json().total_size === emptyConnectorExecuteRecordResponse.totalSize,
      }
    )
  })
  group(`Management Public API: List Connector Execute Chart Records`, () => {
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/connector/charts`
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/charts response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/charts response has connectorExecuteChartRecords`]:
          (r) => r.json().connector_execute_chart_records !== undefined,
      }
    )
    check(
      http.request(
        "GET",
        `${constant.mgmtPublicHost}/metrics/vdp/connector/charts?filter=status=STATUS_COMPLETED%20AND%20connector_id=%22a%22`
      ),
      {
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/charts with filter response status is 200`]:
          (r) => r.status === 200,
        [`GET /${constant.mgmtVersion}/metrics/vdp/connector/charts with filter response connectorExecuteChartRecords length is 0`]:
          (r) => r.json().connector_execute_chart_records.length === 0,
      }
    )
  })
}
