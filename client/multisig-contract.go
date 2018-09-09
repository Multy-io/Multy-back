// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package client

import (
	"math/big"
	"strings"

	//	"github.com/dzeckelev/ethereum-front/abi"
	ethereum "github.com/ethereum/go-ethereum"
	abi "github.com/ethereum/go-ethereum/accounts/abi"
	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// FactoryABI is the input ABI used to generate the binding from.
const FactoryABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"isInstantiation\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"instantiations\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"creator\",\"type\":\"address\"}],\"name\":\"getInstantiationCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"instantiation\",\"type\":\"address\"}],\"name\":\"ContractInstantiation\",\"type\":\"event\"}]"

// FactoryBin is the compiled bytecode used for deploying new contracts.
const FactoryBin = `0x608060405234801561001057600080fd5b506101e4806100206000396000f3006080604052600436106100565763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416632f4f3316811461005b57806357183c821461009d5780638f838478146100f7575b600080fd5b34801561006757600080fd5b5061008973ffffffffffffffffffffffffffffffffffffffff60043516610137565b604080519115158252519081900360200190f35b3480156100a957600080fd5b506100ce73ffffffffffffffffffffffffffffffffffffffff6004351660243561014c565b6040805173ffffffffffffffffffffffffffffffffffffffff9092168252519081900360200190f35b34801561010357600080fd5b5061012573ffffffffffffffffffffffffffffffffffffffff60043516610190565b60408051918252519081900360200190f35b60006020819052908152604090205460ff1681565b60016020528160005260406000208181548110151561016757fe5b60009182526020909120015473ffffffffffffffffffffffffffffffffffffffff169150829050565b73ffffffffffffffffffffffffffffffffffffffff16600090815260016020526040902054905600a165627a7a723058208cb2b1ed9a49c706d9186ce5cc78d5b7763f7c7a419bade324544e7870d7e3720029`

// DeployFactory deploys a new Ethereum contract, binding an instance of Factory to it.
func DeployFactory(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Factory, error) {
	parsed, err := abi.JSON(strings.NewReader(FactoryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(FactoryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Factory{FactoryCaller: FactoryCaller{contract: contract}, FactoryTransactor: FactoryTransactor{contract: contract}, FactoryFilterer: FactoryFilterer{contract: contract}}, nil
}

// Factory is an auto generated Go binding around an Ethereum contract.
type Factory struct {
	FactoryCaller     // Read-only binding to the contract
	FactoryTransactor // Write-only binding to the contract
	FactoryFilterer   // Log filterer for contract events
}

// FactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type FactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FactorySession struct {
	Contract     *Factory          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FactoryCallerSession struct {
	Contract *FactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// FactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FactoryTransactorSession struct {
	Contract     *FactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// FactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type FactoryRaw struct {
	Contract *Factory // Generic contract binding to access the raw methods on
}

// FactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FactoryCallerRaw struct {
	Contract *FactoryCaller // Generic read-only contract binding to access the raw methods on
}

// FactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FactoryTransactorRaw struct {
	Contract *FactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFactory creates a new instance of Factory, bound to a specific deployed contract.
func NewFactory(address common.Address, backend bind.ContractBackend) (*Factory, error) {
	contract, err := bindFactory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Factory{FactoryCaller: FactoryCaller{contract: contract}, FactoryTransactor: FactoryTransactor{contract: contract}, FactoryFilterer: FactoryFilterer{contract: contract}}, nil
}

// NewFactoryCaller creates a new read-only instance of Factory, bound to a specific deployed contract.
func NewFactoryCaller(address common.Address, caller bind.ContractCaller) (*FactoryCaller, error) {
	contract, err := bindFactory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FactoryCaller{contract: contract}, nil
}

// NewFactoryTransactor creates a new write-only instance of Factory, bound to a specific deployed contract.
func NewFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*FactoryTransactor, error) {
	contract, err := bindFactory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FactoryTransactor{contract: contract}, nil
}

// NewFactoryFilterer creates a new log filterer instance of Factory, bound to a specific deployed contract.
func NewFactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*FactoryFilterer, error) {
	contract, err := bindFactory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FactoryFilterer{contract: contract}, nil
}

// bindFactory binds a generic wrapper to an already deployed contract.
func bindFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FactoryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Factory *FactoryRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Factory.Contract.FactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Factory *FactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Factory.Contract.FactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Factory *FactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Factory.Contract.FactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Factory *FactoryCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Factory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Factory *FactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Factory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Factory *FactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Factory.Contract.contract.Transact(opts, method, params...)
}

// GetInstantiationCount is a free data retrieval call binding the contract method 0x8f838478.
//
// Solidity: function getInstantiationCount(creator address) constant returns(uint256)
func (_Factory *FactoryCaller) GetInstantiationCount(opts *bind.CallOpts, creator common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Factory.contract.Call(opts, out, "getInstantiationCount", creator)
	return *ret0, err
}

// GetInstantiationCount is a free data retrieval call binding the contract method 0x8f838478.
//
// Solidity: function getInstantiationCount(creator address) constant returns(uint256)
func (_Factory *FactorySession) GetInstantiationCount(creator common.Address) (*big.Int, error) {
	return _Factory.Contract.GetInstantiationCount(&_Factory.CallOpts, creator)
}

// GetInstantiationCount is a free data retrieval call binding the contract method 0x8f838478.
//
// Solidity: function getInstantiationCount(creator address) constant returns(uint256)
func (_Factory *FactoryCallerSession) GetInstantiationCount(creator common.Address) (*big.Int, error) {
	return _Factory.Contract.GetInstantiationCount(&_Factory.CallOpts, creator)
}

// Instantiations is a free data retrieval call binding the contract method 0x57183c82.
//
// Solidity: function instantiations( address,  uint256) constant returns(address)
func (_Factory *FactoryCaller) Instantiations(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Factory.contract.Call(opts, out, "instantiations", arg0, arg1)
	return *ret0, err
}

// Instantiations is a free data retrieval call binding the contract method 0x57183c82.
//
// Solidity: function instantiations( address,  uint256) constant returns(address)
func (_Factory *FactorySession) Instantiations(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _Factory.Contract.Instantiations(&_Factory.CallOpts, arg0, arg1)
}

// Instantiations is a free data retrieval call binding the contract method 0x57183c82.
//
// Solidity: function instantiations( address,  uint256) constant returns(address)
func (_Factory *FactoryCallerSession) Instantiations(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _Factory.Contract.Instantiations(&_Factory.CallOpts, arg0, arg1)
}

// IsInstantiation is a free data retrieval call binding the contract method 0x2f4f3316.
//
// Solidity: function isInstantiation( address) constant returns(bool)
func (_Factory *FactoryCaller) IsInstantiation(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Factory.contract.Call(opts, out, "isInstantiation", arg0)
	return *ret0, err
}

// IsInstantiation is a free data retrieval call binding the contract method 0x2f4f3316.
//
// Solidity: function isInstantiation( address) constant returns(bool)
func (_Factory *FactorySession) IsInstantiation(arg0 common.Address) (bool, error) {
	return _Factory.Contract.IsInstantiation(&_Factory.CallOpts, arg0)
}

// IsInstantiation is a free data retrieval call binding the contract method 0x2f4f3316.
//
// Solidity: function isInstantiation( address) constant returns(bool)
func (_Factory *FactoryCallerSession) IsInstantiation(arg0 common.Address) (bool, error) {
	return _Factory.Contract.IsInstantiation(&_Factory.CallOpts, arg0)
}

