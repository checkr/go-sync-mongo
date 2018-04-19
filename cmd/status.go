package cmd

import (
	"fmt"
	"os"
	"strconv"

	db "github.com/checkr/go-sync-mongo/db"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type LastRecord struct {
	ID bson.ObjectId `bson:"_id"`
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Shows all databases and counts of all the records accross collections",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		srcConfig := db.Config{
			URI: viper.GetString("src"),
			SSL: viper.GetBool("src-ssl"),
			Creds: mgo.Credential{
				Username: viper.GetString("src-username"),
				Password: viper.GetString("src-password"),
			},
		}
		src, err := db.NewConnection(srcConfig)
		if err != nil {
			fmt.Printf("Error: new src connection - %s\n", err)
			os.Exit(1)
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
			fmt.Printf("Error: new dst connection - %s\n", err)
			os.Exit(1)
		}

		data := [][]string{}

		dbnames, err := src.Databases()
		if err != nil {
			fmt.Printf("Error: get src databases - %s\n", err)
			os.Exit(1)
		}

		for _, dbname := range dbnames {
			var (
				total    int
				srcTotal int
				dstTotal int
			)

			collnames, err := src.Session.DB(dbname).CollectionNames()
			if err != nil {
				fmt.Printf("Error: get collections of %s - %s\n", dbname, err)
				os.Exit(1)
			}
			for _, collname := range collnames {
				dstColl := dst.Session.DB(dbname).C(collname)
				var dstLastRecord LastRecord
				_ = dstColl.Find(nil).Sort("-$natural").Limit(1).One(&dstLastRecord)

				dstQuery := dstColl.Find(bson.M{"_id": bson.M{"$lt": dstLastRecord.ID}})
				total, _ = dstQuery.Count()
				dstTotal += total

				srcColl := src.Session.DB(dbname).C(collname)
				srcQuery := srcColl.Find(bson.M{"_id": bson.M{"$lt": dstLastRecord.ID}})
				total, _ = srcQuery.Count()
				srcTotal += total
			}
			row := []string{dbname, strconv.Itoa(srcTotal), strconv.Itoa(dstTotal), strconv.Itoa(srcTotal - dstTotal)}
			data = append(data, row)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"DB", "Source", "Destination", "Diff"})

		for _, v := range data {
			table.Append(v)
		}

		table.Render()
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
