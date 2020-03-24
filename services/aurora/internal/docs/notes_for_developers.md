## Notes for Developers

This document contains additional information related to the development of Aurora. For a detailed discussion of how to build and develop against Aurora, see the [Aurora development guide](developing.md).

- [Initial set up](#setup)
- [Regenerating generated code](#regen)
- [Adding and rebuilding test scenarios](#scenarios)
- [Running tests](#tests)
- [Logging](#logging)


---
## <a name="setup"></a> Initial set up
Compile and install Aurora as described in the [Aurora development guide](developing.md).

## <a name="regen"></a> Regenerating generated code

Aurora uses two Go tools you'll need to install:
1. [go-bindata](https://github.com/jteeuwen/go-bindata) is used to bundle test data
2. [go-codegen](https://github.com/nullstyle/go-codegen) is used to generate some boilerplate code

After the above are installed, run `go generate github.com/hcnet/go/services/aurora/...`. This will look for any `.tmpl` files in the directory and use them to generate code when annotated structs are found in the package source.

## <a name="scenarios"></a> Adding, rebuilding and using test scenarios

In order to simulate ledgers Aurora uses [`hcnet-core-commander`](https://github.com/hcnet/hcnet_core_commander) recipe files to add transactions and operations to ledgers using the hcnet-core test framework.

In order to add a new scenario or rebuild existing scenarios you need:

1. [`hcnet-core-commander`](https://github.com/hcnet/hcnet_core_commander) (in short: `scc`) installed and [configured](https://github.com/hcnet/hcnet_core_commander#assumptions-about-environment).
2. [`hcnet-core`](https://github.com/hcnet/hcnet-core) binary.
3. This repository cloned locally.

`scc` allows you to write scripts/recipes that are later executed in `hcnet-core` isolated network. After executing a recipe you can then export the `hcnet-core` database to be able to run Aurora ingestion system against it (this repository contains a script that does this for you - read below).

### Example recipe

Here's an example of recipe file with comments:
```rb
# Define two accounts test accounts
account :scott, HcNet::KeyPair.from_seed("SBZWG33UOQQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAPSA")
account :bartek, HcNet::KeyPair.from_seed("SBRGC4TUMVVSAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCBDHV")

# use_manual_close causes scc to run a process with MANUAL_CLOSE=true
use_manual_close

# Create 2 accounts `scott` and `bartek`
create_account :scott,  :master, 100
create_account :bartek, :master, 100

# Close ledger
close_ledger

# Send 5 XLM from `scott` to `bartek`
payment :scott, :bartek,  [:native, 5]
```

You can find more recipes in [`scc` examples](https://github.com/hcnet/hcnet_core_commander/tree/84d5ffb97202ecc3a0ed34a739c98e69536c0c2c/examples) and [aurora test scenarios](https://github.com/hcnet/go/tree/master/services/aurora/internal/test/scenarios).

### Rebuilding scenarios

1. Create a new or modify existing recipe. All new recipes should be added to [aurora test scenarios](https://github.com/hcnet/go/tree/master/services/aurora/internal/test/scenarios) directory.
2. In `hcnet/go` repository root directory run `./services/aurora/internal/scripts/build_test_scenarios.bash`.
3. The command above will rebuild all test scenarios. If you need to rebuild only one scenario modify `PACKAGES` environment variable temporarily in the script.

### Using test scenarios

In your `Test*` function execute:

```go
ht := StartHTTPTest(t, scenarioName)
defer ht.Finish()
```
where `scenarioName` is the name of the scenario you want to use. This will start test Aurora server with data loaded from the recipe.

When testing ingestion you can load scenario data without Aurora database like:

```go
tt := test.Start(t).ScenarioWithoutAurora("kahuna")
defer tt.Finish()
s := ingest(tt, true)
```

Check existing tests for more examples.

## <a name="tests"></a> Running Tests

start a redis server on port `6379`

```bash
redis-server
```

then, run the all the Go monorepo tests like so (assuming you are at hcnet/go, or run from hcnet/go/services/aurora for just the Aurora subset):

```bash
bash ./support/scripts/run_tests
```

or run individual Aurora tests like so, providing the expected arguments:

```bash
go test github.com/hcnet/go/services/aurora/...
```

## <a name="logging"></a> Logging

All logging infrastructure is in the `github.com/hcnet/go/tree/master/services/aurora/internal/log` package.  This package provides "level-based" logging:  Each logging statement has a severity, one of "Debug", "Info", "Warn", "Error" or "Panic".  The Aurora server has a configured level "filter", specified either using the `--log-level` command line flag or the `LOG_LEVEL` environment variable.  When a logging statement is executed, the statements declared severity is checked against the filter and will only be emitted if the severity of the statement is equal or higher severity than the filter.

In addition, the logging subsystem has support for fields: Arbitrary key-value pairs that will be associated with an entry to allow for filtering and additional contextual information.

### Making logging statements

Assuming that you've imports the log package, making a simple logging call is just:

```go

log.Info("my log line")
log.Infof("I take a %s", "format string")

```

Adding fields to a statement happens with a call to `WithField` or `WithFields`

```go
log.WithField("pid", 1234).Warn("i'm scared")

log.WithFields(log.F{
	"some_field": 123,
	"second_field": "hello",
}).Debug("here")
```

The return value from `WithField` or `WithFields` is a `*log.Entry`, which you can save to emit multiple logging
statements that all share the same field.  For example, the action system for Aurora attaches a log entry to `action.Log` on every request that can be used to emit log entries that have the request's id attached as a field.

### Logging and Context

The logging package provides the root logger at `log.DefaultLogger` and the package level funcs such as `log.Info` operate against the default logger.  However, often it is important to include request-specific fields in a logging statement that are not available in the local scope.  For example, it is useful to include an http request's id in every log statement that is emitted by code running on behalf of the request.  This allows for easier debugging, as an operator can filter the log stream to a specific request id and not have to wade through the entirety of the log.

Unfortunately, it is not prudent to thread an `*http.Request` parameter to every downstream subroutine and so we need another way to make that information available.  The idiomatic way to do this in Go is with a context parameter, as describe [on the Go blog](https://blog.golang.org/context).  The logging provides a func to bind a logger to a context using `log.Set` and allows you to retrieve a bound logger using `log.Ctx(ctx)`.  Functions that need to log on behalf of an server request should take a context parameter.

Here's an example of using context:

```go

// create a new sublogger
sub := log.WithField("val", 1)

// bind it to a context
ctx := log.Set(context.Background(), sub)

log.Info("no fields on this statement")
log.Ctx(ctx).Info("This statement will use the sub logger")

```

### Logging Best Practices

It's recommended that you try to avoid contextual information in your logging messages.  Instead, use fields to establish context and use a static string for your message.  This practice allows Aurora operators to more easily filter log lines to provide better insight into the health of the server.  Lets take an example:

```go
// BAD
log.Infof("running initializer: %s", i.Name)

//GOOD
log.WithField("init_name", i.Name).Info("running initializer")
```

With the "bad" form of the logging example above, an operator can filter on both the message as well as the initializer name independently.  This gets more powerful when multiple fields are combined, allowing for all sorts of slicing and dicing.


## <a name="TLS"></a> Enabling TLS on your local workstation

Aurora support HTTP/2 when served using TLS.  To enable TLS on your local workstation, you must generate a certificate and configure Aurora to use it.  We've written a helper script at `tls/regen.sh` to make this simple.  Run the script from your terminal, and simply choose all the default options.  This will create two files: `tls/server.crt` and `tls/server.key`.  

Now you must configure Aurora to use them: You can simply add `--tls-cert tls/server.crt --tls-key tls/server.key` to your command line invocations of Aurora, or you may specify `TLS_CERT` and `TLS_KEY` environment variables.