// FactoryContractInstantiationIterator is returned from FilterContractInstantiation and is used to iterate over the raw logs and unpacked data for ContractInstantiation events raised by the Factory contract.
type FactoryContractInstantiationIterator struct {
	Event *FactoryContractInstantiation // Event containing the contract specifics and raw log

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
func (it *FactoryContractInstantiationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FactoryContractInstantiation)
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
		it.Event = new(FactoryContractInstantiation)
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
func (it *FactoryContractInstantiationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FactoryContractInstantiationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FactoryContractInstantiation represents a ContractInstantiation event raised by the Factory contract.
type FactoryContractInstantiation struct {
	Sender        common.Address
	Instantiation common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterContractInstantiation is a free log retrieval operation binding the contract event 0x4fb057ad4a26ed17a57957fa69c306f11987596069b89521c511fc9a894e6161.
//
// Solidity: event ContractInstantiation(sender address, instantiation address)
func (_Factory *FactoryFilterer) FilterContractInstantiation(opts *bind.FilterOpts) (*FactoryContractInstantiationIterator, error) {

	logs, sub, err := _Factory.contract.FilterLogs(opts, "ContractInstantiation")
	if err != nil {
		return nil, err
	}
	return &FactoryContractInstantiationIterator{contract: _Factory.contract, event: "ContractInstantiation", logs: logs, sub: sub}, nil
}

// WatchContractInstantiation is a free log subscription operation binding the contract event 0x4fb057ad4a26ed17a57957fa69c306f11987596069b89521c511fc9a894e6161.
//
// Solidity: event ContractInstantiation(sender address, instantiation address)
func (_Factory *FactoryFilterer) WatchContractInstantiation(opts *bind.WatchOpts, sink chan<- *FactoryContractInstantiation) (event.Subscription, error) {

	logs, sub, err := _Factory.contract.WatchLogs(opts, "ContractInstantiation")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FactoryContractInstantiation)
				if err := _Factory.contract.UnpackLog(event, "ContractInstantiation", log); err != nil {
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

// MultiSigWalletABI is the input ABI used to generate the binding from.
const MultiSigWalletABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"owners\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"removeOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"revokeConfirmation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"isOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"confirmations\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"pending\",\"type\":\"bool\"},{\"name\":\"executed\",\"type\":\"bool\"}],\"name\":\"getTransactionCount\",\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"addOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"isConfirmed\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"getConfirmationCount\",\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"transactions\",\"outputs\":[{\"name\":\"destination\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"},{\"name\":\"data\",\"type\":\"bytes\"},{\"name\":\"executed\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwners\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"from\",\"type\":\"uint256\"},{\"name\":\"to\",\"type\":\"uint256\"},{\"name\":\"pending\",\"type\":\"bool\"},{\"name\":\"executed\",\"type\":\"bool\"}],\"name\":\"getTransactionIds\",\"outputs\":[{\"name\":\"_transactionIds\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"getConfirmations\",\"outputs\":[{\"name\":\"_confirmations\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transactionCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_required\",\"type\":\"uint256\"}],\"name\":\"changeRequirement\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"confirmTransaction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"destination\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"},{\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"submitTransaction\",\"outputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_OWNER_COUNT\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"required\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"replaceOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"executeTransaction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_owners\",\"type\":\"address[]\"},{\"name\":\"_required\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"Confirmation\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"Revocation\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"Submission\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"Execution\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"ExecutionFailure\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnerAddition\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnerRemoval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"required\",\"type\":\"uint256\"}],\"name\":\"RequirementChange\",\"type\":\"event\"}]"

// MultiSigWalletBin is the compiled bytecode used for deploying new contracts.
const MultiSigWalletBin = `0x60806040523480156200001157600080fd5b50604051620017463803806200174683398101604052805160208201519101805190919060009082603282118015906200004b5750818111155b80156200005757508015155b80156200006357508115155b15156200006f57600080fd5b600092505b845183101562000147576002600086858151811015156200009157fe5b6020908102909101810151600160a060020a031682528101919091526040016000205460ff16158015620000e757508483815181101515620000cf57fe5b90602001906020020151600160a060020a0316600014155b1515620000f357600080fd5b60016002600087868151811015156200010857fe5b602090810291909101810151600160a060020a03168252810191909152604001600020805460ff19169115159190911790556001929092019162000074565b84516200015c9060039060208801906200016e565b50505060049190915550620002029050565b828054828255906000526020600020908101928215620001c6579160200282015b82811115620001c65782518254600160a060020a031916600160a060020a039091161782556020909201916001909101906200018f565b50620001d4929150620001d8565b5090565b620001ff91905b80821115620001d4578054600160a060020a0319168155600101620001df565b90565b61153480620002126000396000f30060806040526004361061011c5763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663025e7c278114610167578063173825d91461019b57806320ea8d86146101bc5780632f54bf6e146101d45780633411c81c14610209578063547415251461022d5780637065cb481461025e578063784547a71461027f5780638b51d13f146102975780639ace38c2146102af578063a0e67e2b1461036a578063a8abe69a146103cf578063b5dc40c3146103f4578063b77bf6001461040c578063ba51a6df14610421578063c01a8c8414610439578063c642747414610451578063d74f8edd146104ba578063dc8452cd146104cf578063e20056e6146104e4578063ee22610b1461050b575b600034111561016557604080513481529051600160a060020a033316917fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c919081900360200190a25b005b34801561017357600080fd5b5061017f600435610523565b60408051600160a060020a039092168252519081900360200190f35b3480156101a757600080fd5b50610165600160a060020a036004351661054b565b3480156101c857600080fd5b506101656004356106d6565b3480156101e057600080fd5b506101f5600160a060020a03600435166107ac565b604080519115158252519081900360200190f35b34801561021557600080fd5b506101f5600435600160a060020a03602435166107c1565b34801561023957600080fd5b5061024c600435151560243515156107e1565b60408051918252519081900360200190f35b34801561026a57600080fd5b50610165600160a060020a036004351661084d565b34801561028b57600080fd5b506101f5600435610986565b3480156102a357600080fd5b5061024c600435610a0a565b3480156102bb57600080fd5b506102c7600435610a79565b6040518085600160a060020a0316600160a060020a031681526020018481526020018060200183151515158152602001828103825284818151815260200191508051906020019080838360005b8381101561032c578181015183820152602001610314565b50505050905090810190601f1680156103595780820380516001836020036101000a031916815260200191505b509550505050505060405180910390f35b34801561037657600080fd5b5061037f610b37565b60408051602080825283518183015283519192839290830191858101910280838360005b838110156103bb5781810151838201526020016103a3565b505050509050019250505060405180910390f35b3480156103db57600080fd5b5061037f60043560243560443515156064351515610b9a565b34801561040057600080fd5b5061037f600435610cd3565b34801561041857600080fd5b5061024c610e4c565b34801561042d57600080fd5b50610165600435610e52565b34801561044557600080fd5b50610165600435610ee5565b34801561045d57600080fd5b50604080516020600460443581810135601f810184900484028501840190955284845261024c948235600160a060020a0316946024803595369594606494920191908190840183828082843750949750610fcc9650505050505050565b3480156104c657600080fd5b5061024c610feb565b3480156104db57600080fd5b5061024c610ff0565b3480156104f057600080fd5b50610165600160a060020a0360043581169060243516610ff6565b34801561051757600080fd5b50610165600435611194565b600380548290811061053157fe5b600091825260209091200154600160a060020a0316905081565b600030600160a060020a031633600160a060020a031614151561056d57600080fd5b600160a060020a038216600090815260026020526040902054829060ff16151561059657600080fd5b600160a060020a0383166000908152600260205260408120805460ff1916905591505b600354600019018210156106715782600160a060020a03166003838154811015156105e057fe5b600091825260209091200154600160a060020a031614156106665760038054600019810190811061060d57fe5b60009182526020909120015460038054600160a060020a03909216918490811061063357fe5b9060005260206000200160006101000a815481600160a060020a030219169083600160a060020a03160217905550610671565b6001909101906105b9565b6003805460001901906106849082611447565b50600354600454111561069d5760035461069d90610e52565b604051600160a060020a038416907f8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b9090600090a2505050565b33600160a060020a03811660009081526002602052604090205460ff1615156106fe57600080fd5b600082815260016020908152604080832033600160a060020a038116855292529091205483919060ff16151561073357600080fd5b600084815260208190526040902060030154849060ff161561075457600080fd5b6000858152600160209081526040808320600160a060020a0333168085529252808320805460ff191690555187927ff6a317157440607f36269043eb55f1287a5a19ba2216afeab88cd46cbcfb88e991a35050505050565b60026020526000908152604090205460ff1681565b600160209081526000928352604080842090915290825290205460ff1681565b6000805b6005548110156108465783801561080e575060008181526020819052604090206003015460ff16155b806108325750828015610832575060008181526020819052604090206003015460ff165b1561083e576001820191505b6001016107e5565b5092915050565b30600160a060020a031633600160a060020a031614151561086d57600080fd5b600160a060020a038116600090815260026020526040902054819060ff161561089557600080fd5b81600160a060020a03811615156108ab57600080fd5b600380549050600101600454603282111580156108c85750818111155b80156108d357508015155b80156108de57508115155b15156108e957600080fd5b600160a060020a038516600081815260026020526040808220805460ff1916600190811790915560038054918201815583527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b01805473ffffffffffffffffffffffffffffffffffffffff191684179055517ff39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d9190a25050505050565b600080805b600354811015610a0357600084815260016020526040812060038054919291849081106109b457fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff16156109e8576001820191505b6004548214156109fb5760019250610a03565b60010161098b565b5050919050565b6000805b600354811015610a735760008381526001602052604081206003805491929184908110610a3757fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff1615610a6b576001820191505b600101610a0e565b50919050565b6000602081815291815260409081902080546001808301546002808501805487516101009582161595909502600019011691909104601f8101889004880284018801909652858352600160a060020a0390931695909491929190830182828015610b245780601f10610af957610100808354040283529160200191610b24565b820191906000526020600020905b815481529060010190602001808311610b0757829003601f168201915b5050506003909301549192505060ff1684565b60606003805480602002602001604051908101604052809291908181526020018280548015610b8f57602002820191906000526020600020905b8154600160a060020a03168152600190910190602001808311610b71575b505050505090505b90565b606080600080600554604051908082528060200260200182016040528015610bcc578160200160208202803883390190505b50925060009150600090505b600554811015610c5357858015610c01575060008181526020819052604090206003015460ff16155b80610c255750848015610c25575060008181526020819052604090206003015460ff165b15610c4b57808383815181101515610c3957fe5b60209081029091010152600191909101905b600101610bd8565b878703604051908082528060200260200182016040528015610c7f578160200160208202803883390190505b5093508790505b86811015610cc8578281815181101515610c9c57fe5b9060200190602002015184898303815181101515610cb657fe5b60209081029091010152600101610c86565b505050949350505050565b606080600080600380549050604051908082528060200260200182016040528015610d08578160200160208202803883390190505b50925060009150600090505b600354811015610dc55760008581526001602052604081206003805491929184908110610d3d57fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff1615610dbd576003805482908110610d7857fe5b6000918252602090912001548351600160a060020a0390911690849084908110610d9e57fe5b600160a060020a03909216602092830290910190910152600191909101905b600101610d14565b81604051908082528060200260200182016040528015610def578160200160208202803883390190505b509350600090505b81811015610e44578281815181101515610e0d57fe5b906020019060200201518482815181101515610e2557fe5b600160a060020a03909216602092830290910190910152600101610df7565b505050919050565b60055481565b30600160a060020a031633600160a060020a0316141515610e7257600080fd5b6003548160328211801590610e875750818111155b8015610e9257508015155b8015610e9d57508115155b1515610ea857600080fd5b60048390556040805184815290517fa3f1ee9126a074d9326c682f561767f710e927faa811f7a99829d49dc421797a9181900360200190a1505050565b33600160a060020a03811660009081526002602052604090205460ff161515610f0d57600080fd5b6000828152602081905260409020548290600160a060020a03161515610f3257600080fd5b600083815260016020908152604080832033600160a060020a038116855292529091205484919060ff1615610f6657600080fd5b6000858152600160208181526040808420600160a060020a0333168086529252808420805460ff1916909317909255905187927f4a504a94899432a9846e1aa406dceb1bcfd538bb839071d49d1e5e23f5be30ef91a3610fc585611194565b5050505050565b6000610fd9848484611357565b9050610fe481610ee5565b9392505050565b603281565b60045481565b600030600160a060020a031633600160a060020a031614151561101857600080fd5b600160a060020a038316600090815260026020526040902054839060ff16151561104157600080fd5b600160a060020a038316600090815260026020526040902054839060ff161561106957600080fd5b600092505b6003548310156110fa5784600160a060020a031660038481548110151561109157fe5b600091825260209091200154600160a060020a031614156110ef57836003848154811015156110bc57fe5b9060005260206000200160006101000a815481600160a060020a030219169083600160a060020a031602179055506110fa565b60019092019161106e565b600160a060020a03808616600081815260026020526040808220805460ff1990811690915593881682528082208054909416600117909355915190917f8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b9091a2604051600160a060020a038516907ff39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d90600090a25050505050565b33600160a060020a03811660009081526002602052604081205490919060ff1615156111bf57600080fd5b600083815260016020908152604080832033600160a060020a038116855292529091205484919060ff1615156111f457600080fd5b600085815260208190526040902060030154859060ff161561121557600080fd5b61121e86610986565b1561134f576000868152602081905260409081902060038101805460ff19166001908117909155815481830154935160028085018054959b50600160a060020a03909316959492939192839285926000199183161561010002919091019091160480156112cc5780601f106112a1576101008083540402835291602001916112cc565b820191906000526020600020905b8154815290600101906020018083116112af57829003601f168201915b505091505060006040518083038185875af192505050156113175760405186907f33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed7590600090a261134f565b60405186907f526441bb6c1aba3c9a4a6ca1d6545da9c2333c8c48343ef398eb858d72b7923690600090a260038501805460ff191690555b505050505050565b600083600160a060020a038116151561136f57600080fd5b60055460408051608081018252600160a060020a0388811682526020808301898152838501898152600060608601819052878152808452959095208451815473ffffffffffffffffffffffffffffffffffffffff1916941693909317835551600183015592518051949650919390926113ef926002850192910190611470565b50606091909101516003909101805460ff191691151591909117905560058054600101905560405182907fc0ba8fe4b176c1714197d43b9cc6bcf797a4a7461c5fe8d0ef6e184ae7601e5190600090a2509392505050565b81548183558181111561146b5760008381526020902061146b9181019083016114ee565b505050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106114b157805160ff19168380011785556114de565b828001600101855582156114de579182015b828111156114de5782518255916020019190600101906114c3565b506114ea9291506114ee565b5090565b610b9791905b808211156114ea57600081556001016114f45600a165627a7a72305820881814adbbbf208b7a25788d1470f331d9add4f0d76cac00167596f6a82a65c70029`

// DeployMultiSigWallet deploys a new Ethereum contract, binding an instance of MultiSigWallet to it.
func DeployMultiSigWallet(auth *bind.TransactOpts, backend bind.ContractBackend, _owners []common.Address, _required *big.Int) (common.Address, *types.Transaction, *MultiSigWallet, error) {
	parsed, err := abi.JSON(strings.NewReader(MultiSigWalletABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MultiSigWalletBin), backend, _owners, _required)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MultiSigWallet{MultiSigWalletCaller: MultiSigWalletCaller{contract: contract}, MultiSigWalletTransactor: MultiSigWalletTransactor{contract: contract}, MultiSigWalletFilterer: MultiSigWalletFilterer{contract: contract}}, nil
}

// MultiSigWallet is an auto generated Go binding around an Ethereum contract.
type MultiSigWallet struct {
	MultiSigWalletCaller     // Read-only binding to the contract
	MultiSigWalletTransactor // Write-only binding to the contract
	MultiSigWalletFilterer   // Log filterer for contract events
}

// MultiSigWalletCaller is an auto generated read-only Go binding around an Ethereum contract.
type MultiSigWalletCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MultiSigWalletTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MultiSigWalletFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MultiSigWalletSession struct {
	Contract     *MultiSigWallet   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MultiSigWalletCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MultiSigWalletCallerSession struct {
	Contract *MultiSigWalletCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// MultiSigWalletTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MultiSigWalletTransactorSession struct {
	Contract     *MultiSigWalletTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// MultiSigWalletRaw is an auto generated low-level Go binding around an Ethereum contract.
type MultiSigWalletRaw struct {
	Contract *MultiSigWallet // Generic contract binding to access the raw methods on
}

// MultiSigWalletCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MultiSigWalletCallerRaw struct {
	Contract *MultiSigWalletCaller // Generic read-only contract binding to access the raw methods on
}

// MultiSigWalletTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MultiSigWalletTransactorRaw struct {
	Contract *MultiSigWalletTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMultiSigWallet creates a new instance of MultiSigWallet, bound to a specific deployed contract.
func NewMultiSigWallet(address common.Address, backend bind.ContractBackend) (*MultiSigWallet, error) {
	contract, err := bindMultiSigWallet(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MultiSigWallet{MultiSigWalletCaller: MultiSigWalletCaller{contract: contract}, MultiSigWalletTransactor: MultiSigWalletTransactor{contract: contract}, MultiSigWalletFilterer: MultiSigWalletFilterer{contract: contract}}, nil
}

// NewMultiSigWalletCaller creates a new read-only instance of MultiSigWallet, bound to a specific deployed contract.
func NewMultiSigWalletCaller(address common.Address, caller bind.ContractCaller) (*MultiSigWalletCaller, error) {
	contract, err := bindMultiSigWallet(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletCaller{contract: contract}, nil
}

// NewMultiSigWalletTransactor creates a new write-only instance of MultiSigWallet, bound to a specific deployed contract.
func NewMultiSigWalletTransactor(address common.Address, transactor bind.ContractTransactor) (*MultiSigWalletTransactor, error) {
	contract, err := bindMultiSigWallet(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletTransactor{contract: contract}, nil
}

// NewMultiSigWalletFilterer creates a new log filterer instance of MultiSigWallet, bound to a specific deployed contract.
func NewMultiSigWalletFilterer(address common.Address, filterer bind.ContractFilterer) (*MultiSigWalletFilterer, error) {
	contract, err := bindMultiSigWallet(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletFilterer{contract: contract}, nil
}

// bindMultiSigWallet binds a generic wrapper to an already deployed contract.
func bindMultiSigWallet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MultiSigWalletABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiSigWallet *MultiSigWalletRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MultiSigWallet.Contract.MultiSigWalletCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiSigWallet *MultiSigWalletRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.MultiSigWalletTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiSigWallet *MultiSigWalletRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.MultiSigWalletTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiSigWallet *MultiSigWalletCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MultiSigWallet.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiSigWallet *MultiSigWalletTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiSigWallet *MultiSigWalletTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.contract.Transact(opts, method, params...)
}

// MAXOWNERCOUNT is a free data retrieval call binding the contract method 0xd74f8edd.
//
// Solidity: function MAX_OWNER_COUNT() constant returns(uint256)
func (_MultiSigWallet *MultiSigWalletCaller) MAXOWNERCOUNT(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "MAX_OWNER_COUNT")
	return *ret0, err
}

// MAXOWNERCOUNT is a free data retrieval call binding the contract method 0xd74f8edd.
//
// Solidity: function MAX_OWNER_COUNT() constant returns(uint256)
func (_MultiSigWallet *MultiSigWalletSession) MAXOWNERCOUNT() (*big.Int, error) {
	return _MultiSigWallet.Contract.MAXOWNERCOUNT(&_MultiSigWallet.CallOpts)
}

// MAXOWNERCOUNT is a free data retrieval call binding the contract method 0xd74f8edd.
//
// Solidity: function MAX_OWNER_COUNT() constant returns(uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) MAXOWNERCOUNT() (*big.Int, error) {
	return _MultiSigWallet.Contract.MAXOWNERCOUNT(&_MultiSigWallet.CallOpts)
}

// Confirmations is a free data retrieval call binding the contract method 0x3411c81c.
//
// Solidity: function confirmations( uint256,  address) constant returns(bool)
func (_MultiSigWallet *MultiSigWalletCaller) Confirmations(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "confirmations", arg0, arg1)
	return *ret0, err
}

// Confirmations is a free data retrieval call binding the contract method 0x3411c81c.
//
// Solidity: function confirmations( uint256,  address) constant returns(bool)
func (_MultiSigWallet *MultiSigWalletSession) Confirmations(arg0 *big.Int, arg1 common.Address) (bool, error) {
	return _MultiSigWallet.Contract.Confirmations(&_MultiSigWallet.CallOpts, arg0, arg1)
}

// Confirmations is a free data retrieval call binding the contract method 0x3411c81c.
//
// Solidity: function confirmations( uint256,  address) constant returns(bool)
func (_MultiSigWallet *MultiSigWalletCallerSession) Confirmations(arg0 *big.Int, arg1 common.Address) (bool, error) {
	return _MultiSigWallet.Contract.Confirmations(&_MultiSigWallet.CallOpts, arg0, arg1)
}

// GetConfirmationCount is a free data retrieval call binding the contract method 0x8b51d13f.
//
// Solidity: function getConfirmationCount(transactionId uint256) constant returns(count uint256)
func (_MultiSigWallet *MultiSigWalletCaller) GetConfirmationCount(opts *bind.CallOpts, transactionId *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "getConfirmationCount", transactionId)
	return *ret0, err
}

// GetConfirmationCount is a free data retrieval call binding the contract method 0x8b51d13f.
//
// Solidity: function getConfirmationCount(transactionId uint256) constant returns(count uint256)
func (_MultiSigWallet *MultiSigWalletSession) GetConfirmationCount(transactionId *big.Int) (*big.Int, error) {
	return _MultiSigWallet.Contract.GetConfirmationCount(&_MultiSigWallet.CallOpts, transactionId)
}

// GetConfirmationCount is a free data retrieval call binding the contract method 0x8b51d13f.
//
// Solidity: function getConfirmationCount(transactionId uint256) constant returns(count uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) GetConfirmationCount(transactionId *big.Int) (*big.Int, error) {
	return _MultiSigWallet.Contract.GetConfirmationCount(&_MultiSigWallet.CallOpts, transactionId)
}

// GetConfirmations is a free data retrieval call binding the contract method 0xb5dc40c3.
//
// Solidity: function getConfirmations(transactionId uint256) constant returns(_confirmations address[])
func (_MultiSigWallet *MultiSigWalletCaller) GetConfirmations(opts *bind.CallOpts, transactionId *big.Int) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "getConfirmations", transactionId)
	return *ret0, err
}

// GetConfirmations is a free data retrieval call binding the contract method 0xb5dc40c3.
//
// Solidity: function getConfirmations(transactionId uint256) constant returns(_confirmations address[])
func (_MultiSigWallet *MultiSigWalletSession) GetConfirmations(transactionId *big.Int) ([]common.Address, error) {
	return _MultiSigWallet.Contract.GetConfirmations(&_MultiSigWallet.CallOpts, transactionId)
}

// GetConfirmations is a free data retrieval call binding the contract method 0xb5dc40c3.
//
// Solidity: function getConfirmations(transactionId uint256) constant returns(_confirmations address[])
func (_MultiSigWallet *MultiSigWalletCallerSession) GetConfirmations(transactionId *big.Int) ([]common.Address, error) {
	return _MultiSigWallet.Contract.GetConfirmations(&_MultiSigWallet.CallOpts, transactionId)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() constant returns(address[])
func (_MultiSigWallet *MultiSigWalletCaller) GetOwners(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "getOwners")
	return *ret0, err
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() constant returns(address[])
func (_MultiSigWallet *MultiSigWalletSession) GetOwners() ([]common.Address, error) {
	return _MultiSigWallet.Contract.GetOwners(&_MultiSigWallet.CallOpts)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() constant returns(address[])
func (_MultiSigWallet *MultiSigWalletCallerSession) GetOwners() ([]common.Address, error) {
	return _MultiSigWallet.Contract.GetOwners(&_MultiSigWallet.CallOpts)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x54741525.
//
// Solidity: function getTransactionCount(pending bool, executed bool) constant returns(count uint256)
func (_MultiSigWallet *MultiSigWalletCaller) GetTransactionCount(opts *bind.CallOpts, pending bool, executed bool) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "getTransactionCount", pending, executed)
	return *ret0, err
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x54741525.
//
// Solidity: function getTransactionCount(pending bool, executed bool) constant returns(count uint256)
func (_MultiSigWallet *MultiSigWalletSession) GetTransactionCount(pending bool, executed bool) (*big.Int, error) {
	return _MultiSigWallet.Contract.GetTransactionCount(&_MultiSigWallet.CallOpts, pending, executed)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x54741525.
//
// Solidity: function getTransactionCount(pending bool, executed bool) constant returns(count uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) GetTransactionCount(pending bool, executed bool) (*big.Int, error) {
	return _MultiSigWallet.Contract.GetTransactionCount(&_MultiSigWallet.CallOpts, pending, executed)
}

// GetTransactionIds is a free data retrieval call binding the contract method 0xa8abe69a.
//
// Solidity: function getTransactionIds(from uint256, to uint256, pending bool, executed bool) constant returns(_transactionIds uint256[])
func (_MultiSigWallet *MultiSigWalletCaller) GetTransactionIds(opts *bind.CallOpts, from *big.Int, to *big.Int, pending bool, executed bool) ([]*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "getTransactionIds", from, to, pending, executed)
	return *ret0, err
}

// GetTransactionIds is a free data retrieval call binding the contract method 0xa8abe69a.
//
// Solidity: function getTransactionIds(from uint256, to uint256, pending bool, executed bool) constant returns(_transactionIds uint256[])
func (_MultiSigWallet *MultiSigWalletSession) GetTransactionIds(from *big.Int, to *big.Int, pending bool, executed bool) ([]*big.Int, error) {
	return _MultiSigWallet.Contract.GetTransactionIds(&_MultiSigWallet.CallOpts, from, to, pending, executed)
}

// GetTransactionIds is a free data retrieval call binding the contract method 0xa8abe69a.
//
// Solidity: function getTransactionIds(from uint256, to uint256, pending bool, executed bool) constant returns(_transactionIds uint256[])
func (_MultiSigWallet *MultiSigWalletCallerSession) GetTransactionIds(from *big.Int, to *big.Int, pending bool, executed bool) ([]*big.Int, error) {
	return _MultiSigWallet.Contract.GetTransactionIds(&_MultiSigWallet.CallOpts, from, to, pending, executed)
}

// IsConfirmed is a free data retrieval call binding the contract method 0x784547a7.
//
// Solidity: function isConfirmed(transactionId uint256) constant returns(bool)
func (_MultiSigWallet *MultiSigWalletCaller) IsConfirmed(opts *bind.CallOpts, transactionId *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "isConfirmed", transactionId)
	return *ret0, err
}

// IsConfirmed is a free data retrieval call binding the contract method 0x784547a7.
//
// Solidity: function isConfirmed(transactionId uint256) constant returns(bool)
func (_MultiSigWallet *MultiSigWalletSession) IsConfirmed(transactionId *big.Int) (bool, error) {
	return _MultiSigWallet.Contract.IsConfirmed(&_MultiSigWallet.CallOpts, transactionId)
}

// IsConfirmed is a free data retrieval call binding the contract method 0x784547a7.
//
// Solidity: function isConfirmed(transactionId uint256) constant returns(bool)
func (_MultiSigWallet *MultiSigWalletCallerSession) IsConfirmed(transactionId *big.Int) (bool, error) {
	return _MultiSigWallet.Contract.IsConfirmed(&_MultiSigWallet.CallOpts, transactionId)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner( address) constant returns(bool)
func (_MultiSigWallet *MultiSigWalletCaller) IsOwner(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "isOwner", arg0)
	return *ret0, err
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner( address) constant returns(bool)
func (_MultiSigWallet *MultiSigWalletSession) IsOwner(arg0 common.Address) (bool, error) {
	return _MultiSigWallet.Contract.IsOwner(&_MultiSigWallet.CallOpts, arg0)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner( address) constant returns(bool)
func (_MultiSigWallet *MultiSigWalletCallerSession) IsOwner(arg0 common.Address) (bool, error) {
	return _MultiSigWallet.Contract.IsOwner(&_MultiSigWallet.CallOpts, arg0)
}

// Owners is a free data retrieval call binding the contract method 0x025e7c27.
//
// Solidity: function owners( uint256) constant returns(address)
func (_MultiSigWallet *MultiSigWalletCaller) Owners(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "owners", arg0)
	return *ret0, err
}

// Owners is a free data retrieval call binding the contract method 0x025e7c27.
//
// Solidity: function owners( uint256) constant returns(address)
func (_MultiSigWallet *MultiSigWalletSession) Owners(arg0 *big.Int) (common.Address, error) {
	return _MultiSigWallet.Contract.Owners(&_MultiSigWallet.CallOpts, arg0)
}

// Owners is a free data retrieval call binding the contract method 0x025e7c27.
//
// Solidity: function owners( uint256) constant returns(address)
func (_MultiSigWallet *MultiSigWalletCallerSession) Owners(arg0 *big.Int) (common.Address, error) {
	return _MultiSigWallet.Contract.Owners(&_MultiSigWallet.CallOpts, arg0)
}

// Required is a free data retrieval call binding the contract method 0xdc8452cd.
//
// Solidity: function required() constant returns(uint256)
func (_MultiSigWallet *MultiSigWalletCaller) Required(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "required")
	return *ret0, err
}

// Required is a free data retrieval call binding the contract method 0xdc8452cd.
//
// Solidity: function required() constant returns(uint256)
func (_MultiSigWallet *MultiSigWalletSession) Required() (*big.Int, error) {
	return _MultiSigWallet.Contract.Required(&_MultiSigWallet.CallOpts)
}

// Required is a free data retrieval call binding the contract method 0xdc8452cd.
//
// Solidity: function required() constant returns(uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) Required() (*big.Int, error) {
	return _MultiSigWallet.Contract.Required(&_MultiSigWallet.CallOpts)
}

