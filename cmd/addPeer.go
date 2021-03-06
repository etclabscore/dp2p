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
	"fmt"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
	"log"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	elog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/spf13/cobra"
)

var spInProg bool

// addPeerCmd represents the addPeer command
var addPeerCmd = &cobra.Command{
	Aliases: []string{"addPeer"},
	Use:     "addpeer <enode>",
	Short:   "Add an ethereum enode as a peer",
	Long: `
    Spins up a memory-backed p2p server and attempts to make a very basic connection with an enode.
`,
	Run: func(cmd *cobra.Command, args []string) {

		en := mustEnodeArg(args)

		pEventCh := make(chan *p2p.PeerEvent)
		resCh := make(chan int)
		quitCh := make(chan bool)

		nodekey, _ := crypto.GenerateKey()
		c := p2p.Config{
			PrivateKey:      nodekey,
			MaxPeers:        25,
			MaxPendingPeers: 50,
			NoDiscovery:     true,
			Name:            "dp2p",
			Protocols: []p2p.Protocol{p2p.Protocol{
				Name:    eth.ProtocolName,
				Version: eth.ProtocolVersions[0], // just eth/63 for now, no immediate need for backwards compat
				Length:  eth.ProtocolLengths[0],
				Run: func(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
					log.Println(peer.String())
					log.Println(spew.Sdump(peer.Info()))

					if statusProto {

						log.Println("attempting status proto exchange")

						// Send out own handshake in a new thread
						errc := make(chan error, 2)
						spInProg = true
						defer func() { spInProg = false }()
						var status statusData
						go func() {
							errc <- p2p.Send(ws, eth.StatusMsg, &statusData{
								ProtocolVersion: uint32(eth.ProtocolVersions[0]),
								NetworkId:       1,
								TD:              core.DefaultGenesisBlock().Difficulty,
								CurrentBlock:    params.MainnetGenesisHash,
								GenesisBlock:    params.MainnetGenesisHash,
							})
						}()
						go func() {
							errc <- readStatus(ws, &status)
						}()
						timeout := time.NewTimer(time.Second * 2)
						defer timeout.Stop()
						for i := 0; i < 2; i++ {
							select {
							case err := <-errc:
								if err != nil {
									return err
								}
							case <-timeout.C:
								return p2p.DiscReadTimeout
							}
						}
						log.Println(spew.Sdump(status))

					} else {
						log.Println("status proto exchange not enabled")
					}
					peer.Disconnect(p2p.DiscQuitting)
					resCh <- 0
					return nil
				},
			}},
			ListenAddr:      listenAddr,
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
			t := time.NewTicker(time.Duration(int32(connectTimeout)) * time.Second)
			defer serv.Stop()
			for {
				select {
				case ev := <-pEventCh:
					log.Println(ev)
					if ev.Type == p2p.PeerEventTypeAdd {

					}
					if ev.Type == p2p.PeerEventTypeDrop {
						log.Println("peer dropped")
						resCh <- 1
						return
					}
				case err := <-pSub.Err():
					log.Println("peer event sub error", err)
					resCh <- 1
					return
				case <-t.C:
					log.Println("ticker expired", time.Now().Format(time.RFC3339))
					resCh <- 1
					return
				case <-quitCh:
					return
				}
			}
		}()
		go serv.AddPeer(en)
		for {
			select {
			case c := <-resCh:
				// log.Println("got res code", c)
				if c == 0 {
					fmt.Println("OK")
				} else {
					fmt.Println("FAIL")
				}
				//quitCh <- true
				for(spInProg) {}
				os.Exit(c)
			}
		}
	},
}

func init() {
	addPeerCmd.PersistentFlags().IntVarP(&connectTimeout, "timeout", "t", 30, "time in seconds to wait for node to dial a connection")
	addPeerCmd.PersistentFlags().StringVarP(&listenAddr, "listenaddr", "a", ":30301", "address:port to listen at")
	addPeerCmd.PersistentFlags().BoolVarP(&statusProto, "statusproto", "s", true,"if adding peer succeeds, attempt to exchange status messages")

	rootCmd.AddCommand(addPeerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addPeerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addPeerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
