package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/gozap/csi-nfs/pkg/nfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	Version   string
	BuildDate string
	CommitID  string

	versionTpl = `
Name: %s
Version: %s
Arch: %s
BuildDate: %s
CommitID: %s
`
)

var (
	debug              bool
	name               string
	endpoint           string
	nodeID             string
	maxStorageCapacity string

	nfsServer            string
	nfsSharePoint        string
	nfsLocalMountPoint   string
	nfsLocalMountOptions string
	nfsSnapshotPath      string

	enableIdentityServer   bool
	enableControllerServer bool
	enableNodeServer       bool
)

var rootCmd = &cobra.Command{
	Use:     "csi-nfs",
	Short:   "CSI based NFS driver",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		nfs.NewCSIDriver(
			name,
			strings.TrimPrefix(Version, "v"),
			nodeID,
			endpoint,
			maxStorageCapacity,
			nfsServer,
			nfsSharePoint,
			nfsLocalMountPoint,
			nfsLocalMountOptions,
			nfsSnapshotPath,
			enableIdentityServer,
			enableControllerServer,
			enableNodeServer,
		).Run()
	},
}

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug log")

	rootCmd.PersistentFlags().BoolVar(&enableIdentityServer, "enable-identity-server", false, "Enable Identity gRPC Server")
	rootCmd.PersistentFlags().BoolVar(&enableControllerServer, "enable-controller-server", false, "Enable Controller gRPC Server")
	rootCmd.PersistentFlags().BoolVar(&enableNodeServer, "enable-node-server", false, "Enable Node gRPC Server")

	rootCmd.PersistentFlags().StringVar(&nodeID, "nodeid", "", "CSI Node ID")
	_ = rootCmd.MarkPersistentFlagRequired("nodeid")

	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "unix:///csi/csi.sock", "CSI gRPC Server Endpoint")
	_ = rootCmd.MarkPersistentFlagRequired("endpoint")

	rootCmd.PersistentFlags().StringVar(&name, "name", "csi-nfs", "CSI Driver Name")
	_ = rootCmd.PersistentFlags().MarkHidden("name")

	rootCmd.PersistentFlags().StringVar(&nfsServer, "nfs-server", "", "NFS Server Address")
	rootCmd.PersistentFlags().StringVar(&nfsSharePoint, "nfs-server-share-point", "/", "NFS Server Share Point")
	rootCmd.PersistentFlags().StringVar(&nfsLocalMountPoint, "nfs-local-mount-point", "/nfs", "NFS Local Mount Point")
	rootCmd.PersistentFlags().StringVar(&nfsLocalMountOptions, "nfs-local-mount-options", "rw,vers=4,soft,timeo=10,retry=3", "NFS Local Mount Options")
	rootCmd.PersistentFlags().StringVar(&nfsSnapshotPath, "nfs-local-snapshot-mount-point", "/snapshot", "NFS Local Snapshot Mount Point")
	rootCmd.PersistentFlags().StringVar(&maxStorageCapacity, "max-storage-capacity", "50G", "Volume Max Storage Capacity")

	rootCmd.SetVersionTemplate(fmt.Sprintf(versionTpl, name, Version, runtime.GOOS+"/"+runtime.GOARCH, BuildDate, CommitID))
}

func initLog() {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
