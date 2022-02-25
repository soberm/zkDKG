// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package dkg

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// KeyVerifierProof is an auto generated low-level Go binding around an user-defined struct.
type KeyVerifierProof struct {
	A PairingG1Point
	B PairingG2Point
	C PairingG1Point
}

// PairingG1Point is an auto generated low-level Go binding around an user-defined struct.
type PairingG1Point struct {
	X *big.Int
	Y *big.Int
}

// PairingG2Point is an auto generated low-level Go binding around an user-defined struct.
type PairingG2Point struct {
	X [2]*big.Int
	Y [2]*big.Int
}

// ShareVerifierProof is an auto generated low-level Go binding around an user-defined struct.
type ShareVerifierProof struct {
	A PairingG1Point
	B PairingG2Point
	C PairingG1Point
}

// ZKDKGParticipant is an auto generated low-level Go binding around an user-defined struct.
type ZKDKGParticipant struct {
	Index     *big.Int
	PublicKey [2]*big.Int
}

// ZKDKGContractABI is the input ABI used to generate the binding from.
const ZKDKGContractABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_shareVerifier\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_keyVerifier\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_noParticipants\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"BroadcastSharesLog\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"result\",\"type\":\"bool\"}],\"name\":\"DisputeShare\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"RegistrationEndLog\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MIN_STAKE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"addresses\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[2][]\",\"name\":\"commitments\",\"type\":\"uint256[2][]\"},{\"internalType\":\"uint256[]\",\"name\":\"shares\",\"type\":\"uint256[]\"}],\"name\":\"broadcastShares\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"commitmentHashes\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"countParticipants\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[2]\",\"name\":\"_publicKey\",\"type\":\"uint256[2]\"},{\"components\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structPairing.G1Point\",\"name\":\"a\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structPairing.G2Point\",\"name\":\"b\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structPairing.G1Point\",\"name\":\"c\",\"type\":\"tuple\"}],\"internalType\":\"structKeyVerifier.Proof\",\"name\":\"proof\",\"type\":\"tuple\"}],\"name\":\"derivePublicKey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dealer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"share\",\"type\":\"uint256\"},{\"components\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structPairing.G1Point\",\"name\":\"a\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structPairing.G2Point\",\"name\":\"b\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structPairing.G1Point\",\"name\":\"c\",\"type\":\"tuple\"}],\"internalType\":\"structShareVerifier.Proof\",\"name\":\"proof\",\"type\":\"tuple\"}],\"name\":\"disputeShare\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_index\",\"type\":\"uint256\"}],\"name\":\"findParticipantByIndex\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"internalType\":\"uint256[2]\",\"name\":\"publicKey\",\"type\":\"uint256[2]\"}],\"internalType\":\"structZKDKG.Participant\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"firstCoefficients\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_hash\",\"type\":\"bytes32\"}],\"name\":\"hashToUint128\",\"outputs\":[{\"internalType\":\"uint256[2]\",\"name\":\"\",\"type\":\"uint256[2]\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"isRegistered\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"masterPublicKey\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"noParticipants\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"participants\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[2]\",\"name\":\"publicKey\",\"type\":\"uint256[2]\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"shareHashes\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"threshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ZKDKGContract is an auto generated Go binding around an Ethereum contract.
type ZKDKGContract struct {
	ZKDKGContractCaller     // Read-only binding to the contract
	ZKDKGContractTransactor // Write-only binding to the contract
	ZKDKGContractFilterer   // Log filterer for contract events
}

// ZKDKGContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ZKDKGContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZKDKGContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ZKDKGContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZKDKGContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ZKDKGContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZKDKGContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ZKDKGContractSession struct {
	Contract     *ZKDKGContract    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ZKDKGContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ZKDKGContractCallerSession struct {
	Contract *ZKDKGContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ZKDKGContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ZKDKGContractTransactorSession struct {
	Contract     *ZKDKGContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ZKDKGContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ZKDKGContractRaw struct {
	Contract *ZKDKGContract // Generic contract binding to access the raw methods on
}

// ZKDKGContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ZKDKGContractCallerRaw struct {
	Contract *ZKDKGContractCaller // Generic read-only contract binding to access the raw methods on
}

// ZKDKGContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ZKDKGContractTransactorRaw struct {
	Contract *ZKDKGContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewZKDKGContract creates a new instance of ZKDKGContract, bound to a specific deployed contract.
func NewZKDKGContract(address common.Address, backend bind.ContractBackend) (*ZKDKGContract, error) {
	contract, err := bindZKDKGContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ZKDKGContract{ZKDKGContractCaller: ZKDKGContractCaller{contract: contract}, ZKDKGContractTransactor: ZKDKGContractTransactor{contract: contract}, ZKDKGContractFilterer: ZKDKGContractFilterer{contract: contract}}, nil
}

// NewZKDKGContractCaller creates a new read-only instance of ZKDKGContract, bound to a specific deployed contract.
func NewZKDKGContractCaller(address common.Address, caller bind.ContractCaller) (*ZKDKGContractCaller, error) {
	contract, err := bindZKDKGContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ZKDKGContractCaller{contract: contract}, nil
}

// NewZKDKGContractTransactor creates a new write-only instance of ZKDKGContract, bound to a specific deployed contract.
func NewZKDKGContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ZKDKGContractTransactor, error) {
	contract, err := bindZKDKGContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ZKDKGContractTransactor{contract: contract}, nil
}

// NewZKDKGContractFilterer creates a new log filterer instance of ZKDKGContract, bound to a specific deployed contract.
func NewZKDKGContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ZKDKGContractFilterer, error) {
	contract, err := bindZKDKGContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ZKDKGContractFilterer{contract: contract}, nil
}

// bindZKDKGContract binds a generic wrapper to an already deployed contract.
func bindZKDKGContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ZKDKGContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ZKDKGContract *ZKDKGContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ZKDKGContract.Contract.ZKDKGContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ZKDKGContract *ZKDKGContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.ZKDKGContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ZKDKGContract *ZKDKGContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.ZKDKGContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ZKDKGContract *ZKDKGContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ZKDKGContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ZKDKGContract *ZKDKGContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ZKDKGContract *ZKDKGContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.contract.Transact(opts, method, params...)
}

// MINSTAKE is a free data retrieval call binding the contract method 0xcb1c2b5c.
//
// Solidity: function MIN_STAKE() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCaller) MINSTAKE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "MIN_STAKE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINSTAKE is a free data retrieval call binding the contract method 0xcb1c2b5c.
//
// Solidity: function MIN_STAKE() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractSession) MINSTAKE() (*big.Int, error) {
	return _ZKDKGContract.Contract.MINSTAKE(&_ZKDKGContract.CallOpts)
}

// MINSTAKE is a free data retrieval call binding the contract method 0xcb1c2b5c.
//
// Solidity: function MIN_STAKE() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCallerSession) MINSTAKE() (*big.Int, error) {
	return _ZKDKGContract.Contract.MINSTAKE(&_ZKDKGContract.CallOpts)
}

