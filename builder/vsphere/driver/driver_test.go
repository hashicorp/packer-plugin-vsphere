// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common/utils"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

func newTestDriver(t *testing.T) Driver {
	vcenter := utils.GetenvOrDefault(utils.EnvVcenterServer, utils.DefaultVcenterServer)
	username := utils.GetenvOrDefault(utils.EnvVsphereUsername, utils.DefaultVsphereUsername)
	password := utils.GetenvOrDefault(utils.EnvVspherePassword, utils.DefaultVspherePassword)

	d, err := NewDriver(&ConnectConfig{
		VCenterServer:      vcenter,
		Username:           username,
		Password:           password,
		InsecureConnection: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	return d
}

func newVMName() string {
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	return fmt.Sprintf("test-%v", r.Intn(1000))
}

type VCenterSimulator struct {
	model  *simulator.Model
	server *simulator.Server
	driver *VCenterDriver
}

func NewCustomVCenterSimulator(model *simulator.Model) (*VCenterSimulator, error) {
	sim := new(VCenterSimulator)
	sim.model = model

	server, err := sim.NewSimulatorServer()
	if err != nil {
		sim.Close()
		return nil, err
	}
	sim.server = server

	driver, err := sim.NewSimulatorDriver()
	if err != nil {
		sim.Close()
		return nil, err
	}
	sim.driver = driver
	return sim, nil
}

func NewVCenterSimulator() (*VCenterSimulator, error) {
	model := simulator.VPX()
	model.Machine = 1
	return NewCustomVCenterSimulator(model)
}

func (s *VCenterSimulator) Close() {
	if s.model != nil {
		s.model.Remove()
	}
	if s.server != nil {
		s.server.Close()
	}
}

// Simulator shortcut to choose any pre-created virtual machine.
func (s *VCenterSimulator) ChooseSimulatorPreCreatedVM() (VirtualMachine, *simulator.VirtualMachine) {
	machine := s.model.Map().Any("VirtualMachine").(*simulator.VirtualMachine)
	ref := machine.Reference()
	vm := s.driver.NewVM(&ref)
	return vm, machine
}

// Simulator shortcut to choose any pre-created datastore.
func (s *VCenterSimulator) ChooseSimulatorPreCreatedDatastore() (Datastore, *simulator.Datastore) {
	ds := s.model.Map().Any("Datastore").(*simulator.Datastore)
	ref := ds.Reference()
	datastore := s.driver.NewDatastore(&ref)
	return datastore, ds
}

// Simulator shortcut to choose any pre-created EXSi host.
func (s *VCenterSimulator) ChooseSimulatorPreCreatedHost() (*Host, *simulator.HostSystem) {
	h := s.model.Map().Any("HostSystem").(*simulator.HostSystem)
	ref := h.Reference()
	host := s.driver.NewHost(&ref)
	return host, h
}

func (s *VCenterSimulator) NewSimulatorServer() (*simulator.Server, error) {
	err := s.model.Create()
	if err != nil {
		return nil, err
	}

	s.model.Service.RegisterEndpoints = true
	s.model.Service.TLS = new(tls.Config)
	s.model.Service.ServeMux = http.NewServeMux()
	return s.model.Service.NewServer(), nil
}

func (s *VCenterSimulator) NewSimulatorDriver() (*VCenterDriver, error) {
	ctx := context.TODO()
	user := &url.Userinfo{}
	s.server.URL.User = user

	soapClient := soap.NewClient(s.server.URL, true)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	vimClient.RoundTripper = session.KeepAlive(vimClient.RoundTripper, 10*time.Minute)
	client := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	err = client.SessionManager.Login(ctx, user)
	if err != nil {
		return nil, err
	}

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, "")
	if err != nil {
		return nil, err
	}
	finder.SetDatacenter(datacenter)

	d := &VCenterDriver{
		ctx:       ctx,
		client:    client,
		vimClient: vimClient,
		restClient: &RestClient{
			client:      rest.NewClient(vimClient),
			credentials: user,
		},
		datacenter: datacenter,
		finder:     finder,
	}
	return d, nil
}
