// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"log"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/etclabscore/devp2p-disco/lib"
	elog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		log.SetPrefix("")

		lg := elog.NewGlogHandler(elog.StreamHandler(os.Stderr, elog.TerminalFormat(false)))
		lg.Verbosity(elog.Lvl(100))
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

		c := lib.ServerConfigEth63_Default()
		c.Protocols[0].Run = func(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
			log.Println(peer.String())
			log.Println(spew.Sdump(peer.Info()))
			peer.Disconnect(p2p.DiscQuitting)
			resCh <- 0
			// time.Sleep(200 * time.Millisecond)
			// peer.Disconnect(p2p.DiscQuitting)
			return nil
		}

		serv := lib.NewServer(c)
		pSub := serv.SubscribeEvents(pEventCh)
		if err := serv.Start(); err != nil {
			log.Println("failed to start p2p server", err)
			os.Exit(1)
		}
		go func() {
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
				log.Println("got res", c)
				quitCh <- true
				os.Exit(c)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