// TransactionCount is a free data retrieval call binding the contract method 0xb77bf600.
//
// Solidity: function transactionCount() constant returns(uint256)
func (_MultiSigWallet *MultiSigWalletCaller) TransactionCount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWallet.contract.Call(opts, out, "transactionCount")
	return *ret0, err
}

// TransactionCount is a free data retrieval call binding the contract method 0xb77bf600.
//
// Solidity: function transactionCount() constant returns(uint256)
func (_MultiSigWallet *MultiSigWalletSession) TransactionCount() (*big.Int, error) {
	return _MultiSigWallet.Contract.TransactionCount(&_MultiSigWallet.CallOpts)
}

// TransactionCount is a free data retrieval call binding the contract method 0xb77bf600.
//
// Solidity: function transactionCount() constant returns(uint256)
func (_MultiSigWallet *MultiSigWalletCallerSession) TransactionCount() (*big.Int, error) {
	return _MultiSigWallet.Contract.TransactionCount(&_MultiSigWallet.CallOpts)
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions( uint256) constant returns(destination address, value uint256, data bytes, executed bool)
func (_MultiSigWallet *MultiSigWalletCaller) Transactions(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Destination common.Address
	Value       *big.Int
	Data        []byte
	Executed    bool
}, error) {
	ret := new(struct {
		Destination common.Address
		Value       *big.Int
		Data        []byte
		Executed    bool
	})
	out := ret
	err := _MultiSigWallet.contract.Call(opts, out, "transactions", arg0)
	return *ret, err
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions( uint256) constant returns(destination address, value uint256, data bytes, executed bool)
func (_MultiSigWallet *MultiSigWalletSession) Transactions(arg0 *big.Int) (struct {
	Destination common.Address
	Value       *big.Int
	Data        []byte
	Executed    bool
}, error) {
	return _MultiSigWallet.Contract.Transactions(&_MultiSigWallet.CallOpts, arg0)
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions( uint256) constant returns(destination address, value uint256, data bytes, executed bool)
func (_MultiSigWallet *MultiSigWalletCallerSession) Transactions(arg0 *big.Int) (struct {
	Destination common.Address
	Value       *big.Int
	Data        []byte
	Executed    bool
}, error) {
	return _MultiSigWallet.Contract.Transactions(&_MultiSigWallet.CallOpts, arg0)
}

// AddOwner is a paid mutator transaction binding the contract method 0x7065cb48.
//
// Solidity: function addOwner(owner address) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) AddOwner(opts *bind.TransactOpts, owner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "addOwner", owner)
}

// AddOwner is a paid mutator transaction binding the contract method 0x7065cb48.
//
// Solidity: function addOwner(owner address) returns()
func (_MultiSigWallet *MultiSigWalletSession) AddOwner(owner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.AddOwner(&_MultiSigWallet.TransactOpts, owner)
}

// AddOwner is a paid mutator transaction binding the contract method 0x7065cb48.
//
// Solidity: function addOwner(owner address) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) AddOwner(owner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.AddOwner(&_MultiSigWallet.TransactOpts, owner)
}

// ChangeRequirement is a paid mutator transaction binding the contract method 0xba51a6df.
//
// Solidity: function changeRequirement(_required uint256) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) ChangeRequirement(opts *bind.TransactOpts, _required *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "changeRequirement", _required)
}

// ChangeRequirement is a paid mutator transaction binding the contract method 0xba51a6df.
//
// Solidity: function changeRequirement(_required uint256) returns()
func (_MultiSigWallet *MultiSigWalletSession) ChangeRequirement(_required *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ChangeRequirement(&_MultiSigWallet.TransactOpts, _required)
}

// ChangeRequirement is a paid mutator transaction binding the contract method 0xba51a6df.
//
// Solidity: function changeRequirement(_required uint256) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) ChangeRequirement(_required *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ChangeRequirement(&_MultiSigWallet.TransactOpts, _required)
}

// ConfirmTransaction is a paid mutator transaction binding the contract method 0xc01a8c84.
//
// Solidity: function confirmTransaction(transactionId uint256) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) ConfirmTransaction(opts *bind.TransactOpts, transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "confirmTransaction", transactionId)
}

// ConfirmTransaction is a paid mutator transaction binding the contract method 0xc01a8c84.
//
// Solidity: function confirmTransaction(transactionId uint256) returns()
func (_MultiSigWallet *MultiSigWalletSession) ConfirmTransaction(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ConfirmTransaction(&_MultiSigWallet.TransactOpts, transactionId)
}

// ConfirmTransaction is a paid mutator transaction binding the contract method 0xc01a8c84.
//
// Solidity: function confirmTransaction(transactionId uint256) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) ConfirmTransaction(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ConfirmTransaction(&_MultiSigWallet.TransactOpts, transactionId)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0xee22610b.
//
// Solidity: function executeTransaction(transactionId uint256) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) ExecuteTransaction(opts *bind.TransactOpts, transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "executeTransaction", transactionId)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0xee22610b.
//
// Solidity: function executeTransaction(transactionId uint256) returns()
func (_MultiSigWallet *MultiSigWalletSession) ExecuteTransaction(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ExecuteTransaction(&_MultiSigWallet.TransactOpts, transactionId)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0xee22610b.
//
// Solidity: function executeTransaction(transactionId uint256) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) ExecuteTransaction(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ExecuteTransaction(&_MultiSigWallet.TransactOpts, transactionId)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0x173825d9.
//
// Solidity: function removeOwner(owner address) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) RemoveOwner(opts *bind.TransactOpts, owner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "removeOwner", owner)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0x173825d9.
//
// Solidity: function removeOwner(owner address) returns()
func (_MultiSigWallet *MultiSigWalletSession) RemoveOwner(owner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.RemoveOwner(&_MultiSigWallet.TransactOpts, owner)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0x173825d9.
//
// Solidity: function removeOwner(owner address) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) RemoveOwner(owner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.RemoveOwner(&_MultiSigWallet.TransactOpts, owner)
}

// ReplaceOwner is a paid mutator transaction binding the contract method 0xe20056e6.
//
// Solidity: function replaceOwner(owner address, newOwner address) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) ReplaceOwner(opts *bind.TransactOpts, owner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "replaceOwner", owner, newOwner)
}

// ReplaceOwner is a paid mutator transaction binding the contract method 0xe20056e6.
//
// Solidity: function replaceOwner(owner address, newOwner address) returns()
func (_MultiSigWallet *MultiSigWalletSession) ReplaceOwner(owner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ReplaceOwner(&_MultiSigWallet.TransactOpts, owner, newOwner)
}

// ReplaceOwner is a paid mutator transaction binding the contract method 0xe20056e6.
//
// Solidity: function replaceOwner(owner address, newOwner address) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) ReplaceOwner(owner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.ReplaceOwner(&_MultiSigWallet.TransactOpts, owner, newOwner)
}

// RevokeConfirmation is a paid mutator transaction binding the contract method 0x20ea8d86.
//
// Solidity: function revokeConfirmation(transactionId uint256) returns()
func (_MultiSigWallet *MultiSigWalletTransactor) RevokeConfirmation(opts *bind.TransactOpts, transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "revokeConfirmation", transactionId)
}

// RevokeConfirmation is a paid mutator transaction binding the contract method 0x20ea8d86.
//
// Solidity: function revokeConfirmation(transactionId uint256) returns()
func (_MultiSigWallet *MultiSigWalletSession) RevokeConfirmation(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.RevokeConfirmation(&_MultiSigWallet.TransactOpts, transactionId)
}

// RevokeConfirmation is a paid mutator transaction binding the contract method 0x20ea8d86.
//
// Solidity: function revokeConfirmation(transactionId uint256) returns()
func (_MultiSigWallet *MultiSigWalletTransactorSession) RevokeConfirmation(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.RevokeConfirmation(&_MultiSigWallet.TransactOpts, transactionId)
}

// SubmitTransaction is a paid mutator transaction binding the contract method 0xc6427474.
//
// Solidity: function submitTransaction(destination address, value uint256, data bytes) returns(transactionId uint256)
func (_MultiSigWallet *MultiSigWalletTransactor) SubmitTransaction(opts *bind.TransactOpts, destination common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _MultiSigWallet.contract.Transact(opts, "submitTransaction", destination, value, data)
}

// SubmitTransaction is a paid mutator transaction binding the contract method 0xc6427474.
//
// Solidity: function submitTransaction(destination address, value uint256, data bytes) returns(transactionId uint256)
func (_MultiSigWallet *MultiSigWalletSession) SubmitTransaction(destination common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.SubmitTransaction(&_MultiSigWallet.TransactOpts, destination, value, data)
}

// SubmitTransaction is a paid mutator transaction binding the contract method 0xc6427474.
//
// Solidity: function submitTransaction(destination address, value uint256, data bytes) returns(transactionId uint256)
func (_MultiSigWallet *MultiSigWalletTransactorSession) SubmitTransaction(destination common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _MultiSigWallet.Contract.SubmitTransaction(&_MultiSigWallet.TransactOpts, destination, value, data)
}

