package nfs

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/utils/mount"
)

func NewIdentityServer(d *nfsDriver) *IdentityServer {
	return &IdentityServer{
		Driver: d,
	}
}

func NewControllerServer(d *nfsDriver) *ControllerServer {
	if d.nfsLocalMountOptions == "" {
		// default mount options
		d.nfsLocalMountOptions = "rw,soft,timeo=10,retry=3,vers=4"
	}
	if strings.Contains(d.nfsLocalMountOptions, "rw") {
		logrus.Warn("nfs server is not mounted with rw mode, volume creation may fail")
	}
	mounter := mount.New("")
	source := fmt.Sprintf("%s:%s", d.nfsServer, d.nfsSharePoint)
	logrus.Infof("mount local nfs: %s => %s(%s)", source, d.nfsLocalMountPoint, d.nfsLocalMountOptions)
	err := mounter.Mount(source, d.nfsLocalMountPoint, "nfs", strings.Split(d.nfsLocalMountOptions, ","))
	if err != nil {
		logrus.Fatalf("failed to mount [%s:%s] to local mount point: %s:%s", d.nfsServer, d.nfsSharePoint, d.nfsLocalMountPoint, err)
	}

	return &ControllerServer{
		Driver:  d,
		mounter: mounter,
	}
}

func NewNodeServer(n *nfsDriver) *NodeServer {
	return &NodeServer{
		Driver:  n,
		mounter: mount.New(""),
	}
}

func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func ParseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("invalid endpoint: %v", ep)
}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logrus.Infof("GRPC call: %s", info.FullMethod)
	logrus.Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	resp, err := handler(ctx, req)
	if err != nil {
		logrus.Errorf("GRPC error: %v", err)
	} else {
		logrus.Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}
	return resp, err
}
