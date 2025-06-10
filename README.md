# cash

A lightweight, portable cache database made with Golang

## Installation

```bash
go install github.com/grqphical/cash@latest
```

## Usage

```bash
cash
```

CLI Flags:

`-port`: Port to use for server (default 6400)

`-host`: IP to host on (default 0.0.0.0)

`-file`: File to persist data to. Pass as empty string to disable persistent values (default cache.cashlog)

## Query Syntax

cash uses a simple query language similar to SQL. Multiple commands can be sent in one packet, just seperate them with semicolons

### Set data

```
SET [key] [value] [?COMPRESS]
```

Sets ket to value, if COMPRESS is added as the third argument, the value will be compressed with gzip

### Get data

```
GET [key]
```

Retrieves a key from the cache, if the key is compressed it will be automatically decompressed

### Delete data

```
DELETE [key]
```

Deletes a key from the cache

### Key Expiration

```
EXPIRES [key] [duration]
```

Sets a key to expire after the given duration. Duration is in seconds

## License

cash is licensed under the MIT license
