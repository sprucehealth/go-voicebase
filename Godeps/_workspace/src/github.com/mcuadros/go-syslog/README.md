go-syslog [![Build Status](https://travis-ci.org/mcuadros/go-syslog.png?branch=master)](https://travis-ci.org/mcuadros/go-syslog) [![GoDoc](http://godoc.org/github.com/mcuadros/go-syslog?status.png)](http://godoc.org/github.com/mcuadros/go-syslog)
==============================

Syslog server library for go, build easy your custom syslog server over UDP, TCP or Unix sockets using RFC3164 or RFC5424

Installation
------------

The recommended way to install go-syslog

```
go get github.com/mcuadros/go-syslog
```

Examples
--------

How import the package

```go
import "github.com/mcuadros/go-syslog"
```

Example of a basic syslog [UDP server](example/basic_udp.go):    

```go
channel := make(syslog.LogPartsChannel)
handler := syslog.NewChannelHandler(channel)

server := syslog.NewServer()
server.SetFormat(syslog.RFC5424)
server.SetHandler(handler)
server.ListenUDP("0.0.0.0:514")
server.Boot()

go func(channel syslog.LogPartsChannel) {
    for logParts := range channel {
        fmt.Println(logParts)
    }
}(channel)

server.Wait()
```

License
-------

MIT, see [LICENSE](LICENSE)
