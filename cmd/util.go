package cmd

import (
	"fmt"
	"github.com/etclabscore/dp2p/discover"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"log"
	"math/big"
	"net"
	"os"
)

var (
	connectTimeout int
	respTimeout int
	listenAddr string
	statusProto bool
)

type statusData struct {
	ProtocolVersion uint32
	NetworkId uint64
	TD *big.Int
	CurrentBlock common.Hash
	GenesisBlock common.Hash
}

type errCode int

const (
	ErrMsgTooLarge = iota
	ErrDecode
	ErrInvalidMsgCode
	ErrProtocolVersionMismatch
	ErrNetworkIdMismatch
	ErrGenesisBlockMismatch
	ErrNoStatusMsg
	ErrExtraStatusMsg
	ErrSuspendedPeer
)

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

func readStatus(ws p2p.MsgReadWriter, status *statusData) error {
	msg, err := ws.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != eth.StatusMsg {
		return errResp(eth.ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, eth.StatusMsg)
	}
	if msg.Size > eth.ProtocolMaxMsgSize {
		return errResp(eth.ErrMsgTooLarge, "%v > %v", msg.Size, eth.ProtocolMaxMsgSize)
	}
	// Decode the handshake
	if err := msg.Decode(&status); err != nil {
		return errResp(eth.ErrDecode, "msg %v: %v", msg, err)
	}
	return nil
}

func mustEnodeArg(args []string) *enode.Node {
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
	return en
}

func mustUdp() *discover.Udp {
	nodeKey, _ := crypto.GenerateKey()

	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		utils.Fatalf("-ResolveUDPAddr: %v", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		utils.Fatalf("-ListenUDP: %v", err)
	}

	db, _ := enode.OpenDB("")
	ln := enode.NewLocalNode(db, nodeKey)
	cfg := discover.Config{
		PrivateKey:  nodeKey,
		//NetRestrict: restrictList,
	}
	_, u, err := discover.ListenUDP(conn, ln, cfg)
	if err != nil {
		utils.Fatalf("%v", err)
	}
	return u
}