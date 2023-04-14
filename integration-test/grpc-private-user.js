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
client.load(['proto/vdp/mgmt/v1alpha'], 'mgmt.proto');
client.load(['proto/vdp/mgmt/v1alpha'], 'mgmt_private_service.proto');

export function CheckPrivateListUsersAdmin() {
  group("Management Private API: List users", () => {

    client.connect(constant.mgmtPrivateGRPCHost, {
      plaintext: true
    });

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin', {}), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin status': (r) => r && r.status == grpc.StatusOK,
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin response body has user array': (r) => r && Array.isArray(r.message.users),
    });


    var res = client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin', {})
    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin', {
      page_size: 0
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin page_size=0 status': (r) => r && r.status == grpc.StatusOK,
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin page_size=0 response all records': (r) => r && r.message.users.length === res.message.users.length,
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin page_size=0 response total_size 1': (r) => r && r.message.totalSize === res.message.totalSize,
    });

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin', {
      page_size: 5
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin page_size=0 status': (r) => r && r.status == grpc.StatusOK,
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin page_size=0 response all records size 1': (r) => r && r.message.users.length === 1,
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin page_size=0 response totalSize 1': (r) => r && r.message.totalSize == 1,
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin page_size=0 response nextPageToken is empty': (r) => r && r.message.nextPageToken === "",
    });

    var invalidNextPageToken = `${randomString(10)}`;

    var res = client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin', {})
    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin', {
      page_size: 1,
      page_token: invalidNextPageToken
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/ListUsersAdmin page_size: 1 page_token: invalidNextPageToken status StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

    client.close();
  });
}

export function CheckPrivateGetUserAdmin() {

  client.connect(constant.mgmtPrivateGRPCHost, {
    plaintext: true
  });

  group(`Management Private API: Get default user`, () => {

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin', {
      name: `users/${constant.defaultUser.id}`
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin status': (r) => r && r.status == grpc.StatusOK,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin status': (r) => r && r.status == grpc.StatusOK,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response name': (r) => r && r.message.user.name !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response uid is UUID': (r) => r && helper.isUUID(r.message.user.uid),
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response id': (r) => r && r.message.user.id !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response id': (r) => r && r.message.user.id === constant.defaultUser.id,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response type': (r) => r && r.message.user.type === "OWNER_TYPE_USER",
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response email': (r) => r && r.message.user.email !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response customerId': (r) => r && r.message.user.customerId !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response firstName': (r) => r && r.message.user.firstName !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response lastName': (r) => r && r.message.user.lastName !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response orgName': (r) => r && r.message.user.orgName !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response role': (r) => r && r.message.user.role !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response newsletterSubscription': (r) => r && r.message.user.newsletterSubscription !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response cookieToken': (r) => r && r.message.user.cookieToken !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response createTime': (r) => r && r.message.user.createTime !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin response updateTime': (r) => r && r.message.user.updateTime !== undefined,
    });

  });

  var nonExistID = "non-exist";
  group(`Management Private API: Get non-exist user`, () => {

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin', {
      name: "users/" + nonExistID
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin status StatusNotFound': (r) => r && r.status == grpc.StatusNotFound,
    });

  });

  client.close();
}

export function CheckPrivateLookUpUserAdmin() {

  client.connect(constant.mgmtPrivateGRPCHost, {
    plaintext: true
  });

  // Get the uid of the default user
  var res = client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin', {
    name: `users/${constant.defaultUser.id}`
  })
  var defaultUid = res.message.user.uid;

  group(`Management Private API: Look up default user by permalink`, () => {

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin', {
      permalink: `users/${defaultUid}`
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin status': (r) => r && r.status == grpc.StatusOK,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin status': (r) => r && r.status == grpc.StatusOK,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response name': (r) => r && r.message.user.name !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response uid is UUID': (r) => r && helper.isUUID(r.message.user.uid),
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response id': (r) => r && r.message.user.id !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response id': (r) => r && r.message.user.id === constant.defaultUser.id,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response type': (r) => r && r.message.user.type === "OWNER_TYPE_USER",
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response email': (r) => r && r.message.user.email !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response customerId': (r) => r && r.message.user.customerId !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response firstName': (r) => r && r.message.user.firstName !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response lastName': (r) => r && r.message.user.lastName !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response orgName': (r) => r && r.message.user.orgName !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response role': (r) => r && r.message.user.role !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response newsletterSubscription': (r) => r && r.message.user.newsletterSubscription !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response cookieToken': (r) => r && r.message.user.cookieToken !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response createTime': (r) => r && r.message.user.createTime !== undefined,
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin response updateTime': (r) => r && r.message.user.updateTime !== undefined,
    });

  });

  var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
  group(`Management Private API: Look up non-exist user by permalink`, () => {

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin', {
      permalink: `users/${nonExistUID}`
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/LookUpUserAdmin status StatusNotFound': (r) => r && r.status == grpc.StatusNotFound,
    });
  });

  client.close()
}