// Addresses is a free data retrieval call binding the contract method 0xedf26d9b.
//
// Solidity: function addresses(uint256 ) view returns(address)
func (_ZKDKGContract *ZKDKGContractCaller) Addresses(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "addresses", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Addresses is a free data retrieval call binding the contract method 0xedf26d9b.
//
// Solidity: function addresses(uint256 ) view returns(address)
func (_ZKDKGContract *ZKDKGContractSession) Addresses(arg0 *big.Int) (common.Address, error) {
	return _ZKDKGContract.Contract.Addresses(&_ZKDKGContract.CallOpts, arg0)
}

// Addresses is a free data retrieval call binding the contract method 0xedf26d9b.
//
// Solidity: function addresses(uint256 ) view returns(address)
func (_ZKDKGContract *ZKDKGContractCallerSession) Addresses(arg0 *big.Int) (common.Address, error) {
	return _ZKDKGContract.Contract.Addresses(&_ZKDKGContract.CallOpts, arg0)
}

// CommitmentHashes is a free data retrieval call binding the contract method 0x8a48b163.
//
// Solidity: function commitmentHashes(address ) view returns(bytes32)
func (_ZKDKGContract *ZKDKGContractCaller) CommitmentHashes(opts *bind.CallOpts, arg0 common.Address) ([32]byte, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "commitmentHashes", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CommitmentHashes is a free data retrieval call binding the contract method 0x8a48b163.
//
// Solidity: function commitmentHashes(address ) view returns(bytes32)
func (_ZKDKGContract *ZKDKGContractSession) CommitmentHashes(arg0 common.Address) ([32]byte, error) {
	return _ZKDKGContract.Contract.CommitmentHashes(&_ZKDKGContract.CallOpts, arg0)
}

// CommitmentHashes is a free data retrieval call binding the contract method 0x8a48b163.
//
// Solidity: function commitmentHashes(address ) view returns(bytes32)
func (_ZKDKGContract *ZKDKGContractCallerSession) CommitmentHashes(arg0 common.Address) ([32]byte, error) {
	return _ZKDKGContract.Contract.CommitmentHashes(&_ZKDKGContract.CallOpts, arg0)
}

// CountParticipants is a free data retrieval call binding the contract method 0x5c52bba7.
//
// Solidity: function countParticipants() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCaller) CountParticipants(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "countParticipants")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CountParticipants is a free data retrieval call binding the contract method 0x5c52bba7.
//
// Solidity: function countParticipants() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractSession) CountParticipants() (*big.Int, error) {
	return _ZKDKGContract.Contract.CountParticipants(&_ZKDKGContract.CallOpts)
}

// CountParticipants is a free data retrieval call binding the contract method 0x5c52bba7.
//
// Solidity: function countParticipants() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCallerSession) CountParticipants() (*big.Int, error) {
	return _ZKDKGContract.Contract.CountParticipants(&_ZKDKGContract.CallOpts)
}

// FindParticipantByIndex is a free data retrieval call binding the contract method 0x9586f96a.
//
// Solidity: function findParticipantByIndex(uint256 _index) view returns((uint256,uint256[2]))
func (_ZKDKGContract *ZKDKGContractCaller) FindParticipantByIndex(opts *bind.CallOpts, _index *big.Int) (ZKDKGParticipant, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "findParticipantByIndex", _index)

	if err != nil {
		return *new(ZKDKGParticipant), err
	}

	out0 := *abi.ConvertType(out[0], new(ZKDKGParticipant)).(*ZKDKGParticipant)

	return out0, err

}

// FindParticipantByIndex is a free data retrieval call binding the contract method 0x9586f96a.
//
// Solidity: function findParticipantByIndex(uint256 _index) view returns((uint256,uint256[2]))
func (_ZKDKGContract *ZKDKGContractSession) FindParticipantByIndex(_index *big.Int) (ZKDKGParticipant, error) {
	return _ZKDKGContract.Contract.FindParticipantByIndex(&_ZKDKGContract.CallOpts, _index)
}

// FindParticipantByIndex is a free data retrieval call binding the contract method 0x9586f96a.
//
// Solidity: function findParticipantByIndex(uint256 _index) view returns((uint256,uint256[2]))
func (_ZKDKGContract *ZKDKGContractCallerSession) FindParticipantByIndex(_index *big.Int) (ZKDKGParticipant, error) {
	return _ZKDKGContract.Contract.FindParticipantByIndex(&_ZKDKGContract.CallOpts, _index)
}

// FirstCoefficients is a free data retrieval call binding the contract method 0xf5e3e4c8.
//
// Solidity: function firstCoefficients(uint256 , uint256 ) view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCaller) FirstCoefficients(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "firstCoefficients", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FirstCoefficients is a free data retrieval call binding the contract method 0xf5e3e4c8.
//
// Solidity: function firstCoefficients(uint256 , uint256 ) view returns(uint256)
func (_ZKDKGContract *ZKDKGContractSession) FirstCoefficients(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _ZKDKGContract.Contract.FirstCoefficients(&_ZKDKGContract.CallOpts, arg0, arg1)
}

// FirstCoefficients is a free data retrieval call binding the contract method 0xf5e3e4c8.
//
// Solidity: function firstCoefficients(uint256 , uint256 ) view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCallerSession) FirstCoefficients(arg0 *big.Int, arg1 *big.Int) (*big.Int, error) {
	return _ZKDKGContract.Contract.FirstCoefficients(&_ZKDKGContract.CallOpts, arg0, arg1)
}

