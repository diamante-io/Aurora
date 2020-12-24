# bifrost

> In Norse mythology, Bifröst (/ˈbɪvrɒst/ or sometimes Bilröst or Bivrost) is a burning rainbow bridge that reaches between Midgard (Earth) and Asgard, the realm of the gods. [**Wikipedia**](https://en.wikipedia.org/wiki/Bifr%C3%B6st)

Bifrost is highly available and secure Bitcoin/Ethereum → DiamNet bridge. It allows users to move BTC/ETH to DiamNet network and then trade them to other tokens or participate in ICOs (initial coin offering).

It solves many problems connected to moving tokens to DiamNet network:

* Security:
  * Developers don’t have access to users’ DiamNet keys.
  * No Bitcoin/Ethereum private keys for receiving accounts uploaded to any application machine.
* High availability:
  * Can be deployed to multiple availability zones (one or more machines can fail).
  * Handles high load.
* Easy to use:
  * Simple process for users.
  * Easy installation for developers.

We are releasing the **alpha version** of this software. We encourage our community of developers to test and improve it.

Download the binary from [the release page](https://github.com/diamnet/go/releases/tag/bifrost-v0.0.2) and use it with it's [Bifrost JS SDK](https://github.com/diamnet/bifrost-js-sdk).

## How it works

1. User opens the web app implemented using [Bifrost JS SDK](https://github.com/diamnet/bifrost-js-sdk).
1. User is presented with her public and private DiamNet keys where Bitcoin/Ethereum will be sent.
1. User selects what cryptocurrency she wants to move to DiamNet network.
1. A receiving Bitcoin/Ethereum address is generated.
1. User sends funds in Bitcoin/Ethereum network.
1. Bifrost listens to Bitcoin and Ethereum network events. When payment arrives it creates a DiamNet [account](https://www.diamnet.org/developers/guides/concepts/accounts.html) for the user who later adds a temporary signer on the account.
1. Using a temporary signer Bifrost creates necessary [trust lines](https://www.diamnet.org/developers/guides/concepts/assets.html) and sends corresponding amount of BTC/ETH to user's DiamNet account. It also exchanges BTC/ETH to the final token at a given rate.
1. Finally a temporary signer is removed.

## Demo

We use Ethereum Ropsten test network for this demo because it's faster that Bitcoin testnet.

https://bifrost.diamnet.org/

1. First you need some ETH on Ethereum testnet.
1. Create an account at https://www.myetherwallet.com/, then switch Network (top-right dropdown) to "Ropsten (MyEtherWallet)". Write down/copy your Ethereum address somewhere.
1. Use http://faucet.ropsten.be:3001/ to send 3 ETH to your testnet account.
1. Now open Bifrost demo: https://bifrost.diamnet.org/ It will display an Ethereum address for sending your test ETH.
1. Go back to MyEtherWallet and from your testnet account send ETH to the address displayed by Bifrost. Sometimes Ropsten network is overloaded so monitor the Etherscan to check if your tx was included in a block. If not, increase the gas price (this can be done in "Send Offline" tab).
1. Switch back to Bifrost demo and check the progress.

## Config

* `port` - bifrost server listening port
* `using_proxy` (default `false`) - set to `true` if bifrost lives behind a proxy or load balancer
* `access_control_allow_origin_header` (default `*`) - value of `Access-Control-Allow-Origin` headers sent by Bifrost
* `bitcoin`
  * `master_public_key` - master public key for bitcoin keys derivation (read more in [BIP-0032](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki))
  * `rpc_server` - URL of [bitcoin-core](https://github.com/bitcoin/bitcoin) >= 0.15.0 RPC server
  * `rpc_user` (default empty) - username for RPC server (if any)
  * `rpc_pass` (default empty) - password for RPC server (if any)
  * `testnet` (default `false`) - set to `true` if you're testing bifrost in ethereum
  * `minimum_value_btc` - minimum transaction value in BTC that will be accepted by Bifrost, everything below will be ignored.
  * `token_price` - a price of one BTC in terms of final token (`diamnet.token_asset_code`)
* `ethereum`
  * `master_public_key` - master public key for ethereum keys derivation (read more in [BIP-0032](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki))
  * `rpc_server` - URL of [geth](https://github.com/ethereum/go-ethereum) >= 1.7.1 RPC server
  * `network_id` - network ID (`3` - Ropsten testnet, `1` - live Ethereum network)
  * `minimum_value_eth` - minimum transaction value in ETH that will be accepted by Bifrost, everything below will be ignored.
  * `token_price` - a price of one ETH in terms of final token (`diamnet.token_asset_code`)
* `diamnet`
  * `token_asset_code` - asset code for the token that will be distributed
  * `issuer_public_key` - public key of the assets issuer (see "Account configuration" section),
  * `distribution_public_key` - public key of the distribution account, it can be the same as issuer but it's recommended to use a separate account, it's also used to fund new accounts (see "Account configuration" section)
  * `signer_secret_key` - distribuions accounts's secret key if only one instance of Bifrost is deployed OR [channel](https://www.diamnet.org/developers/guides/channels.html)'s secret key if more than one instance of Bifrost is deployed, signer is also used as a temporary signer in new accounts (see "Account configuration" section)
  * `aurora` - URL to [aurora](https://github.com/diamnet/go/tree/master/services/aurora) server
  * `network_passphrase` - DiamNet network passphrase (`Public Global DiamNet Network ; September 2015` for production network, `Test SDF Network ; September 2015` for test network)
  * `starting_balance` - DiamNet XLM amount issued to created account (41 by default)
  * `lock_unix_timestamp` - Unix timestamp (in seconds) of a date until funds will be locked in a new account (helpful if you want to disallow trading during token sale)
* `database`
  * `type` - currently the only supported database type is: `postgres`
  * `dsn` - data source name for postgres connection (`postgres://user:password@host/dbname?sslmode=sslmode` - [more info](https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters))

## Deployment

There are two ways you can deploy Bifrost:
* **Simple/single-instance deployment** - Bifrost will live on a single instance along with geth, bitcoin-core and database. This is the easiest way to deploy Bifrost but if this instance goes down, so does Bifrost and all the other components.
* **High-availability deployment** - Each of the components is deployed to at least 2 instances so even if random instances go down, Bifrost will still be online accepting payments.

We recommend the latter.

### High-availability deployment

Here's the proposed architecture diagram of high-availability deployment:

![Architecture](./images/architecture.png)

## Account configuration

Issuing, Distribution and Signer Accounts can be configured in many different ways, it all depends on your needs.

### Issuing Account Only

This is the most basic configuration but it is not recommended. In this configuration only one account is responsible for everything: issuing and distributing assets and creating new accounts. In this configuration `diamnet.issuer_public_key`, `diamnet.distribution_public_key` and `diamnet.signer_secret_key` represent a single account.

### Issuing and Distribution Accounts

Sometimes you want to have a limited supply of your token so you create a distribution account, send all the tokens to distribution (and other accounts, ex. vesting) and lock the issuing account so no new tokens can be created. In this configuration `diamnet.distribution_public_key` and `diamnet.signer_secret_key` represent a separate account used to distribute funds in a token sale. This account will also be used to create new accounts.

### Distribution Account with multiple signers

If you expect really high load you should deploy multiple Bifrost servers. To prevent Bifrost instances from consuming the same sequence numbers of the Distribution account you should use [channels](https://www.diamnet.org/developers/guides/channels.html). Each Bifrost server will contain a secret key of one of the signers of Distribution account and will use Signer account as a [transaction source](https://www.diamnet.org/developers/guides/concepts/transactions.html#source-account) when creating new accounts.

In this configuration all servers have the same `diamnet.issuer_public_key` and `diamnet.distribution_public_key` values but each server has a separate `diamnet.signer_secret_key` which is a signer of `diamnet.distribution_public_key` account (and needs to exist in DiamNet ledger).

![Accounts](./images/accounts.png)

## Going to production

* Remember that everyone with master public key and **any** child private key can recover your **master** private key. Do not share your master public key and obviously any private keys. Treat your master public key as if it was a private key. Read more in BIP-0032 [Security](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki#security) section.
* Make sure "Sell [your token] for BTC" and/or "Sell [your token] for ETH" exist in DiamNet production network. If not, you can create an offer by sending a transaction with `manage_offer` operation.
* Make sure you don't use account from `diamnet.issuer_secret_key` anywhere else than bifrost. Otherwise, sequence numbers will go out of sync and bifrost will stop working. It's good idea to create a new signer on issuing account.
* Check if public master key is correct. Use CLI tool (`bifrost check-keys`) to generate a few addresses and ensure you have corresponding private keys! You should probably send test transactions to some of these addresses and check if you can withdraw funds.
* Make sure `using_proxy` variable is set to correct value. Otherwise you will see your proxy IP instead of users' IPs in logs.
* Make sure you're not connecting to testnets.
* Deploy at least 2 bifrost, bitcoin-core, geth, diamnet-core and aurora servers. Use multi-AZ database.
* Do not use SDF's aurora servers. There is no SLA and we cannot guarantee it will handle your load.
* Make sure bifrost <-> bitcoin-core and bifrost <-> geth connections are not public or are encrypted (mitm attacks).
* It's a good idea to set "Authorization required" and "Authorization revocable" [flags](https://www.diamnet.org/developers/guides/concepts/accounts.html#flags) on issuing account during ICO stage to remove trustlines to accounts with lost keys. You need to set `diamnet.needs_authorize` in config to `true` if you want to use this feature.
* Monitor bifrost logs and react to all WARN and ERROR entries.
* Make sure you are using geth >= 1.7.1 and bitcoin-core >= 0.15.0.
* Increase aurora rate limiting to handle expected load.
* Make sure you configured minimum accepted value for Bitcoin and Ethereum transactions to the value you really want.
* Make sure you start from a fresh Bifrost DB in production. If Bifrost was running, you stopped it and then started it again, all the Bitcoin and Ethereum blocks mined during that period will be processed which can take a lot of time.
* Make sure that `lock_unix_timestamp` is correct and test it. If the value is too far in the future users will be locked out of their funds.

## FAQ

#### Transactions are not processed by Bifrost

Here is a list of possible reasons why Bifrost may not process transactions:

* Bifrost is detecting transactions but there's a proxy between client and Bifrost server and it's blocking SSE (Server-Sent Events) sent by Bifrost. Solution for this can be found [here](https://serverfault.com/a/801629).
* There's a long queue of blocks Bifrost needs to process. During the first start of Bifrost, it starts processing blocks from the latest block available. If Bifrost was turned off for a longer time or `geth` is still catchup up it may cause delays. To solve this, recreate Bifrost DB so it resets data or (if you don't want to delete data in a DB) set `ethereum_last_block` value in `key_value_store` table to `0`.
* You are sending a transaction with amount below `minimum_value_eth`.
* Geth is connected to one network and you're sending ETH in another network.

#### How do I unlock the account when `lock_unix_timestamp` is set?

You need to submit a special transaction that "unlocks" the account by removing the temporary signer and adding back user's master key. This can be done by everyone: user or organization that runs Bifrost. Transaction can be found in:

* `Bifrost.ExchangedTimelockedEvent` [event](https://github.com/diamnet/bifrost-js-sdk#oneventevent-data). You can display it to the user if you want your users to unlock accounts.
* In database: `recovery_transaction` table. These transactions are independent so you can submit them in any order / at the same time.

#### I don't see answer to my question. What do I do?

Check [DiamNet Stack Exchange](https://diamnet.stackexchange.com/questions/tagged/bifrost/) to find similar questions or add the new question with the `bifrost` tag.
