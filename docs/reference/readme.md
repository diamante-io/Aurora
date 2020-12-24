---
title: Overview
---

The Go SDK is a set of packages for interacting with most aspects of the DiamNet ecosystem. The primary component is the Aurora SDK, which provides convenient access to Aurora services. There are also packages for other DiamNet services such as [TOML support](https://github.com/diamnet/diamnet-protocol/blob/master/ecosystem/sep-0001.md) and [federation](https://github.com/diamnet/diamnet-protocol/blob/master/ecosystem/sep-0002.md).

## Aurora SDK

The Aurora SDK is composed of two complementary libraries: `txnbuild` + `auroraclient`.
The `txnbuild` ([source](https://github.com/diamnet/go/tree/master/txnbuild), [docs](https://godoc.org/github.com/diamnet/go/txnbuild)) package enables the construction, signing and encoding of DiamNet [transactions](https://www.diamnet.org/developers/guides/concepts/transactions.html) and [operations](https://www.diamnet.org/developers/guides/concepts/list-of-operations.html) in Go. The `auroraclient` ([source](https://github.com/diamnet/go/tree/master/clients/auroraclient), [docs](https://godoc.org/github.com/diamnet/go/clients/auroraclient)) package provides a web client for interfacing with [Aurora](https://www.diamnet.org/developers/guides/get-started/) server REST endpoints to retrieve ledger information, and to submit transactions built with `txnbuild`.

## List of major SDK packages

- `auroraclient` ([source](https://github.com/diamnet/go/tree/master/clients/auroraclient), [docs](https://godoc.org/github.com/diamnet/go/clients/auroraclient)) - programmatic client access to Aurora
- `txnbuild` ([source](https://github.com/diamnet/go/tree/master/txnbuild), [docs](https://godoc.org/github.com/diamnet/go/txnbuild)) - construction, signing and encoding of DiamNet transactions and operations
- `diamnettoml` ([source](https://github.com/diamnet/go/tree/master/clients/auroraclient), [docs](https://godoc.org/github.com/diamnet/go/clients/diamnettoml)) - parse [DiamNet.toml](../../guides/concepts/diamnet-toml.md) files from the internet
- `federation` ([source](https://godoc.org/github.com/diamnet/go/clients/federation)) - resolve federation addresses  into diamnet account IDs, suitable for use within a transaction