// HashToUint128 is a free data retrieval call binding the contract method 0x015d67af.
//
// Solidity: function hashToUint128(bytes32 _hash) pure returns(uint256[2])
func (_ZKDKGContract *ZKDKGContractCaller) HashToUint128(opts *bind.CallOpts, _hash [32]byte) ([2]*big.Int, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "hashToUint128", _hash)

	if err != nil {
		return *new([2]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([2]*big.Int)).(*[2]*big.Int)

	return out0, err

}

// HashToUint128 is a free data retrieval call binding the contract method 0x015d67af.
//
// Solidity: function hashToUint128(bytes32 _hash) pure returns(uint256[2])
func (_ZKDKGContract *ZKDKGContractSession) HashToUint128(_hash [32]byte) ([2]*big.Int, error) {
	return _ZKDKGContract.Contract.HashToUint128(&_ZKDKGContract.CallOpts, _hash)
}

// HashToUint128 is a free data retrieval call binding the contract method 0x015d67af.
//
// Solidity: function hashToUint128(bytes32 _hash) pure returns(uint256[2])
func (_ZKDKGContract *ZKDKGContractCallerSession) HashToUint128(_hash [32]byte) ([2]*big.Int, error) {
	return _ZKDKGContract.Contract.HashToUint128(&_ZKDKGContract.CallOpts, _hash)
}

// IsRegistered is a free data retrieval call binding the contract method 0xc3c5a547.
//
// Solidity: function isRegistered(address _addr) view returns(bool)
func (_ZKDKGContract *ZKDKGContractCaller) IsRegistered(opts *bind.CallOpts, _addr common.Address) (bool, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "isRegistered", _addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsRegistered is a free data retrieval call binding the contract method 0xc3c5a547.
//
// Solidity: function isRegistered(address _addr) view returns(bool)
func (_ZKDKGContract *ZKDKGContractSession) IsRegistered(_addr common.Address) (bool, error) {
	return _ZKDKGContract.Contract.IsRegistered(&_ZKDKGContract.CallOpts, _addr)
}

// IsRegistered is a free data retrieval call binding the contract method 0xc3c5a547.
//
// Solidity: function isRegistered(address _addr) view returns(bool)
func (_ZKDKGContract *ZKDKGContractCallerSession) IsRegistered(_addr common.Address) (bool, error) {
	return _ZKDKGContract.Contract.IsRegistered(&_ZKDKGContract.CallOpts, _addr)
}

// MasterPublicKey is a free data retrieval call binding the contract method 0x555ba1b6.
//
// Solidity: function masterPublicKey(uint256 ) view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCaller) MasterPublicKey(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "masterPublicKey", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MasterPublicKey is a free data retrieval call binding the contract method 0x555ba1b6.
//
// Solidity: function masterPublicKey(uint256 ) view returns(uint256)
func (_ZKDKGContract *ZKDKGContractSession) MasterPublicKey(arg0 *big.Int) (*big.Int, error) {
	return _ZKDKGContract.Contract.MasterPublicKey(&_ZKDKGContract.CallOpts, arg0)
}

// MasterPublicKey is a free data retrieval call binding the contract method 0x555ba1b6.
//
// Solidity: function masterPublicKey(uint256 ) view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCallerSession) MasterPublicKey(arg0 *big.Int) (*big.Int, error) {
	return _ZKDKGContract.Contract.MasterPublicKey(&_ZKDKGContract.CallOpts, arg0)
}

// NoParticipants is a free data retrieval call binding the contract method 0x3a3b4f62.
//
// Solidity: function noParticipants() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCaller) NoParticipants(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "noParticipants")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NoParticipants is a free data retrieval call binding the contract method 0x3a3b4f62.
//
// Solidity: function noParticipants() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractSession) NoParticipants() (*big.Int, error) {
	return _ZKDKGContract.Contract.NoParticipants(&_ZKDKGContract.CallOpts)
}

// NoParticipants is a free data retrieval call binding the contract method 0x3a3b4f62.
//
// Solidity: function noParticipants() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCallerSession) NoParticipants() (*big.Int, error) {
	return _ZKDKGContract.Contract.NoParticipants(&_ZKDKGContract.CallOpts)
}

