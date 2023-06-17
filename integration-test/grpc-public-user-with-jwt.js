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

export function CheckPublicQueryAuthenticatedUser() {

  group(`Management Public API: Get authenticated user [with random "jwt-sub" header]`, () => {

    client.connect(constant.mgmtPublicGRPCHost, {
      plaintext: true
    });

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser', {}, constant.grpcParamsWithJwtSub), {
      '[with random "jwt-sub" header] vdp.mgmt.v1alpha.MgmtPublicService/QueryAuthenticatedUser status StatusNotFound': (r) => r && r.status == grpc.StatusNotFound,
    });

    client.close();
  })
}

export function CheckPublicPatchAuthenticatedUser() {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Update authenticated user [with random "jwt-sub" header]`, () => {
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

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "email,firstName,lastName,orgName,role,newsletterSubscription,cookieToken"
    }, constant.grpcParamsWithJwtSub), {
      '[with random "jwt-sub" header] vdp.mgmt.v1alpha.MgmtPublicService/PatchAuthenticatedUser status StatusNotFound': (r) => r && r.status == grpc.StatusNotFound,
    });
  });

  client.close();
}
