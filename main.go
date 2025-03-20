package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GenericContract is a generic binding to an Ethereum contract
type GenericContract struct {
	contract *bind.BoundContract
	abi      string
}

// NewGenericContract creates a new instance of a generic contract binding
func NewGenericContract(address common.Address, abiString string, backend bind.ContractBackend) (*GenericContract, error) {
	parsed, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}
	
	contract := bind.NewBoundContract(address, parsed, backend, backend, nil)
	return &GenericContract{
		contract: contract,
		abi:      abiString,
	}, nil
}

// CallViewFunction calls a view function on the contract and returns the result
func (gc *GenericContract) CallViewFunction(functionName string, args ...interface{}) ([]interface{}, error) {
	var out []interface{}
	err := gc.contract.Call(&bind.CallOpts{Context: context.Background()}, &out, functionName, args...)
	return out, err
}

// Example ABIs for common contract types
var (
	// Simple ERC20 token ABI with basic view functions
	ERC20ABI = `[
		{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function"},
		{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"type":"function"},
		{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"type":"function"},
		{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"type":"function"},
		{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}
	]`
	
	// Simple storage contract ABI
	SimpleStorageABI = `[
		{
			"inputs": [],
			"name": "retrieve",
			"outputs": [{"internalType": "uint256", "name": "", "type": "uint256"}],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [{"internalType": "uint256", "name": "num", "type": "uint256"}],
			"name": "store",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]`
	
	// Generic fallback ABI for just function name calls (without knowing the full ABI)
	GenericFallbackABI = `[
		{
			"inputs": [],
			"name": "FUNCTION_PLACEHOLDER",
			"outputs": [{"internalType": "uint256", "name": "", "type": "uint256"}],
			"stateMutability": "view",
			"type": "function"
		}
	]`
)

// Helper function to determine if a string is in a slice
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// Helper function to get contract ABI based on contract type
func getContractABI(contractType string) string {
	switch contractType {
	case "erc20":
		return ERC20ABI
	case "storage":
		return SimpleStorageABI
	default:
		return GenericFallbackABI
	}
}

