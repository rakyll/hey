`hey-consul` is a tiny program that sends some load to a web application.

It is forked from [hey](https://github.com/rakyll/hey) and updated as per the load testing done in [etcd benchmarking tool](https://github.com/etcd-io/etcd)

## Build

* mac `make release-darwin`
* linux `make release-linux`

## Run

* benchmark PUT: `./bin/hey-consul put --t 20000 --disable-keepalive --q 1 --n 500 --c 200 --d 1234 http://192.168.1.5:6555`
* benchmark GET: `./bin/hey-consul get --t 20000 --disable-keepalive --q 1 --n 500 --c 200 http://192.168.1.5:6555`
