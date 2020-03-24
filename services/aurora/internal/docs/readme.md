---
title: Aurora
---

Aurora is the server for the client facing API for the HcNet ecosystem.  It acts as the interface between [hcnet-core](https://www.hcnet.org/developers/learn/hcnet-core) and applications that want to access the HcNet network. It allows you to submit transactions to the network, check the status of accounts, subscribe to event streams, etc. See [an overview of the HcNet ecosystem](https://www.hcnet.org/developers/guides/) for more details.

You can interact directly with aurora via curl or a web browser but SDF provides a [JavaScript SDK](https://www.hcnet.org/developers/js-hcnet-sdk/learn/) for clients to use to interact with Aurora.

SDF runs a instance of Aurora that is connected to the test net [https://aurora-testnet.hcnet.org/](https://aurora-testnet.hcnet.org/).

## Libraries

SDF maintained libraries:<br />
- [JavaScript](https://github.com/hcnet/js-hcnet-sdk)
- [Go](https://github.com/hcnet/go/tree/master/clients/auroraclient)
- [Java](https://github.com/hcnet/java-hcnet-sdk)

Community maintained libraries (in various states of completeness) for interacting with Aurora in other languages:<br>
- [Python](https://github.com/HcNetCN/py-hcnet-base)
- [C# .NET Core 2.x](https://github.com/elucidsoft/dotnetcore-hcnet-sdk)
- [Ruby](https://github.com/bloom-solutions/ruby-hcnet-sdk)
- [iOS and macOS](https://github.com/Soneso/hcnet-ios-mac-sdk)
- [Scala SDK](https://github.com/synesso/scala-hcnet-sdk)
- [C++ SDK](https://github.com/bnogalm/HcNetQtSDK)
