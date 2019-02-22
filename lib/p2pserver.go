package lib

import (
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type DiscoErr error
type DiscoP2PService struct {
	Server      p2p.Server
	peerEventCh chan *p2p.PeerEvent
	ServerSub   event.Subscription
}

var (
	errNoStart = errors.New("failed to start server")
)

type ArgsStart struct{}
type RespStart struct {
	Enode    *enode.Node
	NodeInfo *p2p.NodeInfo
}

func (ds *DiscoP2PService) Start(args *ArgsStart, res *RespStart) error {
	if ds.Server.MaxPeers == 0 {
		ds.Server = NewServer(p2p.Config{})
	}

	ds.peerEventCh = make(chan *p2p.PeerEvent)
	ds.ServerSub = ds.Server.SubscribeEvents(ds.peerEventCh)

	err := ds.Server.Start()
	if err != nil {
		return err
	}
	res = &RespStart{
		Enode:    ds.Server.Self(),
		NodeInfo: ds.Server.NodeInfo(),
	}
	if res.Enode == nil {
		return errNoStart
	}
	return nil
}

type ArgsStop struct{}
type RespStop struct{}

func (ds *DiscoP2PService) Stop(args *ArgsStop, res *RespStop) error {
	// TODO Return error if not running
	ds.Server.Stop()
	return nil
}

type ArgsAddPeer struct {
	Enode string `json:"enode"`
}
type RespAddPeer struct {
	Ok    bool   `json:"ok"`
	Enode string `json:"enode"`
}

func (ds *DiscoP2PService) AddPeer(args *ArgsAddPeer, res *RespAddPeer) error {
	en, err := enode.ParseV4(args.Enode)
	if err != nil {
		return err
	}
	log.Info("disco addPeer", "enode", en)
	// origPeerLen := len(ds.Server.Peers())
	ds.Server.AddPeer(en)
	res = &RespAddPeer{Ok: true, Enode: en.String()}

	t := time.NewTicker(15 * time.Second) // TODO flag me maybe?
	for {
		select {
		case ev := <-ds.peerEventCh:
			log.Info("ds server peer event", "event", ev)
			switch ev.Type {
			case p2p.PeerEventTypeAdd:
				return nil
			}
			return fmt.Errorf("unwanted peer event: %v", ev)
		case err := <-ds.ServerSub.Err():
			res = &RespAddPeer{Ok: false, Enode: en.String()}
			log.Crit("ds server error", "error", err)
			return err
		case <-t.C:
			// could not connect to peer (timeout)
			log.Warn("ticker ticked")
			res = &RespAddPeer{Ok: false, Enode: en.String()}
			return errors.New("failed to connect after 15 seconds")
		}
	}
	// newPeerLen := len(ds.Server.Peers())
	// log.Info("disco", "olen", origPeerLen, "nlen", newPeerLen, "server.peers", ds.Server.Peers())
	// if origPeerLen+1 != newPeerLen {
	// 	res = &RespAddPeer{Ok: false, Enode: en.String()}
	// 	return errors.New("failed to add peer")
	// }

	// peerEventCh := make(chan<- p2p.MeteredPeerEvent)
	// sub := p2p.SubscribeMeteredPeerEvent(peerEventCh)
	// go func() {
	// 	for {
	// 		select {
	// 		case ev := <-peerEventCh:
	// 		case err := <-sub.Err():
	// 			log.Error("peer sub", "error", err)
	// 			return
	// 		}
	// 	}
	// }()
	// sub.Err()

	return nil
}

type ArgsRemovePeer struct {
	Enode string `json:"enode"`
}
type RespRemovePeer struct{}

func (ds *DiscoP2PService) RemovePeer(args *ArgsRemovePeer, res *RespRemovePeer) error {
	en, err := enode.ParseV4(args.Enode)
	if err != nil {
		return err
	}
	log.Info("disco removePeer", "enode", en)
	ds.Server.RemovePeer(en)
	return nil
}

func ProtocolEth63Disco() p2p.Protocol {
	return p2p.Protocol{
		Name:    eth.ProtocolName,
		Version: eth.ProtocolVersions[0], // just eth/63 for now, no immediate need for backwards compat
		Length:  eth.ProtocolLengths[0],
		Run:     ppfnHandler,
	}
}

func ppfnHandler(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
	log.Info("disco handler", "peer", peer)
	time.Sleep(5 * time.Second)
	return nil
}

func ServerConfigEth63_Default() p2p.Config {
	nodekey, _ := crypto.GenerateKey()
	return p2p.Config{
		PrivateKey:      nodekey,
		MaxPeers:        25,
		MaxPendingPeers: 50,
		NoDiscovery:     true,
		Name:            "disco",
		Protocols:       []p2p.Protocol{ProtocolEth63Disco()},
		ListenAddr:      ":30301",
		Logger:          log.Root(),
		NodeDatabase:    "", // empty for memory
		EnableMsgEvents: true,
	}
}

func NewServer(config p2p.Config) p2p.Server {
	// use zero-value MaxPeers field as proxy for empty (unset) config
	if config.MaxPeers == 0 {
		config = ServerConfigEth63_Default()
	}
	return p2p.Server{
		Config: config,
	}
}
