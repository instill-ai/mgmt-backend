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
client.load(['proto/base/mgmt/v1alpha'], 'mgmt.proto');
client.load(['proto/base/mgmt/v1alpha'], 'mgmt_public_service.proto');

export function CheckHealth() {
  // Health check
  group("Management API: Health check", () => {

    client.connect(constant.mgmtPublicGRPCHost, {
      plaintext: true
    });

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/Liveness', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/Liveness status': (r) => r && r.status == grpc.StatusOK,
    });

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/Readiness', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/Readiness status': (r) => r && r.status == grpc.StatusOK,
    });

    client.close();
  });
}

export function CheckPublicQueryAuthenticatedUser() {

  group(`Management Public API: Get authenticated user`, () => {

    client.connect(constant.mgmtPublicGRPCHost, {
      plaintext: true
    });

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response name': (r) => r && r.message.user.name !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response uid is UUID': (r) => r && helper.isUUID(r.message.user.uid),
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response id': (r) => r && r.message.user.id !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response id': (r) => r && r.message.user.id === constant.defaultUser.id,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response type': (r) => r && r.message.user.type === "OWNER_TYPE_USER",
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response email': (r) => r && r.message.user.email !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response customerId': (r) => r && r.message.user.customerId !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response firstName': (r) => r && r.message.user.firstName !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response lastName': (r) => r && r.message.user.lastName !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response orgName': (r) => r && r.message.user.orgName !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response role': (r) => r && r.message.user.role !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response newsletterSubscription': (r) => r && r.message.user.newsletterSubscription !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response cookieToken': (r) => r && r.message.user.cookieToken !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response createTime': (r) => r && r.message.user.createTime !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser response updateTime': (r) => r && r.message.user.updateTime !== undefined,
    });

    client.close();
  })
}

export function CheckPublicPatchAuthenticatedUser() {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Update authenticated user`, () => {
    var userUpdate = {
      name: `users/${constant.defaultUser.id}`,
      type: "OWNER_TYPE_ORGANIZATION",
      email: "test@foo.bar",
      customer_id: "new_customer_id",
      first_name: "test",
      last_name: "foo",
      org_name: "company",
      role: "ai-engineer",
      newsletter_subscription: true,
      cookie_token: "f5730f62-7026-4e11-917a-d890da315d3b",
    };

    var res = client.invoke('base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser', {})

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "email,firstName,lastName,orgName,role,newsletterSubscription,cookieToken"
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response name unchanged': (r) => r && r.message.user.name === res.message.user.name,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response uid unchanged': (r) => r && r.message.user.uid === res.message.user.uid,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response id unchanged': (r) => r && r.message.user.id === res.message.user.id,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response type unchanged': (r) => r && r.message.user.type === res.message.user.type,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response email updated': (r) => r && r.message.user.email === userUpdate.email,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response customerId unchanged': (r) => r && r.message.user.customerId === res.message.user.customerId,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response firstName updated': (r) => r && r.message.user.firstName === userUpdate.first_name,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response lastName updated': (r) => r && r.message.user.lastName === userUpdate.last_name,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response orgName updated': (r) => r && r.message.user.orgName === userUpdate.org_name,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response role updated': (r) => r && r.message.user.role === userUpdate.role,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response newsletterSubscription updated': (r) => r && r.message.user.newsletterSubscription === userUpdate.newsletter_subscription,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response cookieToken updated': (r) => r && r.message.user.cookieToken === userUpdate.cookie_token,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response createTime unchanged': (r) => r && r.message.user.createTime === res.message.user.createTime,
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser response updateTime updated': (r) => r && r.message.user.updateTime !== res.message.user.updateTime,
    });

    // Restore to default user
    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: constant.defaultUser,
      update_mask: "email,firstName,lastName,orgName,role,newsletterSubscription,cookieToken"
    }), {
      [`[restore the default user] base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser status`]: (r) => r && r.status == grpc.StatusOK,
    });

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
    });
  });

  group(`Management Public API: Update authenticated user with a non-exist role`, () => {
    var nonExistRole = "non-exist-role";
    var userUpdate = {
      role: nonExistRole,
    };

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "role"
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser nonExistRole StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

  });

  group(`Management Public API: Update authenticated user ID [not allowed]`, () => {
    var userUpdate = {
      id: `test_${randomString(10)}`,
    };

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "id"
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser update ID StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

  });

  group(`Management Public API: Update authenticated user UID [not allowed]`, () => {
    var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
    var userUpdate = {
      uid: nonExistUID,
    };

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "uid"
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser nonExistUID StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });
  });

  client.close();
}

export function CheckPublicCreateToken() {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Create API token`, () => {

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/CreateToken', {
      token: {
        id: `${constant.testToken.id}`,
        ttl: 86400
      }
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/CreateToken status StatusUnimplemented': (r) => r && r.status == grpc.StatusUnimplemented,
    });

  });

  client.close();
}

export function CheckPublicListTokens() {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: List API tokens`, () => {

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListTokens', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/ListTokens status StatusUnimplemented': (r) => r && r.status == grpc.StatusUnimplemented,
    });

  });

  client.close();
}

export function CheckPublicGetToken() {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Get API token`, () => {

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/GetToken', {
      name: `tokens/${constant.testToken.id}`,
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/GetToken status StatusUnimplemented': (r) => r && r.status == grpc.StatusUnimplemented,
    });

  });

  client.close();
}