// Participants is a free data retrieval call binding the contract method 0x09e69ede.
//
// Solidity: function participants(address ) view returns(uint256 index)
func (_ZKDKGContract *ZKDKGContractCaller) Participants(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "participants", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Participants is a free data retrieval call binding the contract method 0x09e69ede.
//
// Solidity: function participants(address ) view returns(uint256 index)
func (_ZKDKGContract *ZKDKGContractSession) Participants(arg0 common.Address) (*big.Int, error) {
	return _ZKDKGContract.Contract.Participants(&_ZKDKGContract.CallOpts, arg0)
}

// Participants is a free data retrieval call binding the contract method 0x09e69ede.
//
// Solidity: function participants(address ) view returns(uint256 index)
func (_ZKDKGContract *ZKDKGContractCallerSession) Participants(arg0 common.Address) (*big.Int, error) {
	return _ZKDKGContract.Contract.Participants(&_ZKDKGContract.CallOpts, arg0)
}

// ShareHashes is a free data retrieval call binding the contract method 0xfec140fc.
//
// Solidity: function shareHashes(address ) view returns(bytes32)
func (_ZKDKGContract *ZKDKGContractCaller) ShareHashes(opts *bind.CallOpts, arg0 common.Address) ([32]byte, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "shareHashes", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ShareHashes is a free data retrieval call binding the contract method 0xfec140fc.
//
// Solidity: function shareHashes(address ) view returns(bytes32)
func (_ZKDKGContract *ZKDKGContractSession) ShareHashes(arg0 common.Address) ([32]byte, error) {
	return _ZKDKGContract.Contract.ShareHashes(&_ZKDKGContract.CallOpts, arg0)
}

// ShareHashes is a free data retrieval call binding the contract method 0xfec140fc.
//
// Solidity: function shareHashes(address ) view returns(bytes32)
func (_ZKDKGContract *ZKDKGContractCallerSession) ShareHashes(arg0 common.Address) ([32]byte, error) {
	return _ZKDKGContract.Contract.ShareHashes(&_ZKDKGContract.CallOpts, arg0)
}

// Threshold is a free data retrieval call binding the contract method 0x42cde4e8.
//
// Solidity: function threshold() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCaller) Threshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ZKDKGContract.contract.Call(opts, &out, "threshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Threshold is a free data retrieval call binding the contract method 0x42cde4e8.
//
// Solidity: function threshold() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractSession) Threshold() (*big.Int, error) {
	return _ZKDKGContract.Contract.Threshold(&_ZKDKGContract.CallOpts)
}

// Threshold is a free data retrieval call binding the contract method 0x42cde4e8.
//
// Solidity: function threshold() view returns(uint256)
func (_ZKDKGContract *ZKDKGContractCallerSession) Threshold() (*big.Int, error) {
	return _ZKDKGContract.Contract.Threshold(&_ZKDKGContract.CallOpts)
}

// BroadcastShares is a paid mutator transaction binding the contract method 0x7a357ebf.
//
// Solidity: function broadcastShares(uint256[2][] commitments, uint256[] shares) returns()
func (_ZKDKGContract *ZKDKGContractTransactor) BroadcastShares(opts *bind.TransactOpts, commitments [][2]*big.Int, shares []*big.Int) (*types.Transaction, error) {
	return _ZKDKGContract.contract.Transact(opts, "broadcastShares", commitments, shares)
}

// BroadcastShares is a paid mutator transaction binding the contract method 0x7a357ebf.
//
// Solidity: function broadcastShares(uint256[2][] commitments, uint256[] shares) returns()
func (_ZKDKGContract *ZKDKGContractSession) BroadcastShares(commitments [][2]*big.Int, shares []*big.Int) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.BroadcastShares(&_ZKDKGContract.TransactOpts, commitments, shares)
}

// BroadcastShares is a paid mutator transaction binding the contract method 0x7a357ebf.
//
// Solidity: function broadcastShares(uint256[2][] commitments, uint256[] shares) returns()
func (_ZKDKGContract *ZKDKGContractTransactorSession) BroadcastShares(commitments [][2]*big.Int, shares []*big.Int) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.BroadcastShares(&_ZKDKGContract.TransactOpts, commitments, shares)
}