// MultiSigWalletConfirmationIterator is returned from FilterConfirmation and is used to iterate over the raw logs and unpacked data for Confirmation events raised by the MultiSigWallet contract.
type MultiSigWalletConfirmationIterator struct {
	Event *MultiSigWalletConfirmation // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletConfirmationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletConfirmation)
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
		it.Event = new(MultiSigWalletConfirmation)
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
func (it *MultiSigWalletConfirmationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletConfirmationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletConfirmation represents a Confirmation event raised by the MultiSigWallet contract.
type MultiSigWalletConfirmation struct {
	Sender        common.Address
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterConfirmation is a free log retrieval operation binding the contract event 0x4a504a94899432a9846e1aa406dceb1bcfd538bb839071d49d1e5e23f5be30ef.
//
// Solidity: event Confirmation(sender indexed address, transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterConfirmation(opts *bind.FilterOpts, sender []common.Address, transactionId []*big.Int) (*MultiSigWalletConfirmationIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "Confirmation", senderRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletConfirmationIterator{contract: _MultiSigWallet.contract, event: "Confirmation", logs: logs, sub: sub}, nil
}

// WatchConfirmation is a free log subscription operation binding the contract event 0x4a504a94899432a9846e1aa406dceb1bcfd538bb839071d49d1e5e23f5be30ef.
//
// Solidity: event Confirmation(sender indexed address, transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchConfirmation(opts *bind.WatchOpts, sink chan<- *MultiSigWalletConfirmation, sender []common.Address, transactionId []*big.Int) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "Confirmation", senderRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletConfirmation)
				if err := _MultiSigWallet.contract.UnpackLog(event, "Confirmation", log); err != nil {
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

// MultiSigWalletDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the MultiSigWallet contract.
type MultiSigWalletDepositIterator struct {
	Event *MultiSigWalletDeposit // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletDeposit)
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
		it.Event = new(MultiSigWalletDeposit)
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
func (it *MultiSigWalletDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletDeposit represents a Deposit event raised by the MultiSigWallet contract.
type MultiSigWalletDeposit struct {
	Sender common.Address
	Value  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(sender indexed address, value uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterDeposit(opts *bind.FilterOpts, sender []common.Address) (*MultiSigWalletDepositIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "Deposit", senderRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletDepositIterator{contract: _MultiSigWallet.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(sender indexed address, value uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *MultiSigWalletDeposit, sender []common.Address) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "Deposit", senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletDeposit)
				if err := _MultiSigWallet.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// MultiSigWalletExecutionIterator is returned from FilterExecution and is used to iterate over the raw logs and unpacked data for Execution events raised by the MultiSigWallet contract.
type MultiSigWalletExecutionIterator struct {
	Event *MultiSigWalletExecution // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletExecutionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletExecution)
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
		it.Event = new(MultiSigWalletExecution)
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
func (it *MultiSigWalletExecutionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletExecutionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletExecution represents a Execution event raised by the MultiSigWallet contract.
type MultiSigWalletExecution struct {
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterExecution is a free log retrieval operation binding the contract event 0x33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed75.
//
// Solidity: event Execution(transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterExecution(opts *bind.FilterOpts, transactionId []*big.Int) (*MultiSigWalletExecutionIterator, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "Execution", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletExecutionIterator{contract: _MultiSigWallet.contract, event: "Execution", logs: logs, sub: sub}, nil
}

// WatchExecution is a free log subscription operation binding the contract event 0x33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed75.
//
// Solidity: event Execution(transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchExecution(opts *bind.WatchOpts, sink chan<- *MultiSigWalletExecution, transactionId []*big.Int) (event.Subscription, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "Execution", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletExecution)
				if err := _MultiSigWallet.contract.UnpackLog(event, "Execution", log); err != nil {
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

// MultiSigWalletExecutionFailureIterator is returned from FilterExecutionFailure and is used to iterate over the raw logs and unpacked data for ExecutionFailure events raised by the MultiSigWallet contract.
type MultiSigWalletExecutionFailureIterator struct {
	Event *MultiSigWalletExecutionFailure // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletExecutionFailureIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletExecutionFailure)
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
		it.Event = new(MultiSigWalletExecutionFailure)
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
func (it *MultiSigWalletExecutionFailureIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletExecutionFailureIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletExecutionFailure represents a ExecutionFailure event raised by the MultiSigWallet contract.
type MultiSigWalletExecutionFailure struct {
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterExecutionFailure is a free log retrieval operation binding the contract event 0x526441bb6c1aba3c9a4a6ca1d6545da9c2333c8c48343ef398eb858d72b79236.
//
// Solidity: event ExecutionFailure(transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterExecutionFailure(opts *bind.FilterOpts, transactionId []*big.Int) (*MultiSigWalletExecutionFailureIterator, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "ExecutionFailure", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletExecutionFailureIterator{contract: _MultiSigWallet.contract, event: "ExecutionFailure", logs: logs, sub: sub}, nil
}

// WatchExecutionFailure is a free log subscription operation binding the contract event 0x526441bb6c1aba3c9a4a6ca1d6545da9c2333c8c48343ef398eb858d72b79236.
//
// Solidity: event ExecutionFailure(transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchExecutionFailure(opts *bind.WatchOpts, sink chan<- *MultiSigWalletExecutionFailure, transactionId []*big.Int) (event.Subscription, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "ExecutionFailure", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletExecutionFailure)
				if err := _MultiSigWallet.contract.UnpackLog(event, "ExecutionFailure", log); err != nil {
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

// MultiSigWalletOwnerAdditionIterator is returned from FilterOwnerAddition and is used to iterate over the raw logs and unpacked data for OwnerAddition events raised by the MultiSigWallet contract.
type MultiSigWalletOwnerAdditionIterator struct {
	Event *MultiSigWalletOwnerAddition // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletOwnerAdditionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletOwnerAddition)
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
		it.Event = new(MultiSigWalletOwnerAddition)
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
func (it *MultiSigWalletOwnerAdditionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletOwnerAdditionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletOwnerAddition represents a OwnerAddition event raised by the MultiSigWallet contract.
type MultiSigWalletOwnerAddition struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOwnerAddition is a free log retrieval operation binding the contract event 0xf39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d.
//
// Solidity: event OwnerAddition(owner indexed address)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterOwnerAddition(opts *bind.FilterOpts, owner []common.Address) (*MultiSigWalletOwnerAdditionIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "OwnerAddition", ownerRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletOwnerAdditionIterator{contract: _MultiSigWallet.contract, event: "OwnerAddition", logs: logs, sub: sub}, nil
}

// WatchOwnerAddition is a free log subscription operation binding the contract event 0xf39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d.
//
// Solidity: event OwnerAddition(owner indexed address)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchOwnerAddition(opts *bind.WatchOpts, sink chan<- *MultiSigWalletOwnerAddition, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "OwnerAddition", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletOwnerAddition)
				if err := _MultiSigWallet.contract.UnpackLog(event, "OwnerAddition", log); err != nil {
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

// MultiSigWalletOwnerRemovalIterator is returned from FilterOwnerRemoval and is used to iterate over the raw logs and unpacked data for OwnerRemoval events raised by the MultiSigWallet contract.
type MultiSigWalletOwnerRemovalIterator struct {
	Event *MultiSigWalletOwnerRemoval // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletOwnerRemovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletOwnerRemoval)
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
		it.Event = new(MultiSigWalletOwnerRemoval)
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
func (it *MultiSigWalletOwnerRemovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletOwnerRemovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletOwnerRemoval represents a OwnerRemoval event raised by the MultiSigWallet contract.
type MultiSigWalletOwnerRemoval struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOwnerRemoval is a free log retrieval operation binding the contract event 0x8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b90.
//
// Solidity: event OwnerRemoval(owner indexed address)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterOwnerRemoval(opts *bind.FilterOpts, owner []common.Address) (*MultiSigWalletOwnerRemovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "OwnerRemoval", ownerRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletOwnerRemovalIterator{contract: _MultiSigWallet.contract, event: "OwnerRemoval", logs: logs, sub: sub}, nil
}

// WatchOwnerRemoval is a free log subscription operation binding the contract event 0x8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b90.
//
// Solidity: event OwnerRemoval(owner indexed address)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchOwnerRemoval(opts *bind.WatchOpts, sink chan<- *MultiSigWalletOwnerRemoval, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "OwnerRemoval", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletOwnerRemoval)
				if err := _MultiSigWallet.contract.UnpackLog(event, "OwnerRemoval", log); err != nil {
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

// MultiSigWalletRequirementChangeIterator is returned from FilterRequirementChange and is used to iterate over the raw logs and unpacked data for RequirementChange events raised by the MultiSigWallet contract.
type MultiSigWalletRequirementChangeIterator struct {
	Event *MultiSigWalletRequirementChange // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletRequirementChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletRequirementChange)
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
		it.Event = new(MultiSigWalletRequirementChange)
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
func (it *MultiSigWalletRequirementChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletRequirementChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletRequirementChange represents a RequirementChange event raised by the MultiSigWallet contract.
type MultiSigWalletRequirementChange struct {
	Required *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRequirementChange is a free log retrieval operation binding the contract event 0xa3f1ee9126a074d9326c682f561767f710e927faa811f7a99829d49dc421797a.
//
// Solidity: event RequirementChange(required uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterRequirementChange(opts *bind.FilterOpts) (*MultiSigWalletRequirementChangeIterator, error) {

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "RequirementChange")
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletRequirementChangeIterator{contract: _MultiSigWallet.contract, event: "RequirementChange", logs: logs, sub: sub}, nil
}

// WatchRequirementChange is a free log subscription operation binding the contract event 0xa3f1ee9126a074d9326c682f561767f710e927faa811f7a99829d49dc421797a.
//
// Solidity: event RequirementChange(required uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchRequirementChange(opts *bind.WatchOpts, sink chan<- *MultiSigWalletRequirementChange) (event.Subscription, error) {

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "RequirementChange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletRequirementChange)
				if err := _MultiSigWallet.contract.UnpackLog(event, "RequirementChange", log); err != nil {
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

// MultiSigWalletRevocationIterator is returned from FilterRevocation and is used to iterate over the raw logs and unpacked data for Revocation events raised by the MultiSigWallet contract.
type MultiSigWalletRevocationIterator struct {
	Event *MultiSigWalletRevocation // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletRevocationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletRevocation)
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
		it.Event = new(MultiSigWalletRevocation)
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
func (it *MultiSigWalletRevocationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletRevocationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletRevocation represents a Revocation event raised by the MultiSigWallet contract.
type MultiSigWalletRevocation struct {
	Sender        common.Address
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRevocation is a free log retrieval operation binding the contract event 0xf6a317157440607f36269043eb55f1287a5a19ba2216afeab88cd46cbcfb88e9.
//
// Solidity: event Revocation(sender indexed address, transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterRevocation(opts *bind.FilterOpts, sender []common.Address, transactionId []*big.Int) (*MultiSigWalletRevocationIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "Revocation", senderRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletRevocationIterator{contract: _MultiSigWallet.contract, event: "Revocation", logs: logs, sub: sub}, nil
}

// WatchRevocation is a free log subscription operation binding the contract event 0xf6a317157440607f36269043eb55f1287a5a19ba2216afeab88cd46cbcfb88e9.
//
// Solidity: event Revocation(sender indexed address, transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchRevocation(opts *bind.WatchOpts, sink chan<- *MultiSigWalletRevocation, sender []common.Address, transactionId []*big.Int) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "Revocation", senderRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletRevocation)
				if err := _MultiSigWallet.contract.UnpackLog(event, "Revocation", log); err != nil {
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

// MultiSigWalletSubmissionIterator is returned from FilterSubmission and is used to iterate over the raw logs and unpacked data for Submission events raised by the MultiSigWallet contract.
type MultiSigWalletSubmissionIterator struct {
	Event *MultiSigWalletSubmission // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletSubmissionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletSubmission)
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
		it.Event = new(MultiSigWalletSubmission)
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
func (it *MultiSigWalletSubmissionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletSubmissionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletSubmission represents a Submission event raised by the MultiSigWallet contract.
type MultiSigWalletSubmission struct {
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSubmission is a free log retrieval operation binding the contract event 0xc0ba8fe4b176c1714197d43b9cc6bcf797a4a7461c5fe8d0ef6e184ae7601e51.
//
// Solidity: event Submission(transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) FilterSubmission(opts *bind.FilterOpts, transactionId []*big.Int) (*MultiSigWalletSubmissionIterator, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.FilterLogs(opts, "Submission", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletSubmissionIterator{contract: _MultiSigWallet.contract, event: "Submission", logs: logs, sub: sub}, nil
}

// WatchSubmission is a free log subscription operation binding the contract event 0xc0ba8fe4b176c1714197d43b9cc6bcf797a4a7461c5fe8d0ef6e184ae7601e51.
//
// Solidity: event Submission(transactionId indexed uint256)
func (_MultiSigWallet *MultiSigWalletFilterer) WatchSubmission(opts *bind.WatchOpts, sink chan<- *MultiSigWalletSubmission, transactionId []*big.Int) (event.Subscription, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWallet.contract.WatchLogs(opts, "Submission", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletSubmission)
				if err := _MultiSigWallet.contract.UnpackLog(event, "Submission", log); err != nil {
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

// MultiSigWalletWithDailyLimitABI is the input ABI used to generate the binding from.
const MultiSigWalletWithDailyLimitABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"owners\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"removeOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"revokeConfirmation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"isOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"confirmations\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"calcMaxWithdraw\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"pending\",\"type\":\"bool\"},{\"name\":\"executed\",\"type\":\"bool\"}],\"name\":\"getTransactionCount\",\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"dailyLimit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastDay\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"addOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"isConfirmed\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"getConfirmationCount\",\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"transactions\",\"outputs\":[{\"name\":\"destination\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"},{\"name\":\"data\",\"type\":\"bytes\"},{\"name\":\"executed\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwners\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"from\",\"type\":\"uint256\"},{\"name\":\"to\",\"type\":\"uint256\"},{\"name\":\"pending\",\"type\":\"bool\"},{\"name\":\"executed\",\"type\":\"bool\"}],\"name\":\"getTransactionIds\",\"outputs\":[{\"name\":\"_transactionIds\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"getConfirmations\",\"outputs\":[{\"name\":\"_confirmations\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"transactionCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_required\",\"type\":\"uint256\"}],\"name\":\"changeRequirement\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"confirmTransaction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"destination\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"},{\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"submitTransaction\",\"outputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_dailyLimit\",\"type\":\"uint256\"}],\"name\":\"changeDailyLimit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_OWNER_COUNT\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"required\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"replaceOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"executeTransaction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"spentToday\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_owners\",\"type\":\"address[]\"},{\"name\":\"_required\",\"type\":\"uint256\"},{\"name\":\"_dailyLimit\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"dailyLimit\",\"type\":\"uint256\"}],\"name\":\"DailyLimitChange\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"Confirmation\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"Revocation\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"Submission\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"Execution\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"transactionId\",\"type\":\"uint256\"}],\"name\":\"ExecutionFailure\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnerAddition\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnerRemoval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"required\",\"type\":\"uint256\"}],\"name\":\"RequirementChange\",\"type\":\"event\"}]"

// MultiSigWalletWithDailyLimitBin is the compiled bytecode used for deploying new contracts.
const MultiSigWalletWithDailyLimitBin = `0x60806040523480156200001157600080fd5b506040516200194c3803806200194c833981016040908152815160208301519183015192018051909290839083906000908260328211801590620000555750818111155b80156200006157508015155b80156200006d57508115155b15156200007957600080fd5b600092505b845183101562000151576002600086858151811015156200009b57fe5b6020908102909101810151600160a060020a031682528101919091526040016000205460ff16158015620000f157508483815181101515620000d957fe5b90602001906020020151600160a060020a0316600014155b1515620000fd57600080fd5b60016002600087868151811015156200011257fe5b602090810291909101810151600160a060020a03168252810191909152604001600020805460ff1916911515919091179055600192909201916200007e565b8451620001669060039060208801906200017d565b505050600491909155505060065550620002119050565b828054828255906000526020600020908101928215620001d5579160200282015b82811115620001d55782518254600160a060020a031916600160a060020a039091161782556020909201916001909101906200019e565b50620001e3929150620001e7565b5090565b6200020e91905b80821115620001e3578054600160a060020a0319168155600101620001ee565b90565b61172b80620002216000396000f3006080604052600436106101535763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663025e7c27811461019e578063173825d9146101d257806320ea8d86146101f35780632f54bf6e1461020b5780633411c81c146102405780634bc9fdc214610264578063547415251461028b57806367eeba0c146102aa5780636b0c932d146102bf5780637065cb48146102d4578063784547a7146102f55780638b51d13f1461030d5780639ace38c214610325578063a0e67e2b146103e0578063a8abe69a14610445578063b5dc40c31461046a578063b77bf60014610482578063ba51a6df14610497578063c01a8c84146104af578063c6427474146104c7578063cea0862114610530578063d74f8edd14610548578063dc8452cd1461055d578063e20056e614610572578063ee22610b14610599578063f059cf2b146105b1575b600034111561019c57604080513481529051600160a060020a033316917fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c919081900360200190a25b005b3480156101aa57600080fd5b506101b66004356105c6565b60408051600160a060020a039092168252519081900360200190f35b3480156101de57600080fd5b5061019c600160a060020a03600435166105ee565b3480156101ff57600080fd5b5061019c600435610779565b34801561021757600080fd5b5061022c600160a060020a036004351661084f565b604080519115158252519081900360200190f35b34801561024c57600080fd5b5061022c600435600160a060020a0360243516610864565b34801561027057600080fd5b50610279610884565b60408051918252519081900360200190f35b34801561029757600080fd5b50610279600435151560243515156108be565b3480156102b657600080fd5b5061027961092a565b3480156102cb57600080fd5b50610279610930565b3480156102e057600080fd5b5061019c600160a060020a0360043516610936565b34801561030157600080fd5b5061022c600435610a6f565b34801561031957600080fd5b50610279600435610af3565b34801561033157600080fd5b5061033d600435610b62565b6040518085600160a060020a0316600160a060020a031681526020018481526020018060200183151515158152602001828103825284818151815260200191508051906020019080838360005b838110156103a257818101518382015260200161038a565b50505050905090810190601f1680156103cf5780820380516001836020036101000a031916815260200191505b509550505050505060405180910390f35b3480156103ec57600080fd5b506103f5610c20565b60408051602080825283518183015283519192839290830191858101910280838360005b83811015610431578181015183820152602001610419565b505050509050019250505060405180910390f35b34801561045157600080fd5b506103f560043560243560443515156064351515610c82565b34801561047657600080fd5b506103f5600435610dbb565b34801561048e57600080fd5b50610279610f34565b3480156104a357600080fd5b5061019c600435610f3a565b3480156104bb57600080fd5b5061019c600435610fcd565b3480156104d357600080fd5b50604080516020600460443581810135601f8101849004840285018401909552848452610279948235600160a060020a03169460248035953695946064949201919081908401838280828437509497506110b49650505050505050565b34801561053c57600080fd5b5061019c6004356110d3565b34801561055457600080fd5b5061027961112e565b34801561056957600080fd5b50610279611133565b34801561057e57600080fd5b5061019c600160a060020a0360043581169060243516611139565b3480156105a557600080fd5b5061019c6004356112d7565b3480156105bd57600080fd5b50610279611500565b60038054829081106105d457fe5b600091825260209091200154600160a060020a0316905081565b600030600160a060020a031633600160a060020a031614151561061057600080fd5b600160a060020a038216600090815260026020526040902054829060ff16151561063957600080fd5b600160a060020a0383166000908152600260205260408120805460ff1916905591505b600354600019018210156107145782600160a060020a031660038381548110151561068357fe5b600091825260209091200154600160a060020a03161415610709576003805460001981019081106106b057fe5b60009182526020909120015460038054600160a060020a0390921691849081106106d657fe5b9060005260206000200160006101000a815481600160a060020a030219169083600160a060020a03160217905550610714565b60019091019061065c565b600380546000190190610727908261163e565b5060035460045411156107405760035461074090610f3a565b604051600160a060020a038416907f8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b9090600090a2505050565b33600160a060020a03811660009081526002602052604090205460ff1615156107a157600080fd5b600082815260016020908152604080832033600160a060020a038116855292529091205483919060ff1615156107d657600080fd5b600084815260208190526040902060030154849060ff16156107f757600080fd5b6000858152600160209081526040808320600160a060020a0333168085529252808320805460ff191690555187927ff6a317157440607f36269043eb55f1287a5a19ba2216afeab88cd46cbcfb88e991a35050505050565b60026020526000908152604090205460ff1681565b600160209081526000928352604080842090915290825290205460ff1681565b6000600754620151800142111561089e57506006546108bb565b60085460065410156108b2575060006108bb565b50600854600654035b90565b6000805b600554811015610923578380156108eb575060008181526020819052604090206003015460ff16155b8061090f575082801561090f575060008181526020819052604090206003015460ff165b1561091b576001820191505b6001016108c2565b5092915050565b60065481565b60075481565b30600160a060020a031633600160a060020a031614151561095657600080fd5b600160a060020a038116600090815260026020526040902054819060ff161561097e57600080fd5b81600160a060020a038116151561099457600080fd5b600380549050600101600454603282111580156109b15750818111155b80156109bc57508015155b80156109c757508115155b15156109d257600080fd5b600160a060020a038516600081815260026020526040808220805460ff1916600190811790915560038054918201815583527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b01805473ffffffffffffffffffffffffffffffffffffffff191684179055517ff39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d9190a25050505050565b600080805b600354811015610aec5760008481526001602052604081206003805491929184908110610a9d57fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff1615610ad1576001820191505b600454821415610ae45760019250610aec565b600101610a74565b5050919050565b6000805b600354811015610b5c5760008381526001602052604081206003805491929184908110610b2057fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff1615610b54576001820191505b600101610af7565b50919050565b6000602081815291815260409081902080546001808301546002808501805487516101009582161595909502600019011691909104601f8101889004880284018801909652858352600160a060020a0390931695909491929190830182828015610c0d5780601f10610be257610100808354040283529160200191610c0d565b820191906000526020600020905b815481529060010190602001808311610bf057829003601f168201915b5050506003909301549192505060ff1684565b60606003805480602002602001604051908101604052809291908181526020018280548015610c7857602002820191906000526020600020905b8154600160a060020a03168152600190910190602001808311610c5a575b5050505050905090565b606080600080600554604051908082528060200260200182016040528015610cb4578160200160208202803883390190505b50925060009150600090505b600554811015610d3b57858015610ce9575060008181526020819052604090206003015460ff16155b80610d0d5750848015610d0d575060008181526020819052604090206003015460ff165b15610d3357808383815181101515610d2157fe5b60209081029091010152600191909101905b600101610cc0565b878703604051908082528060200260200182016040528015610d67578160200160208202803883390190505b5093508790505b86811015610db0578281815181101515610d8457fe5b9060200190602002015184898303815181101515610d9e57fe5b60209081029091010152600101610d6e565b505050949350505050565b606080600080600380549050604051908082528060200260200182016040528015610df0578160200160208202803883390190505b50925060009150600090505b600354811015610ead5760008581526001602052604081206003805491929184908110610e2557fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff1615610ea5576003805482908110610e6057fe5b6000918252602090912001548351600160a060020a0390911690849084908110610e8657fe5b600160a060020a03909216602092830290910190910152600191909101905b600101610dfc565b81604051908082528060200260200182016040528015610ed7578160200160208202803883390190505b509350600090505b81811015610f2c578281815181101515610ef557fe5b906020019060200201518482815181101515610f0d57fe5b600160a060020a03909216602092830290910190910152600101610edf565b505050919050565b60055481565b30600160a060020a031633600160a060020a0316141515610f5a57600080fd5b6003548160328211801590610f6f5750818111155b8015610f7a57508015155b8015610f8557508115155b1515610f9057600080fd5b60048390556040805184815290517fa3f1ee9126a074d9326c682f561767f710e927faa811f7a99829d49dc421797a9181900360200190a1505050565b33600160a060020a03811660009081526002602052604090205460ff161515610ff557600080fd5b6000828152602081905260409020548290600160a060020a0316151561101a57600080fd5b600083815260016020908152604080832033600160a060020a038116855292529091205484919060ff161561104e57600080fd5b6000858152600160208181526040808420600160a060020a0333168086529252808420805460ff1916909317909255905187927f4a504a94899432a9846e1aa406dceb1bcfd538bb839071d49d1e5e23f5be30ef91a36110ad856112d7565b5050505050565b60006110c1848484611506565b90506110cc81610fcd565b9392505050565b30600160a060020a031633600160a060020a03161415156110f357600080fd5b60068190556040805182815290517fc71bdc6afaf9b1aa90a7078191d4fc1adf3bf680fca3183697df6b0dc226bca29181900360200190a150565b603281565b60045481565b600030600160a060020a031633600160a060020a031614151561115b57600080fd5b600160a060020a038316600090815260026020526040902054839060ff16151561118457600080fd5b600160a060020a038316600090815260026020526040902054839060ff16156111ac57600080fd5b600092505b60035483101561123d5784600160a060020a03166003848154811015156111d457fe5b600091825260209091200154600160a060020a0316141561123257836003848154811015156111ff57fe5b9060005260206000200160006101000a815481600160a060020a030219169083600160a060020a0316021790555061123d565b6001909201916111b1565b600160a060020a03808616600081815260026020526040808220805460ff1990811690915593881682528082208054909416600117909355915190917f8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b9091a2604051600160a060020a038516907ff39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d90600090a25050505050565b33600160a060020a0381166000908152600260205260408120549091829160ff16151561130357600080fd5b600084815260016020908152604080832033600160a060020a038116855292529091205485919060ff16151561133857600080fd5b600086815260208190526040902060030154869060ff161561135957600080fd5b6000878152602081905260409020955061137287610a6f565b945084806113a557506002808701546000196101006001831615020116041580156113a557506113a586600101546115f6565b156114f75760038601805460ff191660011790558415156113cf5760018601546008805490910190555b8560000160009054906101000a9004600160a060020a0316600160a060020a0316866001015487600201604051808280546001816001161561010002031660029004801561145e5780601f106114335761010080835404028352916020019161145e565b820191906000526020600020905b81548152906001019060200180831161144157829003601f168201915b505091505060006040518083038185875af192505050156114a95760405187907f33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed7590600090a26114f7565b60405187907f526441bb6c1aba3c9a4a6ca1d6545da9c2333c8c48343ef398eb858d72b7923690600090a260038601805460ff191690558415156114f7576001860154600880549190910390555b50505050505050565b60085481565b600083600160a060020a038116151561151e57600080fd5b60055460408051608081018252600160a060020a0388811682526020808301898152838501898152600060608601819052878152808452959095208451815473ffffffffffffffffffffffffffffffffffffffff19169416939093178355516001830155925180519496509193909261159e926002850192910190611667565b50606091909101516003909101805460ff191691151591909117905560058054600101905560405182907fc0ba8fe4b176c1714197d43b9cc6bcf797a4a7461c5fe8d0ef6e184ae7601e5190600090a2509392505050565b60006007546201518001421115611611574260075560006008555b600654826008540111806116285750600854828101105b1561163557506000611639565b5060015b919050565b815481835581811115611662576000838152602090206116629181019083016116e5565b505050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106116a857805160ff19168380011785556116d5565b828001600101855582156116d5579182015b828111156116d55782518255916020019190600101906116ba565b506116e19291506116e5565b5090565b6108bb91905b808211156116e157600081556001016116eb5600a165627a7a72305820965b5c0e2cab6cc572131d36be2c6f27bc5571a0082b57f07a885634163f92010029`

// DeployMultiSigWalletWithDailyLimit deploys a new Ethereum contract, binding an instance of MultiSigWalletWithDailyLimit to it.
func DeployMultiSigWalletWithDailyLimit(auth *bind.TransactOpts, backend bind.ContractBackend, _owners []common.Address, _required *big.Int, _dailyLimit *big.Int) (common.Address, *types.Transaction, *MultiSigWalletWithDailyLimit, error) {
	parsed, err := abi.JSON(strings.NewReader(MultiSigWalletWithDailyLimitABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MultiSigWalletWithDailyLimitBin), backend, _owners, _required, _dailyLimit)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MultiSigWalletWithDailyLimit{MultiSigWalletWithDailyLimitCaller: MultiSigWalletWithDailyLimitCaller{contract: contract}, MultiSigWalletWithDailyLimitTransactor: MultiSigWalletWithDailyLimitTransactor{contract: contract}, MultiSigWalletWithDailyLimitFilterer: MultiSigWalletWithDailyLimitFilterer{contract: contract}}, nil
}

// MultiSigWalletWithDailyLimit is an auto generated Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimit struct {
	MultiSigWalletWithDailyLimitCaller     // Read-only binding to the contract
	MultiSigWalletWithDailyLimitTransactor // Write-only binding to the contract
	MultiSigWalletWithDailyLimitFilterer   // Log filterer for contract events
}

// MultiSigWalletWithDailyLimitCaller is an auto generated read-only Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletWithDailyLimitTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletWithDailyLimitFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MultiSigWalletWithDailyLimitFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletWithDailyLimitSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MultiSigWalletWithDailyLimitSession struct {
	Contract     *MultiSigWalletWithDailyLimit // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                 // Call options to use throughout this session
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// MultiSigWalletWithDailyLimitCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MultiSigWalletWithDailyLimitCallerSession struct {
	Contract *MultiSigWalletWithDailyLimitCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                       // Call options to use throughout this session
}

// MultiSigWalletWithDailyLimitTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MultiSigWalletWithDailyLimitTransactorSession struct {
	Contract     *MultiSigWalletWithDailyLimitTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                       // Transaction auth options to use throughout this session
}

// MultiSigWalletWithDailyLimitRaw is an auto generated low-level Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitRaw struct {
	Contract *MultiSigWalletWithDailyLimit // Generic contract binding to access the raw methods on
}

// MultiSigWalletWithDailyLimitCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitCallerRaw struct {
	Contract *MultiSigWalletWithDailyLimitCaller // Generic read-only contract binding to access the raw methods on
}

// MultiSigWalletWithDailyLimitTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitTransactorRaw struct {
	Contract *MultiSigWalletWithDailyLimitTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMultiSigWalletWithDailyLimit creates a new instance of MultiSigWalletWithDailyLimit, bound to a specific deployed contract.
func NewMultiSigWalletWithDailyLimit(address common.Address, backend bind.ContractBackend) (*MultiSigWalletWithDailyLimit, error) {
	contract, err := bindMultiSigWalletWithDailyLimit(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimit{MultiSigWalletWithDailyLimitCaller: MultiSigWalletWithDailyLimitCaller{contract: contract}, MultiSigWalletWithDailyLimitTransactor: MultiSigWalletWithDailyLimitTransactor{contract: contract}, MultiSigWalletWithDailyLimitFilterer: MultiSigWalletWithDailyLimitFilterer{contract: contract}}, nil
}

// NewMultiSigWalletWithDailyLimitCaller creates a new read-only instance of MultiSigWalletWithDailyLimit, bound to a specific deployed contract.
func NewMultiSigWalletWithDailyLimitCaller(address common.Address, caller bind.ContractCaller) (*MultiSigWalletWithDailyLimitCaller, error) {
	contract, err := bindMultiSigWalletWithDailyLimit(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitCaller{contract: contract}, nil
}

// NewMultiSigWalletWithDailyLimitTransactor creates a new write-only instance of MultiSigWalletWithDailyLimit, bound to a specific deployed contract.
func NewMultiSigWalletWithDailyLimitTransactor(address common.Address, transactor bind.ContractTransactor) (*MultiSigWalletWithDailyLimitTransactor, error) {
	contract, err := bindMultiSigWalletWithDailyLimit(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitTransactor{contract: contract}, nil
}

// NewMultiSigWalletWithDailyLimitFilterer creates a new log filterer instance of MultiSigWalletWithDailyLimit, bound to a specific deployed contract.
func NewMultiSigWalletWithDailyLimitFilterer(address common.Address, filterer bind.ContractFilterer) (*MultiSigWalletWithDailyLimitFilterer, error) {
	contract, err := bindMultiSigWalletWithDailyLimit(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitFilterer{contract: contract}, nil
}

// bindMultiSigWalletWithDailyLimit binds a generic wrapper to an already deployed contract.
func bindMultiSigWalletWithDailyLimit(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MultiSigWalletWithDailyLimitABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MultiSigWalletWithDailyLimit.Contract.MultiSigWalletWithDailyLimitCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.MultiSigWalletWithDailyLimitTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.MultiSigWalletWithDailyLimitTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MultiSigWalletWithDailyLimit.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.contract.Transact(opts, method, params...)
}

// MAXOWNERCOUNT is a free data retrieval call binding the contract method 0xd74f8edd.
//
// Solidity: function MAX_OWNER_COUNT() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) MAXOWNERCOUNT(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "MAX_OWNER_COUNT")
	return *ret0, err
}

// MAXOWNERCOUNT is a free data retrieval call binding the contract method 0xd74f8edd.
//
// Solidity: function MAX_OWNER_COUNT() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) MAXOWNERCOUNT() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.MAXOWNERCOUNT(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// MAXOWNERCOUNT is a free data retrieval call binding the contract method 0xd74f8edd.
//
// Solidity: function MAX_OWNER_COUNT() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) MAXOWNERCOUNT() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.MAXOWNERCOUNT(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// CalcMaxWithdraw is a free data retrieval call binding the contract method 0x4bc9fdc2.
//
// Solidity: function calcMaxWithdraw() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) CalcMaxWithdraw(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "calcMaxWithdraw")
	return *ret0, err
}

// CalcMaxWithdraw is a free data retrieval call binding the contract method 0x4bc9fdc2.
//
// Solidity: function calcMaxWithdraw() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) CalcMaxWithdraw() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.CalcMaxWithdraw(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// CalcMaxWithdraw is a free data retrieval call binding the contract method 0x4bc9fdc2.
//
// Solidity: function calcMaxWithdraw() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) CalcMaxWithdraw() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.CalcMaxWithdraw(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// Confirmations is a free data retrieval call binding the contract method 0x3411c81c.
//
// Solidity: function confirmations( uint256,  address) constant returns(bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) Confirmations(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "confirmations", arg0, arg1)
	return *ret0, err
}

// Confirmations is a free data retrieval call binding the contract method 0x3411c81c.
//
// Solidity: function confirmations( uint256,  address) constant returns(bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) Confirmations(arg0 *big.Int, arg1 common.Address) (bool, error) {
	return _MultiSigWalletWithDailyLimit.Contract.Confirmations(&_MultiSigWalletWithDailyLimit.CallOpts, arg0, arg1)
}

// Confirmations is a free data retrieval call binding the contract method 0x3411c81c.
//
// Solidity: function confirmations( uint256,  address) constant returns(bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) Confirmations(arg0 *big.Int, arg1 common.Address) (bool, error) {
	return _MultiSigWalletWithDailyLimit.Contract.Confirmations(&_MultiSigWalletWithDailyLimit.CallOpts, arg0, arg1)
}

// DailyLimit is a free data retrieval call binding the contract method 0x67eeba0c.
//
// Solidity: function dailyLimit() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) DailyLimit(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "dailyLimit")
	return *ret0, err
}

// DailyLimit is a free data retrieval call binding the contract method 0x67eeba0c.
//
// Solidity: function dailyLimit() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) DailyLimit() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.DailyLimit(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// DailyLimit is a free data retrieval call binding the contract method 0x67eeba0c.
//
// Solidity: function dailyLimit() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) DailyLimit() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.DailyLimit(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// GetConfirmationCount is a free data retrieval call binding the contract method 0x8b51d13f.
//
// Solidity: function getConfirmationCount(transactionId uint256) constant returns(count uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) GetConfirmationCount(opts *bind.CallOpts, transactionId *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "getConfirmationCount", transactionId)
	return *ret0, err
}

// GetConfirmationCount is a free data retrieval call binding the contract method 0x8b51d13f.
//
// Solidity: function getConfirmationCount(transactionId uint256) constant returns(count uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) GetConfirmationCount(transactionId *big.Int) (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetConfirmationCount(&_MultiSigWalletWithDailyLimit.CallOpts, transactionId)
}

// GetConfirmationCount is a free data retrieval call binding the contract method 0x8b51d13f.
//
// Solidity: function getConfirmationCount(transactionId uint256) constant returns(count uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) GetConfirmationCount(transactionId *big.Int) (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetConfirmationCount(&_MultiSigWalletWithDailyLimit.CallOpts, transactionId)
}

// GetConfirmations is a free data retrieval call binding the contract method 0xb5dc40c3.
//
// Solidity: function getConfirmations(transactionId uint256) constant returns(_confirmations address[])
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) GetConfirmations(opts *bind.CallOpts, transactionId *big.Int) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "getConfirmations", transactionId)
	return *ret0, err
}

// GetConfirmations is a free data retrieval call binding the contract method 0xb5dc40c3.
//
// Solidity: function getConfirmations(transactionId uint256) constant returns(_confirmations address[])
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) GetConfirmations(transactionId *big.Int) ([]common.Address, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetConfirmations(&_MultiSigWalletWithDailyLimit.CallOpts, transactionId)
}

// GetConfirmations is a free data retrieval call binding the contract method 0xb5dc40c3.
//
// Solidity: function getConfirmations(transactionId uint256) constant returns(_confirmations address[])
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) GetConfirmations(transactionId *big.Int) ([]common.Address, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetConfirmations(&_MultiSigWalletWithDailyLimit.CallOpts, transactionId)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() constant returns(address[])
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) GetOwners(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "getOwners")
	return *ret0, err
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() constant returns(address[])
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) GetOwners() ([]common.Address, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetOwners(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// GetOwners is a free data retrieval call binding the contract method 0xa0e67e2b.
//
// Solidity: function getOwners() constant returns(address[])
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) GetOwners() ([]common.Address, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetOwners(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x54741525.
//
// Solidity: function getTransactionCount(pending bool, executed bool) constant returns(count uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) GetTransactionCount(opts *bind.CallOpts, pending bool, executed bool) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "getTransactionCount", pending, executed)
	return *ret0, err
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x54741525.
//
// Solidity: function getTransactionCount(pending bool, executed bool) constant returns(count uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) GetTransactionCount(pending bool, executed bool) (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetTransactionCount(&_MultiSigWalletWithDailyLimit.CallOpts, pending, executed)
}

// GetTransactionCount is a free data retrieval call binding the contract method 0x54741525.
//
// Solidity: function getTransactionCount(pending bool, executed bool) constant returns(count uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) GetTransactionCount(pending bool, executed bool) (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetTransactionCount(&_MultiSigWalletWithDailyLimit.CallOpts, pending, executed)
}

// GetTransactionIds is a free data retrieval call binding the contract method 0xa8abe69a.
//
// Solidity: function getTransactionIds(from uint256, to uint256, pending bool, executed bool) constant returns(_transactionIds uint256[])
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) GetTransactionIds(opts *bind.CallOpts, from *big.Int, to *big.Int, pending bool, executed bool) ([]*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "getTransactionIds", from, to, pending, executed)
	return *ret0, err
}

// GetTransactionIds is a free data retrieval call binding the contract method 0xa8abe69a.
//
// Solidity: function getTransactionIds(from uint256, to uint256, pending bool, executed bool) constant returns(_transactionIds uint256[])
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) GetTransactionIds(from *big.Int, to *big.Int, pending bool, executed bool) ([]*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetTransactionIds(&_MultiSigWalletWithDailyLimit.CallOpts, from, to, pending, executed)
}

// GetTransactionIds is a free data retrieval call binding the contract method 0xa8abe69a.
//
// Solidity: function getTransactionIds(from uint256, to uint256, pending bool, executed bool) constant returns(_transactionIds uint256[])
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) GetTransactionIds(from *big.Int, to *big.Int, pending bool, executed bool) ([]*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.GetTransactionIds(&_MultiSigWalletWithDailyLimit.CallOpts, from, to, pending, executed)
}

// IsConfirmed is a free data retrieval call binding the contract method 0x784547a7.
//
// Solidity: function isConfirmed(transactionId uint256) constant returns(bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) IsConfirmed(opts *bind.CallOpts, transactionId *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "isConfirmed", transactionId)
	return *ret0, err
}

// IsConfirmed is a free data retrieval call binding the contract method 0x784547a7.
//
// Solidity: function isConfirmed(transactionId uint256) constant returns(bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) IsConfirmed(transactionId *big.Int) (bool, error) {
	return _MultiSigWalletWithDailyLimit.Contract.IsConfirmed(&_MultiSigWalletWithDailyLimit.CallOpts, transactionId)
}

// IsConfirmed is a free data retrieval call binding the contract method 0x784547a7.
//
// Solidity: function isConfirmed(transactionId uint256) constant returns(bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) IsConfirmed(transactionId *big.Int) (bool, error) {
	return _MultiSigWalletWithDailyLimit.Contract.IsConfirmed(&_MultiSigWalletWithDailyLimit.CallOpts, transactionId)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner( address) constant returns(bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) IsOwner(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "isOwner", arg0)
	return *ret0, err
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner( address) constant returns(bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) IsOwner(arg0 common.Address) (bool, error) {
	return _MultiSigWalletWithDailyLimit.Contract.IsOwner(&_MultiSigWalletWithDailyLimit.CallOpts, arg0)
}

// IsOwner is a free data retrieval call binding the contract method 0x2f54bf6e.
//
// Solidity: function isOwner( address) constant returns(bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) IsOwner(arg0 common.Address) (bool, error) {
	return _MultiSigWalletWithDailyLimit.Contract.IsOwner(&_MultiSigWalletWithDailyLimit.CallOpts, arg0)
}

// LastDay is a free data retrieval call binding the contract method 0x6b0c932d.
//
// Solidity: function lastDay() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) LastDay(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "lastDay")
	return *ret0, err
}

// LastDay is a free data retrieval call binding the contract method 0x6b0c932d.
//
// Solidity: function lastDay() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) LastDay() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.LastDay(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// LastDay is a free data retrieval call binding the contract method 0x6b0c932d.
//
// Solidity: function lastDay() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) LastDay() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.LastDay(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// Owners is a free data retrieval call binding the contract method 0x025e7c27.
//
// Solidity: function owners( uint256) constant returns(address)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) Owners(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "owners", arg0)
	return *ret0, err
}

// Owners is a free data retrieval call binding the contract method 0x025e7c27.
//
// Solidity: function owners( uint256) constant returns(address)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) Owners(arg0 *big.Int) (common.Address, error) {
	return _MultiSigWalletWithDailyLimit.Contract.Owners(&_MultiSigWalletWithDailyLimit.CallOpts, arg0)
}

// Owners is a free data retrieval call binding the contract method 0x025e7c27.
//
// Solidity: function owners( uint256) constant returns(address)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) Owners(arg0 *big.Int) (common.Address, error) {
	return _MultiSigWalletWithDailyLimit.Contract.Owners(&_MultiSigWalletWithDailyLimit.CallOpts, arg0)
}

// Required is a free data retrieval call binding the contract method 0xdc8452cd.
//
// Solidity: function required() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) Required(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "required")
	return *ret0, err
}

