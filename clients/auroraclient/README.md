# auroraclient


`auroraclient` is a [Diamnet Go SDK](https://developers.diamnet.org/api/) package that provides client access to a aurora server. It supports all endpoints exposed by the [aurora API](https://developers.diamnet.org/api/introduction/).

This project is maintained by the Diamnet Development Foundation.

## Getting Started
This library is aimed at developers building Go applications that interact with the [Diamnet network](https://www.diamnet.org/). It allows users to query the network and submit transactions to the network. The recommended transaction builder for Go programmers is [txnbuild](https://github.com/diamnet/go/tree/master/txnbuild). Together, these two libraries provide a complete Diamnet SDK.

* The [auroraclient API reference](https://godoc.org/github.com/diamnet/go/clients/auroraclient).
* The [txnbuild API reference](https://godoc.org/github.com/diamnet/go/txnbuild).

### Prerequisites
* Go (this repository is officially supported on the last two releases of Go)
* [Modules](https://github.com/golang/go/wiki/Modules) to manage dependencies

### Installing
* `go get github.com/diamnet/go/clients/auroraclient`

### Usage

``` golang
    ...
    import hClient "github.com/diamnet/go/clients/auroraclient"
    ...

    // Use the default pubnet client
    client := hClient.DefaultPublicNetClient

    // Create an account request
    accountRequest := hClient.AccountRequest{AccountID: "GCLWGQPMKXQSPF776IU33AH4PZNOOWNAWGGKVTBQMIC5IMKUNP3E6NVU"}

    // Load the account detail from the network
    account, err := client.AccountDetail(accountRequest)
    if err != nil {
        fmt.Println(err)
        return
    }
    // Account contains information about the diamnet account
    fmt.Print(account)
```
For more examples, refer to the [documentation](https://godoc.org/github.com/diamnet/go/clients/auroraclient).

## Running the tests
Run the unit tests from the package directory: `go test`

## Contributing
Please read [Code of Conduct](https://github.com/diamnet/.github/blob/master/CODE_OF_CONDUCT.md) to understand this project's communication rules.

To submit improvements and fixes to this library, please see [CONTRIBUTING](../CONTRIBUTING.md).

## License
This project is licensed under the Apache License - see the [LICENSE](../../LICENSE-APACHE.txt) file for details.
