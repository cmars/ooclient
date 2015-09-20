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
   --url                 [$OOSTORE_URL]
   --input, -i
   --output, -o
```

### Examples

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
