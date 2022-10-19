![CDL-BOT Logo](https://www.cdl-bot.at/theme/images/logo.png)

# zkDKG

> Distributed Key Generation with Smart Contracts using zk-SNARKs

This project contains a prototypical implementation, i.e. the Go client software, the Distributed Key Generation smart contracts, the Zero Knowledge Proof programs and the evaluation scripts, of its [associated research paper](https://doi.org/xx.xxx/xxx_x).

## Installation

### Prerequisites

- Node LTS (>= 14)
- Docker ([rootless](https://docs.docker.com/engine/security/rootless/))
- Go (>= 1.13)
- Go Ethereum (with developer tools included)

### Setup

Run `npm install` to install the dependencies

## Evaluation

The evaluation will simulate a protocol run with a given number of participants n.
n instances of the Go client software will register and generate their encrypted shares and commitments and broadcast them to the zkDKG smart contract.
One client node will dispute a valid commitment, which forces the disputed node to defend the validity of its broadcast through a ZK proof.
After the successful defense, another ZK proof, confirming the correct calculation of the public key, will be computed and submitted and the protocol run is completed.

```shell
./scripts/evaluation.sh 4,8,16 5
```

to run the above described protocol with 4,8 and 16 participants, each being repeated 5 times.
The gas costs of the smart contract transactions and the runtime and memory usage of the ZK proof generations will be written to `build/$participants/report.csv`.

## Troubleshooting

If you are getting TCP timeouts in Go when running the evaluation scripts (especially for a higher amount of participants), increase the values of either [wsPingInterval](https://github.com/ethereum/go-ethereum/blob/69568c554880b3567bace64f8848ff1be27d084d/rpc/websocket.go#L38) and / or [wsPongTimeout](https://github.com/ethereum/go-ethereum/blob/69568c554880b3567bace64f8848ff1be27d084d/rpc/websocket.go#L40).

## Contributing

This project is a research prototype. We welcome anyone to contribute. File a bug report or submit feature requests through the [issue tracker](https://github.com/soberm/zkDKG/issues). If you want to contribute feel free to submit a pull request.

## Licensing

The code in this project is licensed under the MIT license.
