import grpc from 'k6/net/grpc';
import {
  check,
  group
} from "k6";
import {
  randomString
} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";
import * as constant from "./const.js";
import * as helper from "./helper.js";

const client = new grpc.Client();
client.load(['proto/core/mgmt/v1beta'], 'mgmt.proto');
client.load(['proto/core/mgmt/v1beta'], 'mgmt_public_service.proto');

export function CheckHealth() {
  // Health check
  group("Management API: Health check", () => {

    client.connect(constant.mgmtPublicGRPCHost, {
      plaintext: true
    });

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/Liveness', {}), {
      'core.mgmt.v1beta.MgmtPublicService/Liveness status': (r) => r && r.status == grpc.StatusOK,
    });

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/Readiness', {}), {
      'core.mgmt.v1beta.MgmtPublicService/Readiness status': (r) => r && r.status == grpc.StatusOK,
    });

    client.close();
  });
}

export function CheckPublicGetUser(header) {

  group(`Management Public API: Get authenticated user`, () => {

    client.connect(constant.mgmtPublicGRPCHost, {
      plaintext: true
    });

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/GetUser', {name: "users/me"}, header), {
      'core.mgmt.v1beta.MgmtPublicService/GetUser status': (r) => { return r && r.status == grpc.StatusOK },
      'core.mgmt.v1beta.MgmtPublicService/GetUser response name': (r) => r && r.message.user.name !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response uid is UUID': (r) => r && helper.isUUID(r.message.user.uid),
      'core.mgmt.v1beta.MgmtPublicService/GetUser response id': (r) => r && r.message.user.id !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response id': (r) => r && r.message.user.id === constant.defaultUser.id,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response email': (r) => r && r.message.user.email !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response customerId': (r) => r && r.message.user.customerId !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response firstName': (r) => r && r.message.user.firstName !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response lastName': (r) => r && r.message.user.lastName !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response orgName': (r) => r && r.message.user.orgName !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response role': (r) => r && r.message.user.role !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response newsletterSubscription': (r) => r && r.message.user.newsletterSubscription !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response cookieToken': (r) => r && r.message.user.cookieToken !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response createTime': (r) => r && r.message.user.createTime !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetUser response updateTime': (r) => r && r.message.user.updateTime !== undefined,
    });

    client.close();
  })
}

export function CheckPublicPatchAuthenticatedUser(header) {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Update authenticated user`, () => {
    var userUpdate = {
      name: `users/${constant.defaultUser.id}`,
      email: "test@foo.bar",
      customer_id: "new_customer_id",
      first_name: "test",
      last_name: "foo",
      org_name: "company",
      role: "ai-engineer",
      newsletter_subscription: true,
      cookie_token: "f5730f62-7026-4e11-917a-d890da315d3b",
    };

    var res = client.invoke('core.mgmt.v1beta.MgmtPublicService/GetUser', {name: "users/me"}, header)

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "email,firstName,lastName,orgName,role,newsletterSubscription,cookieToken"
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response name unchanged': (r) => r && r.message.user.name === res.message.user.name,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response uid unchanged': (r) => r && r.message.user.uid === res.message.user.uid,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response id unchanged': (r) => r && r.message.user.id === res.message.user.id,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response email updated': (r) => r && r.message.user.email === userUpdate.email,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response customerId unchanged': (r) => r && r.message.user.customerId === res.message.user.customerId,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response firstName updated': (r) => r && r.message.user.firstName === userUpdate.first_name,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response lastName updated': (r) => r && r.message.user.lastName === userUpdate.last_name,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response orgName updated': (r) => r && r.message.user.orgName === userUpdate.org_name,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response role updated': (r) => r && r.message.user.role === userUpdate.role,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response newsletterSubscription updated': (r) => r && r.message.user.newsletterSubscription === userUpdate.newsletter_subscription,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response cookieToken updated': (r) => r && r.message.user.cookieToken === userUpdate.cookie_token,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response createTime unchanged': (r) => r && r.message.user.createTime === res.message.user.createTime,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response updateTime updated': (r) => r && r.message.user.updateTime !== res.message.user.updateTime,
    });

    // Restore to default user
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser', {
      user: constant.defaultUser,
      update_mask: "email,firstName,lastName,orgName,role,newsletterSubscription,cookieToken"
    }, header), {
      [`[restore the default user] core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser status`]: (r) => r && r.status == grpc.StatusOK,
    });

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/GetUser', {name: "users/me"}, header), {
      'core.mgmt.v1beta.MgmtPublicService/GetUser status': (r) => r && r.status == grpc.StatusOK,
    });
  });

  group(`Management Public API: Update authenticated user ID [not allowed]`, () => {
    var userUpdate = {
      id: `test_${randomString(10)}`,
    };

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "id"
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser update ID StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

  });

  group(`Management Public API: Update authenticated user UID [not allowed]`, () => {
    var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
    var userUpdate = {
      uid: nonExistUID,
    };

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "uid"
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser nonExistUID StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });
  });

  client.close();
}

export function CheckPublicCreateToken(header) {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Create API token`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/CreateToken', {
      token: {
        id: `${constant.testToken.id}`,
        ttl: 86400
      }
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/CreateToken status StatusOK': (r) => r && r.status == grpc.StatusOK,
    });

  });

  client.close();
}

