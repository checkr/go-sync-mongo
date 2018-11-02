package cmd

import (
	"log"

	db "github.com/checkr/go-sync-mongo/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Tails the source oplog and syncs to destination",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		srcConfig := db.Config{
			URI:              viper.GetString("src"),
			SSL:              viper.GetBool("src-ssl"),
			IgnoreApplyError: viper.GetBool("ignore-apply-error"),
			Creds: mgo.Credential{
				Username: viper.GetString("src-username"),
				Password: viper.GetString("src-password"),
			},
		}
		src, err := db.NewConnection(srcConfig)
		if err != nil {
			log.Fatalf("Error: new src connection - %s\n", err)
		}

		dstConfig := db.Config{
			URI: viper.GetString("dst"),
			SSL: viper.GetBool("dst-ssl"),
			Creds: mgo.Credential{
				Username: viper.GetString("dst-username"),
				Password: viper.GetString("dst-password"),
			},
		}
		dst, err := db.NewConnection(dstConfig)
		if err != nil {
			log.Fatalf("Error: new dst connection - %s\n", err)
		}

		err = src.SyncOplog(dst)
		if err != nil {
			log.Fatalf("Error: sync oplog - %s\n", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)
	syncCmd.Flags().Int32("since", 0, "seconds since the Unix epoch")
	syncCmd.Flags().Int32("ordinal", 0, "incrementing ordinal for operations within a given second")
	syncCmd.Flags().Bool("ignore-apply-error", false, "ingore error of applying oplog (true)")
	viper.BindPFlag("since", syncCmd.Flags().Lookup("since"))
	viper.BindPFlag("ordinal", syncCmd.Flags().Lookup("ordinal"))
	viper.BindPFlag("ignore-apply-error", syncCmd.Flags().Lookup("ignore-apply-error"))
}