// Required is a free data retrieval call binding the contract method 0xdc8452cd.
//
// Solidity: function required() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) Required() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.Required(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// Required is a free data retrieval call binding the contract method 0xdc8452cd.
//
// Solidity: function required() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) Required() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.Required(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// SpentToday is a free data retrieval call binding the contract method 0xf059cf2b.
//
// Solidity: function spentToday() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) SpentToday(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "spentToday")
	return *ret0, err
}

// SpentToday is a free data retrieval call binding the contract method 0xf059cf2b.
//
// Solidity: function spentToday() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) SpentToday() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.SpentToday(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// SpentToday is a free data retrieval call binding the contract method 0xf059cf2b.
//
// Solidity: function spentToday() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) SpentToday() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.SpentToday(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// TransactionCount is a free data retrieval call binding the contract method 0xb77bf600.
//
// Solidity: function transactionCount() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) TransactionCount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "transactionCount")
	return *ret0, err
}

// TransactionCount is a free data retrieval call binding the contract method 0xb77bf600.
//
// Solidity: function transactionCount() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) TransactionCount() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.TransactionCount(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// TransactionCount is a free data retrieval call binding the contract method 0xb77bf600.
//
// Solidity: function transactionCount() constant returns(uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) TransactionCount() (*big.Int, error) {
	return _MultiSigWalletWithDailyLimit.Contract.TransactionCount(&_MultiSigWalletWithDailyLimit.CallOpts)
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions( uint256) constant returns(destination address, value uint256, data bytes, executed bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCaller) Transactions(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Destination common.Address
	Value       *big.Int
	Data        []byte
	Executed    bool
}, error) {
	ret := new(struct {
		Destination common.Address
		Value       *big.Int
		Data        []byte
		Executed    bool
	})
	out := ret
	err := _MultiSigWalletWithDailyLimit.contract.Call(opts, out, "transactions", arg0)
	return *ret, err
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions( uint256) constant returns(destination address, value uint256, data bytes, executed bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) Transactions(arg0 *big.Int) (struct {
	Destination common.Address
	Value       *big.Int
	Data        []byte
	Executed    bool
}, error) {
	return _MultiSigWalletWithDailyLimit.Contract.Transactions(&_MultiSigWalletWithDailyLimit.CallOpts, arg0)
}

// Transactions is a free data retrieval call binding the contract method 0x9ace38c2.
//
// Solidity: function transactions( uint256) constant returns(destination address, value uint256, data bytes, executed bool)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitCallerSession) Transactions(arg0 *big.Int) (struct {
	Destination common.Address
	Value       *big.Int
	Data        []byte
	Executed    bool
}, error) {
	return _MultiSigWalletWithDailyLimit.Contract.Transactions(&_MultiSigWalletWithDailyLimit.CallOpts, arg0)
}

