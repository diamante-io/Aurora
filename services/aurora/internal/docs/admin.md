---
title: Aurora Administration Guide
---
## Aurora Administration Guide

Aurora is responsible for providing an HTTP API to data in the DiamNet network. It ingests and re-serves the data produced by the diamnet network in a form that is easier to consume than the performance-oriented data representations used by diamnet-core.

This document describes how to administer a **production** Aurora instance. If you are just starting with Aurora and want to try it out, consider the [Quickstart Guide](quickstart.md) instead. For information about developing on the Aurora codebase, check out the [Development Guide](developing.md).

## Why run Aurora?

The DiamNet Development Foundation runs two Aurora servers, one for the public network and one for the test network, free for anyone's use at https://aurora.diamnet.org and https://aurora-testnet.diamnet.org.  These servers should be fine for development and small scale projects, but it is not recommended that you use them for production services that need strong reliability.  By running Aurora within your own infrastructure provides a number of benefits:

  - Multiple instances can be run for redundancy and scalability.
  - Request rate limiting can be disabled.
  - Full operational control without dependency on the DiamNet Development Foundations operations.

## Prerequisites

Aurora is dependent upon a diamnet-core server.  Aurora needs access to both the SQL database and the HTTP API that is published by diamnet-core. See [the administration guide](https://www.diamnet.org/developers/diamnet-core/learn/admin.html
) to learn how to set up and administer a diamnet-core server.  Secondly, Aurora is dependent upon a postgres server, which it uses to store processed core data for ease of use. Aurora requires postgres version >= 9.5.

In addition to the two prerequisites above, you may optionally install a redis server to be used for rate limiting requests.

## Installing

