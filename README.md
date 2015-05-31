## MiniFS

MiniFS is a simple, distributed, gossip-based filesystem. It supports a
rudimentary REST API over HTTPS and comes with TLS-based authentication.

### Usage:

To run a MiniFS node, compile and build the minifs binary:

```
$ go get github.com/fatlotus/minifs
$ go install github.com/fatlotus/minifs/...
$ $GOROOT/bin/minifs -certfile=server.crt -keyfile=server.key \
   -bind=:8080
...
```

The server will then open a server on https://0.0.0.0:8080/. To inspect
the current state of the cluster, visit https://0.0.0.0:8080/state.json,
which shows the last reached time of each node in the cluster.

If specified with a `-peers` option, MiniFS will use a gossip protocol to
synchronize a consistent hash ring around the cluster. For example:

```
$ $GOROOT/bin/minifs -certfile=server.crt -keyfile=server.key -bind=:8080 &
$ $GOROOT/bin/minifs -certfile=server.crt -keyfile=server.key -bind=:8081 &
$ curl -L --verbose --insecure https://localhost:8080/mykho.txt; ec 
> GET /mykey.txt HTTP/1.1
> Host: localhost:8080
> 
< HTTP/1.1 301 Moved Permanently
< Location: https://valhalla.uchicago.edu:8081/mykey.txt
< 
> GET /mykey.txt HTTP/1.1
> Host: valhalla.uchicago.edu:8081
> 
< HTTP/1.1 200 OK
< 
Hello world!
```

As a result, clients that handle HTTP redirects do not need to manage consistent
hashing.

### License:

MiniFS is available under the MIT license:

Copyright (c) 2015 Jeremy Archer

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