// AddOwner is a paid mutator transaction binding the contract method 0x7065cb48.
//
// Solidity: function addOwner(owner address) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactor) AddOwner(opts *bind.TransactOpts, owner common.Address) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.contract.Transact(opts, "addOwner", owner)
}

// AddOwner is a paid mutator transaction binding the contract method 0x7065cb48.
//
// Solidity: function addOwner(owner address) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) AddOwner(owner common.Address) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.AddOwner(&_MultiSigWalletWithDailyLimit.TransactOpts, owner)
}

// AddOwner is a paid mutator transaction binding the contract method 0x7065cb48.
//
// Solidity: function addOwner(owner address) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorSession) AddOwner(owner common.Address) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.AddOwner(&_MultiSigWalletWithDailyLimit.TransactOpts, owner)
}

// ChangeDailyLimit is a paid mutator transaction binding the contract method 0xcea08621.
//
// Solidity: function changeDailyLimit(_dailyLimit uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactor) ChangeDailyLimit(opts *bind.TransactOpts, _dailyLimit *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.contract.Transact(opts, "changeDailyLimit", _dailyLimit)
}

// ChangeDailyLimit is a paid mutator transaction binding the contract method 0xcea08621.
//
// Solidity: function changeDailyLimit(_dailyLimit uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) ChangeDailyLimit(_dailyLimit *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ChangeDailyLimit(&_MultiSigWalletWithDailyLimit.TransactOpts, _dailyLimit)
}

// ChangeDailyLimit is a paid mutator transaction binding the contract method 0xcea08621.
//
// Solidity: function changeDailyLimit(_dailyLimit uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorSession) ChangeDailyLimit(_dailyLimit *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ChangeDailyLimit(&_MultiSigWalletWithDailyLimit.TransactOpts, _dailyLimit)
}

// ChangeRequirement is a paid mutator transaction binding the contract method 0xba51a6df.
//
// Solidity: function changeRequirement(_required uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactor) ChangeRequirement(opts *bind.TransactOpts, _required *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.contract.Transact(opts, "changeRequirement", _required)
}

// ChangeRequirement is a paid mutator transaction binding the contract method 0xba51a6df.
//
// Solidity: function changeRequirement(_required uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) ChangeRequirement(_required *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ChangeRequirement(&_MultiSigWalletWithDailyLimit.TransactOpts, _required)
}

// ChangeRequirement is a paid mutator transaction binding the contract method 0xba51a6df.
//
// Solidity: function changeRequirement(_required uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorSession) ChangeRequirement(_required *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ChangeRequirement(&_MultiSigWalletWithDailyLimit.TransactOpts, _required)
}

// ConfirmTransaction is a paid mutator transaction binding the contract method 0xc01a8c84.
//
// Solidity: function confirmTransaction(transactionId uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactor) ConfirmTransaction(opts *bind.TransactOpts, transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.contract.Transact(opts, "confirmTransaction", transactionId)
}

// ConfirmTransaction is a paid mutator transaction binding the contract method 0xc01a8c84.
//
// Solidity: function confirmTransaction(transactionId uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) ConfirmTransaction(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ConfirmTransaction(&_MultiSigWalletWithDailyLimit.TransactOpts, transactionId)
}

// ConfirmTransaction is a paid mutator transaction binding the contract method 0xc01a8c84.
//
// Solidity: function confirmTransaction(transactionId uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorSession) ConfirmTransaction(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ConfirmTransaction(&_MultiSigWalletWithDailyLimit.TransactOpts, transactionId)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0xee22610b.
//
// Solidity: function executeTransaction(transactionId uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactor) ExecuteTransaction(opts *bind.TransactOpts, transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.contract.Transact(opts, "executeTransaction", transactionId)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0xee22610b.
//
// Solidity: function executeTransaction(transactionId uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) ExecuteTransaction(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ExecuteTransaction(&_MultiSigWalletWithDailyLimit.TransactOpts, transactionId)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0xee22610b.
//
// Solidity: function executeTransaction(transactionId uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorSession) ExecuteTransaction(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ExecuteTransaction(&_MultiSigWalletWithDailyLimit.TransactOpts, transactionId)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0x173825d9.
//
// Solidity: function removeOwner(owner address) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactor) RemoveOwner(opts *bind.TransactOpts, owner common.Address) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.contract.Transact(opts, "removeOwner", owner)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0x173825d9.
//
// Solidity: function removeOwner(owner address) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) RemoveOwner(owner common.Address) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.RemoveOwner(&_MultiSigWalletWithDailyLimit.TransactOpts, owner)
}

// RemoveOwner is a paid mutator transaction binding the contract method 0x173825d9.
//
// Solidity: function removeOwner(owner address) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorSession) RemoveOwner(owner common.Address) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.RemoveOwner(&_MultiSigWalletWithDailyLimit.TransactOpts, owner)
}

// ReplaceOwner is a paid mutator transaction binding the contract method 0xe20056e6.
//
// Solidity: function replaceOwner(owner address, newOwner address) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactor) ReplaceOwner(opts *bind.TransactOpts, owner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.contract.Transact(opts, "replaceOwner", owner, newOwner)
}

// ReplaceOwner is a paid mutator transaction binding the contract method 0xe20056e6.
//
// Solidity: function replaceOwner(owner address, newOwner address) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) ReplaceOwner(owner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ReplaceOwner(&_MultiSigWalletWithDailyLimit.TransactOpts, owner, newOwner)
}

// ReplaceOwner is a paid mutator transaction binding the contract method 0xe20056e6.
//
// Solidity: function replaceOwner(owner address, newOwner address) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorSession) ReplaceOwner(owner common.Address, newOwner common.Address) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.ReplaceOwner(&_MultiSigWalletWithDailyLimit.TransactOpts, owner, newOwner)
}

// RevokeConfirmation is a paid mutator transaction binding the contract method 0x20ea8d86.
//
// Solidity: function revokeConfirmation(transactionId uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactor) RevokeConfirmation(opts *bind.TransactOpts, transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.contract.Transact(opts, "revokeConfirmation", transactionId)
}

// RevokeConfirmation is a paid mutator transaction binding the contract method 0x20ea8d86.
//
// Solidity: function revokeConfirmation(transactionId uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) RevokeConfirmation(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.RevokeConfirmation(&_MultiSigWalletWithDailyLimit.TransactOpts, transactionId)
}

// RevokeConfirmation is a paid mutator transaction binding the contract method 0x20ea8d86.
//
// Solidity: function revokeConfirmation(transactionId uint256) returns()
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorSession) RevokeConfirmation(transactionId *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.RevokeConfirmation(&_MultiSigWalletWithDailyLimit.TransactOpts, transactionId)
}

// SubmitTransaction is a paid mutator transaction binding the contract method 0xc6427474.
//
// Solidity: function submitTransaction(destination address, value uint256, data bytes) returns(transactionId uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactor) SubmitTransaction(opts *bind.TransactOpts, destination common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.contract.Transact(opts, "submitTransaction", destination, value, data)
}

// SubmitTransaction is a paid mutator transaction binding the contract method 0xc6427474.
//
// Solidity: function submitTransaction(destination address, value uint256, data bytes) returns(transactionId uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitSession) SubmitTransaction(destination common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.SubmitTransaction(&_MultiSigWalletWithDailyLimit.TransactOpts, destination, value, data)
}

// SubmitTransaction is a paid mutator transaction binding the contract method 0xc6427474.
//
// Solidity: function submitTransaction(destination address, value uint256, data bytes) returns(transactionId uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitTransactorSession) SubmitTransaction(destination common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimit.Contract.SubmitTransaction(&_MultiSigWalletWithDailyLimit.TransactOpts, destination, value, data)
}

// MultiSigWalletWithDailyLimitConfirmationIterator is returned from FilterConfirmation and is used to iterate over the raw logs and unpacked data for Confirmation events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitConfirmationIterator struct {
	Event *MultiSigWalletWithDailyLimitConfirmation // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitConfirmationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitConfirmation)
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
		it.Event = new(MultiSigWalletWithDailyLimitConfirmation)
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
func (it *MultiSigWalletWithDailyLimitConfirmationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitConfirmationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitConfirmation represents a Confirmation event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitConfirmation struct {
	Sender        common.Address
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterConfirmation is a free log retrieval operation binding the contract event 0x4a504a94899432a9846e1aa406dceb1bcfd538bb839071d49d1e5e23f5be30ef.
//
// Solidity: event Confirmation(sender indexed address, transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterConfirmation(opts *bind.FilterOpts, sender []common.Address, transactionId []*big.Int) (*MultiSigWalletWithDailyLimitConfirmationIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "Confirmation", senderRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitConfirmationIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "Confirmation", logs: logs, sub: sub}, nil
}

// WatchConfirmation is a free log subscription operation binding the contract event 0x4a504a94899432a9846e1aa406dceb1bcfd538bb839071d49d1e5e23f5be30ef.
//
// Solidity: event Confirmation(sender indexed address, transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchConfirmation(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitConfirmation, sender []common.Address, transactionId []*big.Int) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "Confirmation", senderRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitConfirmation)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "Confirmation", log); err != nil {
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

// MultiSigWalletWithDailyLimitDailyLimitChangeIterator is returned from FilterDailyLimitChange and is used to iterate over the raw logs and unpacked data for DailyLimitChange events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitDailyLimitChangeIterator struct {
	Event *MultiSigWalletWithDailyLimitDailyLimitChange // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitDailyLimitChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitDailyLimitChange)
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
		it.Event = new(MultiSigWalletWithDailyLimitDailyLimitChange)
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
func (it *MultiSigWalletWithDailyLimitDailyLimitChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitDailyLimitChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitDailyLimitChange represents a DailyLimitChange event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitDailyLimitChange struct {
	DailyLimit *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterDailyLimitChange is a free log retrieval operation binding the contract event 0xc71bdc6afaf9b1aa90a7078191d4fc1adf3bf680fca3183697df6b0dc226bca2.
//
// Solidity: event DailyLimitChange(dailyLimit uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterDailyLimitChange(opts *bind.FilterOpts) (*MultiSigWalletWithDailyLimitDailyLimitChangeIterator, error) {

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "DailyLimitChange")
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitDailyLimitChangeIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "DailyLimitChange", logs: logs, sub: sub}, nil
}

// WatchDailyLimitChange is a free log subscription operation binding the contract event 0xc71bdc6afaf9b1aa90a7078191d4fc1adf3bf680fca3183697df6b0dc226bca2.
//
// Solidity: event DailyLimitChange(dailyLimit uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchDailyLimitChange(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitDailyLimitChange) (event.Subscription, error) {

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "DailyLimitChange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitDailyLimitChange)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "DailyLimitChange", log); err != nil {
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

// MultiSigWalletWithDailyLimitDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitDepositIterator struct {
	Event *MultiSigWalletWithDailyLimitDeposit // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitDeposit)
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
		it.Event = new(MultiSigWalletWithDailyLimitDeposit)
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
func (it *MultiSigWalletWithDailyLimitDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitDeposit represents a Deposit event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitDeposit struct {
	Sender common.Address
	Value  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(sender indexed address, value uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterDeposit(opts *bind.FilterOpts, sender []common.Address) (*MultiSigWalletWithDailyLimitDepositIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "Deposit", senderRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitDepositIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(sender indexed address, value uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitDeposit, sender []common.Address) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "Deposit", senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitDeposit)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// MultiSigWalletWithDailyLimitExecutionIterator is returned from FilterExecution and is used to iterate over the raw logs and unpacked data for Execution events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitExecutionIterator struct {
	Event *MultiSigWalletWithDailyLimitExecution // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitExecutionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitExecution)
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
		it.Event = new(MultiSigWalletWithDailyLimitExecution)
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
func (it *MultiSigWalletWithDailyLimitExecutionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitExecutionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitExecution represents a Execution event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitExecution struct {
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterExecution is a free log retrieval operation binding the contract event 0x33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed75.
//
// Solidity: event Execution(transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterExecution(opts *bind.FilterOpts, transactionId []*big.Int) (*MultiSigWalletWithDailyLimitExecutionIterator, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "Execution", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitExecutionIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "Execution", logs: logs, sub: sub}, nil
}

// WatchExecution is a free log subscription operation binding the contract event 0x33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed75.
//
// Solidity: event Execution(transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchExecution(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitExecution, transactionId []*big.Int) (event.Subscription, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "Execution", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitExecution)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "Execution", log); err != nil {
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

// MultiSigWalletWithDailyLimitExecutionFailureIterator is returned from FilterExecutionFailure and is used to iterate over the raw logs and unpacked data for ExecutionFailure events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitExecutionFailureIterator struct {
	Event *MultiSigWalletWithDailyLimitExecutionFailure // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitExecutionFailureIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitExecutionFailure)
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
		it.Event = new(MultiSigWalletWithDailyLimitExecutionFailure)
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
func (it *MultiSigWalletWithDailyLimitExecutionFailureIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitExecutionFailureIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitExecutionFailure represents a ExecutionFailure event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitExecutionFailure struct {
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterExecutionFailure is a free log retrieval operation binding the contract event 0x526441bb6c1aba3c9a4a6ca1d6545da9c2333c8c48343ef398eb858d72b79236.
//
// Solidity: event ExecutionFailure(transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterExecutionFailure(opts *bind.FilterOpts, transactionId []*big.Int) (*MultiSigWalletWithDailyLimitExecutionFailureIterator, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "ExecutionFailure", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitExecutionFailureIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "ExecutionFailure", logs: logs, sub: sub}, nil
}

// WatchExecutionFailure is a free log subscription operation binding the contract event 0x526441bb6c1aba3c9a4a6ca1d6545da9c2333c8c48343ef398eb858d72b79236.
//
// Solidity: event ExecutionFailure(transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchExecutionFailure(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitExecutionFailure, transactionId []*big.Int) (event.Subscription, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "ExecutionFailure", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitExecutionFailure)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "ExecutionFailure", log); err != nil {
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

// MultiSigWalletWithDailyLimitOwnerAdditionIterator is returned from FilterOwnerAddition and is used to iterate over the raw logs and unpacked data for OwnerAddition events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitOwnerAdditionIterator struct {
	Event *MultiSigWalletWithDailyLimitOwnerAddition // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitOwnerAdditionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitOwnerAddition)
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
		it.Event = new(MultiSigWalletWithDailyLimitOwnerAddition)
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
func (it *MultiSigWalletWithDailyLimitOwnerAdditionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitOwnerAdditionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitOwnerAddition represents a OwnerAddition event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitOwnerAddition struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOwnerAddition is a free log retrieval operation binding the contract event 0xf39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d.
//
// Solidity: event OwnerAddition(owner indexed address)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterOwnerAddition(opts *bind.FilterOpts, owner []common.Address) (*MultiSigWalletWithDailyLimitOwnerAdditionIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "OwnerAddition", ownerRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitOwnerAdditionIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "OwnerAddition", logs: logs, sub: sub}, nil
}

// WatchOwnerAddition is a free log subscription operation binding the contract event 0xf39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d.
//
// Solidity: event OwnerAddition(owner indexed address)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchOwnerAddition(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitOwnerAddition, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "OwnerAddition", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitOwnerAddition)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "OwnerAddition", log); err != nil {
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

// MultiSigWalletWithDailyLimitOwnerRemovalIterator is returned from FilterOwnerRemoval and is used to iterate over the raw logs and unpacked data for OwnerRemoval events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitOwnerRemovalIterator struct {
	Event *MultiSigWalletWithDailyLimitOwnerRemoval // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitOwnerRemovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitOwnerRemoval)
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
		it.Event = new(MultiSigWalletWithDailyLimitOwnerRemoval)
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
func (it *MultiSigWalletWithDailyLimitOwnerRemovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitOwnerRemovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitOwnerRemoval represents a OwnerRemoval event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitOwnerRemoval struct {
	Owner common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterOwnerRemoval is a free log retrieval operation binding the contract event 0x8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b90.
//
// Solidity: event OwnerRemoval(owner indexed address)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterOwnerRemoval(opts *bind.FilterOpts, owner []common.Address) (*MultiSigWalletWithDailyLimitOwnerRemovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "OwnerRemoval", ownerRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitOwnerRemovalIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "OwnerRemoval", logs: logs, sub: sub}, nil
}

// WatchOwnerRemoval is a free log subscription operation binding the contract event 0x8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b90.
//
// Solidity: event OwnerRemoval(owner indexed address)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchOwnerRemoval(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitOwnerRemoval, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "OwnerRemoval", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitOwnerRemoval)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "OwnerRemoval", log); err != nil {
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

// MultiSigWalletWithDailyLimitRequirementChangeIterator is returned from FilterRequirementChange and is used to iterate over the raw logs and unpacked data for RequirementChange events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitRequirementChangeIterator struct {
	Event *MultiSigWalletWithDailyLimitRequirementChange // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitRequirementChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitRequirementChange)
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
		it.Event = new(MultiSigWalletWithDailyLimitRequirementChange)
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
func (it *MultiSigWalletWithDailyLimitRequirementChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitRequirementChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitRequirementChange represents a RequirementChange event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitRequirementChange struct {
	Required *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRequirementChange is a free log retrieval operation binding the contract event 0xa3f1ee9126a074d9326c682f561767f710e927faa811f7a99829d49dc421797a.
//
// Solidity: event RequirementChange(required uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterRequirementChange(opts *bind.FilterOpts) (*MultiSigWalletWithDailyLimitRequirementChangeIterator, error) {

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "RequirementChange")
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitRequirementChangeIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "RequirementChange", logs: logs, sub: sub}, nil
}

// WatchRequirementChange is a free log subscription operation binding the contract event 0xa3f1ee9126a074d9326c682f561767f710e927faa811f7a99829d49dc421797a.
//
// Solidity: event RequirementChange(required uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchRequirementChange(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitRequirementChange) (event.Subscription, error) {

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "RequirementChange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitRequirementChange)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "RequirementChange", log); err != nil {
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

// MultiSigWalletWithDailyLimitRevocationIterator is returned from FilterRevocation and is used to iterate over the raw logs and unpacked data for Revocation events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitRevocationIterator struct {
	Event *MultiSigWalletWithDailyLimitRevocation // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitRevocationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitRevocation)
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
		it.Event = new(MultiSigWalletWithDailyLimitRevocation)
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
func (it *MultiSigWalletWithDailyLimitRevocationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitRevocationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitRevocation represents a Revocation event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitRevocation struct {
	Sender        common.Address
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRevocation is a free log retrieval operation binding the contract event 0xf6a317157440607f36269043eb55f1287a5a19ba2216afeab88cd46cbcfb88e9.
//
// Solidity: event Revocation(sender indexed address, transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterRevocation(opts *bind.FilterOpts, sender []common.Address, transactionId []*big.Int) (*MultiSigWalletWithDailyLimitRevocationIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "Revocation", senderRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitRevocationIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "Revocation", logs: logs, sub: sub}, nil
}

// WatchRevocation is a free log subscription operation binding the contract event 0xf6a317157440607f36269043eb55f1287a5a19ba2216afeab88cd46cbcfb88e9.
//
// Solidity: event Revocation(sender indexed address, transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchRevocation(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitRevocation, sender []common.Address, transactionId []*big.Int) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "Revocation", senderRule, transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitRevocation)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "Revocation", log); err != nil {
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

// MultiSigWalletWithDailyLimitSubmissionIterator is returned from FilterSubmission and is used to iterate over the raw logs and unpacked data for Submission events raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitSubmissionIterator struct {
	Event *MultiSigWalletWithDailyLimitSubmission // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitSubmissionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitSubmission)
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
		it.Event = new(MultiSigWalletWithDailyLimitSubmission)
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
func (it *MultiSigWalletWithDailyLimitSubmissionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitSubmissionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitSubmission represents a Submission event raised by the MultiSigWalletWithDailyLimit contract.
type MultiSigWalletWithDailyLimitSubmission struct {
	TransactionId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSubmission is a free log retrieval operation binding the contract event 0xc0ba8fe4b176c1714197d43b9cc6bcf797a4a7461c5fe8d0ef6e184ae7601e51.
//
// Solidity: event Submission(transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) FilterSubmission(opts *bind.FilterOpts, transactionId []*big.Int) (*MultiSigWalletWithDailyLimitSubmissionIterator, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.FilterLogs(opts, "Submission", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitSubmissionIterator{contract: _MultiSigWalletWithDailyLimit.contract, event: "Submission", logs: logs, sub: sub}, nil
}

// WatchSubmission is a free log subscription operation binding the contract event 0xc0ba8fe4b176c1714197d43b9cc6bcf797a4a7461c5fe8d0ef6e184ae7601e51.
//
// Solidity: event Submission(transactionId indexed uint256)
func (_MultiSigWalletWithDailyLimit *MultiSigWalletWithDailyLimitFilterer) WatchSubmission(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitSubmission, transactionId []*big.Int) (event.Subscription, error) {

	var transactionIdRule []interface{}
	for _, transactionIdItem := range transactionId {
		transactionIdRule = append(transactionIdRule, transactionIdItem)
	}

	logs, sub, err := _MultiSigWalletWithDailyLimit.contract.WatchLogs(opts, "Submission", transactionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitSubmission)
				if err := _MultiSigWalletWithDailyLimit.contract.UnpackLog(event, "Submission", log); err != nil {
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

// MultiSigWalletWithDailyLimitFactoryABI is the input ABI used to generate the binding from.
const MultiSigWalletWithDailyLimitFactoryABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"isInstantiation\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owners\",\"type\":\"address[]\"},{\"name\":\"_required\",\"type\":\"uint256\"},{\"name\":\"_dailyLimit\",\"type\":\"uint256\"}],\"name\":\"create\",\"outputs\":[{\"name\":\"wallet\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"instantiations\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"creator\",\"type\":\"address\"}],\"name\":\"getInstantiationCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"instantiation\",\"type\":\"address\"}],\"name\":\"ContractInstantiation\",\"type\":\"event\"}]"

// MultiSigWalletWithDailyLimitFactoryBin is the compiled bytecode used for deploying new contracts.
const MultiSigWalletWithDailyLimitFactoryBin = `0x608060405234801561001057600080fd5b50611c87806100206000396000f3006080604052600436106100615763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416632f4f3316811461006657806353d9d9101461009b57806357183c82146101155780638f83847814610139575b600080fd5b34801561007257600080fd5b50610087600160a060020a036004351661016c565b604080519115158252519081900360200190f35b3480156100a757600080fd5b50604080516020600480358082013583810280860185019096528085526100f9953695939460249493850192918291850190849080828437509497505084359550505060209092013591506101819050565b60408051600160a060020a039092168252519081900360200190f35b34801561012157600080fd5b506100f9600160a060020a0360043516602435610210565b34801561014557600080fd5b5061015a600160a060020a0360043516610247565b60408051918252519081900360200190f35b60006020819052908152604090205460ff1681565b600083838361018e6102ff565b6020808201849052604082018390526060808352855190830152845182916080830191878201910280838360005b838110156101d45781810151838201526020016101bc565b50505050905001945050505050604051809103906000f0801580156101fd573d6000803e3d6000fd5b50905061020981610262565b9392505050565b60016020528160005260406000208181548110151561022b57fe5b600091825260209091200154600160a060020a03169150829050565b600160a060020a031660009081526001602052604090205490565b600160a060020a03808216600081815260208181526040808320805460ff191660019081179091553390951680845285835281842080549687018155845292829020909401805473ffffffffffffffffffffffffffffffffffffffff191684179055835191825281019190915281517f4fb057ad4a26ed17a57957fa69c306f11987596069b89521c511fc9a894e6161929181900390910190a150565b60405161194c8061031083390190560060806040523480156200001157600080fd5b506040516200194c3803806200194c833981016040908152815160208301519183015192018051909290839083906000908260328211801590620000555750818111155b80156200006157508015155b80156200006d57508115155b15156200007957600080fd5b600092505b845183101562000151576002600086858151811015156200009b57fe5b6020908102909101810151600160a060020a031682528101919091526040016000205460ff16158015620000f157508483815181101515620000d957fe5b90602001906020020151600160a060020a0316600014155b1515620000fd57600080fd5b60016002600087868151811015156200011257fe5b602090810291909101810151600160a060020a03168252810191909152604001600020805460ff1916911515919091179055600192909201916200007e565b8451620001669060039060208801906200017d565b505050600491909155505060065550620002119050565b828054828255906000526020600020908101928215620001d5579160200282015b82811115620001d55782518254600160a060020a031916600160a060020a039091161782556020909201916001909101906200019e565b50620001e3929150620001e7565b5090565b6200020e91905b80821115620001e3578054600160a060020a0319168155600101620001ee565b90565b61172b80620002216000396000f3006080604052600436106101535763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663025e7c27811461019e578063173825d9146101d257806320ea8d86146101f35780632f54bf6e1461020b5780633411c81c146102405780634bc9fdc214610264578063547415251461028b57806367eeba0c146102aa5780636b0c932d146102bf5780637065cb48146102d4578063784547a7146102f55780638b51d13f1461030d5780639ace38c214610325578063a0e67e2b146103e0578063a8abe69a14610445578063b5dc40c31461046a578063b77bf60014610482578063ba51a6df14610497578063c01a8c84146104af578063c6427474146104c7578063cea0862114610530578063d74f8edd14610548578063dc8452cd1461055d578063e20056e614610572578063ee22610b14610599578063f059cf2b146105b1575b600034111561019c57604080513481529051600160a060020a033316917fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c919081900360200190a25b005b3480156101aa57600080fd5b506101b66004356105c6565b60408051600160a060020a039092168252519081900360200190f35b3480156101de57600080fd5b5061019c600160a060020a03600435166105ee565b3480156101ff57600080fd5b5061019c600435610779565b34801561021757600080fd5b5061022c600160a060020a036004351661084f565b604080519115158252519081900360200190f35b34801561024c57600080fd5b5061022c600435600160a060020a0360243516610864565b34801561027057600080fd5b50610279610884565b60408051918252519081900360200190f35b34801561029757600080fd5b50610279600435151560243515156108be565b3480156102b657600080fd5b5061027961092a565b3480156102cb57600080fd5b50610279610930565b3480156102e057600080fd5b5061019c600160a060020a0360043516610936565b34801561030157600080fd5b5061022c600435610a6f565b34801561031957600080fd5b50610279600435610af3565b34801561033157600080fd5b5061033d600435610b62565b6040518085600160a060020a0316600160a060020a031681526020018481526020018060200183151515158152602001828103825284818151815260200191508051906020019080838360005b838110156103a257818101518382015260200161038a565b50505050905090810190601f1680156103cf5780820380516001836020036101000a031916815260200191505b509550505050505060405180910390f35b3480156103ec57600080fd5b506103f5610c20565b60408051602080825283518183015283519192839290830191858101910280838360005b83811015610431578181015183820152602001610419565b505050509050019250505060405180910390f35b34801561045157600080fd5b506103f560043560243560443515156064351515610c82565b34801561047657600080fd5b506103f5600435610dbb565b34801561048e57600080fd5b50610279610f34565b3480156104a357600080fd5b5061019c600435610f3a565b3480156104bb57600080fd5b5061019c600435610fcd565b3480156104d357600080fd5b50604080516020600460443581810135601f8101849004840285018401909552848452610279948235600160a060020a03169460248035953695946064949201919081908401838280828437509497506110b49650505050505050565b34801561053c57600080fd5b5061019c6004356110d3565b34801561055457600080fd5b5061027961112e565b34801561056957600080fd5b50610279611133565b34801561057e57600080fd5b5061019c600160a060020a0360043581169060243516611139565b3480156105a557600080fd5b5061019c6004356112d7565b3480156105bd57600080fd5b50610279611500565b60038054829081106105d457fe5b600091825260209091200154600160a060020a0316905081565b600030600160a060020a031633600160a060020a031614151561061057600080fd5b600160a060020a038216600090815260026020526040902054829060ff16151561063957600080fd5b600160a060020a0383166000908152600260205260408120805460ff1916905591505b600354600019018210156107145782600160a060020a031660038381548110151561068357fe5b600091825260209091200154600160a060020a03161415610709576003805460001981019081106106b057fe5b60009182526020909120015460038054600160a060020a0390921691849081106106d657fe5b9060005260206000200160006101000a815481600160a060020a030219169083600160a060020a03160217905550610714565b60019091019061065c565b600380546000190190610727908261163e565b5060035460045411156107405760035461074090610f3a565b604051600160a060020a038416907f8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b9090600090a2505050565b33600160a060020a03811660009081526002602052604090205460ff1615156107a157600080fd5b600082815260016020908152604080832033600160a060020a038116855292529091205483919060ff1615156107d657600080fd5b600084815260208190526040902060030154849060ff16156107f757600080fd5b6000858152600160209081526040808320600160a060020a0333168085529252808320805460ff191690555187927ff6a317157440607f36269043eb55f1287a5a19ba2216afeab88cd46cbcfb88e991a35050505050565b60026020526000908152604090205460ff1681565b600160209081526000928352604080842090915290825290205460ff1681565b6000600754620151800142111561089e57506006546108bb565b60085460065410156108b2575060006108bb565b50600854600654035b90565b6000805b600554811015610923578380156108eb575060008181526020819052604090206003015460ff16155b8061090f575082801561090f575060008181526020819052604090206003015460ff165b1561091b576001820191505b6001016108c2565b5092915050565b60065481565b60075481565b30600160a060020a031633600160a060020a031614151561095657600080fd5b600160a060020a038116600090815260026020526040902054819060ff161561097e57600080fd5b81600160a060020a038116151561099457600080fd5b600380549050600101600454603282111580156109b15750818111155b80156109bc57508015155b80156109c757508115155b15156109d257600080fd5b600160a060020a038516600081815260026020526040808220805460ff1916600190811790915560038054918201815583527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b01805473ffffffffffffffffffffffffffffffffffffffff191684179055517ff39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d9190a25050505050565b600080805b600354811015610aec5760008481526001602052604081206003805491929184908110610a9d57fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff1615610ad1576001820191505b600454821415610ae45760019250610aec565b600101610a74565b5050919050565b6000805b600354811015610b5c5760008381526001602052604081206003805491929184908110610b2057fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff1615610b54576001820191505b600101610af7565b50919050565b6000602081815291815260409081902080546001808301546002808501805487516101009582161595909502600019011691909104601f8101889004880284018801909652858352600160a060020a0390931695909491929190830182828015610c0d5780601f10610be257610100808354040283529160200191610c0d565b820191906000526020600020905b815481529060010190602001808311610bf057829003601f168201915b5050506003909301549192505060ff1684565b60606003805480602002602001604051908101604052809291908181526020018280548015610c7857602002820191906000526020600020905b8154600160a060020a03168152600190910190602001808311610c5a575b5050505050905090565b606080600080600554604051908082528060200260200182016040528015610cb4578160200160208202803883390190505b50925060009150600090505b600554811015610d3b57858015610ce9575060008181526020819052604090206003015460ff16155b80610d0d5750848015610d0d575060008181526020819052604090206003015460ff165b15610d3357808383815181101515610d2157fe5b60209081029091010152600191909101905b600101610cc0565b878703604051908082528060200260200182016040528015610d67578160200160208202803883390190505b5093508790505b86811015610db0578281815181101515610d8457fe5b9060200190602002015184898303815181101515610d9e57fe5b60209081029091010152600101610d6e565b505050949350505050565b606080600080600380549050604051908082528060200260200182016040528015610df0578160200160208202803883390190505b50925060009150600090505b600354811015610ead5760008581526001602052604081206003805491929184908110610e2557fe5b6000918252602080832090910154600160a060020a0316835282019290925260400190205460ff1615610ea5576003805482908110610e6057fe5b6000918252602090912001548351600160a060020a0390911690849084908110610e8657fe5b600160a060020a03909216602092830290910190910152600191909101905b600101610dfc565b81604051908082528060200260200182016040528015610ed7578160200160208202803883390190505b509350600090505b81811015610f2c578281815181101515610ef557fe5b906020019060200201518482815181101515610f0d57fe5b600160a060020a03909216602092830290910190910152600101610edf565b505050919050565b60055481565b30600160a060020a031633600160a060020a0316141515610f5a57600080fd5b6003548160328211801590610f6f5750818111155b8015610f7a57508015155b8015610f8557508115155b1515610f9057600080fd5b60048390556040805184815290517fa3f1ee9126a074d9326c682f561767f710e927faa811f7a99829d49dc421797a9181900360200190a1505050565b33600160a060020a03811660009081526002602052604090205460ff161515610ff557600080fd5b6000828152602081905260409020548290600160a060020a0316151561101a57600080fd5b600083815260016020908152604080832033600160a060020a038116855292529091205484919060ff161561104e57600080fd5b6000858152600160208181526040808420600160a060020a0333168086529252808420805460ff1916909317909255905187927f4a504a94899432a9846e1aa406dceb1bcfd538bb839071d49d1e5e23f5be30ef91a36110ad856112d7565b5050505050565b60006110c1848484611506565b90506110cc81610fcd565b9392505050565b30600160a060020a031633600160a060020a03161415156110f357600080fd5b60068190556040805182815290517fc71bdc6afaf9b1aa90a7078191d4fc1adf3bf680fca3183697df6b0dc226bca29181900360200190a150565b603281565b60045481565b600030600160a060020a031633600160a060020a031614151561115b57600080fd5b600160a060020a038316600090815260026020526040902054839060ff16151561118457600080fd5b600160a060020a038316600090815260026020526040902054839060ff16156111ac57600080fd5b600092505b60035483101561123d5784600160a060020a03166003848154811015156111d457fe5b600091825260209091200154600160a060020a0316141561123257836003848154811015156111ff57fe5b9060005260206000200160006101000a815481600160a060020a030219169083600160a060020a0316021790555061123d565b6001909201916111b1565b600160a060020a03808616600081815260026020526040808220805460ff1990811690915593881682528082208054909416600117909355915190917f8001553a916ef2f495d26a907cc54d96ed840d7bda71e73194bf5a9df7a76b9091a2604051600160a060020a038516907ff39e6e1eb0edcf53c221607b54b00cd28f3196fed0a24994dc308b8f611b682d90600090a25050505050565b33600160a060020a0381166000908152600260205260408120549091829160ff16151561130357600080fd5b600084815260016020908152604080832033600160a060020a038116855292529091205485919060ff16151561133857600080fd5b600086815260208190526040902060030154869060ff161561135957600080fd5b6000878152602081905260409020955061137287610a6f565b945084806113a557506002808701546000196101006001831615020116041580156113a557506113a586600101546115f6565b156114f75760038601805460ff191660011790558415156113cf5760018601546008805490910190555b8560000160009054906101000a9004600160a060020a0316600160a060020a0316866001015487600201604051808280546001816001161561010002031660029004801561145e5780601f106114335761010080835404028352916020019161145e565b820191906000526020600020905b81548152906001019060200180831161144157829003601f168201915b505091505060006040518083038185875af192505050156114a95760405187907f33e13ecb54c3076d8e8bb8c2881800a4d972b792045ffae98fdf46df365fed7590600090a26114f7565b60405187907f526441bb6c1aba3c9a4a6ca1d6545da9c2333c8c48343ef398eb858d72b7923690600090a260038601805460ff191690558415156114f7576001860154600880549190910390555b50505050505050565b60085481565b600083600160a060020a038116151561151e57600080fd5b60055460408051608081018252600160a060020a0388811682526020808301898152838501898152600060608601819052878152808452959095208451815473ffffffffffffffffffffffffffffffffffffffff19169416939093178355516001830155925180519496509193909261159e926002850192910190611667565b50606091909101516003909101805460ff191691151591909117905560058054600101905560405182907fc0ba8fe4b176c1714197d43b9cc6bcf797a4a7461c5fe8d0ef6e184ae7601e5190600090a2509392505050565b60006007546201518001421115611611574260075560006008555b600654826008540111806116285750600854828101105b1561163557506000611639565b5060015b919050565b815481835581811115611662576000838152602090206116629181019083016116e5565b505050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106116a857805160ff19168380011785556116d5565b828001600101855582156116d5579182015b828111156116d55782518255916020019190600101906116ba565b506116e19291506116e5565b5090565b6108bb91905b808211156116e157600081556001016116eb5600a165627a7a72305820965b5c0e2cab6cc572131d36be2c6f27bc5571a0082b57f07a885634163f92010029a165627a7a723058209200b9af036552d5dc8ab972ef129b4e187012c62d579370efcfe543e26c88ba0029`

// DeployMultiSigWalletWithDailyLimitFactory deploys a new Ethereum contract, binding an instance of MultiSigWalletWithDailyLimitFactory to it.
func DeployMultiSigWalletWithDailyLimitFactory(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *MultiSigWalletWithDailyLimitFactory, error) {
	parsed, err := abi.JSON(strings.NewReader(MultiSigWalletWithDailyLimitFactoryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MultiSigWalletWithDailyLimitFactoryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MultiSigWalletWithDailyLimitFactory{MultiSigWalletWithDailyLimitFactoryCaller: MultiSigWalletWithDailyLimitFactoryCaller{contract: contract}, MultiSigWalletWithDailyLimitFactoryTransactor: MultiSigWalletWithDailyLimitFactoryTransactor{contract: contract}, MultiSigWalletWithDailyLimitFactoryFilterer: MultiSigWalletWithDailyLimitFactoryFilterer{contract: contract}}, nil
}

// MultiSigWalletWithDailyLimitFactory is an auto generated Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitFactory struct {
	MultiSigWalletWithDailyLimitFactoryCaller     // Read-only binding to the contract
	MultiSigWalletWithDailyLimitFactoryTransactor // Write-only binding to the contract
	MultiSigWalletWithDailyLimitFactoryFilterer   // Log filterer for contract events
}

// MultiSigWalletWithDailyLimitFactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitFactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletWithDailyLimitFactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitFactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletWithDailyLimitFactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MultiSigWalletWithDailyLimitFactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultiSigWalletWithDailyLimitFactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MultiSigWalletWithDailyLimitFactorySession struct {
	Contract     *MultiSigWalletWithDailyLimitFactory // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                        // Call options to use throughout this session
	TransactOpts bind.TransactOpts                    // Transaction auth options to use throughout this session
}

// MultiSigWalletWithDailyLimitFactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MultiSigWalletWithDailyLimitFactoryCallerSession struct {
	Contract *MultiSigWalletWithDailyLimitFactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                              // Call options to use throughout this session
}

// MultiSigWalletWithDailyLimitFactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MultiSigWalletWithDailyLimitFactoryTransactorSession struct {
	Contract     *MultiSigWalletWithDailyLimitFactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                              // Transaction auth options to use throughout this session
}

// MultiSigWalletWithDailyLimitFactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitFactoryRaw struct {
	Contract *MultiSigWalletWithDailyLimitFactory // Generic contract binding to access the raw methods on
}

// MultiSigWalletWithDailyLimitFactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitFactoryCallerRaw struct {
	Contract *MultiSigWalletWithDailyLimitFactoryCaller // Generic read-only contract binding to access the raw methods on
}

// MultiSigWalletWithDailyLimitFactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MultiSigWalletWithDailyLimitFactoryTransactorRaw struct {
	Contract *MultiSigWalletWithDailyLimitFactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMultiSigWalletWithDailyLimitFactory creates a new instance of MultiSigWalletWithDailyLimitFactory, bound to a specific deployed contract.
func NewMultiSigWalletWithDailyLimitFactory(address common.Address, backend bind.ContractBackend) (*MultiSigWalletWithDailyLimitFactory, error) {
	contract, err := bindMultiSigWalletWithDailyLimitFactory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitFactory{MultiSigWalletWithDailyLimitFactoryCaller: MultiSigWalletWithDailyLimitFactoryCaller{contract: contract}, MultiSigWalletWithDailyLimitFactoryTransactor: MultiSigWalletWithDailyLimitFactoryTransactor{contract: contract}, MultiSigWalletWithDailyLimitFactoryFilterer: MultiSigWalletWithDailyLimitFactoryFilterer{contract: contract}}, nil
}

// NewMultiSigWalletWithDailyLimitFactoryCaller creates a new read-only instance of MultiSigWalletWithDailyLimitFactory, bound to a specific deployed contract.
func NewMultiSigWalletWithDailyLimitFactoryCaller(address common.Address, caller bind.ContractCaller) (*MultiSigWalletWithDailyLimitFactoryCaller, error) {
	contract, err := bindMultiSigWalletWithDailyLimitFactory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitFactoryCaller{contract: contract}, nil
}

// NewMultiSigWalletWithDailyLimitFactoryTransactor creates a new write-only instance of MultiSigWalletWithDailyLimitFactory, bound to a specific deployed contract.
func NewMultiSigWalletWithDailyLimitFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*MultiSigWalletWithDailyLimitFactoryTransactor, error) {
	contract, err := bindMultiSigWalletWithDailyLimitFactory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitFactoryTransactor{contract: contract}, nil
}

// NewMultiSigWalletWithDailyLimitFactoryFilterer creates a new log filterer instance of MultiSigWalletWithDailyLimitFactory, bound to a specific deployed contract.
func NewMultiSigWalletWithDailyLimitFactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*MultiSigWalletWithDailyLimitFactoryFilterer, error) {
	contract, err := bindMultiSigWalletWithDailyLimitFactory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitFactoryFilterer{contract: contract}, nil
}

// bindMultiSigWalletWithDailyLimitFactory binds a generic wrapper to an already deployed contract.
func bindMultiSigWalletWithDailyLimitFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MultiSigWalletWithDailyLimitFactoryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MultiSigWalletWithDailyLimitFactory.Contract.MultiSigWalletWithDailyLimitFactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.MultiSigWalletWithDailyLimitFactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.MultiSigWalletWithDailyLimitFactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MultiSigWalletWithDailyLimitFactory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.contract.Transact(opts, method, params...)
}

// GetInstantiationCount is a free data retrieval call binding the contract method 0x8f838478.
//
// Solidity: function getInstantiationCount(creator address) constant returns(uint256)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryCaller) GetInstantiationCount(opts *bind.CallOpts, creator common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimitFactory.contract.Call(opts, out, "getInstantiationCount", creator)
	return *ret0, err
}

// GetInstantiationCount is a free data retrieval call binding the contract method 0x8f838478.
//
// Solidity: function getInstantiationCount(creator address) constant returns(uint256)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactorySession) GetInstantiationCount(creator common.Address) (*big.Int, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.GetInstantiationCount(&_MultiSigWalletWithDailyLimitFactory.CallOpts, creator)
}

