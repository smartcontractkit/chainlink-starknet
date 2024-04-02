# The starknet-devnet-rs container version - this version 
# supports rpc v0.7 which is needed for snfoundry v0.20.1
CONTAINER_VERSION="7743a089a33beb75d7012e4aa24745bee8ae0d71"

export TESTNET_ACCOUNTS_FILE=~/.starknet_accounts/starknet_open_zeppelin_accounts.json
export TESTNET_ACCOUNT_NAME=testnet-account

export DEVNET_ACCOUNTS_FILE=$(shell pwd)/accounts.json
export DEVNET_ACCOUNT_NAME=devnet-account

# General Commands

test:
	@snforge test

devnet:
	@printf "\nStarting a local starknet devnet docker container:\n\n" \
		&& CONTAINER_VERSION="$(CONTAINER_VERSION)" bash ../../../ops/scripts/devnet-hardhat.sh

# Account Management Commands

create-account:
	@printf "\nCreating starknet account details for testnet...\n\n" \
		&& sncast --profile testnet account create --name "$(TESTNET_ACCOUNT_NAME)" \
	  && printf "\nYour accounts:\n\n" \
		&& cat $(TESTNET_ACCOUNTS_FILE) \
		&& printf "\n\nYou can fund your account here: https://sepolia.starkgate.starknet.io\n"

add-account:
	@printf "\nImporting a prefunded account from starknet devnet container...\n\n" \
		&& sncast --profile devnet account add \
			--name "$(DEVNET_ACCOUNT_NAME)" \
			--address "0x4b3f4ba8c00a02b66142a4b1dd41a4dfab4f92650922a3280977b0f03c75ee1" \
			--private-key "0x57b2f8431c772e647712ae93cc616638" \
	  && printf "\nYour accounts:\n\n" \
		&& cat $(DEVNET_ACCOUNTS_FILE) \

deploy-account:
	@sncast --profile testnet account deploy --name "$(TESTNET_ACCOUNT_NAME)"

# MockAggregator Commands

ma-deploy:
	@cd ./scripts && sncast --profile "$(NETWORK)" script run deploy_mock_aggregator

ma-set-latest-round:
	@cd ./scripts && sncast --profile "$(NETWORK)" script run set_latest_round 

# Aggregator Commands

agg-read-latest-round:
	@cd ./scripts && sncast --profile "$(NETWORK)" script run read_latest_round

agg-read-decimals:
	@cd ./scripts && sncast --profile "$(NETWORK)" script run read_decimals

# AggregatorConsumer commands

ac-deploy:
	@cd ./scripts && sncast --profile "$(NETWORK)" script run deploy_aggregator_consumer

ac-read-answer:
	@cd ./scripts && sncast --profile "$(NETWORK)" script run read_answer

ac-set-answer:
	@cd ./scripts && sncast --profile "$(NETWORK)" script run set_answer

# Helpers

devnet-deploy:
	@make ma-deploy NETWORK=devnet && make ac-deploy NETWORK=devnet

