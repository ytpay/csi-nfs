/*
Copyright 2017 The Kubernetes Authors.

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

package nfs

import (
	"github.com/cloudfoundry/bytefmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
)

type nfsDriver struct {
	name    string
	nodeID  string
	version string

	endpoint string

	maxStorageCapacity uint64
	nfsServer          string
	nfsSharePoint      string
	nfsLocalMountPoint string

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

const (
	driverName = "csi-nfsplugin"
)

var (
	version = "1.0.0"
)

func NewNFSdriver(nodeID, endpoint, maxstoragecapacity, nfsServer, nfsSharePoint, nfsLocalMountPoint string) *nfsDriver {
	glog.Infof("Driver: %v version: %v", driverName, version)

	msc, err := bytefmt.ToBytes(maxstoragecapacity)
	if err != nil {
		glog.Errorf("failed to parse maxstoragecapacity: %s: %s", maxstoragecapacity, err)
		msc = 50 * bytefmt.GIGABYTE
	}

	n := &nfsDriver{
		name:               driverName,
		nodeID:             nodeID,
		version:            version,
		endpoint:           endpoint,
		maxStorageCapacity: msc,
		nfsServer:          nfsServer,
		nfsSharePoint:      nfsSharePoint,
		nfsLocalMountPoint: nfsLocalMountPoint,
	}

	n.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER})
	n.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})

	return n
}

func (n *nfsDriver) Run() {
	server := NewNonBlockingGRPCServer()
	server.Start(
		n.endpoint,
		NewDefaultIdentityServer(n),
		// NFS plugin has not implemented ControllerServer
		// using default controllerserver.
		NewControllerServer(n),
		NewNodeServer(n),
	)
	server.Wait()
}

func (n *nfsDriver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		glog.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, &csi.VolumeCapability_AccessMode{Mode: c})
	}
	n.cap = vca
}

func (n *nfsDriver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		glog.Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}
	n.cscap = csc
}
