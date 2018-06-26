# Multy-back

Back end server for [Multy](https://github.com/Multy-io/Multy/wiki) application

## Usage Instructions

In all cases you would like to set configuration parameters in `multy.config` file with addresses, passwords, tokend and other settings. Note, that configuration file must be located in the same directory as the binary file.

### From source

For building Multy-back, in Multy-back directory run:

```
make all-with-deps
```

After that in `cmd` directory binary file with name `multy` should appear.

To run a server:

```
make run
```

Notice, that program uses NSQ, MongoDB and BTC RPC API. You should install and run it by yourself.

### From docker-compose

In docker-compose file (`multy-back` service) set volumes: `multy.config` and `rpc.cert` (for btc node).

If you want to use btc node from docker container, uncomment btcd-testnet part o fdocker-compose file and set its address in `BTCNodeAddress` field in `cmd/multy.config`.

To run a server:

```
docker-compose up
```

## API

For API documentation, see https://godoc.org/github.com/Multy-io/Multy-back
