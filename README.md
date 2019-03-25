## Install

```shell
$ go get github.com/etclabscore/dp2p/...
```

or

```shell
$ go get github.com/etclabscore/dp2p/...
$ cd $GOPATH/src/github.com/etclabscore/dp2p
$ go install
```

## Help

```
$ dp2p

  Tools for simple interaction and queries on the devp2p protocol.

Usage:
  dp2p [command]

Available Commands:
  addpeer     Add an ethereum enode as a peer
  findnode    Send a devp2p FINDNODE request to an enode (with preliminary PING/PONG)
  help        Help about any command
  ping        Send a PING request to a given enode

Flags:
  -h, --help   help for dp2p

Use "dp2p [command] --help" for more information about a command.
```

## Use

Returns exit code `0` if successful, `1` if failed.  

Will print all logs available from the go-ethereum `p2p` and `discover` libraries in use. As with the go-ethereum client, these go to stderr.
Relevant program output (eg. neighbors) will go to stdout.

#### addpeer
```
$ dp2p addpeer -a ':30301' -t $((60*60)) 'enode://66498ac935f3f54d873de4719bf2d6d61e0c74dd173b547531325bcef331480f9bedece91099810971c8567eeb1ae9f6954b013c47c6dc51355bbbbae65a8c16@54.148.165.1:30303'
```

#### ping

```shell
$ dp2p ping 'enode://66498ac935f3f54d873de4719bf2d6d61e0c74dd173b547531325bcef331480f9bedece91099810971c8567eeb1ae9f6954b013c47c6dc51355bbbbae65a8c16@54.148.165.1:30303'
```

#### findnode

```shell
$ dp2p findnode 'enode://66498ac935f3f54d873de4719bf2d6d61e0c74dd173b547531325bcef331480f9bedece91099810971c8567eeb1ae9f6954b013c47c6dc51355bbbbae65a8c16@54.148.165.1:30303'
```

### Check default go-ethereum/multi-geth bootnodes

If you have a `go-ethereum` source (eg. [ethoxy/multi-geth](https://github.com/ethoxy/multi-geth) or [ethereum/go-ethereum](https://github.com/ethereum/go-ethereum)) available in your $GOPATH, you can run checks for default bootnodes with

```
$ ./examples/check-bootnodes.sh [|<chain> ...]
```

