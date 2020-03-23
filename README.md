# simplealgorandproxy
Very simple proxy golang using Echo and Algorand golang SDK

## To build
go get -u github.com/labstack/echo/...
go get github.com/algorand/go-algorand-sdk/...

go build simplealgorandproxy.go

## To run
We use Daemonize to run the proxy:
(https://software.clapper.org/daemonize/)

daemonize -o /var/log/algorandproxy/access.log -e /var/log/algorandproxy/access.log -v /root/simplealgorandproxy

*WARNING*: modify all data and token, password, etc in the code before build and run the stub.

Code by https://www.gt50.org/
