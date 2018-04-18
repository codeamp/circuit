package cmd

import (
	// "github.com/jinzhu/gorm"
	"fmt"

	"github.com/spf13/cobra"
	// "github.com/checkr/codeflow/server/plugins/codeflow"
	// codeamp "github.com/codeamp/circuit/plugins/codeamp"
	// codeamp_resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
	codeamp_plugins "github.com/codeamp/circuit/plugins"
	// "github.com/jinzhu/gorm/dialects/postgres"
	// "github.com/go-bongo/bongo"
	"github.com/codeamp/transistor"
	// mgo "gopkg.in/mgo.v2"
	"github.com/spf13/viper"	
	"github.com/davecgh/go-spew/spew"
	// uuid "github.com/satori/go.uuid" 
)

// migrateCmd represents the migrate command
var loadBalancersCmd = &cobra.Command{
	Use:   "migrate-loadbalancers",
	Short: "Send Load Balancer events to Kubernetes plugin",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[+] Creating internal load balancers.\n ---------------------------------------- \n")

		// init DB connection for codeamp
		fmt.Println("[*] Initializing CodeAmp DB")
		codeampDB, err := createCodeampDB()
		if err != nil {
			fmt.Println(codeampDB.Error)			
			panic(err.Error())
		}
		fmt.Println("[+] Successfully initialized Codeamp DB")

		fmt.Println("[*] Initializing CodeAmp transistor...")		
		spew.Dump(viper.GetStringMap("codeamp.redis"))
		config := transistor.Config{
			Server:   viper.GetString("codeamp.redis.server"),
			Database: viper.GetString("codeamp.redis.database"),
			Pool:     viper.GetString("codeamp.redis.pool"),
			Process:  viper.GetString("codeamp.redis.process"),
			Queueing: true,
			Plugins:        viper.GetStringMap("codeamp.plugins"),
			EnabledPlugins: []string{"dockerbuilder", "websocket"},
		}
	

		t, err := transistor.NewTransistor(config)
		if err != nil {
			panic(err.Error())
		}

		fmt.Println("[+] Successfully initialized CodeAmp transistor")		
		t.Events <- transistor.NewEvent(codeamp_plugins.WebsocketMsg{
			Event: fmt.Sprintf("test"),
		}, nil)				

		fmt.Println("[+] Finished sending events to all ELBs!")
	},
}

func init() {
	RootCmd.AddCommand(loadBalancersCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
