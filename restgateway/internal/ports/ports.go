package ports

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/kreyyser/transshipment/common/jsonparser"
	"github.com/kreyyser/transshipment/common/router"
	pb "github.com/kreyyser/transshipment/pb/portentries"
	"google.golang.org/grpc"
	"net/http"
	"strconv"
)

const (
	MaxMemoryAllowed = 10 << 20 // 10MB
	ChunkSize        = 20       // 20 ports per time
)

// PortServer is a struct to hold connections to grpc clients and to hold route handlers
type PortServer struct {
	portsConn   *grpc.ClientConn
	portsClient pb.PortsServiceClient
}

// NewServer creates new PortServer instance
func NewServer(portsConn *grpc.ClientConn) *PortServer {
	client := pb.NewPortsServiceClient(portsConn)

	return &PortServer{
		portsConn:   portsConn,
		portsClient: client,
	}
}

// Routes returns set of routes served by PortServer
func Routes(portsConn *grpc.ClientConn) []router.Route {
	srv := NewServer(portsConn)

	return []router.Route{
		{
			Method:  http.MethodGet,
			Path:    "/ports",
			Handler: srv.ListPorts,
		},
		{
			Method:  http.MethodPost,
			Path:    "/ports",
			Handler: srv.CreatePort,
		},
		{
			Method:  http.MethodGet,
			Path:    "/ports/{idOrSlug}",
			Handler: srv.FetchPort,
		},
		{
			Method:  http.MethodPut,
			Path:    "/ports/{idOrSlug}",
			Handler: srv.UpdatePort,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/ports/{idOrSlug}",
			Handler: srv.DeletePort,
		},
		{
			Method:  http.MethodPost,
			Path:    "/upload-ports",
			Handler: srv.UploadPortsState,
		},
	}
}

// ListPorts returns array of Port to the client
func (s *PortServer) ListPorts(w http.ResponseWriter, r *http.Request) {
	res, err := s.portsClient.ListPorts(r.Context(), &pb.ListPortsRequest{})
	if err != nil {
		respondError(err.Error(), w)
		return
	}

	if res != nil && res.Data != nil {
		ports := make([]Port, 0, len(res.Data))
		for _, v := range res.Data {
			ports = append(ports, fromPbPort(v))
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ports); err != nil {
			fmt.Printf("error encoding response payload. err: %v", err)
			respondError(err.Error(), w)

			return
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]struct{}{}); err != nil {
			fmt.Printf("error encoding response payload. err: %v", err)
			respondError(err.Error(), w)

			return
		}
	}
}

