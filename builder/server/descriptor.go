package server

import (
	"github.com/bluekaki/vv/internal/interceptor"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ParseFileDescriptor parse file descriptor
func ParseFileDescriptorP(descriptor protoreflect.FileDescriptor) {
	if descriptor == nil {
		panic("file descriptor required")
	}

	interceptor.FileDescriptor.Parse(descriptor)
}
