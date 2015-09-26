[![Build Status](https://travis-ci.org/cmars/ooclient.svg?branch=master)](https://travis-ci.org/cmars/ooclient)
[![GoDoc](https://godoc.org/github.com/cmars/ooclient?status.svg)](https://godoc.org/github.com/cmars/ooclient)

# ooclient

Client for [oostore](https://github.com/cmars/oostore).

# Build

Install the `oo` binary with:
`go get github.com/cmars/ooclient/cmd/oo`

# Use

```
NAME:
   oo - oo [command] [args]

COMMANDS:
   new                  create a new opaque object with given input, output auth macaroon
   fetch                fetch opaque object contents with auth macaroon
   cond                 place conditional caveats on auth macaroon
   delete, del, rm      delete opaque object with auth macaroon
   help, h              Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h           show help
```

Set `$OOSTORE_URL` to the base URL location of the oostore service, or use the
`--url` option.

```
$ oostore &
2015/09/20 00:08:33 listening for requests on 127.0.0.1:20080
$ export OOSTORE_URL=http://127.0.0.1:20080
```

Commands read input from stdin and write output to stdout, unless otherwise
specified with flags. Exit status will be non-zero on error, with diagnostic
information logged to stderr.

## oo new

```
NAME:
   new - create a new opaque object with given input, output auth macaroon

USAGE:
   command new [command options] [arguments...]

OPTIONS:
   --url                 [$OOSTORE_URL]
   --input, -i
   --output, -o
   --content-type, -t
```

### Example

```
$ echo "hunter2" | oo new
[{"caveats":[{"cid":"object 5zxFasj4FBpBm4nJL5MY7ugWwi3EqgecFgngesFqaMHt"}],"location":"","identifier":"af68ce02fffed6acd80e4eda8bde339b99e60bab252d3fe7","signature":"478ac5c9d76668a02850ebbec63eaed56a93ea70e831bfe8c468efab364d570d"}]
```

## oo fetch

```
NAME:
   fetch - fetch opaque object contents with auth macaroon

USAGE:
   command fetch [command options] [arguments...]

OPTIONS:
   --url                 [$OOSTORE_URL]
   --input, -i
   --output, -o
```

### Example

```
$ echo "hunter2" | oo new | oo fetch
hunter2
```

## oo delete

```
NAME:
   delete - delete opaque object with auth macaroon

USAGE:
   command delete [command options] [arguments...]

OPTIONS:
   --url         [$OOSTORE_URL]
   --input, -i
```

### Example

```
$ echo "hunter2" | oo new > pwd.auth
$ oo fetch < pwd.auth
hunter2
$ oo delete < pwd.auth
$ oo fetch < pwd.auth
2015/09/20 14:00:39 404 Not Found: not found: "AwXgV2LMsBXSv9u5EzM9KrVJrPwoN4b6tVSCGXaB7wX"
```

## oo cond

```
NAME:
   cond - place conditional caveats on auth macaroon

USAGE:
   command cond [command options] [arguments...]

OPTIONS:
   --url                         [$OOSTORE_URL]
   --input, -i 
   --output, -o 
   --location, --loc, -l        location of service for third-party caveat
   --key, -k                    base64-encoded public key of third-party service
```

### First-party caveats

The following conditions are recognized by the oostore service directly.

#### client-ip-addr

`client-ip-addr` takes an allowed IPv4 address as argument. Only requests from
this client IP will be allowed.

Condition met:

```
$ echo "hunter2" | oo new | oo cond client-ip-addr 127.0.0.1 | oo fetch
hunter2
```

Condition not met:

```
$ echo "hunter2" | oo new | oo cond client-ip-addr 1.2.3.4 | oo fetch
2015/09/20 00:45:42 403 Forbidden
verification failed: caveat "client-ip-addr 1.2.3.4" not satisfied: client IP address mismatch, got 127.0.0.1
```

#### time-before

`time-before` sets an expiration on the authorization. Argument is an RFC3339 timestamp.

```
oo cond time-before 2015-11-01T00:00:00Z < auth.json > auth-with-exp.json
```

#### operation

`operation` specifies a comma-separated list of operations allowed. Currently
recognized operations are `fetch` and `delete`.

```
$ oo cond operation fetch < auth.json > auth-fetch-only.json
$ oo delete < auth-fetch-only.json
2015/09/20 14:07:03 403 Forbidden: verification failed: caveat "operation fetch" not satisfied: operation "delete" not allowed
```

### Third-party caveats

For this example, you'll need a third-party caveat discharging service like the
[timestamper](https://github.com/mattyw/timestamper). The timestamper provides
evidence that an authorization was used at a certain time, by declaring a
timestamp in its discharge.

1. `go get` and run the `timestamper` server. It listens on port 8080.
2. Obtain the timestamp service's ephemeral public key with
   `curl http://localhost:8080/publickey`.
3. Add a third-party caveat to an object, requiring requests on it to be timestamped:

```
$ echo "foo biscuits" | oo new | \
	oo cond -l http://localhost:8080 -k aCU6K7U9TpiSjDVYrMMg21P89WjXT0EGmyGcLUeV2G0= is-timestamped | \
	oo fetch
foo biscuits
```

- `oostore` doesn't know anything about `timestamper` or what it does.
- `timestamper` doesn't know anything about `oostore`.
- The object creator is able to require timestamping of requests on the object, just by knowing the public key
  and URL endpoint of the timestamping service.

# License

Copyright 2015 Casey Marshall.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
