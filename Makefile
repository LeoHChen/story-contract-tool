# Makefile for Story Smart Contract CLI Tool

# Configuration variables
BINARY_NAME = ssc-tool
GO_FILES = main.go
RPC_URL = https://mainnet.storyrpc.io

# Contract addresses with labels
# Format: label=address
CONTRACTS = \
	wip=0x1514000000000000000000000000000000000000	\
	usdc=0xF1815bd50389c46847f0Bda824eC8da914045D14

# Common functions
COMMON_FUNCTIONS = \
	erc20_name="name" \
	erc20_symbol="symbol" \
	erc20_decimals="decimals" \
	erc20_totalSupply="totalSupply"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(GO_FILES)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)

# Run with specific contract and function
# Usage: make run CONTRACT=wip FUNCTION=totalSupply [ARGS=0x123...] [CONVERT=true] [TYPE=erc20]
run: build
	@if [ -z "$(CONTRACT)" ]; then \
		echo "Error: CONTRACT parameter is required"; \
		echo "Usage: make run CONTRACT=wip FUNCTION=totalSupply [ARGS=0x123...] [CONVERT=true] [TYPE=erc20]"; \
		exit 1; \
	fi; \
	if [ -z "$(FUNCTION)" ]; then \
		echo "Error: FUNCTION parameter is required"; \
		echo "Usage: make run CONTRACT=wip FUNCTION=totalSupply [ARGS=0x123...] [CONVERT=true] [TYPE=erc20]"; \
		exit 1; \
	fi; \
	CONTRACT_ADDR=$$(echo "$(CONTRACTS)" | grep -o "$(CONTRACT)=[^[:space:]]*" | cut -d= -f2); \
	if [ -z "$$CONTRACT_ADDR" ]; then \
		echo "Error: Contract '$(CONTRACT)' not found in predefined list"; \
		echo "Available contracts: $$(echo "$(CONTRACTS)" | tr ' ' '\n' | cut -d= -f1 | tr '\n' ' ')"; \
		exit 1; \
	fi; \
	CONVERT_FLAG=""; \
	if [ "$(CONVERT)" = "true" ]; then \
		CONVERT_FLAG="-convert"; \
	fi; \
	TYPE_FLAG=""; \
	if [ -n "$(TYPE)" ]; then \
		TYPE_FLAG="-type=$(TYPE)"; \
	else \
		TYPE_FLAG="-type=erc20"; \
	fi; \
	ARGS_FLAG=""; \
	if [ -n "$(ARGS)" ]; then \
		ARGS_FLAG="-args=$(ARGS)"; \
	fi; \
	echo "Running command for $(CONTRACT) ($$CONTRACT_ADDR)..."; \
	./$(BINARY_NAME) -contract=$$CONTRACT_ADDR -function=$(FUNCTION) -rpc=$(RPC_URL) $$TYPE_FLAG $$ARGS_FLAG $$CONVERT_FLAG

# Show all available contracts
list-contracts:
	@echo "Available contracts:"
	@echo "$(CONTRACTS)" | tr ' ' '\n' | sort | column -t -s=

# Show all available common functions
list-functions:
	@echo "Common functions:"
	@echo "$(COMMON_FUNCTIONS)" | tr ' ' '\n' | sort | column -t -s=

