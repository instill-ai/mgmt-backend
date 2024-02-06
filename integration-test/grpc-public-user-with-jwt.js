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

export function CheckPublicGetUser() {

  group(`Management Public API: Get authenticated user [with random "instill-user-uid" header]`, () => {

    client.connect(constant.mgmtPublicGRPCHost, {
      plaintext: true
    });

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser', {}, constant.grpcParamsWithInstillUserUid), {
      '[with random "instill-user-uid" header] core.mgmt.v1beta.MgmtPublicService/GetAuthenticatedUser status StatusUnauthenticated': (r) => r && r.status == grpc.StatusUnauthenticated,
    });

    client.close();
  })
}

export function CheckPublicPatchAuthenticatedUser() {

  client.connect(constant.mgmtPublicGRPCHost, {
    plaintext: true
  });

  group(`Management Public API: Update authenticated user [with random "instill-user-uid" header]`, () => {
    var userUpdate = {
      name: `users/${constant.defaultUser.id}`,
      email: "test@foo.bar",
      customer_id: "new_customer_id",
      profile: {
        display_name: "test",
        company_name: "company",
      },
      role: "ai-engineer",
      newsletter_subscription: true,
    };

    check(client.invoke('core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser', {
      user: userUpdate,
      update_mask: "email,profile,role,newsletterSubscription"
    }, constant.grpcParamsWithInstillUserUid), {
      '[with random "instill-user-uid" header] core.mgmt.v1beta.MgmtPublicService/PatchAuthenticatedUser status StatusUnauthenticated': (r) => r && r.status == grpc.StatusUnauthenticated,
    });
  });

  client.close();
}
