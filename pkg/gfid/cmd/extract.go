package cmd

import (
	"os"

	"github.com/appscode/glusterfs/pkg/gfid/lib"
	"github.com/appscode/go/flags"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
)

func NewCmdExtract() *cobra.Command {
	var dataDir, storageDir string
	var computeChecksum bool
	cmd := &cobra.Command{
		Use: "extract",
		Run: func(cmd *cobra.Command, args []string) {
			flags.EnsureRequiredFlags(cmd, "data-dir", "storage-dir")
			err := lib.Extract(dataDir, storageDir, computeChecksum)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Data to encrypt")
	cmd.Flags().StringVar(&storageDir, "storage-dir", os.TempDir(), "Directory where LevelDB file is stored.")
	cmd.Flags().BoolVar(&computeChecksum, "compute-checksum", false, "Compute md5 checksum for files.")
	return cmd
}
