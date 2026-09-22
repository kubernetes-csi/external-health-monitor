/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mock

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FakeControllerServer is a programmable in-process CSI controller service
// used by tests in place of a real driver. Only the volume health RPCs are
// implemented; every other controller RPC returns codes.Unimplemented via
// the embedded csi.UnimplementedControllerServer.
type FakeControllerServer struct {
	csi.UnimplementedControllerServer

	mu sync.Mutex

	listResp      *csi.ControllerListVolumeHealthResponse
	listPages     map[string]*csi.ControllerListVolumeHealthResponse
	listErr       error
	listReqs      []*csi.ControllerListVolumeHealthRequest
	listDeadlines []time.Time

	getResp      map[string]*csi.ControllerGetVolumeHealthResponse
	getErr       map[string]error
	getReqs      []*csi.ControllerGetVolumeHealthRequest
	getDeadlines []time.Time
	getTimes     []time.Time
}

var _ csi.ControllerServer = &FakeControllerServer{}

func (f *FakeControllerServer) SetListVolumeHealth(resp *csi.ControllerListVolumeHealthResponse, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listResp = resp
	f.listErr = err
}

// SetListVolumeHealthPages serves a paginated listing: the first page answers
// an empty starting token and each page's NextToken selects the next one.
func (f *FakeControllerServer) SetListVolumeHealthPages(pages ...*csi.ControllerListVolumeHealthResponse) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listPages = map[string]*csi.ControllerListVolumeHealthResponse{}
	token := ""
	for _, page := range pages {
		f.listPages[token] = page
		token = page.GetNextToken()
	}
}

func (f *FakeControllerServer) SetGetVolumeHealth(volumeID string, resp *csi.ControllerGetVolumeHealthResponse, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getResp == nil {
		f.getResp = map[string]*csi.ControllerGetVolumeHealthResponse{}
	}
	if f.getErr == nil {
		f.getErr = map[string]error{}
	}
	f.getResp[volumeID] = resp
	f.getErr[volumeID] = err
}

func (f *FakeControllerServer) ListVolumeHealthRequests() []*csi.ControllerListVolumeHealthRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*csi.ControllerListVolumeHealthRequest, len(f.listReqs))
	copy(out, f.listReqs)
	return out
}

func (f *FakeControllerServer) GetVolumeHealthRequests() []*csi.ControllerGetVolumeHealthRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*csi.ControllerGetVolumeHealthRequest, len(f.getReqs))
	copy(out, f.getReqs)
	return out
}

// ListVolumeHealthDeadlines returns the context deadline each received list
// request carried; the zero time means the request had no deadline.
func (f *FakeControllerServer) ListVolumeHealthDeadlines() []time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]time.Time, len(f.listDeadlines))
	copy(out, f.listDeadlines)
	return out
}

// GetVolumeHealthDeadlines is ListVolumeHealthDeadlines for get requests.
func (f *FakeControllerServer) GetVolumeHealthDeadlines() []time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]time.Time, len(f.getDeadlines))
	copy(out, f.getDeadlines)
	return out
}

// GetVolumeHealthTimes returns when each get request arrived, in the same
// order as GetVolumeHealthRequests.
func (f *FakeControllerServer) GetVolumeHealthTimes() []time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]time.Time, len(f.getTimes))
	copy(out, f.getTimes)
	return out
}

func (f *FakeControllerServer) ControllerListVolumeHealth(ctx context.Context, req *csi.ControllerListVolumeHealthRequest) (*csi.ControllerListVolumeHealthResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listReqs = append(f.listReqs, req)
	deadline, _ := ctx.Deadline()
	f.listDeadlines = append(f.listDeadlines, deadline)
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.listPages != nil {
		if resp, ok := f.listPages[req.GetStartingToken()]; ok {
			return resp, nil
		}
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("fake driver: no list page configured for starting token %q", req.GetStartingToken()))
	}
	if f.listResp == nil {
		return nil, status.Error(codes.FailedPrecondition, "fake driver: no ControllerListVolumeHealth response configured")
	}
	return f.listResp, nil
}

func (f *FakeControllerServer) ControllerGetVolumeHealth(ctx context.Context, req *csi.ControllerGetVolumeHealthRequest) (*csi.ControllerGetVolumeHealthResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getReqs = append(f.getReqs, req)
	deadline, _ := ctx.Deadline()
	f.getDeadlines = append(f.getDeadlines, deadline)
	f.getTimes = append(f.getTimes, time.Now())
	if err := f.getErr[req.GetVolumeId()]; err != nil {
		return nil, err
	}
	if resp, ok := f.getResp[req.GetVolumeId()]; ok {
		return resp, nil
	}
	return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("fake driver: no ControllerGetVolumeHealth response configured for volume %q", req.GetVolumeId()))
}

// fakeIdentityServer is registered so a stray identity RPC fails with a clear
// Unimplemented error instead of "unknown service".
type fakeIdentityServer struct {
	csi.UnimplementedIdentityServer
}

// FakeCSIDriver serves a FakeControllerServer on a unix socket.
type FakeCSIDriver struct {
	Controller *FakeControllerServer

	address string
	server  *grpc.Server
}

func (d *FakeCSIDriver) Address() string {
	return d.address
}

// StartFakeDriver runs a fake CSI driver on a unix socket and returns it with
// a client connection. Both are cleaned up when the test ends.
func StartFakeDriver(t *testing.T) (*FakeCSIDriver, *grpc.ClientConn) {
	t.Helper()

	socket := filepath.Join(t.TempDir(), "csi.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatalf("failed to listen on %s: %v", socket, err)
	}

	drv := &FakeCSIDriver{
		Controller: &FakeControllerServer{},
		address:    socket,
		server:     grpc.NewServer(),
	}
	csi.RegisterControllerServer(drv.server, drv.Controller)
	csi.RegisterIdentityServer(drv.server, &fakeIdentityServer{})
	go func() {
		_ = drv.server.Serve(listener)
	}()
	t.Cleanup(drv.server.Stop)

	csiConn, err := New(context.Background(), drv.Address())
	if err != nil {
		t.Fatalf("failed to connect to fake driver: %v", err)
	}
	t.Cleanup(func() {
		_ = csiConn.Close()
	})

	return drv, csiConn
}
