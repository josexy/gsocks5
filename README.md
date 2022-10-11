# gsocks5

[![Go Report Card](https://goreportcard.com/badge/github.com/josexy/gsocks5)](https://goreportcard.com/report/github.com/josexy/gsocks5)
[![License](https://img.shields.io/github/license/josexy/gsocks5)](https://github.com/josexy/gsocks5/blob/main/LICENSE)

A simple socks5 server and client implemented in Golang

## Feature
- Support socks5 server and client
- Support `TCP(CONNECT)` and `UDP(ASSOCIATE)`
- Support `No` and `USERNAME/PASSWORD` authentication
- Support YAML configuration

## Installation
Go mod:
```bash
go get github.com/josexy/gsocks5
```

Git clone:
```bash
git clone https://github.com/josexy/gsocks5.git
cd gsocks5
```

## Example
You can find all examples in the `gsocks5/example` directory

### server
run socks5 server
```bash
go run example/server/main.go -c config.yaml
```

code example
```go
package main

import (
	"github.com/josexy/gsocks5/socks"
)

func main() {
	svr := socks.NewSocks5Server(":10086")
	defer svr.Close()
	svr.Start()
}
```
### client
run socks5 client
```bash
go run example/tcp_client/main.go
```

or `curl` command tool

```bash
curl -v --socks5 127.0.0.1:10086 -U test:12345678 https://www.google.com
curl -v --socks5 127.0.0.1:10086 https://www.google.com
```

or Chrome extension `Proxy SwitchyOmega`

code example
```go
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/josexy/gsocks5/socks"
)

func main() {
	proxyCli := socks.NewSocks5Client("127.0.0.1:10086")
	proxyCli.SetSocksAuth("test", "12345678")
	defer proxyCli.Close()

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return proxyCli.Dial(ctx, addr)
		},
	}
	cli := http.Client{
		Transport: transport,
		Timeout:   time.Second * 10,
	}
	resp, err := cli.Get("https://www.google.com")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
}

```

### yaml config
The socks server supports the following authentication methods:

- `username`: Username/Password authentication
- `none`: No authentication

```yaml
listen_addr: 0.0.0.0:10086
socks_method:
  - username
  - none
auth:
  - test:12345678
  - test2:123
```