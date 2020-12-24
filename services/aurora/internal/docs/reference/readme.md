---
title: Overview
---

Aurora is an API server for the DiamNet ecosystem.  It acts as the interface between [diamnet-core](https://github.com/diamnet/diamnet-core) and applications that want to access the DiamNet network. It allows you to submit transactions to the network, check the status of accounts, subscribe to event streams, etc. See [an overview of the DiamNet ecosystem](https://www.diamnet.org/developers/guides/) for details of where Aurora fits in. You can also watch a [talk on Aurora](https://www.youtube.com/watch?v=AtJ-f6Ih4A4) by DiamNet.org developer Scott Fleckenstein:

[![Aurora: API webserver for the DiamNet network](https://img.youtube.com/vi/AtJ-f6Ih4A4/sddefault.jpg "Aurora: API webserver for the DiamNet network")](https://www.youtube.com/watch?v=AtJ-f6Ih4A4)

Aurora provides a RESTful API to allow client applications to interact with the DiamNet network. You can communicate with Aurora using cURL or just your web browser. However, if you're building a client application, you'll likely want to use a DiamNet SDK in the language of your client.
SDF provides a [JavaScript SDK](https://www.diamnet.org/developers/js-diamnet-sdk/learn/index.html) for clients to use to interact with Aurora.

SDF runs a instance of Aurora that is connected to the test net: [https://aurora-testnet.diamnet.org/](https://aurora-testnet.diamnet.org/) and one that is connected to the public DiamNet network:
[https://aurora.diamnet.org/](https://aurora.diamnet.org/).

## Libraries

SDF maintained libraries:<br />
- [JavaScript](https://github.com/diamnet/js-diamnet-sdk)
- [Go](https://github.com/diamnet/go/tree/master/clients/auroraclient)
- [Java](https://github.com/diamnet/java-diamnet-sdk)

Community maintained libraries (in various states of completeness) for interacting with Aurora in other languages:<br>
- [Python](https://github.com/DiamNetCN/py-diamnet-base)
- [C# .NET Core 2.x](https://github.com/elucidsoft/dotnetcore-diamnet-sdk)
- [Ruby](https://github.com/bloom-solutions/ruby-diamnet-sdk)
- [iOS and macOS](https://github.com/Soneso/diamnet-ios-mac-sdk)
- [Scala SDK](https://github.com/synesso/scala-diamnet-sdk)
- [C++ SDK](https://github.com/bnogalm/DiamNetQtSDK)