// CreatePort creates new Port record
func (s *PortServer) CreatePort(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit just in case...
	dec := json.NewDecoder(r.Body)
	var port Port
	if err := dec.Decode(&port); err != nil {
		respondError(err.Error(), w)
		return
	}

	res, err := s.portsClient.CreatePort(r.Context(), &pb.CreatePortRequest{Data: toPbPort(port)})
	if err != nil {
		respondError(err.Error(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fromPbPort(res.Data)); err != nil {
		fmt.Printf("error encoding response payload. err: %v", err)
		respondError(err.Error(), w)

		return
	}

	w.WriteHeader(http.StatusCreated)
}

// UpdatePort updates existing Port record
func (s *PortServer) UpdatePort(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "idOrSlug")
	var (
		id   *int64
		slug *string
	)

	if len(idOrSlug) == 0 {
		respondError("wrong identifier or slug provided", w)
		return
	}

	if n, err := strconv.ParseInt(idOrSlug, 10, 64); err == nil {
		id = &n
	} else {
		slug = &idOrSlug
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit just in case...
	dec := json.NewDecoder(r.Body)
	var port Port
	if err := dec.Decode(&port); err != nil {
		respondError(err.Error(), w)
		return
	}

	updReq := &pb.UpdatePortRequest{}
	if id != nil {
		updReq.Id = &wrappers.Int64Value{Value: *id}
	}

	if slug != nil {
		updReq.Slug = &wrappers.StringValue{Value: *slug}
	}

	updReq.Data = toPbPortUpdatable(port)
	res, err := s.portsClient.UpdatePort(r.Context(), updReq)
	if err != nil {
		respondError(err.Error(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fromPbPort(res.Data)); err != nil {
		fmt.Printf("error encoding response payload. err: %v", err)
		respondError(err.Error(), w)

		return
	}
}

// FetchPort returns existing Port record by id or slug
func (s *PortServer) FetchPort(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "idOrSlug")
	var (
		id   *int64
		slug *string
	)

	if len(idOrSlug) == 0 {
		respondError("wrong identifier or slug provided", w)
		return
	}

	if n, err := strconv.ParseInt(idOrSlug, 10, 64); err == nil {
		id = &n
	} else {
		slug = &idOrSlug
	}

	req := &pb.PortRequest{}
	if id != nil {
		req.Id = &wrappers.Int64Value{Value: *id}
	}

	if slug != nil {
		req.Slug = &wrappers.StringValue{Value: *slug}
	}

	res, err := s.portsClient.FetchPort(r.Context(), req)
	if err != nil {
		respondError(err.Error(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fromPbPort(res.Data)); err != nil {
		fmt.Printf("error encoding response payload. err: %v", err)
		respondError(err.Error(), w)

		return
	}
}

// DeletePort deletes existing Port record by id or slug
func (s *PortServer) DeletePort(w http.ResponseWriter, r *http.Request) {
	idOrSlug := chi.URLParam(r, "idOrSlug")
	var (
		id   *int64
		slug *string
	)

	if len(idOrSlug) == 0 {
		respondError("wrong identifier or slug provided", w)
		return
	}

	slug = &idOrSlug
	if n, err := strconv.ParseInt(idOrSlug, 10, 64); err == nil {
		id = &n
	}

	req := &pb.PortRequest{}
	if id != nil {
		req.Id = &wrappers.Int64Value{Value: *id}
	} else {
		req.Slug = &wrappers.StringValue{Value: *slug}
	}

	_, err := s.portsClient.DeletePort(r.Context(), req)
	if err != nil {
		respondError(err.Error(), w)
		return
	}
}

// UploadPortsState allows you to upload file with Port instances and create or update existing Port by slug
// it should have a json object structure
// {
//    "PORT1": {
//      ...
//    },
//    "PORT2": {
//      ...
//    },
// }
func (s *PortServer) UploadPortsState(w http.ResponseWriter, r *http.Request) {
	//go mem.PrintUsage()
	_ = r.ParseMultipartForm(MaxMemoryAllowed) // 10MB
	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(err.Error(), w)
		return
	}
	defer func() { _ = file.Close() }()

	parser := jsonparser.New(MaxMemoryAllowed, ChunkSize) // 10MB and 20 ports at once
	portsChan, errsChan, err := parser.ParseFile(file)
	if err != nil {
		respondError(err.Error(), w)
		return
	}

	var parseErr bool
	var finished bool
	for {
		if parseErr || finished {
			break
		}

		select {
		case data, ok := <-portsChan:
			if !ok && len(data) == 0 {
				finished = true
				break
			}

			ports := make([]Port, 0, len(data))
			for slug, portData := range data {
				bytes, err := json.Marshal(portData)
				if err != nil {
					parseErr = true
					fmt.Println(err)
					break
				}

				var p Port
				if err := json.Unmarshal(bytes, &p); err != nil {
					parseErr = true
					fmt.Println(err)
					break
				}
				p.Slug = slug

				ports = append(ports, p)
			}

			_, err := s.portsClient.CreateOrUpdatePortBulk(r.Context(), &pb.UpsertPortBulkRequest{Data: toPbPorts(ports)})
			if err != nil {
				respondError(err.Error(), w)
				return
			}

			if !ok {
				finished = true
				break
			}
		case _ = <-errsChan:
			parseErr = true
			break
		}
	}
	if parseErr {
		respondInternal(w)
		return
	}
}

// Port is an temporary struct to transfer data from request from user to grpc client
type Port struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	City        string    `json:"city"`
	Province    string    `json:"province"`
	Country     string    `json:"country"`
	Alias       []string  `json:"alias"`
	Regions     []string  `json:"regions"`
	Coordinates []float64 `json:"coordinates"`
	Timezone    string    `json:"timezone"`
	Unlocks     []string  `json:"unlocks"`
	Code        string    `json:"code"`
}

func fromPbPort(proto *pb.Port) Port {
	port := Port{
		Slug:     proto.Slug,
		Name:     proto.Name,
		City:     proto.City,
		Province: proto.Province,
		Country:  proto.Country,
		Alias:    proto.Alias,
		Regions:  proto.Regions,
		Coordinates: []float64{
			proto.Coordinates.Lng,
			proto.Coordinates.Lat,
		},
		Timezone: proto.Timezone,
		Unlocks:  proto.Unlocks,
		Code:     proto.Code,
	}

	if proto.Id != nil {
		port.ID = proto.Id.Value
	}

	return port
}

func toPbPorts(ports []Port) []*pb.Port {
	proto := make([]*pb.Port, 0, len(ports))

	for _, v := range ports {
		proto = append(proto, toPbPort(v))
	}

	return proto
}

func toPbPort(port Port) *pb.Port {
	proto := &pb.Port{
		Slug:     port.Slug,
		Name:     port.Name,
		City:     port.City,
		Province: port.Province,
		Country:  port.Country,
		Alias:    port.Alias,
		Regions:  port.Regions,
		Timezone: port.Timezone,
		Unlocks:  port.Unlocks,
		Code:     port.Country,
	}

	if len(port.Coordinates) == 2 {
		proto.Coordinates = &pb.Coordinates{
			Lng: port.Coordinates[0],
			Lat: port.Coordinates[1],
		}
	}

	return proto
}

func toPbPortUpdatable(port Port) *pb.PortUpdatable {
	proto := &pb.PortUpdatable{
		Name:        toWString(port.Name),
		City:        toWString(port.City),
		Province:    toWString(port.Province),
		Country:     toWString(port.Country),
		Coordinates: nil,
		Timezone:    toWString(port.Timezone),
		Code:        toWString(port.Code),
	}

	if port.Alias != nil {
		proto.Alias = port.Alias
	}

	if port.Regions != nil {
		proto.Regions = port.Regions
	}

	if port.Unlocks != nil {
		proto.Unlocks = port.Unlocks
	}

	if len(port.Coordinates) == 2 {
		proto.Coordinates = &pb.Coordinates{
			Lng: port.Coordinates[0],
			Lat: port.Coordinates[1],
		}
	}

	return proto
}

func toWString(s string) *wrappers.StringValue {
	if s != "" {
		return &wrappers.StringValue{Value: s}
	}

	return nil
}

func respondError(message string, w http.ResponseWriter) {
	payload := map[string]string{
		"message": message,
	}
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Printf("error encoding response payload. err: %v", err)
	}
}

func respondInternal(w http.ResponseWriter) {
	respondError("internal error occured", w)
}
