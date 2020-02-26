package nfs

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/pborman/uuid"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/mount"
)

type ControllerServer struct {
	Driver  *nfsDriver
	mounter mount.Interface
}

func (cs *ControllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}
	caps := req.GetVolumeCapabilities()
	if caps == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}
	for _, cap := range caps {
		if cap.GetBlock() != nil {
			return nil, status.Error(codes.Unimplemented, "Block Volume not supported")
		}
	}

	volID := uuid.NewUUID().String()
	logrus.Infof("create volume: %s", volID)

	volPath := filepath.Join(cs.Driver.nfsLocalMountPoint, volID)
	_, err := os.Stat(volPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(volPath, 0755)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	capacity := uint64(req.GetCapacityRange().GetRequiredBytes())
	if capacity >= cs.Driver.maxStorageCapacity {
		return nil, status.Errorf(codes.OutOfRange, "Requested capacity %d exceeds maximum allowed %d", capacity, cs.Driver.maxStorageCapacity)
	}

	volContext := req.GetParameters()
	volContext["server"] = cs.Driver.nfsServer
	volContext["share"] = filepath.Join(cs.Driver.nfsSharePoint, volID)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volID,
			VolumeContext: volContext,
			CapacityBytes: int64(capacity),
		},
	}, nil

}

func (cs *ControllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	logrus.Infof("remove volume: %s", req.VolumeId)
	volPath := filepath.Join(cs.Driver.nfsLocalMountPoint, req.VolumeId)
	_, err := os.Stat(volPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &csi.DeleteVolumeResponse{}, nil
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		err := os.RemoveAll(volPath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		} else {
			return &csi.DeleteVolumeResponse{}, nil
		}
	}
}

func (cs *ControllerServer) ControllerPublishVolume(_ context.Context, _ *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ControllerPublishVolume")
}

func (cs *ControllerServer) ControllerUnpublishVolume(_ context.Context, _ *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ControllerUnpublishVolume")
}

func (cs *ControllerServer) ValidateVolumeCapabilities(_ context.Context, _ *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ValidateVolumeCapabilities")
}

func (cs *ControllerServer) ListVolumes(_ context.Context, _ *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ListVolumes")
}

func (cs *ControllerServer) GetCapacity(_ context.Context, _ *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented GetCapacity")
}

// ControllerGetCapabilities implements the default GRPC callout.
// Default supports all capabilities
func (cs *ControllerServer) ControllerGetCapabilities(_ context.Context, _ *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	logrus.Infof("Using default ControllerGetCapabilities")

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.cscap,
	}, nil
}

func (cs *ControllerServer) CreateSnapshot(_ context.Context, _ *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented CreateSnapshot")
}

func (cs *ControllerServer) DeleteSnapshot(_ context.Context, _ *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented DeleteSnapshot")
}

func (cs *ControllerServer) ListSnapshots(_ context.Context, _ *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ListSnapshots")
}

func (cs *ControllerServer) ControllerExpandVolume(_ context.Context, _ *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ControllerExpandVolume")
}
