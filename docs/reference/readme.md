---
title: Overview
---

The Go SDK is a set of packages for interacting with most aspects of the Diamnet ecosystem. The primary component is the Aurora SDK, which provides convenient access to Aurora services. There are also packages for other Diamnet services such as [TOML support](https://github.com/diamnet/diamnet-protocol/blob/master/ecosystem/sep-0001.md) and [federation](https://github.com/diamnet/diamnet-protocol/blob/master/ecosystem/sep-0002.md).

## Aurora SDK

The Aurora SDK is composed of two complementary libraries: `txnbuild` + `auroraclient`.
The `txnbuild` ([source](https://github.com/diamnet/go/tree/master/txnbuild), [docs](https://godoc.org/github.com/diamnet/go/txnbuild)) package enables the construction, signing and encoding of Diamnet [transactions](https://developers.diamnet.org/docs/glossary/transactions/) and [operations](https://developers.diamnet.org/docs/start/list-of-operations/) in Go. The `auroraclient` ([source](https://github.com/diamnet/go/tree/master/clients/auroraclient), [docs](https://godoc.org/github.com/diamnet/go/clients/auroraclient)) package provides a web client for interfacing with [Aurora](https://developers.diamnet.org/docs/start/introduction/) server REST endpoints to retrieve ledger information, and to submit transactions built with `txnbuild`.

## List of major SDK packages

- `auroraclient` ([source](https://github.com/diamnet/go/tree/master/clients/auroraclient), [docs](https://godoc.org/github.com/diamnet/go/clients/auroraclient)) - programmatic client access to Aurora
- `txnbuild` ([source](https://github.com/diamnet/go/tree/master/txnbuild), [docs](https://godoc.org/github.com/diamnet/go/txnbuild)) - construction, signing and encoding of Diamnet transactions and operations
- `diamnettoml` ([source](https://github.com/diamnet/go/tree/master/clients/diamnettoml), [docs](https://godoc.org/github.com/diamnet/go/clients/diamnettoml)) - parse [Diamnet.toml](../../guides/concepts/diamnet-toml.md) files from the internet
- `federation` ([source](https://godoc.org/github.com/diamnet/go/clients/federation)) - resolve federation addresses  into diamnet account IDs, suitable for use within a transaction