To install Aurora, you have a choice: either downloading a [prebuilt release for your target architecture](https://github.com/diamnet/go/releases) and operation system, or [building Aurora yourself](#Building).  When either approach is complete, you will find yourself with a directory containing a file named `aurora`.  This file is a native binary.

After building or unpacking Aurora, you simply need to copy the native binary into a directory that is part of your PATH.  Most unix-like systems have `/usr/local/bin` in PATH by default, so unless you have a preference or know better, we recommend you copy the binary there.

To test the installation, simply run `aurora --help` from a terminal.  If the help for Aurora is displayed, your installation was successful. Note: some shells, such as zsh, cache PATH lookups.  You may need to clear your cache  (by using `rehash` in zsh, for example) before trying to run `aurora --help`.


## Building

Should you decide not to use one of our prebuilt releases, you may instead build Aurora from source.  To do so, you need to install some developer tools:

- A unix-like operating system with the common core commands (cp, tar, mkdir, bash, etc.)
- A compatible distribution of Go (we officially support Go 1.10 and later)
- [go-dep](https://golang.github.io/dep/)
- [git](https://git-scm.com/)
- [mercurial](https://www.mercurial-scm.org/)


1. Set your [GOPATH](https://github.com/golang/go/wiki/GOPATH) environment variable, if you haven't already. The default `GOPATH` is `$HOME/go`.
2. Clone the DiamNet Go monorepo:  `go get github.com/diamnet/go`. You should see the repository cloned at `$GOPATH/src/github.com/diamnet/go`.
3. Enter the source dir: `cd $GOPATH/src/github.com/diamnet/go`, and download external dependencies: `dep ensure -v`. You should see the downloaded third party dependencies in `$GOPATH/pkg`.
4. Compile the Aurora binary: `cd $GOPATH; go install github.com/diamnet/go/services/aurora`. You should see the `aurora` binary in `$GOPATH/bin`.
5. Add Go binaries to your PATH in your `bashrc` or equivalent, for easy access: `export PATH=${GOPATH//://bin:}/bin:$PATH`

Open a new terminal. Confirm everything worked by running `aurora --help` successfully.

Note:  Building directly on windows is not supported.


## Configuring

Aurora is configured using command line flags or environment variables.  To see the list of command line flags that are available (and their default values) for your version of Aurora, run:

`aurora --help`

As you will see if you run the command above, Aurora defines a large number of flags, however only three are required:

| flag                    | envvar                      | example                              |
|-------------------------|-----------------------------|--------------------------------------|
| `--db-url`              | `DATABASE_URL`              | postgres://localhost/aurora_testnet |
| `--diamnet-core-db-url` | `HCNET_CORE_DATABASE_URL` | postgres://localhost/core_testnet    |
| `--diamnet-core-url`    | `HCNET_CORE_URL`          | http://localhost:11626               |

`--db-url` specifies the Aurora database, and its value should be a valid [PostgreSQL Connection URI](http://www.postgresql.org/docs/9.2/static/libpq-connect.html#AEN38419).  `--diamnet-core-db-url` specifies a diamnet-core database which will be used to load data about the diamnet ledger.  Finally, `--diamnet-core-url` specifies the HTTP control port for an instance of diamnet-core.  This URL should be associated with the diamnet-core that is writing to the database at `--diamnet-core-db-url`.

Specifying command line flags every time you invoke Aurora can be cumbersome, and so we recommend using environment variables.  There are many tools you can use to manage environment variables:  we recommend either [direnv](http://direnv.net/) or [dotenv](https://github.com/bkeepers/dotenv).  A template configuration that is compatible with dotenv can be found in the [Aurora git repo](https://github.com/diamnet/go/blob/master/services/aurora/.env.template).



## Preparing the database

Before the Aurora server can be run, we must first prepare the Aurora database.  This database will be used for all of the information produced by Aurora, notably historical information about successful transactions that have occurred on the diamnet network.

To prepare a database for Aurora's use, first you must ensure the database is blank.  It's easiest to simply create a new database on your postgres server specifically for Aurora's use.  Next you must install the schema by running `aurora db init`.  Remember to use the appropriate command line flags or environment variables to configure Aurora as explained in [Configuring ](#Configuring).  This command will log any errors that occur.

### Postgres configuration

It is recommended to set `random_page_cost=1` in Postgres configuration if you are using SSD storage. With this setting Query Planner will make a better use of indexes, expecially for `JOIN` queries. We have noticed a huge speed improvement for some queries.

## Running

Once your Aurora database is configured, you're ready to run Aurora.  To run Aurora you simply run `aurora` or `aurora serve`, both of which start the HTTP server and start logging to standard out.  When run, you should see some output that similar to:

```
INFO[0000] Starting aurora on :8000                     pid=29013
```

The log line above announces that Aurora is ready to serve client requests. Note: the numbers shown above may be different for your installation.  Next we can confirm that Aurora is responding correctly by loading the root resource.  In the example above, that URL would be [http://127.0.0.1:8000/] and simply running `curl http://127.0.0.1:8000/` shows you that the root resource can be loaded correctly.

If you didn't set up a diamnet-core yet, you may see an error like this:
```
ERRO[2019-05-06T16:21:14.126+08:00] Error getting core latest ledger err="get failed: pq: relation \"ledgerheaders\" does not exist"
```
Aurora requires a functional diamnet-core. Go back and set up diamnet-core as described in the admin guide. In particular, you need to initialise the database as [described here](https://www.diamnet.org/developers/diamnet-core/software/admin.html#database-and-local-state).

## Ingesting live diamnet-core data

Aurora provides most of its utility through ingested data.  Your Aurora server can be configured
to listen for and ingest transaction results from the connected diamnet-core.  We recommend that
within your infrastructure you run one (and only one) Aurora process that is configured in this
way. While running multiple ingestion processes will not corrupt the Aurora database, your error
logs will quickly fill up as the two instances race to ingest the data from diamnet-core. A notable
exception to this is when you are reingesting data, which we recommend using multiple processes for
speed (more on this below).

To enable ingestion, you must either pass `--ingest=true` on the command line or set the `INGEST`
environment variable to "true".

### Ingesting historical data

To enable ingestion of historical data from diamnet-core you need to run `aurora db backfill NUM_LEDGERS`. If you're running a full validator with published history archive, for example, you might want to ingest all of history. In this case your `NUM_LEDGERS` should be slightly higher than the current ledger id on the network. You can run this process in the background while your Aurora server is up. This continuously decrements the `history.elder_ledger` in your /metrics endpoint until `NUM_LEDGERS` is reached and the backfill is complete.

### Reingesting Ledgers
A notable exception to running only a single Aurora process is when you are reingesting ledgers,
which we recommend you run multiple processes for in order to dramatically speed up re-ingestion
time. This is done through the `aurora db range [START_LEDGER] [END_LEDGER]` command, which could
be run as follows:

```
aurora1> aurora db reingest range 1 10000
aurora2> aurora db reingest range 10001 20000
aurora3> aurora db reingest range 20001 30000
# ... etc.
```

This allows reingestion to be split up and done in parallel by multiple Aurora processes, and is
available as of Aurora [0.17.4](https://github.com/diamnet/go/releases/tag/aurora-v0.17.4).

### Managing storage for historical data

Over time, the recorded network history will grow unbounded, increasing storage used by the database. Aurora expands the data ingested from diamnet-core and needs sufficient disk space. Unless you need to maintain a history archive you may configure Aurora to only retain a certain number of ledgers in the database. This is done using the `--history-retention-count` flag or the `HISTORY_RETENTION_COUNT` environment variable. Set the value to the number of recent ledgers you wish to keep around, and every hour the Aurora subsystem will reap expired data.  Alternatively, you may execute the command `aurora db reap` to force a collection.

### Surviving diamnet-core downtime

Aurora tries to maintain a gap-free window into the history of the diamnet-network.  This reduces the number of edge cases that Aurora-dependent software must deal with, aiming to make the integration process simpler.  To maintain a gap-free history, Aurora needs access to all of the metadata produced by diamnet-core in the process of closing a ledger, and there are instances when this metadata can be lost.  Usually, this loss of metadata occurs because the diamnet-core node went offline and performed a catchup operation when restarted.

To ensure that the metadata required by Aurora is maintained, you have several options: You may either set the `CATCHUP_COMPLETE` diamnet-core configuration option to `true` or configure `CATCHUP_RECENT` to determine the amount of time your diamnet-core can be offline without having to rebuild your Aurora database.

Unless your node is a full validator and archive publisher we _do not_ recommend using the `CATCHUP_COMPLETE` method, as this will force diamnet-core to apply every transaction from the beginning of the ledger, which will take an ever increasing amount of time. Instead, we recommend you set the `CATCHUP_RECENT` config value. To do this, determine how long of a downtime you would like to survive (expressed in seconds) and divide by ten.  This roughly equates to the number of ledgers that occur within your desired grace period (ledgers roughly close at a rate of one every ten seconds).  With this value set, diamnet-core will replay transactions for ledgers that are recent enough, ensuring that the metadata needed by Aurora is present.

### Correcting gaps in historical data

In the section above, we mentioned that Aurora _tries_ to maintain a gap-free window.  Unfortunately, it cannot directly control the state of diamnet-core and [so gaps may form](https://www.diamnet.org/developers/software/known-issues.html#gaps-detected) due to extended down time.  When a gap is encountered, Aurora will stop ingesting historical data and complain loudly in the log with error messages (log lines will include "ledger gap detected").  To resolve this situation, you must re-establish the expected state of the diamnet-core database and purge historical data from Aurora's database.  We leave the details of this process up to the reader as it is dependent upon your operating needs and configuration, but we offer one potential solution:

We recommend you configure the HISTORY_RETENTION_COUNT in Aurora to a value less than or equal to the configured value for CATCHUP_RECENT in diamnet-core.  Given this situation any downtime that would cause a ledger gap will require a downtime greater than the amount of historical data retained by Aurora.  To re-establish continuity:

1.  Stop Aurora.
2.  Run `aurora db reap` to clear the historical database.
3.  Clear the cursor for Aurora by running `diamnet-core -c "dropcursor?id=HORIZON"` (ensure capitilization is maintained).
4.  Clear ledger metadata from before the gap by running `diamnet-core -c "maintenance?queue=true"`.
5.  Restart Aurora.

## Managing Stale Historical Data

Aurora ingests ledger data from a connected instance of diamnet-core.  In the event that diamnet-core stops running (or if Aurora stops ingesting data for any other reason), the view provided by Aurora will start to lag behind reality.  For simpler applications, this may be fine, but in many cases this lag is unacceptable and the application should not continue operating until the lag is resolved.

To help applications that cannot tolerate lag, Aurora provides a configurable "staleness" threshold.  Given that enough lag has accumulated to surpass this threshold (expressed in number of ledgers), Aurora will only respond with an error: [`stale_history`](./errors/stale-history.md).  To configure this option, use either the `--history-stale-threshold` command line flag or the `HISTORY_STALE_THRESHOLD` environment variable.  NOTE:  non-historical requests (such as submitting transactions or finding payment paths) will not error out when the staleness threshold is surpassed.

## Monitoring

To ensure that your instance of Aurora is performing correctly we encourage you to monitor it, and provide both logs and metrics to do so.

Aurora will output logs to standard out.  Information about what requests are coming in will be reported, but more importantly, warnings or errors will also be emitted by default.  A correctly running Aurora instance will not output any warning or error log entries.

Metrics are collected while a Aurora process is running and they are exposed at the `/metrics` path.  You can see an example at (https://aurora-testnet.diamnet.org/metrics).

## I'm Stuck! Help!

If any of the above steps don't work or you are otherwise prevented from correctly setting up
Aurora, please come to our community and tell us. Either
[post a question at our Stack Exchange](https://diamnet.stackexchange.com/) or
[chat with us on Keybase in #dev_discussion](https://keybase.io/team/diamnet.public) to ask for
help.
