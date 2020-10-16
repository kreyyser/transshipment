package portentries

import (
	"context"
	"errors"
	pb "github.com/kreyyser/transshipment/pb/portentries"
	"github.com/stretchr/testify/require"
	"testing"
)

type DBMock struct {
	err  error
	resp interface{}
}

func (m *DBMock) Reset() {
	m.err = nil
	m.resp = nil
}

func (m *DBMock) SetErr(err error) {
	m.err = err
}

func (m *DBMock) SetResp(resp interface{}) {
	m.resp = resp
}

func (m *DBMock) List(_ context.Context) ([]PortEntry, error) {
	if m.err != nil {
		return nil, m.err
	}

	if res, ok := m.resp.([]PortEntry); ok {
		return res, nil
	}

	return []PortEntry{}, nil
}

func (m *DBMock) Upsert(_ context.Context, port PortEntry) (PortEntry, error) {
	if m.err != nil {
		return port, m.err
	}

	if res, ok := m.resp.(PortEntry); ok {
		return res, nil
	}

	return port, nil
}

func (m *DBMock) BulkUpsert(_ context.Context, ports []PortEntry) ([]PortEntry, error) {
	if m.err != nil {
		return ports, m.err
	}

	if res, ok := m.resp.([]PortEntry); ok {
		return res, nil
	}

	return ports, nil
}

func (m *DBMock) Fetch(_ context.Context, id *int64, slug *string) (PortEntry, error) {
	if m.err != nil {
		return PortEntry{}, m.err
	}

	if res, ok := m.resp.(PortEntry); ok {
		return res, nil
	}

	return PortEntry{}, nil
}

func (m *DBMock) Store(_ context.Context, port *PortEntry) error {
	if m.err != nil {
		return m.err
	}

	return nil
}

func (m *DBMock) Update(_ context.Context, port *PortEntry) error {
	if m.err != nil {
		return m.err
	}

	return nil
}

func (m *DBMock) Delete(_ context.Context, id *int64, slug *string) error {
	if m.err != nil {
		return m.err
	}

	return nil
}

func TestService_CreateOrUpdatePort(t *testing.T) {
	r := require.New(t)
	db := &DBMock{}
	service := Service{
		store: db,
	}
	cases := []struct{
		name  string
		req   *pb.CreatePortRequest
		res   *pb.EmptyResponse
		err   error
		dbErr error
		dbRes *PortEntry
	}{
		{
			name:  "db err",
			req:   &pb.CreatePortRequest{Data: &pb.Port{}},
			err:   errors.New("some db error occured"),
			dbErr: errors.New("some db error occured"),
		},
	}
	
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			defer db.Reset()
			if test.dbRes != nil {
				db.SetResp(*test.dbRes)
			}
			if test.dbErr != nil {
				db.SetErr(test.dbErr)
			}

			resp, err := service.CreateOrUpdatePort(context.Background(), test.req)
			if test.err != nil {
				r.Error(err)
				return
			}
			r.Equal(resp, test.res)
		})
	}
}