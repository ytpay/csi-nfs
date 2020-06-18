package nfs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/ptypes"

	"k8s.io/utils/exec"

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

func (cs *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	reqVolName := req.GetName()
	if reqVolName == "" {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}
	logrus.Infof("create volume: %s", reqVolName)

	caps := req.GetVolumeCapabilities()
	if caps == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}
	for _, c := range caps {
		if c.GetBlock() != nil {
			return nil, status.Error(codes.Unimplemented, "Block Volume not supported")
		}
	}

	volPath := filepath.Join(cs.Driver.nfsLocalMountPoint, reqVolName)
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
	volContext["share"] = filepath.Join(cs.Driver.nfsSharePoint, reqVolName)

	if req.GetVolumeContentSource() != nil {
		contentSource := req.GetVolumeContentSource()
		if contentSource.GetSnapshot() != nil {
			snapID := contentSource.GetSnapshot().GetSnapshotId()
			snapPath := filepath.Join(cs.Driver.nfsLocalMountPoint, cs.Driver.nfsSnapshotPath)
			_, err := os.Stat(snapPath)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			targetPath := filepath.Join(snapPath, snapID+".tar.gz")
			_, err = os.Stat(targetPath)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			outBs, err := exec.New().CommandContext(ctx, "tar", "-zxpf", targetPath, "-C", volPath).CombinedOutput()
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("%s: %s", err.Error(), string(outBs)))
			}
		}

	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      reqVolName,
			VolumeContext: volContext,
			CapacityBytes: int64(capacity),
		},
	}, nil

}

func (cs *ControllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	logrus.Infof("delete volume: %s", req.VolumeId)
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

func (cs *ControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	volPath := filepath.Join(cs.Driver.nfsLocalMountPoint, req.SourceVolumeId)
	_, err := os.Stat(volPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	snapPath := filepath.Join(cs.Driver.nfsLocalMountPoint, cs.Driver.nfsSnapshotPath)
	_, err = os.Stat(snapPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	snapID := uuid.NewUUID().String()
	logrus.Infof("create volume [%s] snapshot: %s", req.SourceVolumeId, snapID)
	targetPath := filepath.Join(snapPath, snapID+".tar.gz")

	outBs, err := exec.New().CommandContext(ctx, "tar", "-zcpf", targetPath, volPath).CombinedOutput()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("%s: %s", err.Error(), string(outBs)))
	}

	stat, err := os.Stat(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.CreateSnapshotResponse{Snapshot: &csi.Snapshot{
		SizeBytes:      stat.Size(),
		SnapshotId:     snapID,
		SourceVolumeId: req.SourceVolumeId,
		CreationTime:   ptypes.TimestampNow(),
		ReadyToUse:     true,
	}}, nil

}

func (cs *ControllerServer) DeleteSnapshot(_ context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	logrus.Infof("delete volume snapshot: %s", req.SnapshotId)
	snapPath := filepath.Join(cs.Driver.nfsLocalMountPoint, cs.Driver.nfsSnapshotPath)
	_, err := os.Stat(snapPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	targetPath := filepath.Join(snapPath, req.SnapshotId+".tar.gz")
	err = os.Remove(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.DeleteSnapshotResponse{}, nil
}

func (cs *ControllerServer) ListSnapshots(_ context.Context, _ *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ListSnapshots")
}

func (cs *ControllerServer) ControllerExpandVolume(_ context.Context, _ *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ControllerExpandVolume")
}

func (cs *ControllerServer) ControllerGetVolume(_ context.Context, _ *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ControllerGetVolume")
}