// GetInstantiationCount is a free data retrieval call binding the contract method 0x8f838478.
//
// Solidity: function getInstantiationCount(creator address) constant returns(uint256)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryCallerSession) GetInstantiationCount(creator common.Address) (*big.Int, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.GetInstantiationCount(&_MultiSigWalletWithDailyLimitFactory.CallOpts, creator)
}

// Instantiations is a free data retrieval call binding the contract method 0x57183c82.
//
// Solidity: function instantiations( address,  uint256) constant returns(address)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryCaller) Instantiations(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimitFactory.contract.Call(opts, out, "instantiations", arg0, arg1)
	return *ret0, err
}

// Instantiations is a free data retrieval call binding the contract method 0x57183c82.
//
// Solidity: function instantiations( address,  uint256) constant returns(address)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactorySession) Instantiations(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.Instantiations(&_MultiSigWalletWithDailyLimitFactory.CallOpts, arg0, arg1)
}

// Instantiations is a free data retrieval call binding the contract method 0x57183c82.
//
// Solidity: function instantiations( address,  uint256) constant returns(address)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryCallerSession) Instantiations(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.Instantiations(&_MultiSigWalletWithDailyLimitFactory.CallOpts, arg0, arg1)
}

// IsInstantiation is a free data retrieval call binding the contract method 0x2f4f3316.
//
// Solidity: function isInstantiation( address) constant returns(bool)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryCaller) IsInstantiation(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MultiSigWalletWithDailyLimitFactory.contract.Call(opts, out, "isInstantiation", arg0)
	return *ret0, err
}

// IsInstantiation is a free data retrieval call binding the contract method 0x2f4f3316.
//
// Solidity: function isInstantiation( address) constant returns(bool)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactorySession) IsInstantiation(arg0 common.Address) (bool, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.IsInstantiation(&_MultiSigWalletWithDailyLimitFactory.CallOpts, arg0)
}

// IsInstantiation is a free data retrieval call binding the contract method 0x2f4f3316.
//
// Solidity: function isInstantiation( address) constant returns(bool)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryCallerSession) IsInstantiation(arg0 common.Address) (bool, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.IsInstantiation(&_MultiSigWalletWithDailyLimitFactory.CallOpts, arg0)
}

// Create is a paid mutator transaction binding the contract method 0x53d9d910.
//
// Solidity: function create(_owners address[], _required uint256, _dailyLimit uint256) returns(wallet address)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryTransactor) Create(opts *bind.TransactOpts, _owners []common.Address, _required *big.Int, _dailyLimit *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimitFactory.contract.Transact(opts, "create", _owners, _required, _dailyLimit)
}

// Create is a paid mutator transaction binding the contract method 0x53d9d910.
//
// Solidity: function create(_owners address[], _required uint256, _dailyLimit uint256) returns(wallet address)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactorySession) Create(_owners []common.Address, _required *big.Int, _dailyLimit *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.Create(&_MultiSigWalletWithDailyLimitFactory.TransactOpts, _owners, _required, _dailyLimit)
}

// Create is a paid mutator transaction binding the contract method 0x53d9d910.
//
// Solidity: function create(_owners address[], _required uint256, _dailyLimit uint256) returns(wallet address)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryTransactorSession) Create(_owners []common.Address, _required *big.Int, _dailyLimit *big.Int) (*types.Transaction, error) {
	return _MultiSigWalletWithDailyLimitFactory.Contract.Create(&_MultiSigWalletWithDailyLimitFactory.TransactOpts, _owners, _required, _dailyLimit)
}

// MultiSigWalletWithDailyLimitFactoryContractInstantiationIterator is returned from FilterContractInstantiation and is used to iterate over the raw logs and unpacked data for ContractInstantiation events raised by the MultiSigWalletWithDailyLimitFactory contract.
type MultiSigWalletWithDailyLimitFactoryContractInstantiationIterator struct {
	Event *MultiSigWalletWithDailyLimitFactoryContractInstantiation // Event containing the contract specifics and raw log

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
func (it *MultiSigWalletWithDailyLimitFactoryContractInstantiationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultiSigWalletWithDailyLimitFactoryContractInstantiation)
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
		it.Event = new(MultiSigWalletWithDailyLimitFactoryContractInstantiation)
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
func (it *MultiSigWalletWithDailyLimitFactoryContractInstantiationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultiSigWalletWithDailyLimitFactoryContractInstantiationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultiSigWalletWithDailyLimitFactoryContractInstantiation represents a ContractInstantiation event raised by the MultiSigWalletWithDailyLimitFactory contract.
type MultiSigWalletWithDailyLimitFactoryContractInstantiation struct {
	Sender        common.Address
	Instantiation common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterContractInstantiation is a free log retrieval operation binding the contract event 0x4fb057ad4a26ed17a57957fa69c306f11987596069b89521c511fc9a894e6161.
//
// Solidity: event ContractInstantiation(sender address, instantiation address)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryFilterer) FilterContractInstantiation(opts *bind.FilterOpts) (*MultiSigWalletWithDailyLimitFactoryContractInstantiationIterator, error) {

	logs, sub, err := _MultiSigWalletWithDailyLimitFactory.contract.FilterLogs(opts, "ContractInstantiation")
	if err != nil {
		return nil, err
	}
	return &MultiSigWalletWithDailyLimitFactoryContractInstantiationIterator{contract: _MultiSigWalletWithDailyLimitFactory.contract, event: "ContractInstantiation", logs: logs, sub: sub}, nil
}

// WatchContractInstantiation is a free log subscription operation binding the contract event 0x4fb057ad4a26ed17a57957fa69c306f11987596069b89521c511fc9a894e6161.
//
// Solidity: event ContractInstantiation(sender address, instantiation address)
func (_MultiSigWalletWithDailyLimitFactory *MultiSigWalletWithDailyLimitFactoryFilterer) WatchContractInstantiation(opts *bind.WatchOpts, sink chan<- *MultiSigWalletWithDailyLimitFactoryContractInstantiation) (event.Subscription, error) {

	logs, sub, err := _MultiSigWalletWithDailyLimitFactory.contract.WatchLogs(opts, "ContractInstantiation")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultiSigWalletWithDailyLimitFactoryContractInstantiation)
				if err := _MultiSigWalletWithDailyLimitFactory.contract.UnpackLog(event, "ContractInstantiation", log); err != nil {
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
