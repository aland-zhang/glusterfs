package main

import (
	"flag"
	"log"
	_ "net/http/pprof"
	"os"

	gfidCmd "github.com/appscode/glusterfs/cmd/gfider/cmd"
	"github.com/appscode/go/version"
	logs "github.com/appscode/log/golog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	var rootCmd = &cobra.Command{
		Use: "gfider",
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	rootCmd.AddCommand(version.NewCmdVersion())
	rootCmd.AddCommand(gfidCmd.NewCmdExtract())
	rootCmd.AddCommand(gfidCmd.NewCmdGet())
	rootCmd.AddCommand(gfidCmd.NewCmdDescribe())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