// DerivePublicKey is a paid mutator transaction binding the contract method 0xe2439dda.
//
// Solidity: function derivePublicKey(uint256[2] _publicKey, ((uint256,uint256),(uint256[2],uint256[2]),(uint256,uint256)) proof) returns()
func (_ZKDKGContract *ZKDKGContractTransactor) DerivePublicKey(opts *bind.TransactOpts, _publicKey [2]*big.Int, proof KeyVerifierProof) (*types.Transaction, error) {
	return _ZKDKGContract.contract.Transact(opts, "derivePublicKey", _publicKey, proof)
}

// DerivePublicKey is a paid mutator transaction binding the contract method 0xe2439dda.
//
// Solidity: function derivePublicKey(uint256[2] _publicKey, ((uint256,uint256),(uint256[2],uint256[2]),(uint256,uint256)) proof) returns()
func (_ZKDKGContract *ZKDKGContractSession) DerivePublicKey(_publicKey [2]*big.Int, proof KeyVerifierProof) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.DerivePublicKey(&_ZKDKGContract.TransactOpts, _publicKey, proof)
}

// DerivePublicKey is a paid mutator transaction binding the contract method 0xe2439dda.
//
// Solidity: function derivePublicKey(uint256[2] _publicKey, ((uint256,uint256),(uint256[2],uint256[2]),(uint256,uint256)) proof) returns()
func (_ZKDKGContract *ZKDKGContractTransactorSession) DerivePublicKey(_publicKey [2]*big.Int, proof KeyVerifierProof) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.DerivePublicKey(&_ZKDKGContract.TransactOpts, _publicKey, proof)
}

// DisputeShare is a paid mutator transaction binding the contract method 0x7af6b904.
//
// Solidity: function disputeShare(address dealer, uint256 share, ((uint256,uint256),(uint256[2],uint256[2]),(uint256,uint256)) proof) returns()
func (_ZKDKGContract *ZKDKGContractTransactor) DisputeShare(opts *bind.TransactOpts, dealer common.Address, share *big.Int, proof ShareVerifierProof) (*types.Transaction, error) {
	return _ZKDKGContract.contract.Transact(opts, "disputeShare", dealer, share, proof)
}

// DisputeShare is a paid mutator transaction binding the contract method 0x7af6b904.
//
// Solidity: function disputeShare(address dealer, uint256 share, ((uint256,uint256),(uint256[2],uint256[2]),(uint256,uint256)) proof) returns()
func (_ZKDKGContract *ZKDKGContractSession) DisputeShare(dealer common.Address, share *big.Int, proof ShareVerifierProof) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.DisputeShare(&_ZKDKGContract.TransactOpts, dealer, share, proof)
}

// DisputeShare is a paid mutator transaction binding the contract method 0x7af6b904.
//
// Solidity: function disputeShare(address dealer, uint256 share, ((uint256,uint256),(uint256[2],uint256[2]),(uint256,uint256)) proof) returns()
func (_ZKDKGContract *ZKDKGContractTransactorSession) DisputeShare(dealer common.Address, share *big.Int, proof ShareVerifierProof) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.DisputeShare(&_ZKDKGContract.TransactOpts, dealer, share, proof)
}

// Register is a paid mutator transaction binding the contract method 0x3442af5c.
//
// Solidity: function register(uint256[2] publicKey) payable returns()
func (_ZKDKGContract *ZKDKGContractTransactor) Register(opts *bind.TransactOpts, publicKey [2]*big.Int) (*types.Transaction, error) {
	return _ZKDKGContract.contract.Transact(opts, "register", publicKey)
}

// Register is a paid mutator transaction binding the contract method 0x3442af5c.
//
// Solidity: function register(uint256[2] publicKey) payable returns()
func (_ZKDKGContract *ZKDKGContractSession) Register(publicKey [2]*big.Int) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.Register(&_ZKDKGContract.TransactOpts, publicKey)
}

// Register is a paid mutator transaction binding the contract method 0x3442af5c.
//
// Solidity: function register(uint256[2] publicKey) payable returns()
func (_ZKDKGContract *ZKDKGContractTransactorSession) Register(publicKey [2]*big.Int) (*types.Transaction, error) {
	return _ZKDKGContract.Contract.Register(&_ZKDKGContract.TransactOpts, publicKey)
}

