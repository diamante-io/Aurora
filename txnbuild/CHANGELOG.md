# Changelog

All notable changes to this project will be documented in this
file.  This project adheres to [Semantic Versioning](http://semver.org/).

## Unreleased

* Add `Transaction.BuildChallengeTx` method for building [SEP-10](https://github.com/hcnet/hcnet-protocol/blob/master/ecosystem/sep-0010.md) challenge transaction.


## [v1.3.0](https://github.com/hcnet/go/releases/tag/auroraclient-v1.3.0) - 2019-07-08

* Add support for getting the hex-encoded transaction hash with `Transaction.HashHex` method.
* `TransactionEnvelope` is now available after building a transaction(`Transaction.Build`). Previously, this was only available after signing a transaction. ([#1376](https://github.com/hcnet/go/pull/1376)) 
* Add support for getting the `TransactionEnvelope` struct with `Transaction.TxEnvelope` method ([#1415](https://github.com/hcnet/go/issues/1415)).
* `AllowTrust` operations no longer requires the asset issuer, only asset code is required ([#1330](https://github.com/hcnet/go/issues/1330)).
* `Transaction.SetDefaultFee` method is deprecated and will be removed in the next major release ([#1221](https://github.com/hcnet/go/issues/1221)).
* `Transaction.TransactionFee` method has been added to get the fee that will be paid for a transaction.
* `Transaction.SignHashX` method adds support for signing transactions with hash(x) signature types.

## [v1.2.0](https://github.com/hcnet/go/releases/tag/auroraclient-v1.2.0) - 2019-05-16

* In addition to account responses from aurora, transactions and operations can now be built with txnbuild.SimpleAccount structs constructed locally ([#1266](https://github.com/hcnet/go/issues/1266)). 
* Added `MaxTrustlineLimit` which represents the maximum value for a trustline limit ([#1265](https://github.com/hcnet/go/issues/1265)).
* ChangeTrust operation with no `Limit` field set now defaults to `MaxTrustlineLimit` ([#1265](https://github.com/hcnet/go/issues/1265)).
* Add support for building `ManageBuyOffer` operation ([#1165](https://github.com/hcnet/go/issues/1165)).
* Fix bug in ChangeTrust operation builder ([1296](https://github.com/hcnet/go/issues/1296)).

## [v1.1.0](https://github.com/hcnet/go/releases/tag/auroraclient-v1.1.0) - 2019-02-02

* Support for multiple signatures ([#1198](https://github.com/hcnet/go/pull/1198))

## [v1.0.0](https://github.com/hcnet/go/releases/tag/auroraclient-v1.0) - 2019-04-26

* Initial release
