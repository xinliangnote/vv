package interceptor

import (
	"fmt"
	"sync"

	"github.com/bluekaki/vv/options"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const _ = grpc.SupportPackageIsVersion7

// FileDescriptor protobuf file descriptor
var FileDescriptor = &fileDescriptor{
	options: make(map[string]protoreflect.ProtoMessage),
}

type fileDescriptor struct {
	sync.RWMutex
	options map[string]protoreflect.ProtoMessage // FullMethod : Options
}

func (f *fileDescriptor) ParseP(descriptor protoreflect.FileDescriptor) {
	f.Lock()
	defer f.Unlock()

	serivces := descriptor.Services()
	for i := 0; i < serivces.Len(); i++ {
		serivce := serivces.Get(i)
		methods := serivce.Methods()

		for k := 0; k < methods.Len(); k++ {
			method := methods.Get(k)
			fullMethod := fmt.Sprintf("/%s/%s", serivce.FullName(), method.Name())
			f.options[fullMethod] = method.Options()

			if option := proto.GetExtension(method.Options(), options.E_Authorization).(*options.Validator); option != nil &&
				Validator.AuthorizationValidator(option.Name) == nil {
				panic(fmt.Sprintf("%s options.authorization validator: [%s] not found", fullMethod, option.Name))
			}

			if option := proto.GetExtension(method.Options(), options.E_ProxyAuthorization).(*options.Validator); option != nil &&
				Validator.ProxyAuthorizationValidator(option.Name) == nil {
				panic(fmt.Sprintf("%s options.proxy_authorization validator: [%s] not found", fullMethod, option.Name))
			}
		}
	}
}

func (f *fileDescriptor) Options(fullMethod string) protoreflect.ProtoMessage {
	f.RLock()
	defer f.RUnlock()

	return f.options[fullMethod]
}
