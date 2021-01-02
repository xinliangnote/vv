package interceptor

import (
	"sync"
)

type userinfoHandler func(authorization string, payload Payload) (userinfo interface{}, err error)
type signatureHandler func(proxyAuthorization string, payload Payload) (ok bool, err error)

// Validator authorization & proxy_authorization validator
var Validator = &validator{
	auth:      make(map[string]userinfoHandler),
	proxyAuth: make(map[string]signatureHandler),
}

type validator struct {
	sync.RWMutex
	auth      map[string]userinfoHandler
	proxyAuth map[string]signatureHandler
}

// RegisteAuthorizationValidator some handler(s) for validate authorization and return userinfo
func (v *validator) RegisteAuthorizationValidator(name string, handler userinfoHandler) {
	v.Lock()
	defer v.Unlock()

	v.auth[name] = handler
}

func (v *validator) RegisteProxyAuthorizationValidator(name string, handler signatureHandler) {
	v.Lock()
	defer v.Unlock()

	v.proxyAuth[name] = handler
}

// RegisteProxyAuthorizationValidator some handler(s) for validate signature
func (v *validator) AuthorizationValidator(name string) userinfoHandler {
	v.RLock()
	defer v.RUnlock()

	return v.auth[name]
}

func (v *validator) ProxyAuthorizationValidator(name string) signatureHandler {
	v.RLock()
	defer v.RUnlock()

	return v.proxyAuth[name]
}