func main() {
	// Define command line flags
	contractAddrPtr := flag.String("contract", "", "Ethereum smart contract address (required)")
	functionNamePtr := flag.String("function", "", "Smart contract function name to call (required)")
	rpcURLPtr := flag.String("rpc", "https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID", "Ethereum RPC URL")
	contractTypePtr := flag.String("type", "generic", "Contract type (erc20, storage, generic)")
	abiFilePtr := flag.String("abi", "", "Path to ABI JSON file (optional)")
	argsPtr := flag.String("args", "", "Function arguments (comma separated)")
	convertPtr := flag.Bool("convert", false, "Convert big.Int results to decimal using 10^18 denominator (for tokens)")
	helpPtr := flag.Bool("help", false, "Display help information")
	
	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -contract=0xContractAddress -function=functionName [-rpc=https://your-ethereum-node] [-type=erc20|storage|generic] [-abi=path/to/abi.json] [-args=arg1,arg2,...] [-convert]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  Call a function without arguments:\n")
		fmt.Fprintf(os.Stderr, "    %s -contract=0x123... -function=totalSupply -type=erc20\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Call a function with an address argument:\n")
		fmt.Fprintf(os.Stderr, "    %s -contract=0x123... -function=balanceOf -type=erc20 -args=0x456...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Call a function and convert result to decimal:\n")
		fmt.Fprintf(os.Stderr, "    %s -contract=0x123... -function=balanceOf -type=erc20 -args=0x456... -convert\n", os.Args[0])
	}
	
	// Parse command line arguments
	flag.Parse()
	
	// Check if help flag is set
	if *helpPtr {
		flag.Usage()
		os.Exit(0)
	}
	
	// Validate required parameters
	if *contractAddrPtr == "" {
		fmt.Println("Error: Contract address is required")
		flag.Usage()
		os.Exit(1)
	}
	
	if *functionNamePtr == "" {
		fmt.Println("Error: Function name is required")
		flag.Usage()
		os.Exit(1)
	}
	
	// Validate contract address format
	if !common.IsHexAddress(*contractAddrPtr) {
		fmt.Printf("Error: Invalid contract address format: %s\n", *contractAddrPtr)
		os.Exit(1)
	}
	
	// Connect to an Ethereum node
	client, err := ethclient.Dial(*rpcURLPtr)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	
	fmt.Printf("Connected to Ethereum node at %s\n", *rpcURLPtr)
	
	// Convert the address string to the correct format
	contractAddress := common.HexToAddress(*contractAddrPtr)
	fmt.Printf("Using contract address: %s\n", contractAddress.Hex())
	
	// Get the appropriate ABI
	var abiString string
	
	if *abiFilePtr != "" {
		// Read ABI from file
		abiData, err := os.ReadFile(*abiFilePtr)
		if err != nil {
			log.Fatalf("Failed to read ABI file: %v", err)
		}
		abiString = string(abiData)
	} else {
		// Use predefined ABI based on contract type
		abiString = getContractABI(*contractTypePtr)
		
		// If using generic ABI, replace function placeholder with actual function name
		if *contractTypePtr == "generic" || *contractTypePtr == "" {
			abiString = strings.Replace(abiString, "FUNCTION_PLACEHOLDER", *functionNamePtr, 1)
		}
	}
	
	// Create a new instance of the generic contract bound to the specific deployed contract
	contract, err := NewGenericContract(contractAddress, abiString, client)
	if err != nil {
		log.Fatalf("Failed to instantiate the contract: %v", err)
	}
	
	// Prepare function arguments
	var functionArgs []interface{}
	if *argsPtr != "" {
		argsList := strings.Split(*argsPtr, ",")
		for _, arg := range argsList {
			// Try to detect argument type and convert accordingly
			arg = strings.TrimSpace(arg)
			
			// Check if it's an address
			if common.IsHexAddress(arg) {
				functionArgs = append(functionArgs, common.HexToAddress(arg))
			} else if strings.HasPrefix(arg, "0x") {
				// Assume it's a hex number
				bigInt, success := new(big.Int).SetString(arg[2:], 16)
				if !success {
					log.Fatalf("Failed to parse hex argument: %s", arg)
				}
				functionArgs = append(functionArgs, bigInt)
			} else {
				// Try to parse as integer
				bigInt, success := new(big.Int).SetString(arg, 10)
				if success {
					functionArgs = append(functionArgs, bigInt)
				} else {
					// Assume it's a string
					functionArgs = append(functionArgs, arg)
				}
			}
		}
	}
	
	// Call the view function
	fmt.Printf("Calling function '%s' with %d arguments\n", *functionNamePtr, len(functionArgs))
	results, err := contract.CallViewFunction(*functionNamePtr, functionArgs...)
	if err != nil {
		log.Fatalf("Failed to call function '%s': %v", *functionNamePtr, err)
	}
	
	// Display results
	fmt.Println("Function returned successfully!")
	
	// Create 10^18 constant for potential conversion
	exp18 := big.NewInt(10)
	exp18.Exp(exp18, big.NewInt(18), nil)
	
	for i, result := range results {
		fmt.Printf("Result[%d]: ", i)
		
		switch v := result.(type) {
		case *big.Int:
			fmt.Printf("%s (big.Int)\n", v.String())
			
			// Also display value / 10^18 (common for token amounts)
			if *convertPtr && v.Cmp(big.NewInt(0)) > 0 {
				// Calculate integer part (v / 10^18)
				intPart := new(big.Int).Div(new(big.Int).Set(v), exp18)
				
				// Calculate fractional part (v % 10^18)
				fracPart := new(big.Int).Mod(new(big.Int).Set(v), exp18)
				
				// Format the fractional part with leading zeros
				fracStr := fracPart.String()
				for len(fracStr) < 18 {
					fracStr = "0" + fracStr
				}
				
				// Trim trailing zeros
				for len(fracStr) > 0 && fracStr[len(fracStr)-1] == '0' {
					fracStr = fracStr[:len(fracStr)-1]
				}
				
				if len(fracStr) > 0 {
					fmt.Printf("           = %s.%s (decimal)\n", intPart.String(), fracStr)
				} else {
					fmt.Printf("           = %s (decimal)\n", intPart.String())
				}
			}
		case string:
			fmt.Printf("%s (string)\n", v)
		case []byte:
			fmt.Printf("0x%x (bytes)\n", v)
		case common.Address:
			fmt.Printf("%s (address)\n", v.Hex())
		case bool:
			fmt.Printf("%t (bool)\n", v)
		default:
			fmt.Printf("%v (type: %T)\n", v, v)
		}
	}
}