# Run a common function for a specific contract
# Usage: make run-common ABI=abi/contract.json CONTRACT=wip COMMON_FUNC=erc20_symbol [CONVERT=true]
run-common: build
	@if [ -z "$(CONTRACT)" ]; then \
		echo "Error: CONTRACT parameter is required"; \
		echo "Usage: make run-common CONTRACT=wip COMMON_FUNC=erc20_symbol [CONVERT=true]"; \
		exit 1; \
	fi; \
	if [ -z "$(COMMON_FUNC)" ]; then \
		echo "Error: COMMON_FUNC parameter is required"; \
		echo "Usage: make run-common CONTRACT=wip COMMON_FUNC=erc20_symbol [CONVERT=true]"; \
		exit 1; \
	fi; \
	CONTRACT_ADDR=$$(echo "$(CONTRACTS)" | grep -o "$(CONTRACT)=[^[:space:]]*" | cut -d= -f2); \
	if [ -z "$$CONTRACT_ADDR" ]; then \
		echo "Error: Contract '$(CONTRACT)' not found in predefined list"; \
		echo "Available contracts: $$(echo "$(CONTRACTS)" | tr ' ' '\n' | cut -d= -f1 | tr '\n' ' ')"; \
		exit 1; \
	fi; \
	FUNC_NAME=$$(echo "$(COMMON_FUNCTIONS)" | grep -o "$(COMMON_FUNC)=[^[:space:]]*" | cut -d= -f2 | tr -d '"'); \
	if [ -z "$$FUNC_NAME" ]; then \
		echo "Error: Common function '$(COMMON_FUNC)' not found"; \
		echo "Available common functions: $$(echo "$(COMMON_FUNCTIONS)" | tr ' ' '\n' | cut -d= -f1 | tr '\n' ' ')"; \
		exit 1; \
	fi; \
	CONVERT_FLAG=""; \
	if [ "$(CONVERT)" = "true" ]; then \
		CONVERT_FLAG="-convert"; \
	fi; \
	ABI_FLAG=""; \
	if [ -n "$(ABI)" ]; then \
		ABI_FLAG="-abi=$(ABI)"; \
	fi; \
	echo "Running function '$$FUNC_NAME' for $(CONTRACT) ($$CONTRACT_ADDR)..."; \
	./$(BINARY_NAME) -contract=$$CONTRACT_ADDR -function=$$FUNC_NAME -rpc=$(RPC_URL) -type=erc20 $$CONVERT_FLAG $$ABI_FLAG

# Display balance of an address in a token contract
# Usage: make balance CONTRACT=wip ADDRESS=0x123...
balance: build
	@if [ -z "$(CONTRACT)" ]; then \
		echo "Error: CONTRACT parameter is required"; \
		echo "Usage: make balance CONTRACT=wip ADDRESS=0x123..."; \
		exit 1; \
	fi; \
	if [ -z "$(ADDRESS)" ]; then \
		echo "Error: ADDRESS parameter is required"; \
		echo "Usage: make balance CONTRACT=wip ADDRESS=0x123..."; \
		exit 1; \
	fi; \
	CONTRACT_ADDR=$$(echo "$(CONTRACTS)" | grep -o "$(CONTRACT)=[^[:space:]]*" | cut -d= -f2); \
	if [ -z "$$CONTRACT_ADDR" ]; then \
		echo "Error: Contract '$(CONTRACT)' not found in predefined list"; \
		echo "Available contracts: $$(echo "$(CONTRACTS)" | tr ' ' '\n' | cut -d= -f1 | tr '\n' ' ')"; \
		exit 1; \
	fi; \
	echo "Checking balance of $(ADDRESS) in $(CONTRACT) ($$CONTRACT_ADDR)..."; \
	./$(BINARY_NAME) -contract=$$CONTRACT_ADDR -function=balanceOf -args=$(ADDRESS) -rpc=$(RPC_URL) -type=erc20 -convert

# Quick commands for common operations
wip-info: build
	@echo "WIP Token Info:"
	@$(MAKE) run-common CONTRACT=wip COMMON_FUNC=erc20_name
	@$(MAKE) run-common CONTRACT=wip COMMON_FUNC=erc20_symbol
	@$(MAKE) run-common CONTRACT=wip COMMON_FUNC=erc20_decimals
	@$(MAKE) run-common CONTRACT=wip COMMON_FUNC=erc20_totalSupply CONVERT=true

help:
	@echo "Story Smart Contract CLI Tool Makefile"
	@echo ""
	@echo "Commands:"
	@echo "  make build                   - Build the binary"
	@echo "  make clean                   - Remove the binary"
	@echo "  make list-contracts          - List all configured contracts"
	@echo "  make list-functions          - List all common functions"
	@echo "  make run CONTRACT=wip FUNCTION=totalSupply [ARGS=0x123] [CONVERT=true] [TYPE=erc20]"
	@echo "                               - Run a specific function on a contract"
	@echo "  make run-common CONTRACT=wip COMMON_FUNC=erc20_symbol [CONVERT=true]"
	@echo "                               - Run a predefined common function on a contract"
	@echo "  make balance CONTRACT=wip ADDRESS=0x123..."
	@echo "                               - Check token balance of an address"
	@echo "  make wip-info               - Show DAI token information"
	@echo ""
	@echo "Examples:"
	@echo "  make run CONTRACT=wip FUNCTION=totalSupply CONVERT=true"
	@echo "  make run-common CONTRACT=usdc COMMON_FUNC=erc20_symbol"
	@echo "  make balance CONTRACT=wip ADDRESS=0x3EF98543F9772DC959255545B717a61D408e7b61"

.PHONY: build clean run list-contracts list-functions run-common balance wip-info help
