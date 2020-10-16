package portentries

import (
	"context"
	"github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/kreyyser/transshipment/pb/portentries"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

var _ interface {
	pb.PortsServiceServer
} = &Service{}

// Service encapsulates ports operations
type Service struct {
	db           *gorm.DB
	store        PortsDB
}

// NewPortsService returns pointer to newly created Ports service
func NewPortsService(db *gorm.DB) *Service {
	return &Service{
		db:    db,
		store: NewDatastore(db),
	}
}

// GRPCRegister adds this service to the gRPC server
func (s *Service) GRPCRegister(srvr *grpc.Server) {
	pb.RegisterPortsServiceServer(srvr, s)
}

// CreateOrUpdatePort creates port based on request or updates it
func (s Service) CreateOrUpdatePort(ctx context.Context, req *pb.CreatePortRequest) (*pb.EmptyResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.store.Upsert(ctx, pbPortToModel(req.Data)); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

// CreateOrUpdatePortBulk creates ports based on request or updates them
func (s Service) CreateOrUpdatePortBulk(ctx context.Context, req *pb.UpsertPortBulkRequest) (*pb.EmptyResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.store.BulkUpsert(ctx, pbToPorts(req.Data)); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

// ListPorts returns list of ports
func (s Service) ListPorts(ctx context.Context, req *pb.ListPortsRequest) (*pb.ListPortsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	ports, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ListPortsResponse{Data: portsToPB(ports)}, nil
}

// FetchPort returns ports by id or slug
func (s Service) FetchPort(ctx context.Context, req *pb.PortRequest) (*pb.PortResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	id, slug := extractIdOrSlug(req)
	port, err := s.store.Fetch(ctx, id, slug)
	if err != nil {
		return nil, err
	}

	return &pb.PortResponse{Data: portToPB(port)}, nil
}

// CreatePort creates new port
func (s Service) CreatePort(ctx context.Context, req *pb.CreatePortRequest) (*pb.PortResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	port := pbPortToModel(req.Data)
	if err := s.store.Store(ctx, &port); err != nil {
		return nil, err
	}

	return &pb.PortResponse{Data: portToPB(port)}, nil
}

// UpdatePort updates port by id or slug
func (s Service) UpdatePort(ctx context.Context, req *pb.UpdatePortRequest) (*pb.PortResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	id, slug := extractIdOrSlug(req)
	port, err := s.store.Fetch(ctx, id, slug)
	if err != nil {
		return nil, err
	}

	if req.Data.Code != nil {
		port.Code = req.Data.Code.Value
	}

	if req.Data.Unlocks != nil {
		port.Unlocks = req.Data.Unlocks
	}

	if req.Data.Timezone != nil {
		port.Timezone = req.Data.Timezone.Value
	}

	if req.Data.Coordinates != nil {
		port.Longitude = req.Data.Coordinates.Lng
		port.Latitude = req.Data.Coordinates.Lat
	}

	if req.Data.Regions != nil {
		port.Regions = req.Data.Regions
	}

	if req.Data.Alias != nil {
		port.Alias = req.Data.Alias
	}

	if req.Data.Country != nil {
		port.Country = req.Data.Country.Value
	}

	if req.Data.Province != nil {
		port.Province = req.Data.Province.Value
	}

	if req.Data.City != nil {
		port.City = req.Data.City.Value
	}

	if req.Data.Name != nil {
		port.Name = req.Data.Name.Value
	}

	if err := s.store.Update(ctx, &port); err != nil {
		return nil, err
	}

	return &pb.PortResponse{Data: portToPB(port)}, nil
}

// DeletePort deletes port by id or slug
func (s Service) DeletePort(ctx context.Context, req *pb.PortRequest) (*pb.EmptyResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	id, slug := extractIdOrSlug(req)

	if err := s.store.Delete(ctx, id, slug); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

type IDSlugger interface {
	GetId() *wrappers.Int64Value
	GetSlug() *wrappers.StringValue
}

func extractIdOrSlug(req IDSlugger) (*int64, *string) {
	var (
		id int64
		slug string
	)

	if req.GetId() != nil {
		id = req.GetId().Value
	}

	if req.GetSlug() != nil {
		slug = req.GetSlug().Value
	}

	return &id, &slug
}

func pbToPorts(protos []*pb.Port) []PortEntry {
	ports := make([]PortEntry, 0, len(protos))

	for _, v := range protos {
		if v != nil {
			ports = append(ports, pbPortToModel(v))
		}
	}

	return ports
}

func portsToPB(ports []PortEntry) []*pb.Port {
	protos := make([]*pb.Port, 0, len(ports))

	for _, v := range ports {
		protos = append(protos, portToPB(v))
	}

	return protos
}

func pbPortToModel(proto *pb.Port) PortEntry {
	pe := PortEntry{
		Slug:      proto.Slug,
		Code:      proto.Code,
		Name:      proto.Name,
		City:      proto.City,
		Province:  proto.Province,
		Country:   proto.Country,
		Alias:     proto.Alias,
		Regions:   proto.Regions,
		Timezone:  proto.Timezone,
		Unlocks:   proto.Unlocks,
	}

	if proto.Coordinates != nil {
		pe.Latitude =  proto.Coordinates.Lat
		pe.Longitude = proto.Coordinates.Lng
	}

	return pe
}

func portToPB(port PortEntry) *pb.Port {
	return &pb.Port{
		Slug:        port.Slug,
		Id:          &wrappers.Int64Value{Value: port.ID},
		Name:        port.Name,
		City:        port.City,
		Province:    port.Province,
		Country:     port.Country,
		Alias:       port.Alias,
		Regions:     port.Regions,
		Coordinates: &pb.Coordinates{
			Lng: port.Longitude,
			Lat: port.Latitude,
		},
		Timezone:    port.Timezone,
		Unlocks:     port.Unlocks,
		Code:        port.Code,
	}
}
