package cmd

import (
	"github.com/appscode/glusterfs/pkg/gfid/lib"
	"github.com/appscode/log"
	"github.com/spf13/cobra"
)

func NewCmdGet() *cobra.Command {
	gfidDir, err := GetGFIDDir()
	if err != nil {
		log.Fatalln(err)
	}

	var vol, brick string
	var gfids []string
	cmd := &cobra.Command{
		Use: "get",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if brick != "" {
				gfids, err = lib.GetGFIDs(vol, brick)
				if err != nil {
					log.Fatalln(err)
				}
			}
			err = lib.Get(gfidDir, gfids)
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	cmd.Flags().StringVar(&gfidDir, "gfid-dir", gfidDir, "Directory where LevelDB file is stored.")
	cmd.Flags().StringSliceVar(&gfids, "gfids", []string{}, "Comma separated list of GFIDs.")
	cmd.Flags().StringVar(&vol, "volume", "", "Volume name used to detect split-brain gfids eg, vol")
	cmd.Flags().StringVar(&brick, "brick", "", "Brick name used to detect split-brain gfids eg, ip:/dir")
	return cmd
}
