ACCOUNT_NAME="test-acct"

devnet:
	@printf "\nStarting a local starknet devnet docker container:\n\n" \
		&& bash ../../../ops/scripts/devnet-hardhat-down.sh \
		&& bash ../../../ops/scripts/devnet-hardhat.sh \
		&& printf "\n\nIf you intend to use a local starknet devnet container to run these examples,\nadd \"url = http://localhost:5050/rpc\" in the snfoundry.toml file\n\n"

create-account:
	@printf "\nCreating an account named \"$(ACCOUNT_NAME)\"...\n\n" \
		&& sncast account create --name $(ACCOUNT_NAME) \
	  && printf "\nYour accounts:\n\n" \
		&& cat ~/.starknet_accounts/starknet_open_zeppelin_accounts.json \
		&& printf "\n\nYou can fund your account here: https://sepolia.starkgate.starknet.io\n"

add-account:
	@printf "\nAdding an account named \"$(ACCOUNT_NAME)\"...\n\n" \
		&& sncast \
			--url "http://127.0.0.1:5050/rpc" \
			account add \
			--name $(ACCOUNT_NAME) \
			--address "0x4b3f4ba8c00a02b66142a4b1dd41a4dfab4f92650922a3280977b0f03c75ee1" \
			--private-key "0x57b2f8431c772e647712ae93cc616638" \
	  && printf "\nYour accounts:\n\n" \
		&& cat ~/.starknet_accounts/starknet_open_zeppelin_accounts.json

view-accounts:
	@cat ~/.starknet_accounts/starknet_open_zeppelin_accounts.json

deploy-account:
	@sncast account deploy --name $(ACCOUNT_NAME)

delete-account:
	@sncast account delete --name $(ACCOUNT_NAME)

tc-deploy:
	@cd ./scripts && sncast --account "$(ACCOUNT_NAME)" script run deploy_test_consumer

tc-set-latest-round:
	@cd ./scripts && sncast --account "$(ACCOUNT_NAME)" script run set_latest_round 

tc-read-latest-round:
	@cd ./scripts && sncast script run read_latest_round

tc-read-decimals:
	@cd ./scripts && sncast script run read_decimals

