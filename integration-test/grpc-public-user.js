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
client.load(['proto/vdp/mgmt/v1alpha'], 'mgmt.proto');
client.load(['proto/vdp/mgmt/v1alpha'], 'mgmt_public_service.proto');

export function CheckHealth() {
  // Health check
  group("Management API: Health check", () => {

    client.connect(constant.mgmtPublicGRPCHost, {
      plaintext: true
    });

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/Liveness', {}), {
      'vdp.model.v1alpha.MgmtPublicService/Liveness status': (r) => r && r.status == grpc.StatusOK,
    });

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/Readiness', {}), {
      'vdp.model.v1alpha.MgmtPublicService/Readiness status': (r) => r && r.status == grpc.StatusOK,
    });

    client.close();
  });
}

export function CheckPublicGet() {

  group(`Management Public API: Get authenticated user`, () => {

    client.connect(constant.mgmtPublicGRPCHost, {
      plaintext: true
    });

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser', {}), {
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response name': (r) => r && r.message.user.name !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response uid is UUID': (r) => r && helper.isUUID(r.message.user.uid),
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response id': (r) => r && r.message.user.id !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response id': (r) => r && r.message.user.id === constant.defaultUser.id,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response type': (r) => r && r.message.user.type === "OWNER_TYPE_USER",
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response email': (r) => r && r.message.user.email !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response customerId': (r) => r && r.message.user.customerId !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response firstName': (r) => r && r.message.user.firstName !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response lastName': (r) => r && r.message.user.lastName !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response orgName': (r) => r && r.message.user.orgName !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response role': (r) => r && r.message.user.role !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response newsletterSubscription': (r) => r && r.message.user.newsletterSubscription !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response cookieToken': (r) => r && r.message.user.cookieToken !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response createTime': (r) => r && r.message.user.createTime !== undefined,
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser response updateTime': (r) => r && r.message.user.updateTime !== undefined,
    });

    client.close();
  })
}

export function CheckPublicUpdate() {

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

    var res = client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser', {})

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "email,firstName,lastName,orgName,role,newsletterSubscription,cookieToken"
    }), {
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response name unchanged': (r) => r && r.message.user.name === res.message.user.name,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response uid unchanged': (r) => r && r.message.user.uid === res.message.user.uid,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response id unchanged': (r) => r && r.message.user.id === res.message.user.id,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response type unchanged': (r) => r && r.message.user.type === res.message.user.type,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response email updated': (r) => r && r.message.user.email === userUpdate.email,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response customerId unchanged': (r) => r && r.message.user.customerId === res.message.user.customerId,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response firstName updated': (r) => r && r.message.user.firstName === userUpdate.first_name,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response lastName updated': (r) => r && r.message.user.lastName === userUpdate.last_name,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response orgName updated': (r) => r && r.message.user.orgName === userUpdate.org_name,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response role updated': (r) => r && r.message.user.role === userUpdate.role,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response newsletterSubscription updated': (r) => r && r.message.user.newsletterSubscription === userUpdate.newsletter_subscription,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response cookieToken updated': (r) => r && r.message.user.cookieToken === userUpdate.cookie_token,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response createTime unchanged': (r) => r && r.message.user.createTime === res.message.user.createTime,
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser response updateTime updated': (r) => r && r.message.user.updateTime !== res.message.user.updateTime,
    });

    // Restore to default user
    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: constant.defaultUser,
      update_mask: "email,firstName,lastName,orgName,role,newsletterSubscription,cookieToken"
    }), {
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
    });

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser', {}), {
      'vdp.model.v1alpha.MgmtPublicService/QueryAuthenticatedUser status': (r) => r && r.status == grpc.StatusOK,
    });
  });

  group(`Management Public API: Update authenticated user with a non-exist role`, () => {
    var nonExistRole = "non-exist-role";
    var userUpdate = {
      role: nonExistRole,
    };

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "role"
    }), {
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser nonExistRole StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

  });

  group(`Management Public API: Update authenticated user ID [not allowed]`, () => {
    var userUpdate = {
      id: `test_${randomString(10)}`,
    };

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "id"
    }), {
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser update ID StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

  });

  group(`Management Public API: Update authenticated user UID [not allowed]`, () => {
    var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
    var userUpdate = {
      uid: nonExistUID,
    };

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "uid"
    }), {
      'vdp.model.v1alpha.MgmtPublicService/PatchAuthenticatedUser nonExistUID StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });
  });

  client.close();
}
