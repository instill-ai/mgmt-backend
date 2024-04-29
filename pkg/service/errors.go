package service

import "errors"

var ErrNoPermission = errors.New("no permission")
var ErrUnauthenticated = errors.New("unauthenticated")
var ErrInvalidTokenTTL = errors.New("invalid token ttl")
var ErrInvalidRole = errors.New("invalid role")
var ErrInvalidOwnerNamespace = errors.New("invalid owner namepsace format")
var ErrStateCanOnlyBeActive = errors.New("state can only be active")
var ErrCanNotRemoveOwnerFromOrganization = errors.New("can not remove owner from organization")
var ErrCanNotSetAnotherOwner = errors.New("can not set another user as owner")
var ErrPasswordNotMatch = errors.New("password not match")
