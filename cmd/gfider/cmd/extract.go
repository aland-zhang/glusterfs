package cmd

import (
	"github.com/appscode/glusterfs/pkg/gfid"
	"github.com/appscode/go/flags"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
)

func NewCmdExtract() *cobra.Command {
	gfidDir, err := GetGFIDDir()
	if err != nil {
		log.Fatalln(err)
	}

	var dataDir string
	var computeChecksum bool
	cmd := &cobra.Command{
		Use: "extract",
		Run: func(cmd *cobra.Command, args []string) {
			flags.EnsureRequiredFlags(cmd, "data-dir")
			err := gfid.Extract(dataDir, gfidDir, computeChecksum)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	cmd.Flags().StringVar(&dataDir, "data-dir", "", "Data to encrypt")
	cmd.Flags().StringVar(&gfidDir, "gfid-dir", gfidDir, "Directory where LevelDB file is stored.")
	cmd.Flags().BoolVar(&computeChecksum, "compute-checksum", false, "Compute md5 checksum for files.")
	return cmd
}
