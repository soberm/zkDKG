#!/bin/bash

cd "$(dirname $0)"/.. || exit 1

generateOnly=false
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

    mkdir -p "$buildRoot"

    if ! $generateOnly; then
        npx hardhat compile
        (cd ./dkg/; go build -o "$buildRoot" ./cmd/full_node)
    fi

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

    for participants in ${participantsSizes[@]}; do

        ./scripts/build.sh $participants

        buildDir="$buildRoot"/$participants
        containerPipe="$buildDir"/container_pipe
        declare -a ethPrivs

        if [[ ! -p $containerPipe ]]; then
            rm -f "$containerPipe"
            mkfifo "$containerPipe"
        fi

        if ! $generateOnly; then
            log="$buildDir"/hardhat.log

            mkdir -p "$buildDir"/nodes

            npx ts-node ./scripts/collectStats.ts "$containerPipe" "$buildRoot"/cadvisor.log "$log" $repetitions > "$buildDir"/report.csv &
            scriptPid=$!
        fi

        for ((repetition = 1; repetition <= repetitions; repetition++)); do
            echo "Starting to measure stats for run no. $repetition/$repetitions for $participants participants"

            if ! $generateOnly; then
                npx hardhat launch $participants > "$log" &
                nodePid=$!

                # Retrieve the private keys for the accounts from the log of the Hardhat node
                ethPrivs=( $(tail -f "$log" | awk 'BEGIN{i=0; ORS=" "} match($0, /Private Key: 0x([[:alnum:]]+)/, res){print res[1]; if (++i == n) exit}' n=$participants) )

                npx hardhat --network localhost deploy $participants
            fi

            local goPids=()

            if $generateOnly; then
                generate_config 1 $participants | go run ./dkg/cmd/generator -c /dev/stdin --participants $participants --id-pipe="$containerPipe" |& tee "$buildDir"/generator.log &
                goPids[0]=$!
            else
                for ((i = 1; i <= participants; i++)); do
                    flags=()
                    if (( i == 2 )); then # The 2nd node should dispute the 1st node's broadcast
                        flags+=("--dispute-valid" )
                    fi

                    if (( i == 1 )); then # Report container IDs of the containers running the zokrates commands
                        flags+=("--id-pipe=$containerPipe")
                    fi

                    if (( i != 1 && i != 2 )); then
                        flags+=("--broadcast-only")
                    fi

                    generate_config $i $participants | ./build/full_node -c /dev/stdin ${flags[@]} |& tee "$buildDir"/nodes/node_$i.log &
                    goPids[$i]=$!

                    if (( i == 1 || i == 2 )); then
                        sleep 3 # Ensure that the first two started nodes really have the same index in the contract
                    fi
                done
            fi

            for pid in ${goPids[@]}; do
                wait $pid
            done

            if ! $generateOnly; then
                kill $nodePid
                unset nodePid
            fi
        done

        if ! $generateOnly; then
            wait $scriptPid
            unset scriptPid
        fi
    done
}

parse_input() {
    local args=$(getopt --name run --options g --longoptions generate-only -- "$@")
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

    local regex="^[0-9]+(,[0-9]+)*$"

    if [[ $1 =~ $regex ]]; then
        IFS=',' read -a participantsSizes <<< $1
    else
        usage
    fi

    if [[ $2 =~ [0-9]+ ]]; then
        repetitions=$2
    elif [[ -z $2 ]]; then
        repetitions=1
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
        \"MountSource\":        \"$(readlink -e ./build/$2/zk)\"
    }"
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
    echo "Usage: run [ -g | --generate-only ] participant-range [ repetitions ]"
    exit 2
}

main "$@"
