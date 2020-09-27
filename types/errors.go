package types

import "errors"

var UnAuthorizedScope = errors.New("User is not authorized for this scope.")
