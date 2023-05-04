---
title: Overview
---

Aurora is an API server for the Diamnet ecosystem.  It acts as the interface between [diamnet-core](https://github.com/diamnet/diamnet-core) and applications that want to access the Diamnet network. It allows you to submit transactions to the network, check the status of accounts, subscribe to event streams, etc. See [an overview of the Diamnet ecosystem](https://www.diamnet.org/developers/guides/) for details of where Aurora fits in.

Aurora provides a RESTful API to allow client applications to interact with the Diamnet network. You can communicate with Aurora using cURL or just your web browser. However, if you're building a client application, you'll likely want to use a Diamnet SDK in the language of your client.
SDF provides a [JavaScript SDK](https://www.diamnet.org/developers/js-diamnet-sdk/reference/index.html) for clients to use to interact with Aurora.

SDF runs a instance of Aurora that is connected to the test net: [https://aurora-testnet.diamnet.org/](https://aurora-testnet.diamnet.org/) and one that is connected to the public Diamnet network:
[https://aurora.diamnet.org/](https://aurora.diamnet.org/).

## Libraries

SDF maintained libraries:<br />
- [JavaScript](https://github.com/diamnet/js-diamnet-sdk)
- [Go](https://github.com/diamnet/go/tree/master/clients/auroraclient)
- [Java](https://github.com/diamnet/java-diamnet-sdk)

Community maintained libraries for interacting with Aurora in other languages:<br>
- [Python](https://github.com/DiamnetCN/py-diamnet-base)
- [C# .NET Core 2.x](https://github.com/elucidsoft/dotnetcore-diamnet-sdk)
- [Ruby](https://github.com/astroband/ruby-diamnet-sdk)
- [iOS and macOS](https://github.com/Soneso/diamnet-ios-mac-sdk)
- [Scala SDK](https://github.com/synesso/scala-diamnet-sdk)
- [C++ SDK](https://github.com/bnogalm/DiamnetQtSDK)
