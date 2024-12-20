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

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser status': (r) => { return r && r.status == grpc.StatusOK },
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response name': (r) => r && r.message.user.name !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response uid is UUID': (r) => r && helper.isUUID(r.message.user.uid),
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response id': (r) => r && r.message.user.id !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response id': (r) => r && r.message.user.id === constant.defaultUser.id,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response email': (r) => r && r.message.user.email !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response customerId': (r) => r && r.message.user.customerId !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response displayName': (r) => r && r.message.user.profile.displayName !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response companyName': (r) => r && r.message.user.profile.companyName !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response role': (r) => r && r.message.user.role !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response newsletterSubscription': (r) => r && r.message.user.newsletterSubscription !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response createTime': (r) => r && r.message.user.createTime !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser response updateTime': (r) => r && r.message.user.updateTime !== undefined,
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
      customerId: "new_customer_id",
      profile: {
        displayName: "test",
        companyName: "company",
      },
      role: "ai-engineer",
      newsletterSubscription: true,
    };

    var res = client.invoke('core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser', {}, header)

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "email,profile,role,newsletterSubscription"
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response name unchanged': (r) => r && r.message.user.name === res.message.user.name,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response uid unchanged': (r) => r && r.message.user.uid === res.message.user.uid,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response id unchanged': (r) => r && r.message.user.id === res.message.user.id,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response email updated': (r) => r && r.message.user.email === userUpdate.email,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response displayName updated': (r) => r && r.message.user.profile.displayName === userUpdate.profile.displayName,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response companyName updated': (r) => r && r.message.user.profile.companyName === userUpdate.profile.companyName,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response role updated': (r) => r && r.message.user.role === userUpdate.role,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response newsletterSubscription updated': (r) => r && r.message.user.newsletterSubscription === userUpdate.newsletterSubscription,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response createTime unchanged': (r) => r && r.message.user.createTime === res.message.user.createTime,
      'core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser response updateTime updated': (r) => r && r.message.user.updateTime !== res.message.user.updateTime,
    });

    // Restore to default user
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser', {
      user: constant.defaultUser,
      update_mask: "email,profile,role,newsletterSubscription"
    }, header), {
      [`[restore the default user] core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser status`]: (r) => r && r.status == grpc.StatusOK,
    });

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
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
      tokenId: constant.testToken.id,
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
      tokenId: constant.testToken.id,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/DeleteToken status StatusOK': (r) => r && r.status == grpc.StatusOK,
    });

  });

  client.close();
}

export function CheckPublicGetRemainingCredit(header) {
  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Get remaining credit`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/GetRemainingCredit', {
      namespaceId: constant.defaultUser.id,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/GetRemainingCredit status Unimplemented': (r) => r && r.status == grpc.StatusUnimplemented,
    });

  });

  client.close();
}

export function CheckPublicMetrics(header) {
  let pipeline_id = randomString(10);

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: List Pipeline Trigger Records`, () => {
    let emptyPipelineTriggerRecordResponse = {
      "pipelineTriggerRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    };

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords response has pipelineTriggerRecords': (r) => r && r.message.pipelineTriggerRecords !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords response has totalSize': (r) => r && r.message.totalSize !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords response has nextPageToken': (r) => r && r.message.nextPageToken !== undefined,
    });

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords', {
      filter: `pipelineId="${pipeline_id}" AND triggerMode=MODE_SYNC`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords with filter status': (r) =>  r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords with filter response pipelineTriggerRecords length is 0': (r) => r && r.message.pipelineTriggerRecords.length === 0,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords with filter response totalSize is 0': (r) => r && r.message.totalSize === emptyPipelineTriggerRecordResponse.totalSize,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerRecords with filter response nextPageToken is empty': (r) => r && r.message.nextPageToken === emptyPipelineTriggerRecordResponse.nextPageToken,
    });
  });

  group(`Management Public API: List Pipeline Trigger Table Records`, () => {
    let emptyPipelineTriggerTableRecordResponse = {
      "pipelineTriggerTableRecords": [],
      "nextPageToken": "",
      "totalSize": 0
    };

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords response has pipelineTriggerTableRecords': (r) => r && r.message.pipelineTriggerTableRecords !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords response has totalSize': (r) => r && r.message.totalSize !== undefined,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords response has nextPageToken': (r) => r && r.message.nextPageToken !== undefined,
    });
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords', {
      filter: `pipelineId="${pipeline_id}"`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords with filter status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords with filter response pipelineTriggerTableRecords length is 0': (r) => r && r.message.pipelineTriggerTableRecords.length === 0,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords with filter response totalSize is 0': (r) => r && r.message.totalSize === emptyPipelineTriggerTableRecordResponse.totalSize,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerTableRecords with filter response nextPageToken is empty': (r) => r && r.message.nextPageToken === emptyPipelineTriggerTableRecordResponse.nextPageToken,
    });
  });

  group(`Management Public API: List Pipeline Trigger Chart Records`, () => {
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecordsV0', {}, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecordsV0 status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecordsV0 response has pipelineTriggerChartRecords': (r) => r && r.message.pipelineTriggerChartRecords !== undefined,
    });
    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecordsV0', {
      filter: `pipelineId="${pipeline_id}" AND triggerMode=MODE_SYNC`,
    }, header), {
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecordsV0 with filter status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPublicService/ListPipelineTriggerChartRecordsV0 with filter response pipelineTriggerChartRecords lenght is 0': (r) => r && r.message.pipelineTriggerChartRecords.length === 0,
    });
  });

  client.close();
}
