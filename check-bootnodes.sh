#!/usr/bin/env bash

# Checks bootnode(enode) reachability for networks supported by a geth client, eg. multi-geth.
#
# * Requires tool 'json' to be installed. Try 'npm install -g json-cli'.
#
# Use:
#  cd $GOPATH/src/github.com/etclabscore/devp2ping &&
#  ./check-bootnodes.sh [|network names...]
#
# If no network names are passed, then all available are used.
# - See 'data_dir' var below if you want to change where resulting data goes.
# - See 'timeout_lim' var if you want to change how long (in seconds) before ping attempt resigns.
# - See 'base_addr' for floor udp port for ping/discovery listen addresses. Currently, only local ports are supported.

timeout_lim=$((60*60)) # m * n_seconds
base_addr=30300
data_dir="data/$(date +%s)-${timeout_lim}s"
all_networks=(mainnet testnet classic social ethersocial mix rinkeby kotti goerli)


set -e


# I think this is npm's json-cli
# I'd prefer to use jj, but couldn't get it to do array->multi line quite as nicely.
command -v json || { echo "Requires tool 'json' to be installed. Maybe try 'npm install -g json-cli'" && exit 1; }


# Build executables geth and devp2ping, we'll remove them after use.
curd="$(pwd)"
pushd $GOPATH/src/github.com/ethereum/go-ethereum
make all
cp build/bin/geth $curd/geth
popd
go build -o devp2ping

trap "rm geth devp2ping" EXIT # Remove executable bins on exit.

mkdir -p "$data_dir"

networks=()
for a in "$@"; do
    networks+=("$a")
done
# Set default network list in case none are argued.
if [[ "${#networks[@]}" -eq 0 ]]; then networks=("${all_networks[@]}"); fi


# Collect lists of bootnodes from geth's 'dumpconfig' command.
for net in "${networks[@]}"
do
    [[ -z "$net" ]] && echo 'empty net, continuing' && continue

    echo "Found chain: $net"
    mkdir -p "$data_dir/$net"

    nf="--$net"
    [[ "$net" == mainnet ]] && nf=

    ./geth 2>/dev/null $nf dumpconfig |
        egrep "^BootstrapNodes " |
        cut -d'=' -f2 |
        sed 's/^ //g' |
        json -a |
        tee "$data_dir/$net/bootnodes.list"

    ./geth version > "$data_dir/$net/geth.version" 2>&1

    [[ "$(wc -l < $data_dir/$net/bootnodes.list)" -eq 0 ]] && echo "$net: no bootnodes" && exit 1
done

# Run ping attempts for each enode for each network.
# Note that subshells are spawned for each net, then each enode.
# Subshells are not disowned, and we wait for them all to exit.
for net in "${networks[@]}"; do
    echo "Running chain: $net"
    base_addr=$((base_addr+20))
    (
        while read enode; do
            base_addr=$((base_addr+1))
            enode_id="$(echo $enode | cut -d'/' -f2- | cut -d '@' -f1)"
            (
                set +e # Must allow devp2ping to 'fail', ie exit w/ 1
                echo "Running $net $enode"
                start="$(date +%s)"
                ./devp2ping -a ":$base_addr" -t $timeout_lim "$enode" > "$data_dir/$net$enode_id.log" 2>&1
                res="$?"
                end="$(date +%s)"
                echo "$res $net-$enode "'('$((end-start))'s)' >> "$data_dir/$net/outcomes"
            )&
        done < "$data_dir/$net/bootnodes.list"
        wait
    )&
done
wait

