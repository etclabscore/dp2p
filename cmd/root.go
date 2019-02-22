// Copyright © 2019 Isaac Ardis isaac.ardis@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	elog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "devp2ping",
	Short: "Ping an ethereum enode",
	Long: `
    Spins up a memory-backed p2p server and attempts to make a very basic connection with an enode.
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		log.SetPrefix("")

		lg := elog.NewGlogHandler(elog.StreamHandler(os.Stderr, elog.TerminalFormat(false)))
		lg.Verbosity(elog.Lvl(11)) // turn it up to... #loud
		elog.Root().SetHandler(lg)

		if len(args) == 0 {
			log.Println("need enode as first argument")
			os.Exit(1)
		}
		eni := args[0]
		en, err := enode.ParseV4(eni)
		if err != nil {
			log.Println("malformed enode", eni)
			os.Exit(1)
		}

		pEventCh := make(chan *p2p.PeerEvent)
		resCh := make(chan int)
		quitCh := make(chan bool)

		nodekey, _ := crypto.GenerateKey()
		c := p2p.Config{
			PrivateKey:      nodekey,
			MaxPeers:        25,
			MaxPendingPeers: 50,
			NoDiscovery:     true,
			Name:            "devp2ping",
			Protocols: []p2p.Protocol{p2p.Protocol{
				Name:    eth.ProtocolName,
				Version: eth.ProtocolVersions[0], // just eth/63 for now, no immediate need for backwards compat
				Length:  eth.ProtocolLengths[0],
				Run: func(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
					log.Println(peer.String())
					log.Println(spew.Sdump(peer.Info()))
					peer.Disconnect(p2p.DiscQuitting)
					resCh <- 0
					return nil
				},
			}},
			ListenAddr:      ":30301",
			Logger:          elog.Root(),
			NodeDatabase:    "", // empty for memory
			EnableMsgEvents: true,
		}
		serv := p2p.Server{Config: c}
		pSub := serv.SubscribeEvents(pEventCh)
		if err := serv.Start(); err != nil {
			log.Println("failed to start p2p server", err)
			os.Exit(1)
		}
		go func() {
			// We can't listen for dial failures w/o too much hacking.
			// So we just set a timeout to wait for a positive connection.
			t := time.NewTicker(30 * time.Second)
			defer serv.Stop()
			for {
				select {
				case ev := <-pEventCh:
					if ev.Type == p2p.PeerEventTypeAdd {
						log.Println(ev)
					}
				case err := <-pSub.Err():
					log.Println("peer event sub error", err)
					resCh <- 1
				case <-t.C:
					log.Println("ticker expired")
					resCh <- 1
				case <-quitCh:
					return
				}
			}
		}()
		go serv.AddPeer(en)
		for {
			select {
			case c := <-resCh:
				log.Println("got res code", c)
				if c == 0 {
					log.Println("OK")
				} else {
					log.Println("FAIL")
				}
				quitCh <- true
				os.Exit(c)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.devp2ping")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".devp2ping" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".devp2ping")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
