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

	nfsServer          string
	nfsSharePoint      string
	nfsLocalMountPoint string
	nfsSnapshotPath    string
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
			nfsSnapshotPath,
		).Run()
	},
}

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug log")

	rootCmd.PersistentFlags().StringVar(&nodeID, "nodeid", "", "csi node id")
	_ = rootCmd.MarkPersistentFlagRequired("nodeid")

	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "csi endpoint")
	_ = rootCmd.MarkPersistentFlagRequired("endpoint")

	rootCmd.PersistentFlags().StringVar(&name, "name", "csi-nfs", "csi name")
	_ = rootCmd.PersistentFlags().MarkHidden("name")
	_ = rootCmd.MarkPersistentFlagRequired("name")

	rootCmd.PersistentFlags().StringVar(&maxStorageCapacity, "maxstoragecapacity", "50G", "nfs maxStorageCapacity")
	rootCmd.PersistentFlags().StringVar(&nfsServer, "nfsserver", "", "nfs server")
	rootCmd.PersistentFlags().StringVar(&nfsSharePoint, "nfssharepoint", "/", "nfs server share point")
	rootCmd.PersistentFlags().StringVar(&nfsLocalMountPoint, "nfslocalmountpoint", "/nfs", "nfs local mount point")
	rootCmd.PersistentFlags().StringVar(&nfsSnapshotPath, "nfssnapshotpath", "/snapshot", "nfs snapshot path")

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
