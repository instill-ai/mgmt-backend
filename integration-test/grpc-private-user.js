import http from "k6/http";
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
client.load(['proto/core/mgmt/v1beta'], 'mgmt_private_service.proto');

export function CheckPrivateListUsersAdmin() {
  group("Management Private API: List users", () => {

    client.connect(constant.mgmtPrivateGRPCHost, {
      plaintext: true
    });

    check(client.invoke('core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin', {}), {
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin response body has user array': (r) => r && Array.isArray(r.message.users),
    });


    var res = client.invoke('core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin', {})
    check(client.invoke('core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin', {
      page_size: 0
    }), {
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin page_size=0 status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin page_size=0 response all records': (r) => r && r.message.users.length === res.message.users.length,
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin page_size=0 response total_size 1': (r) => r && r.message.totalSize === res.message.totalSize,
    });

    check(client.invoke('core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin', {
      page_size: 5
    }), {
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin page_size=0 status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin page_size=0 response all records size 1': (r) => r && r.message.users.length === 1,
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin page_size=0 response totalSize 1': (r) => r && r.message.totalSize == 1,
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin page_size=0 response nextPageToken is empty': (r) => r && r.message.nextPageToken === "",
    });

    var invalidNextPageToken = `${randomString(10)}`;

    var res = client.invoke('core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin', {})
    check(client.invoke('core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin', {
      page_size: 1,
      page_token: invalidNextPageToken
    }), {
      'core.mgmt.v1beta.MgmtPrivateService/ListUsersAdmin page_size: 1 page_token: invalidNextPageToken status StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

    client.close();
  });
}

export function CheckPrivateGetUserAdmin() {

  client.connect(constant.mgmtPrivateGRPCHost, {
    plaintext: true
  });

  group(`Management Private API: Get default user`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin', {
      name: `users/${constant.defaultUser.id}`
    }), {
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response name': (r) => r && r.message.user.name !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response uid is UUID': (r) => r && helper.isUUID(r.message.user.uid),
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response id': (r) => r && r.message.user.id !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response id': (r) => r && r.message.user.id === constant.defaultUser.id,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response email': (r) => r && r.message.user.email !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response customerId': (r) => r && r.message.user.customerId !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response firstName': (r) => r && r.message.user.firstName !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response lastName': (r) => r && r.message.user.lastName !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response createTime': (r) => r && r.message.user.createTime !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin response updateTime': (r) => r && r.message.user.updateTime !== undefined,
    });

  });

  var nonExistID = "non-exist";
  group(`Management Private API: Get non-exist user`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin', {
      name: "users/" + nonExistID
    }), {
      'core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin status StatusNotFound': (r) => r && r.status == grpc.StatusNotFound,
    });

  });

  client.close();
}

export function CheckPrivateLookUpUserAdmin() {

  client.connect(constant.mgmtPrivateGRPCHost, {
    plaintext: true
  });

  // Get the uid of the default user
  var res = client.invoke('core.mgmt.v1beta.MgmtPrivateService/GetUserAdmin', {
    name: `users/${constant.defaultUser.id}`
  })
  var defaultUid = res.message.user.uid;

  group(`Management Private API: Look up default user by permalink`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin', {
      permalink: `users/${defaultUid}`
    }), {
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin status': (r) => r && r.status == grpc.StatusOK,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response name': (r) => r && r.message.user.name !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response uid is UUID': (r) => r && helper.isUUID(r.message.user.uid),
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response id': (r) => r && r.message.user.id !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response id': (r) => r && r.message.user.id === constant.defaultUser.id,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response email': (r) => r && r.message.user.email !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response customerId': (r) => r && r.message.user.customerId !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response firstName': (r) => r && r.message.user.firstName !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response lastName': (r) => r && r.message.user.lastName !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response companyName': (r) => r && r.message.user.companyName !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response role': (r) => r && r.message.user.role !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response newsletterSubscription': (r) => r && r.message.user.newsletterSubscription !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response createTime': (r) => r && r.message.user.createTime !== undefined,
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin response updateTime': (r) => r && r.message.user.updateTime !== undefined,
    });

  });

  var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
  group(`Management Private API: Look up non-exist user by permalink`, () => {

    check(client.invoke('core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin', {
      permalink: `users/${nonExistUID}`
    }), {
      'core.mgmt.v1beta.MgmtPrivateService/LookUpUserAdmin status StatusNotFound': (r) => r && r.status == grpc.StatusNotFound,
    });
  });

  client.close()
}
