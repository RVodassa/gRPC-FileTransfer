// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package file_transfer

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FileTransferClient is the client API for FileTransfer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FileTransferClient interface {
	UploadFile(ctx context.Context, opts ...grpc.CallOption) (FileTransfer_UploadFileClient, error)
	ListFiles(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ListFilesResponse, error)
	GetFile(ctx context.Context, in *GetFileRequest, opts ...grpc.CallOption) (FileTransfer_GetFileClient, error)
}

type fileTransferClient struct {
	cc grpc.ClientConnInterface
}

func NewFileTransferClient(cc grpc.ClientConnInterface) FileTransferClient {
	return &fileTransferClient{cc}
}

func (c *fileTransferClient) UploadFile(ctx context.Context, opts ...grpc.CallOption) (FileTransfer_UploadFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &FileTransfer_ServiceDesc.Streams[0], "/file_transfer.FileTransfer/UploadFile", opts...)
	if err != nil {
		return nil, err
	}
	x := &fileTransferUploadFileClient{stream}
	return x, nil
}

type FileTransfer_UploadFileClient interface {
	Send(*UploadFileRequest) error
	CloseAndRecv() (*UploadFileResponse, error)
	grpc.ClientStream
}

type fileTransferUploadFileClient struct {
	grpc.ClientStream
}

func (x *fileTransferUploadFileClient) Send(m *UploadFileRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *fileTransferUploadFileClient) CloseAndRecv() (*UploadFileResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(UploadFileResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *fileTransferClient) ListFiles(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ListFilesResponse, error) {
	out := new(ListFilesResponse)
	err := c.cc.Invoke(ctx, "/file_transfer.FileTransfer/ListFiles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileTransferClient) GetFile(ctx context.Context, in *GetFileRequest, opts ...grpc.CallOption) (FileTransfer_GetFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &FileTransfer_ServiceDesc.Streams[1], "/file_transfer.FileTransfer/GetFile", opts...)
	if err != nil {
		return nil, err
	}
	x := &fileTransferGetFileClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type FileTransfer_GetFileClient interface {
	Recv() (*GetFileResponse, error)
	grpc.ClientStream
}

type fileTransferGetFileClient struct {
	grpc.ClientStream
}

func (x *fileTransferGetFileClient) Recv() (*GetFileResponse, error) {
	m := new(GetFileResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// FileTransferServer is the server API for FileTransfer service.
// All implementations must embed UnimplementedFileTransferServer
// for forward compatibility
type FileTransferServer interface {
	UploadFile(FileTransfer_UploadFileServer) error
	ListFiles(context.Context, *Empty) (*ListFilesResponse, error)
	GetFile(*GetFileRequest, FileTransfer_GetFileServer) error
	mustEmbedUnimplementedFileTransferServer()
}

// UnimplementedFileTransferServer must be embedded to have forward compatible implementations.
type UnimplementedFileTransferServer struct {
}

func (UnimplementedFileTransferServer) UploadFile(FileTransfer_UploadFileServer) error {
	return status.Errorf(codes.Unimplemented, "method UploadFile not implemented")
}
func (UnimplementedFileTransferServer) ListFiles(context.Context, *Empty) (*ListFilesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFiles not implemented")
}
func (UnimplementedFileTransferServer) GetFile(*GetFileRequest, FileTransfer_GetFileServer) error {
	return status.Errorf(codes.Unimplemented, "method GetFile not implemented")
}
func (UnimplementedFileTransferServer) mustEmbedUnimplementedFileTransferServer() {}

// UnsafeFileTransferServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FileTransferServer will
// result in compilation errors.
type UnsafeFileTransferServer interface {
	mustEmbedUnimplementedFileTransferServer()
}

func RegisterFileTransferServer(s grpc.ServiceRegistrar, srv FileTransferServer) {
	s.RegisterService(&FileTransfer_ServiceDesc, srv)
}

func _FileTransfer_UploadFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(FileTransferServer).UploadFile(&fileTransferUploadFileServer{stream})
}

type FileTransfer_UploadFileServer interface {
	SendAndClose(*UploadFileResponse) error
	Recv() (*UploadFileRequest, error)
	grpc.ServerStream
}

type fileTransferUploadFileServer struct {
	grpc.ServerStream
}

func (x *fileTransferUploadFileServer) SendAndClose(m *UploadFileResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *fileTransferUploadFileServer) Recv() (*UploadFileRequest, error) {
	m := new(UploadFileRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _FileTransfer_ListFiles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileTransferServer).ListFiles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/file_transfer.FileTransfer/ListFiles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileTransferServer).ListFiles(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileTransfer_GetFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetFileRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FileTransferServer).GetFile(m, &fileTransferGetFileServer{stream})
}

type FileTransfer_GetFileServer interface {
	Send(*GetFileResponse) error
	grpc.ServerStream
}

type fileTransferGetFileServer struct {
	grpc.ServerStream
}

func (x *fileTransferGetFileServer) Send(m *GetFileResponse) error {
	return x.ServerStream.SendMsg(m)
}

// FileTransfer_ServiceDesc is the grpc.ServiceDesc for FileTransfer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FileTransfer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "file_transfer.FileTransfer",
	HandlerType: (*FileTransferServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListFiles",
			Handler:    _FileTransfer_ListFiles_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "UploadFile",
			Handler:       _FileTransfer_UploadFile_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "GetFile",
			Handler:       _FileTransfer_GetFile_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/protos/file_transfer.proto",
}
