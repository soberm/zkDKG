#!/bin/bash

cd "$(dirname $0)"/../.. || exit 1

generateOnly=false
start=3
end=3
stepSize=3
containerIndex=0
containerCsvFiles=("poly_eval_witness" "poly_eval_proof" "key_deriv_witness" "key_deriv_proof")
root="$(pwd)"

declare cadvisorId
declare nodePid
declare -A goPids=()

trap cleanup EXIT

main() {
    parse_input "$@"

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

    until [[ -f $cidFile ]]; do
        sleep 1
    done

    cadvisorId=$(<"$cidFile")
    rm "$cidFile"

    cd ./contracts/

    for ((participants = start; participants <= end; participants++)) do
        echo "Starting to measure runtime for $participants participants"

        ../zk/scripts/build.sh $participants

        buildDir="$buildRoot"/$participants
        containerPipe="$buildDir"/container_pipe
        declare -a ethPrivs

        if ! $generateOnly; then
            log="$buildDir"/hardhat.log
            NODE_PATH=./node_modules npx hardhat launch $participants > "$log" &
            nodePid=$!

            # Retrieve the private keys for the accounts from the log of the Hardhat node
            ethPrivs=( $(tail -f "$log" | awk 'BEGIN{i=0; ORS=" "} match($0, /Private Key: 0x([[:alnum:]]+)/, res){print res[1]; if (++i == n) exit}' n=$participants) )

            npx hardhat --network localhost deploy $participants
        fi

        if [[ ! -p $containerPipe ]]; then
            rm -f "$containerPipe"
            mkfifo "$containerPipe"
        fi
        containerPipe=$(readlink -e "$containerPipe")

        local goPids=()
        cd ../dkg/

        if $generateOnly; then
            generate_config 1 $participants | go run ./cmd/generator -c /dev/stdin --participants $participants --id-pipe="$containerPipe" |& tee "$buildDir"/generator.log &
            goPids[0]=$!
        else
            mkdir -p "$buildDir"/nodes

            for ((i = 1; i <= participants; i++)); do
                flags=()
                if (( i == 1 )); then # The 1st node emits invalid commitments
                    flags+=("--rogue" )
                    flags+=("--id-pipe=$containerPipe")
                fi

                if (( i != 2 )); then # The 2nd node is the only node that should dispute the invalid broadcast
                    flags+=("--ignore-invalid")
                fi

                if (( i != 1 && i != 2 )); then
                    flags+=("--broadcast-only")
                fi

                generate_config $i $participants | go run ./cmd/full_node -c /dev/stdin ${flags[@]} |& tee "$buildDir"/nodes/node_$i.log &
                goPids[$i]=$!
            done
        fi

        while read dockerId; do
            collect_container_stats "$buildDir" $dockerId
        done < "$containerPipe"

        cd ../contracts/

        for pid in ${goPids[@]}; do
            wait $pid
        done

        if ! $generateOnly; then
            kill $nodePid
            unset nodePid
        fi
    done
}

parse_input() {
    local args=$(getopt -a -n run -o g --long generate-only: -- "$@")
    if [[ $? != 0 ]]; then
        usage
    fi
    eval set -- "$args"

    while :
    do
        case "$1" in
            -g | --generate-only)   generateOnly=true; shift;;
            --)                     shift; break ;;
        esac
    done

    local singleRegex="^[0-9]+$"
    local rangeRegex="^\[([[:digit:]]+),([[:digit:]]+)(,([[:digit:]]+))?\]$"

    if [[ $1 =~ $singleRegex ]]; then
        start=$1
        end=$1
    elif [[ $1 =~ $rangeRegex ]]; then
        start=${BASH_REMATCH[1]}
        end=${BASH_REMATCH[1]}
        if [[ -n ${BASH_REMATCH[4]} ]]; then
            stepSize=${BASH_REMATCH[4]}
        fi
    else
        usage
    fi
}

generate_config() {
    # Use a constant hex string and add the participant index
    dkgPriv=$(echo "obase=16;ibase=16;147F0309B0587059C68AE43949192C6DC2222210D5105777A512DCDD373CE1AA + $(echo "obase=16;$1" | bc)" | bc)

    echo "{
        \"EthereumNode\":       \"ws://127.0.0.1:8545\",
        \"EthereumPrivateKey\": \"${ethPrivs[$1 - 1]}\",
        \"DkgPrivateKey\":      \"$dkgPriv\",
        \"ContractAddress\":    \"0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0\",
        \"MountSource\":        \"$(readlink -e ../build/$2/zk)\"
    }"
}

collect_container_stats() {
    local csv="$1"/${containerCsvFiles[containerIndex++]}.csv
    echo "time,memory_usage" > "$csv"
    perl -ne "print if s/^cName=$2(?=.*timestamp=([[:digit:]]+))(?=.*memory_usage=([[:digit:]]+)).*/\1,\2/" "$buildRoot"/cadvisor.log >> "$csv"
}

cleanup() {
    if [[ -n $nodePid ]]; then
        kill $nodePid
    fi

    if [[ -n $cadvisorId ]]; then
        docker kill $cadvisorId > /dev/null
    fi    
}

usage() {
    echo "Usage: run [ -g | --generate-only ] participant-range"
    exit 2
}

main "$@"