export function CheckPublicDeleteToken() {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Delete API token`, () => {

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/DeleteToken', {
      name: `tokens/${constant.testToken.id}`,
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/DeleteToken status StatusUnimplemented': (r) => r && r.status == grpc.StatusUnimplemented,
    });

  });

  client.close();
}

export function CheckPublicMetrics() {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: List Pipeline Trigger Records`, () => {

    let emptyPipelineTriggerRecordResponse = {
      "pipelineTriggerRecords":[],
      "nextPageToken":"",
      "totalSize":"0"
    }

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords response has pipelineTriggerRecords': (r) => r && r.message.pipelineTriggerRecords !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords response has total_size': (r) => r && r.message.totalSize !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords response has next_page_token': (r) => r && r.message.nextPageToken !== undefined,
    });
    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords', {
      filter: "pipeline_id=\"a\" AND trigger_mode=MODE_SYNC",
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords with filter response pipelineTriggerRecords length is 0': (r) => r && r.message.pipelineTriggerRecords.length === 0,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords with filter response total_size is 0': (r) => r && r.message.totalSize === emptyPipelineTriggerRecordResponse.totalSize,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerRecords with filter response next_page_token is empty': (r) => r && r.message.nextPageToken === emptyPipelineTriggerRecordResponse.nextPageToken,
    });

  });

  group(`Management Public API: List Pipeline Trigger Table Records`, () => {

    let emptyPipelineTriggerTableRecordResponse = {
      "pipelineTriggerTableRecords":[],
      "nextPageToken":"",
      "totalSize":"0"
    }

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords response has pipelineTriggerTableRecords': (r) => r && r.message.pipelineTriggerTableRecords !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords response has total_size': (r) => r && r.message.totalSize !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords response has next_page_token': (r) => r && r.message.nextPageToken !== undefined,
    });
    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords', {
      filter: "pipeline_id=\"iloveinstill\"",
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords with filter response pipelineTriggerTableRecords length is 0': (r) => r && r.message.pipelineTriggerTableRecords.length === 0,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords with filter response total_size is 0': (r) => r && r.message.totalSize === emptyPipelineTriggerTableRecordResponse.totalSize,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerTableRecords with filter response next_page_token is empty': (r) => r && r.message.nextPageToken === emptyPipelineTriggerTableRecordResponse.nextPageToken,
    });

  });

  group(`Management Public API: List Pipeline Trigger Chart Records`, () => {

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerChartRecords', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerChartRecords status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerChartRecords response has pipelineTriggerChartRecords': (r) => r && r.message.pipelineTriggerChartRecords !== undefined,
    });
    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerChartRecords', {
      filter: "pipeline_id=\"a\" AND trigger_mode=MODE_SYNC",
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerChartRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListPipelineTriggerChartRecords with filter response pipelineTriggerChartRecords lenght is 0': (r) => r && r.message.pipelineTriggerChartRecords.length === 0,
    });

  });

  group(`Management Public API: List Connector Execute Records`, () => {

    let emptyConnectorExecuteRecordResponse = {
      "connectorExecuteRecords":[],
      "nextPageToken":"",
      "totalSize":"0"
    }

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords response has connectorExecuteRecords': (r) => r && r.message.connectorExecuteRecords !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords response has total_size': (r) => r && r.message.totalSize !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords response has next_page_token': (r) => r && r.message.nextPageToken !== undefined,
    });
    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords', {
      filter: "connector_id=\"a\" AND status=STATUS_COMPLETED",
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords with filter response connectorExecuteRecords length is 0': (r) => r && r.message.connectorExecuteRecords.length === 0,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords with filter response total_size is 0': (r) => r && r.message.totalSize === emptyConnectorExecuteRecordResponse.totalSize,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteRecords with filter response next_page_token is empty': (r) => r && r.message.nextPageToken === emptyConnectorExecuteRecordResponse.nextPageToken,
    });

  });

  group(`Management Public API: List Connector Execute Table Records`, () => {

    let emptyConnectorExecuteTableRecordResponse = {
      "connectorExecuteTableRecords":[],
      "nextPageToken":"",
      "totalSize":"0"
    }

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords response has connectorExecuteTableRecords': (r) => r && r.message.connectorExecuteTableRecords !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords response has total_size': (r) => r && r.message.totalSize !== undefined,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords response has next_page_token': (r) => r && r.message.nextPageToken !== undefined,
    });
    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords', {
      filter: "connector_id=\"iloveinstill\"",
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords with filter response connectorExecuteTableRecords length is 0': (r) => r && r.message.connectorExecuteTableRecords.length === 0,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords with filter response total_size is 0': (r) => r && r.message.totalSize === emptyConnectorExecuteTableRecordResponse.totalSize,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteTableRecords with filter response next_page_token is empty': (r) => r && r.message.nextPageToken === emptyConnectorExecuteTableRecordResponse.nextPageToken,
    });

  });

  group(`Management Public API: List Connector Execute Chart Records`, () => {

    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteChartRecords', {}), {
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteChartRecords status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteChartRecords response has connectorExecuteChartRecords': (r) => r && r.message.connectorExecuteChartRecords !== undefined,
    });
    check(client.invoke('base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteChartRecords', {
      filter: "connector_id=\"a\" AND status=STATUS_COMPLETED",
    }), {
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteChartRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'base.mgmt.v1alpha.MgmtPublicService/ListConnectorExecuteChartRecords with filter response connectorExecuteChartRecords lenght is 0': (r) => r && r.message.connectorExecuteChartRecords.length === 0,
    });

  });

  client.close();
}