export function CheckPublicListTokens(header) {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: List API tokens`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListTokens', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListTokens status StatusOK': (r) => r && r.status == grpc.StatusOK,
    });

  });

  client.close();
}

export function CheckPublicGetToken(header) {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Get API token`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/GetToken', {
      name: `tokens/${constant.testToken.id}`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/GetToken status StatusOK': (r) => r && r.status == grpc.StatusOK,
    });

  });

  client.close();
}

export function CheckPublicDeleteToken(header) {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Delete API token`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/DeleteToken', {
      name: `tokens/${constant.testToken.id}`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/DeleteToken status StatusOK': (r) => r && r.status == grpc.StatusOK,
    });

  });

  client.close();
}

export function CheckPublicMetrics(header) {

  let pipeline_id = randomString(10)
  let connector_id = randomString(10)

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: List Pipeline Trigger Records`, () => {

    let emptyPipelineTriggerRecordResponse = {
      "pipelineTriggerRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    }

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords response has pipelineTriggerRecords': (r) => r && r.message.pipelineTriggerRecords !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords response has total_size': (r) => r && r.message.totalSize !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords response has next_page_token': (r) => r && r.message.nextPageToken !== undefined,
    });
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords', {
      filter: `pipeline_id="${pipeline_id}" AND trigger_mode=MODE_SYNC`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords with filter response pipelineTriggerRecords length is 0': (r) => r && r.message.pipelineTriggerRecords.length === 0,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords with filter response total_size is 0': (r) => r && r.message.totalSize === emptyPipelineTriggerRecordResponse.totalSize,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords with filter response next_page_token is empty': (r) => r && r.message.nextPageToken === emptyPipelineTriggerRecordResponse.nextPageToken,
    });

  });

  group(`Management Public API: List Pipeline Trigger Table Records`, () => {

    let emptyPipelineTriggerTableRecordResponse = {
      "pipelineTriggerTableRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    }

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords response has pipelineTriggerTableRecords': (r) => r && r.message.pipelineTriggerTableRecords !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords response has total_size': (r) => r && r.message.totalSize !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords response has next_page_token': (r) => r && r.message.nextPageToken !== undefined,
    });
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords', {
      filter: `pipeline_id="${pipeline_id}"`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords with filter response pipelineTriggerTableRecords length is 0': (r) => r && r.message.pipelineTriggerTableRecords.length === 0,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords with filter response total_size is 0': (r) => r && r.message.totalSize === emptyPipelineTriggerTableRecordResponse.totalSize,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords with filter response next_page_token is empty': (r) => r && r.message.nextPageToken === emptyPipelineTriggerTableRecordResponse.nextPageToken,
    });

  });

  group(`Management Public API: List Pipeline Trigger Chart Records`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecords', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecords status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecords response has pipelineTriggerChartRecords': (r) => r && r.message.pipelineTriggerChartRecords !== undefined,
    });
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecords', {
      filter: `pipeline_id="${pipeline_id}" AND trigger_mode=MODE_SYNC`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecords with filter response pipelineTriggerChartRecords lenght is 0': (r) => r && r.message.pipelineTriggerChartRecords.length === 0,
    });

  });

  group(`Management Public API: List Connector Execute Records`, () => {

    let emptyConnectorExecuteRecordResponse = {
      "connectorExecuteRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    }

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords response has connectorExecuteRecords': (r) => r && r.message.connectorExecuteRecords !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords response has total_size': (r) => r && r.message.totalSize !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords response has next_page_token': (r) => r && r.message.nextPageToken !== undefined,
    });
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords', {
      filter: `connector_id="${connector_id}" AND status=STATUS_COMPLETED`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords with filter response connectorExecuteRecords length is 0': (r) => r && r.message.connectorExecuteRecords.length === 0,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords with filter response total_size is 0': (r) => r && r.message.totalSize === emptyConnectorExecuteRecordResponse.totalSize,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteRecords with filter response next_page_token is empty': (r) => r && r.message.nextPageToken === emptyConnectorExecuteRecordResponse.nextPageToken,
    });

  });

  group(`Management Public API: List Connector Execute Table Records`, () => {

    let emptyConnectorExecuteTableRecordResponse = {
      "connectorExecuteTableRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    }

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords response has connectorExecuteTableRecords': (r) => r && r.message.connectorExecuteTableRecords !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords response has total_size': (r) => r && r.message.totalSize !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords response has next_page_token': (r) => r && r.message.nextPageToken !== undefined,
    });
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords', {
      filter: `connector_id="${connector_id}"`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords with filter response connectorExecuteTableRecords length is 0': (r) => r && r.message.connectorExecuteTableRecords.length === 0,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords with filter response total_size is 0': (r) => r && r.message.totalSize === emptyConnectorExecuteTableRecordResponse.totalSize,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteTableRecords with filter response next_page_token is empty': (r) => r && r.message.nextPageToken === emptyConnectorExecuteTableRecordResponse.nextPageToken,
    });

  });

  group(`Management Public API: List Connector Execute Chart Records`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteChartRecords', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteChartRecords status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteChartRecords response has connectorExecuteChartRecords': (r) => r && r.message.connectorExecuteChartRecords !== undefined,
    });
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteChartRecords', {
      filter: `connector_id="${connector_id}" AND status=STATUS_COMPLETED`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteChartRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListConnectorExecuteChartRecords with filter response connectorExecuteChartRecords lenght is 0': (r) => r && r.message.connectorExecuteChartRecords.length === 0,
    });

  });

  client.close();
}
