package aipmiddleware

import (
	"context"
	"net"
	"sync"
	"testing"

	examplefreightv1 "go.einride.tech/aip/proto/gen/einride/example/freight/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"gotest.tools/v3/assert"
)

func TestParentValidator(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		server  mockFreightServer
		logOnly bool
		f       func(*testing.T, context.Context, *mockFreightServer, *grpc.ClientConn)
	}{
		{
			name: "non-parent request",
			server: mockFreightServer{
				getShipperResponse: &examplefreightv1.Shipper{Name: "shippers/1"},
			},
			f: func(t *testing.T, ctx context.Context, server *mockFreightServer, conn *grpc.ClientConn) {
				client := examplefreightv1.NewFreightServiceClient(conn)
				response, err := client.GetShipper(ctx, &examplefreightv1.GetShipperRequest{
					Name: "shippers/1",
				})
				assert.NilError(t, err)
				assert.DeepEqual(t, server.getShipperResponse, response, protocmp.Transform())
			},
		},

		{
			name: "valid parent response",
			server: mockFreightServer{
				listSitesResponse: &examplefreightv1.ListSitesResponse{
					Sites: []*examplefreightv1.Site{
						{Name: "shippers/1/sites/1"},
						{Name: "shippers/1/sites/2"},
						{Name: "shippers/1/sites/3"},
					},
				},
			},
			f: func(t *testing.T, ctx context.Context, server *mockFreightServer, conn *grpc.ClientConn) {
				client := examplefreightv1.NewFreightServiceClient(conn)
				response, err := client.ListSites(ctx, &examplefreightv1.ListSitesRequest{
					Parent: "shippers/1",
				})
				assert.NilError(t, err)
				assert.DeepEqual(t, server.listSitesResponse, response, protocmp.Transform())
			},
		},

		{
			name: "error response for parent method",
			server: mockFreightServer{
				listSitesError: status.Error(codes.Internal, "boom"),
			},
			f: func(t *testing.T, ctx context.Context, server *mockFreightServer, conn *grpc.ClientConn) {
				client := examplefreightv1.NewFreightServiceClient(conn)
				response, err := client.ListSites(ctx, &examplefreightv1.ListSitesRequest{
					Parent: "shippers/1",
				})
				assert.Equal(t, codes.Internal, status.Code(err))
				assert.Assert(t, response == nil)
			},
		},

		{
			name: "error response for non-parent method",
			server: mockFreightServer{
				getShipperError: status.Error(codes.Internal, "boom"),
			},
			f: func(t *testing.T, ctx context.Context, server *mockFreightServer, conn *grpc.ClientConn) {
				client := examplefreightv1.NewFreightServiceClient(conn)
				response, err := client.GetShipper(ctx, &examplefreightv1.GetShipperRequest{
					Name: "shippers/1",
				})
				assert.Equal(t, codes.Internal, status.Code(err))
				assert.Assert(t, response == nil)
			},
		},

		{
			name: "valid wildcard parent response",
			server: mockFreightServer{
				listSitesResponse: &examplefreightv1.ListSitesResponse{
					Sites: []*examplefreightv1.Site{
						{Name: "shippers/1/sites/1"},
						{Name: "shippers/2/sites/2"},
						{Name: "shippers/3/sites/3"},
					},
				},
			},
			f: func(t *testing.T, ctx context.Context, server *mockFreightServer, conn *grpc.ClientConn) {
				client := examplefreightv1.NewFreightServiceClient(conn)
				response, err := client.ListSites(ctx, &examplefreightv1.ListSitesRequest{
					Parent: "shippers/-",
				})
				assert.NilError(t, err)
				assert.DeepEqual(t, server.listSitesResponse, response, protocmp.Transform())
			},
		},

		{
			name: "invalid parent response",
			server: mockFreightServer{
				listSitesResponse: &examplefreightv1.ListSitesResponse{
					Sites: []*examplefreightv1.Site{
						{Name: "shippers/1/sites/1"},
						{Name: "shippers/1/sites/2"},
						{Name: "shippers/2/sites/3"}, // invalid
					},
				},
			},
			f: func(t *testing.T, ctx context.Context, server *mockFreightServer, conn *grpc.ClientConn) {
				client := examplefreightv1.NewFreightServiceClient(conn)
				response, err := client.ListSites(ctx, &examplefreightv1.ListSitesRequest{
					Parent: "shippers/1",
				})
				assert.Assert(t, response == nil)
				assert.Equal(t, codes.Internal, status.Code(err))
			},
		},

		{
			name: "invalid parent response with log only",
			server: mockFreightServer{
				listSitesResponse: &examplefreightv1.ListSitesResponse{
					Sites: []*examplefreightv1.Site{
						{Name: "shippers/1/sites/1"},
						{Name: "shippers/1/sites/2"},
						{Name: "shippers/2/sites/3"}, // invalid
					},
				},
			},
			logOnly: true,
			f: func(t *testing.T, ctx context.Context, server *mockFreightServer, conn *grpc.ClientConn) {
				client := examplefreightv1.NewFreightServiceClient(conn)
				response, err := client.ListSites(ctx, &examplefreightv1.ListSitesRequest{
					Parent: "shippers/1",
				})
				assert.NilError(t, err)
				assert.DeepEqual(t, server.listSitesResponse, response, protocmp.Transform())
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			lis, err := net.Listen("tcp", "localhost:0")
			assert.NilError(t, err)
			parentValidator := NewParentValidator(ParentValidatorConfig{
				ErrorLogHook: func(info *grpc.UnaryServerInfo, request, response proto.Message, parent, name string) {
					t.Log(
						info.FullMethod,
						prototext.MarshalOptions{}.Format(request),
						prototext.MarshalOptions{}.Format(response),
						parent,
						name,
					)
				},
				LogOnly: tt.logOnly,
			})
			grpcServer := grpc.NewServer(grpc.UnaryInterceptor(parentValidator.UnaryServerInterceptor))
			examplefreightv1.RegisterFreightServiceServer(grpcServer, &tt.server)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NilError(t, grpcServer.Serve(lis))
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ctx.Done()
				grpcServer.GracefulStop()
			}()
			t.Cleanup(wg.Wait)
			conn, err := grpc.DialContext(
				ctx,
				lis.Addr().String(),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			assert.NilError(t, err)
			tt.f(t, ctx, &tt.server, conn)
		})
	}
}

type mockFreightServer struct {
	examplefreightv1.UnimplementedFreightServiceServer
	// GetShipper
	getShipperResponse *examplefreightv1.Shipper
	getShipperError    error
	// ListSites
	listSitesResponse *examplefreightv1.ListSitesResponse
	listSitesError    error
}

func (m *mockFreightServer) GetShipper(
	_ context.Context,
	_ *examplefreightv1.GetShipperRequest,
) (*examplefreightv1.Shipper, error) {
	return m.getShipperResponse, m.getShipperError
}

func (m *mockFreightServer) ListSites(
	_ context.Context,
	_ *examplefreightv1.ListSitesRequest,
) (*examplefreightv1.ListSitesResponse, error) {
	return m.listSitesResponse, m.listSitesError
}
