package interceptor

import (
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const _ = grpc.SupportPackageIsVersion7

var FileDescriptor = &fileDescriptor{
	options: make(map[string]protoreflect.ProtoMessage),
}

type fileDescriptor struct {
	sync.RWMutex
	options map[string]protoreflect.ProtoMessage // FullMethod : Options
}

func (f *fileDescriptor) Parse(descriptor protoreflect.FileDescriptor) {
	f.Lock()
	defer f.Unlock()

	serivces := descriptor.Services()
	for i := 0; i < serivces.Len(); i++ {
		serivce := serivces.Get(i)
		methods := serivce.Methods()

		for k := 0; k < methods.Len(); k++ {
			method := methods.Get(k)
			f.options[fmt.Sprintf("/%s/%s", serivce.FullName(), method.Name())] = method.Options()
		}
	}
}

func (f *fileDescriptor) Options(fullMethod string) protoreflect.ProtoMessage {
	f.RLock()
	defer f.RUnlock()

	return f.options[fullMethod]
}
