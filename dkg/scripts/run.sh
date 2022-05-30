#!/bin/bash

start=3
end=3
stepSize=3
containerIndex=1

cd "$(dirname $0)"/../.. || exit 1
root="$(pwd)"

main() {
    parse_input "$1"

    buildRoot="$root"/build
    cidFile="$buildRoot"/cid
    cadvisorVer=v0.44.0

    docker run \
    --volume=/:/rootfs:ro \
    --volume=/var/run:/var/run:rw \
    --volume=/sys:/sys:ro \
    --volume=/var/lib/docker/:/var/lib/docker:ro \
    --publish=8080:8080 \
    --cidfile="$cidFile" \
    gcr.io/cadvisor/cadvisor:$cadvisorVer \
    --docker_only=true \
    --disable_root_cgroup_stats=true \
    --storage_driver=stdout \
    --allow_dynamic_housekeeping=false \
    &> "$buildRoot"/cadvisor.log &

    until [[ -f "$cidFile" ]]; do
        sleep 1
    done

    cadvisorId=$(<"$cidFile")
    rm "$cidFile"

    cd ./contracts/

    for ((participants = start; participants <= end; participants++)) do
        echo "Starting to measure runtime for $participants participants"

        ../zk/scripts/build.sh $participants

        buildDir="$buildRoot"/$participants
        config="$buildDir"/hardhat.config.js
        log="$buildDir"/hardhat.log
        containerPipe="$buildDir"/container_pipe

        echo "$(cat ./hardhat.config.js)

module.exports.networks = module.exports.networks || {};
module.exports.networks.hardhat = module.exports.networks.hardhat || {};
module.exports.networks.hardhat.accounts = {count: $participants};" > "$config"

        NODE_PATH=./node_modules npx hardhat node --config "$config" > "$log" &
        nodePid=$!

        trap "kill $nodePid" EXIT

        # Retrieve the private keys for the accounts from the log of the Hardhat node
        ethPrivs=( $(tail -f "$log" | awk 'BEGIN{i=0; ORS=" "} match($0, /Private Key: 0x([[:alnum:]]+)/, res){print res[1]; if (++i == n) exit}' n=$participants) )

        npx hardhat --network localhost run ./scripts/deploy.js

        if [[ ! -p "$containerPipe" ]]; then
            rm -f "$containerPipe"
            mkfifo "$containerPipe"
        fi
        containerPipe=$(readlink -e "$containerPipe")

        goPids=()
        cd ../dkg/

        for ((i = 0; i < participants; i++)); do

            # Delay node starts so that gas estimation for the register transaction is accurate
            if (( i != 0 )); then
                sleep 2
            fi

            # Use a constant hex string and add the participant index
            dkgPriv=$(echo "obase=16;ibase=16;147F0309B0587059C68AE43949192C6DC2222210D5105777A512DCDD373CE1AA + $(echo "obase=16;$i" | bc)" | bc)

            config="{
                \"EthereumNode\":       \"ws://127.0.0.1:8545\",
                \"EthereumPrivateKey\": \"${ethPrivs[$i]}\",
                \"DkgPrivateKey\":      \"$dkgPriv\",
                \"ContractAddress\":    \"0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9\",
                \"MountSource\":        \"$(readlink -e ../build/$participants/)\"
            }"

            flags=()
            if (( i == 0 )); then # The 0th node emits invalid commitments
                flags+=("--rogue")
            fi

            if (( i == 1 )); then # The 1st node is the only node that should compute the proof
                flags+=("--id-pipe=$containerPipe")
            else
                flags+=("--ignore-invalid")
            fi

            echo "$config" | go run ./cmd/ -c /dev/stdin ${flags[@]} |& tee "$buildDir"/node_$i.log &
            goPids[$i]=$!
        done

        while read dockerId; do
            collect_container_stats "$buildDir" $dockerId
        done < "$containerPipe"

        cd ../contracts/

        for pid in ${goPids[@]}; do
            wait $pid
        done

        kill $nodePid
        trap - EXIT
    done
}

parse_input() {
    local singleRegex="^[0-9]+$"
    local rangeRegex="^\[([[:digit:]]+),([[:digit:]]+)(,([[:digit:]]+))?\]$"

    if [[ "$1" =~ $singleRegex ]]; then
        start=$1
        end=$1
    elif [[ "$1" =~ $rangeRegex ]]; then
        start=${BASH_REMATCH[1]}
        end=${BASH_REMATCH[1]}
        if [[ -n ${BASH_REMATCH[4]} ]]; then
            stepSize=${BASH_REMATCH[4]}
        fi
    else
        echo "Wrong input" >&2 && exit 1
    fi
}

collect_container_stats() {
    local csv="$1"/$((containerIndex++)).csv
    echo "time,memory_usage" > "$csv"
    perl -ne "print if s/^cName=$2(?=.*timestamp=([[:digit:]]+))(?=.*memory_usage=([[:digit:]]+)).*/\1,\2/" "$buildRoot"/cadvisor.log >> "$csv"
}

main "$@"
