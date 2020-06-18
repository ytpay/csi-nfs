package nfs

import (
	"code.cloudfoundry.org/bytefmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
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
	nfsSnapshotPath    string

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

const (
	driverName = "csi-nfs"
)

func NewNFSDriver(version, nodeID, endpoint, maxstoragecapacity, nfsServer, nfsSharePoint, nfsLocalMountPoint, nfsSnapshotPath string) *nfsDriver {
	logrus.Infof("Driver: %s version: %s", driverName, version)

	msc, err := bytefmt.ToBytes(maxstoragecapacity)
	if err != nil {
		logrus.Errorf("failed to parse maxstoragecapacity: %s: %s", maxstoragecapacity, err)
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
		nfsSnapshotPath:    nfsSnapshotPath,
	}

	n.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	})

	n.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
	})

	return n
}

func (n *nfsDriver) Run() {
	server := NewNonBlockingGRPCServer()
	server.Start(
		n.endpoint,
		NewIdentityServer(n),
		NewControllerServer(n),
		NewNodeServer(n),
	)
	server.Wait()
}

func (n *nfsDriver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		logrus.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, &csi.VolumeCapability_AccessMode{Mode: c})
	}
	n.cap = vca
}

func (n *nfsDriver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		logrus.Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}
	n.cscap = csc
}