export function CheckPrivateUpdateUserAdmin() {

  client.connect(constant.mgmtPrivateGRPCHost, {
    plaintext: true
  });

  group(`Management Private API: Update default user`, () => {
    var userUpdate = {
      name: `users/${constant.defaultUser.id}`,
      type: "OWNER_TYPE_ORGANIZATION",
      email: "test@foo.bar",
      customer_id: "new_customer_id",
      first_name: "test",
      last_name: "foo",
      org_name: "company",
      role: "ai-researcher",
      newsletter_subscription: true,
      cookie_token: "f5730f62-7026-4e11-917a-d890da315d3b",
    };

    var res = client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin', {
      name: `users/${constant.defaultUser.id}`
    })

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin', {
      user: userUpdate,
      update_mask: "email,firstName,lastName,orgName,role,newsletterSubscription,cookieToken"
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin status': (r) => r && r.status == grpc.StatusOK,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response name unchanged': (r) => r && r.message.user.name === res.message.user.name,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response uid unchanged': (r) => r && r.message.user.uid === res.message.user.uid,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response id unchanged': (r) => r && r.message.user.id === res.message.user.id,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response type unchanged': (r) => r && r.message.user.type === res.message.user.type,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response email updated': (r) => r && r.message.user.email === userUpdate.email,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response customerId unchanged': (r) => r && r.message.user.customerId === res.message.user.customerId,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response firstName updated': (r) => r && r.message.user.firstName === userUpdate.first_name,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response lastName updated': (r) => r && r.message.user.lastName === userUpdate.last_name,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response orgName updated': (r) => r && r.message.user.orgName === userUpdate.org_name,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response role updated': (r) => r && r.message.user.role === userUpdate.role,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response newsletterSubscription updated': (r) => r && r.message.user.newsletterSubscription === userUpdate.newsletter_subscription,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response cookieToken updated': (r) => r && r.message.user.cookieToken === userUpdate.cookie_token,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response createTime unchanged': (r) => r && r.message.user.createTime === res.message.user.createTime,
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin response updateTime updated': (r) => r && r.message.user.updateTime !== res.message.user.updateTime,
    });

    // Restore to default user
    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin', {
      user: constant.defaultUser,
      update_mask: "email,firstName,lastName,orgName,role,newsletterSubscription,cookieToken"
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin status': (r) => r && r.status == grpc.StatusOK,
    });

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin', {
      name: `users/${constant.defaultUser.id}`
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/GetUserAdmin status': (r) => r && r.status == grpc.StatusOK,
    });
  });

  group(`Management Private API: Update user with a non-exist role`, () => {
    var nonExistRole = "non-exist-role";
    var userUpdate = {
      name: `users/${constant.defaultUser.id}`,
      role: nonExistRole,
    };
    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin', {
      user: userUpdate,
      update_mask: "role"
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin status StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });
  });

  group(`Management Private API: Update user ID [not allowed]`, () => {
    var userUpdate = {
      name: `users/${constant.defaultUser.id}`,
      id: `test_${randomString(10)}`,
    };
    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin', {
      user: userUpdate,
      update_mask: "role"
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin status StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });
  });

  group(`Management Private API: Update user UID [not allowed]`, () => {
    var nonExistUID = "2a06c2f7-8da9-4046-91ea-240f88a5d000";
    var userUpdate = {
      name: `users/${constant.defaultUser.id}`,
      uid: nonExistUID,
    };
    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin', {
      user: userUpdate,
      update_mask: "role"
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin status StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });
  })

  var nonExistID = "non-exist";
  group(`Management Private API: Update non-exist user`, () => {
    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin', {
      user: {
        name: `users/${nonExistID}`,
        id: nonExistID,
        role: "admin"
      },
      update_mask: "role"
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/UpdateUserAdmin status StatusNotFound': (r) => r && r.status == grpc.StatusNotFound,
    });
  });

  client.close()
}

export function CheckPrivateCreateUserAdmin() {

  client.connect(constant.mgmtPrivateGRPCHost, {
    plaintext: true
  });

  group("Management Private API: Create user with UUID as id", () => {

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/CreateUserAdmin', {
      user: {
        id: "2a06c2f7-8da9-4046-91ea-240f88a5d000"
      }
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/CreateUserAdmin status StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

  });
  group("Management Private API: Create user with invalid id", () => {

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/CreateUserAdmin', {
      user: {
        id: "local user"
      }
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/CreateUserAdmin status StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

  });
  group("Management Private API: Create user", () => {

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/CreateUserAdmin', {
      user: {
        id: "instill",
      }
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/CreateUserAdmin status StatusInvalidArgument': (r) => r && r.status == grpc.StatusInvalidArgument,
    });

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/CreateUserAdmin', {
      user: {
        id: "test-user",
        email: "test-user@instill.tech"
      }
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/CreateUserAdmin status StatusUnimplemented': (r) => r && r.status == grpc.StatusUnimplemented,
    });

  });

  client.close();
}

export function CheckPrivateDeleteUserAdmin() {

  client.connect(constant.mgmtPrivateGRPCHost, {
    plaintext: true
  });

  group(`Management Private API: Delete user`, () => {

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/DeleteUserAdmin', {
      name: `users/${constant.defaultUser.id}`,
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/DeleteUserAdmin status StatusUnimplemented': (r) => r && r.status == grpc.StatusUnimplemented,
    });

  });

  client.close();
}

export function CheckPrivateValidateToken() {

  client.connect(constant.mgmtPrivateGRPCHost, {
    plaintext: true
  });

  group(`Management Private API: Validate API token`, () => {

    check(client.invoke('vdp.mgmt.v1alpha.MgmtPrivateService/ValidateToken', {
      name: `tokens/${constant.testToken.id}`,
    }), {
      'vdp.mgmt.v1alpha.MgmtPrivateService/ValidateToken status StatusUnimplemented': (r) => r && r.status == grpc.StatusUnimplemented,
    });

  });

  client.close();
}
