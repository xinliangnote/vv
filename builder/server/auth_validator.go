package server

import (
	"github.com/bluekaki/vv/internal/interceptor"
)

// Payload rest or grpc payload
type Payload = interceptor.Payload

func RegisteAuthorizationValidator(name string, handler func(authorization string, payload Payload) (userinfo interface{}, err error)) {
	interceptor.Validator.RegisteAuthorizationValidator(name, handler)
}

func RegisteProxyAuthorizationValidator(name string, handler func(proxyAuthorization string, payload Payload) (ok bool, err error)) {
	interceptor.Validator.RegisteProxyAuthorizationValidator(name, handler)
}
