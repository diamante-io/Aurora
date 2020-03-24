# auroraclient


`auroraclient` is a [HcNet Go SDK](https://www.hcnet.org/developers/reference/) package that provides client access to a aurora server. It supports all endpoints exposed by the [aurora API](https://www.hcnet.org/developers/aurora/reference/index.html).

This project is maintained by the HcNet Development Foundation.

## Getting Started
This library is aimed at developers building Go applications that interact with the [HcNet network](https://www.hcnet.org/). It allows users to query the network and submit transactions to the network. The recommended transaction builder for Go programmers is [txnbuild](https://github.com/hcnet/go/tree/master/txnbuild). Together, these two libraries provide a complete HcNet SDK.

* The [auroraclient API reference](https://godoc.org/github.com/hcnet/go/clients/auroraclient).
* The [txnbuild API reference](https://godoc.org/github.com/hcnet/go/txnbuild).

### Prerequisites
* Go 1.10 or greater

### Installing
* Download the HcNet Go monorepo: `git clone git@github.com:hcnet/go.git`
* Enter the source directory: `cd $GOPATH/src/github.com/hcnet/go`
* Download external dependencies: `dep ensure -v`

### Usage

``` golang
    ...
    import hClient "github.com/hcnet/go/clients/auroraclient"
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
    // Account contains information about the hcnet account
    fmt.Print(account)
```
For more examples, refer to the [documentation](https://godoc.org/github.com/hcnet/go/clients/auroraclient).

## Running the tests
Run the unit tests from the package directory: `go test`

## Contributing
Please read [Code of Conduct](https://github.com/hcnet/.github/blob/master/CODE_OF_CONDUCT.md) to understand this project's communication rules.

To submit improvements and fixes to this library, please see [CONTRIBUTING](../CONTRIBUTING.md).

## License
This project is licensed under the Apache License - see the [LICENSE](../../LICENSE-APACHE.txt) file for details.
