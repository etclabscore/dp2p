package cmd

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/etclabscore/dp2p/discover"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

// getBlockHeadersData represents a block header query.
type getBlockHeadersData struct {
	Origin  uint64 // Block from which to retrieve headers
	Amount  uint64       // Maximum number of headers to retrieve
	Skip    uint64       // Blocks to skip between consecutive headers
	Reverse bool         // Query direction (false = rising towards latest, true = falling towards genesis)
}

// newBlockHashesData is the network packet for the block announcements.
type newBlockHashesData []struct {
	Hash   common.Hash // Hash of one particular block being announced
	Number uint64      // Number of one particular block being announced
}

// newBlockData is the network packet for the block propagation message.
type newBlockData struct {
	Block *types.Block
	TD    *big.Int
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

var errPar = errors.New("parity EOF")
func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

func readStatus(ws p2p.MsgReadWriter, status *statusData) error {
	msg, err := ws.ReadMsg()
	if err != nil {
		if msg.Size == 0 || msg.Payload == nil {
			log.Println("got nil message (parity?)")
			return errPar
		}

		b := []byte{}
		n, e := msg.Payload.Read(b)
		if err.Error() == "EOF" || n == 0 {
			log.Println("got EOF (parity?)")
			log.Println(string(b), b, e)
			log.Println(spew.Sdump(msg))
			return msg.Discard()
		}
		return err
	}
	if msg.Code != eth.StatusMsg {
		return errResp(ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, eth.StatusMsg)
	}
	if msg.Size > eth.ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, eth.ProtocolMaxMsgSize)
	}
	// Decode the handshake
	if err := msg.Decode(&status); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	log.Println(spew.Sdump(status))
	return nil
}

func readBlockHeadersMsg(ws p2p.MsgReadWriter, headers []*types.Header) (error) {
	msg, err := ws.ReadMsg()
	if err != nil {
		if msg.Size == 0 || msg.Payload == nil {
			log.Println("got nil message (parity?)")
			return errPar
		}
		b := []byte{}
		n, e := msg.Payload.Read(b)
		if err.Error() == "EOF" || msg.Size == 0 || n == 0 {
			log.Println("got EOF (parity?)")
			log.Println(string(b), b, e)
			log.Println(spew.Sdump(msg))
			return msg.Discard()
		}
		return err
	}
	if msg.Code != eth.BlockHeadersMsg {
		if msg.Code == eth.StatusMsg {
			var status *statusData
			return readStatus(ws, status)
		} else if msg.Code == eth.NewBlockHashesMsg {
			var announces newBlockHashesData
			if err := msg.Decode(&announces); err != nil {
				return errResp(ErrDecode, "%v: %v", msg, err)
			}
			log.Println(spew.Sdump(announces))
		} else if msg.Code == eth.TxMsg {
			// Transactions can be processed, parse all of them and deliver to the pool
			var txs []*types.Transaction
			if err := msg.Decode(&txs); err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			log.Println(spew.Sdump(txs))
			for _, t := range txs {
				bs, _ := t.MarshalJSON()
				log.Println("tx" ,"hash", t.Hash().Hex(), "chainid", t.ChainId(), "json", string(bs))
			}
		} else if msg.Code == eth.NewBlockMsg {
			// Retrieve and decode the propagated block
			var request newBlockData
			if err := msg.Decode(&request); err != nil {
				return errResp(ErrDecode, "%v: %v", msg, err)
			}
			log.Println(spew.Sdump(request))
		}
		log.Println("got unqueried msg, discarding", msg.String())

		return msg.Discard()
		//return errResp(ErrInvalidMsgCode, "want: %x, got: %x", eth.BlockHeadersMsg, msg.Code)
	}
	if err := msg.Decode(&headers); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	log.Println(spew.Sdump(headers))
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