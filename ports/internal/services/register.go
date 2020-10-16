package services

import "google.golang.org/grpc"

// GRPCService is service which can be registered with a gRPC server
type GRPCService interface {
	GRPCRegister(srvr *grpc.Server)
}

// TODO might be added any other possible interfaces like JSONProxy or HTTP servers, etc.