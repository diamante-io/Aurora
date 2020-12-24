---
title: Aurora Quickstart
---
## Aurora Quickstart
This document describes how to quickly set up a **test** DiamNet Core + Aurora node, that you can play around with to get a feel for how a diamnet node operates. **This configuration is not secure!** It is **not** intended as a guide for production administration.

For detailed information about running Aurora and DiamNet Core safely in production see the [Aurora Administration Guide](admin.md) and the [DiamNet Core Administration Guide](https://www.diamnet.org/developers/diamnet-core/software/admin.html).

If you're ready to roll up your sleeves and dig into the code, check out the [Developer Guide](developing.md).

### Install and run the Quickstart Docker Image
The fastest way to get up and running is using the [DiamNet Quickstart Docker Image](https://github.com/diamnet/docker-diamnet-core-aurora). This is a Docker container that provides both `diamnet-core` and `aurora`, pre-configured for testing.

1. Install [Docker](https://www.docker.com/get-started).
2. Verify your Docker installation works: `docker run hello-world`
3. Create a local directory that the container can use to record state. This is helpful because it can take a few minutes to sync a new `diamnet-core` with enough data for testing, and because it allows you to inspect and modify the configuration if needed. Here, we create a directory called `diamnet` to use as the persistent volume:
`cd $HOME; mkdir diamnet`
4. Download and run the DiamNet Quickstart container, replacing `USER` with your username:

```bash
docker run --rm -it -p "8000:8000" -p "11626:11626" -p "11625:11625" -p"8002:5432" -v $HOME/diamnet:/opt/diamnet --name diamnet diamnet/quickstart --testnet
```

You can check out DiamNet Core status by browsing to http://localhost:11626.

You can check out your Aurora instance by browsing to http://localhost:8000.

You can tail logs within the container to see what's going on behind the scenes:
```bash
docker exec -it diamnet /bin/bash
supervisorctl tail -f diamnet-core
supervisorctl tail -f aurora stderr
```

On a modern laptop this test setup takes about 15 minutes to synchronise with the last couple of days of testnet ledgers. At that point Aurora will be available for querying. 

See the [Quickstart Docker Image](https://github.com/diamnet/docker-diamnet-core-aurora) documentation for more details, and alternative ways to run the container. 

You can test your Aurora instance with a query like: http://localhost:8001/transactions?cursor=&limit=10&order=asc. Use the [DiamNet Laboratory](https://www.diamnet.org/laboratory/) to craft other queries to try out,
and read about the available endpoints and see examples in the [Aurora API reference](https://www.diamnet.org/developers/aurora/reference/).

