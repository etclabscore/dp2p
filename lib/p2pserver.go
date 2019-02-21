package lib

import (
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type DiscoErr error
type DiscoP2PService struct {
	Server p2p.Server
}

var (
	errNoStart = errors.New("failed to start server")
)

type ArgsStart struct{}
type RespStart struct {
	Enode *enode.Node
}

func (ds *DiscoP2PService) Start(args *ArgsStart, res *RespStart) error {
	if ds.Server.MaxPeers == 0 {
		ds.Server = NewServer(p2p.Config{})
	}
	err := ds.Server.Start()
	if err != nil {
		return err
	}
	res = &RespStart{
		Enode: ds.Server.Self(),
	}
	if res.Enode != nil {
		return nil
	}
	return errNoStart
}

type ArgsStop struct{}
type RespStop struct{}

func (ds *DiscoP2PService) Stop(args *ArgsStop, res *RespStop) error {
	// TODO Return error if not running
	ds.Server.Stop()
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
