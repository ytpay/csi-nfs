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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kubernetes-csi/csi-driver-nfs/pkg/nfs"
)

var (
	endpoint           string
	nodeID             string
	maxStorageCapacity string
	sharePoint         string
)

func init() {
	_ = flag.Set("logtostderr", "true")
}

func main() {

	_ = flag.CommandLine.Parse([]string{})

	cmd := &cobra.Command{
		Use:   "NFS",
		Short: "CSI based NFS driver",
		Run: func(cmd *cobra.Command, args []string) {
			handle()
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(&nodeID, "nodeid", "", "node id")
	_ = cmd.MarkPersistentFlagRequired("nodeid")

	cmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "csi endpoint")
	_ = cmd.MarkPersistentFlagRequired("endpoint")

	cmd.PersistentFlags().StringVar(&maxStorageCapacity, "maxstoragecapacity", "50G", "nfs maxStorageCapacity")
	cmd.PersistentFlags().StringVar(&sharePoint, "sharepoint", "/nfs", "nfs mount point")

	_ = cmd.ParseFlags(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

}

func handle() {
	nfs.NewNFSdriver(nodeID, endpoint, maxStorageCapacity, sharePoint).Run()
}