// ZKDKGContractBroadcastSharesLogIterator is returned from FilterBroadcastSharesLog and is used to iterate over the raw logs and unpacked data for BroadcastSharesLog events raised by the ZKDKGContract contract.
type ZKDKGContractBroadcastSharesLogIterator struct {
	Event *ZKDKGContractBroadcastSharesLog // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ZKDKGContractBroadcastSharesLogIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZKDKGContractBroadcastSharesLog)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ZKDKGContractBroadcastSharesLog)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ZKDKGContractBroadcastSharesLogIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZKDKGContractBroadcastSharesLogIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZKDKGContractBroadcastSharesLog represents a BroadcastSharesLog event raised by the ZKDKGContract contract.
type ZKDKGContractBroadcastSharesLog struct {
	Sender common.Address
	Index  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBroadcastSharesLog is a free log retrieval operation binding the contract event 0xe9f361b9d754d638c67ee5a79969234d86c3c510db254b8903929d5f0bcf1fc8.
//
// Solidity: event BroadcastSharesLog(address sender, uint256 index)
func (_ZKDKGContract *ZKDKGContractFilterer) FilterBroadcastSharesLog(opts *bind.FilterOpts) (*ZKDKGContractBroadcastSharesLogIterator, error) {

	logs, sub, err := _ZKDKGContract.contract.FilterLogs(opts, "BroadcastSharesLog")
	if err != nil {
		return nil, err
	}
	return &ZKDKGContractBroadcastSharesLogIterator{contract: _ZKDKGContract.contract, event: "BroadcastSharesLog", logs: logs, sub: sub}, nil
}

// WatchBroadcastSharesLog is a free log subscription operation binding the contract event 0xe9f361b9d754d638c67ee5a79969234d86c3c510db254b8903929d5f0bcf1fc8.
//
// Solidity: event BroadcastSharesLog(address sender, uint256 index)
func (_ZKDKGContract *ZKDKGContractFilterer) WatchBroadcastSharesLog(opts *bind.WatchOpts, sink chan<- *ZKDKGContractBroadcastSharesLog) (event.Subscription, error) {

	logs, sub, err := _ZKDKGContract.contract.WatchLogs(opts, "BroadcastSharesLog")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZKDKGContractBroadcastSharesLog)
				if err := _ZKDKGContract.contract.UnpackLog(event, "BroadcastSharesLog", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBroadcastSharesLog is a log parse operation binding the contract event 0xe9f361b9d754d638c67ee5a79969234d86c3c510db254b8903929d5f0bcf1fc8.
//
// Solidity: event BroadcastSharesLog(address sender, uint256 index)
func (_ZKDKGContract *ZKDKGContractFilterer) ParseBroadcastSharesLog(log types.Log) (*ZKDKGContractBroadcastSharesLog, error) {
	event := new(ZKDKGContractBroadcastSharesLog)
	if err := _ZKDKGContract.contract.UnpackLog(event, "BroadcastSharesLog", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ZKDKGContractDisputeShareIterator is returned from FilterDisputeShare and is used to iterate over the raw logs and unpacked data for DisputeShare events raised by the ZKDKGContract contract.
type ZKDKGContractDisputeShareIterator struct {
	Event *ZKDKGContractDisputeShare // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ZKDKGContractDisputeShareIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZKDKGContractDisputeShare)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ZKDKGContractDisputeShare)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ZKDKGContractDisputeShareIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZKDKGContractDisputeShareIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZKDKGContractDisputeShare represents a DisputeShare event raised by the ZKDKGContract contract.
type ZKDKGContractDisputeShare struct {
	Result bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDisputeShare is a free log retrieval operation binding the contract event 0xb2d2b8f9b4fc8422ebc1a150457ebb30eff244875f0da6180ef47749ffbd74da.
//
// Solidity: event DisputeShare(bool result)
func (_ZKDKGContract *ZKDKGContractFilterer) FilterDisputeShare(opts *bind.FilterOpts) (*ZKDKGContractDisputeShareIterator, error) {

	logs, sub, err := _ZKDKGContract.contract.FilterLogs(opts, "DisputeShare")
	if err != nil {
		return nil, err
	}
	return &ZKDKGContractDisputeShareIterator{contract: _ZKDKGContract.contract, event: "DisputeShare", logs: logs, sub: sub}, nil
}

// WatchDisputeShare is a free log subscription operation binding the contract event 0xb2d2b8f9b4fc8422ebc1a150457ebb30eff244875f0da6180ef47749ffbd74da.
//
// Solidity: event DisputeShare(bool result)
func (_ZKDKGContract *ZKDKGContractFilterer) WatchDisputeShare(opts *bind.WatchOpts, sink chan<- *ZKDKGContractDisputeShare) (event.Subscription, error) {

	logs, sub, err := _ZKDKGContract.contract.WatchLogs(opts, "DisputeShare")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZKDKGContractDisputeShare)
				if err := _ZKDKGContract.contract.UnpackLog(event, "DisputeShare", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDisputeShare is a log parse operation binding the contract event 0xb2d2b8f9b4fc8422ebc1a150457ebb30eff244875f0da6180ef47749ffbd74da.
//
// Solidity: event DisputeShare(bool result)
func (_ZKDKGContract *ZKDKGContractFilterer) ParseDisputeShare(log types.Log) (*ZKDKGContractDisputeShare, error) {
	event := new(ZKDKGContractDisputeShare)
	if err := _ZKDKGContract.contract.UnpackLog(event, "DisputeShare", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ZKDKGContractRegistrationEndLogIterator is returned from FilterRegistrationEndLog and is used to iterate over the raw logs and unpacked data for RegistrationEndLog events raised by the ZKDKGContract contract.
type ZKDKGContractRegistrationEndLogIterator struct {
	Event *ZKDKGContractRegistrationEndLog // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ZKDKGContractRegistrationEndLogIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ZKDKGContractRegistrationEndLog)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ZKDKGContractRegistrationEndLog)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ZKDKGContractRegistrationEndLogIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ZKDKGContractRegistrationEndLogIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ZKDKGContractRegistrationEndLog represents a RegistrationEndLog event raised by the ZKDKGContract contract.
type ZKDKGContractRegistrationEndLog struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterRegistrationEndLog is a free log retrieval operation binding the contract event 0x4bdb43f822bd6cc36c8e0ae7be9183af9b7abc30c6d42bb71e156fe987e2b858.
//
// Solidity: event RegistrationEndLog()
func (_ZKDKGContract *ZKDKGContractFilterer) FilterRegistrationEndLog(opts *bind.FilterOpts) (*ZKDKGContractRegistrationEndLogIterator, error) {

	logs, sub, err := _ZKDKGContract.contract.FilterLogs(opts, "RegistrationEndLog")
	if err != nil {
		return nil, err
	}
	return &ZKDKGContractRegistrationEndLogIterator{contract: _ZKDKGContract.contract, event: "RegistrationEndLog", logs: logs, sub: sub}, nil
}

// WatchRegistrationEndLog is a free log subscription operation binding the contract event 0x4bdb43f822bd6cc36c8e0ae7be9183af9b7abc30c6d42bb71e156fe987e2b858.
//
// Solidity: event RegistrationEndLog()
func (_ZKDKGContract *ZKDKGContractFilterer) WatchRegistrationEndLog(opts *bind.WatchOpts, sink chan<- *ZKDKGContractRegistrationEndLog) (event.Subscription, error) {

	logs, sub, err := _ZKDKGContract.contract.WatchLogs(opts, "RegistrationEndLog")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ZKDKGContractRegistrationEndLog)
				if err := _ZKDKGContract.contract.UnpackLog(event, "RegistrationEndLog", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRegistrationEndLog is a log parse operation binding the contract event 0x4bdb43f822bd6cc36c8e0ae7be9183af9b7abc30c6d42bb71e156fe987e2b858.
//
// Solidity: event RegistrationEndLog()
func (_ZKDKGContract *ZKDKGContractFilterer) ParseRegistrationEndLog(log types.Log) (*ZKDKGContractRegistrationEndLog, error) {
	event := new(ZKDKGContractRegistrationEndLog)
	if err := _ZKDKGContract.contract.UnpackLog(event, "RegistrationEndLog", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
