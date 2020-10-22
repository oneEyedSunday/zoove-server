package errors

import "errors"

var UnAuthorized = errors.New("Error authorizing this guy.")
var NotFound = errors.New("Not Found")
var IncompleteRequest = errors.New("The request is incomplete. An import part is missing")
var BadOrInvalidJwt = errors.New("malformed authorization token")
