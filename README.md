# consul-service-replica-count - a consul-template plugin

This is a consul-template plugin that fetches the number of 
service inside a backend(replicas generally) and returns that count * 2 
in stdout in go `String` format.

I had to work on this because i could not find any existing go-template helper 
function that gives the count of a `List` like  `[]map[String]interface{}` .
The scope of this is to use this alongside an haproxy consul-template instance 
to allocate the haproxy golang chans/slots on each backend.
A typical configuration looks like this in consul-template/haproxy conf.
```
backend http_back
    balance roundrobin
    server-template mywebapp 10 _web._tcp.service.consul resolvers consul    resolve-opts allow-dup-ip resolve-prefer ipv4 check
```

Here `10` is the number of hardcoded slots allocated to this backend. If you need something that can dynamcially check the number of 
running replicas of your service and then return that `(count-of-replicas * 2)` as your slot count for each backend. You can think of 
using this plugin.

## Installation
```
cd <your-target-dir>
git clone https://github.com/mainak90/consul-service-replica-count.git
cd consul-service-replica-count/
make build
```

## Usage
Please ensure that this env variable `CONSUL_TEMPLATE_OPTS` is set on the host/container running the consul-template/plugin.
The value should be `CONSUL_TEMPLATE_OPTS=" --addr=<consul-agent/server-ip>:<port>"`

```bash
service-replica-count <version/service-name>
service-replica --no-argument-- returns the usage
```

## Example Consul Template Usage

```
{{ range services }}{{ $servicename := .Name }}
backend {{ $servicename }}_backend
    balance leastconn
    # uncomment to enable bath based balancing
    #reqrep ^([^\ ]*\ /){{ .Name }}[/]?(.*)     \1\2
    server-template {{ $servicename }} {{ plugin "service-replica-count" $servicename }} _{{ $servicename }}._tcp.service.consul resolvers consul resolve-prefer ipv4 check
{{end}}

```

## License

MIT
