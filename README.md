# ooclient

Client for [oostore](https://github.com/cmars/oostore).

# Build

Install the `oo` binary with:
`go get github.com/cmars/ooclient/cmd/oo`

# Use

Set `$OOSTORE_URL` to the base URL location of the oostore service, or use the
`--url` option.

```
$ oostore &
2015/09/20 00:08:33 listening for requests on 127.0.0.1:20080
$ export OOSTORE_URL=http://127.0.0.1:20080
```

Commands read input from stdin and write output to stdout, unless otherwise
specified with flags.

## oo new

Create a new opaque object, obtaining an auth token.

```
NAME:
   new - oo new [-i|--input file] [-o|--output file] [-t|--content-type type]

USAGE:
   command new [command options] [arguments...]

OPTIONS:
   --url 		 [$OOSTORE_URL]
   --input, -i 		
   --output, -o 	
   --content-type, -t 	
```   

### Example

```
$ echo "hunter2" | oonew
[{"caveats":[{"cid":"object 5zxFasj4FBpBm4nJL5MY7ugWwi3EqgecFgngesFqaMHt"}],"location":"","identifier":"af68ce02fffed6acd80e4eda8bde339b99e60bab252d3fe7","signature":"478ac5c9d76668a02850ebbec63eaed56a93ea70e831bfe8c468efab364d570d"}]
```

## oo fetch

Retrieve the opaque object with an auth token.

```
NAME:
   fetch - oo fetch [-i|--input file] [-o|--output file]

USAGE:
   command fetch [command options] [arguments...]

OPTIONS:
   --url 		 [$OOSTORE_URL]
   --input, -i 		
   --output, -o 	
```

### Example

```
$ echo "hunter2" | oo new | oo fetch
hunter2
```

## oo cond

Add conditions to auth token.

```
NAME:
   cond - oo cond [-i|--input file] [-o|--output file] condition

USAGE:
   command cond [command options] [arguments...]

OPTIONS:
   --url 		 [$OOSTORE_URL]
   --input, -i 		
   --output, -o 	
```

### Example

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
