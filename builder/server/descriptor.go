package server

import (
	"github.com/bluekaki/vv/internal/interceptor"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ParseFileDescriptorP parse file descriptor
func ParseFileDescriptorP(descriptor protoreflect.FileDescriptor) {
	if descriptor == nil {
		panic("file descriptor required")
	}

	interceptor.FileDescriptor.ParseP(descriptor)
}
