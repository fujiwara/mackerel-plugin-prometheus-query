# mackerel-plugin-prometheus-query

Prometheus query plugin for Mackerel.

## Usage

```
Usage of mackerel-plugin-prometheus-query:
Usage of ./mackerel-plugin-prometheus-query:
  -address string
    	Prometheus address (default "http://localhost:9090")
  -emit-zero
    	emit 0 when query returns no result
  -metric-key-format string
    	Metric key format
  -query string
    	PromQL query
  -timeout string
    	timeout for query (default "10s")
```

## Example

mackerel-plugin-prometheus-query runs PromQL on prometheus and outputs result as a mackerel plugin metric format.

```console
$ mackerel-plugin-prometheus-query -query "up" -metric-key-format "promq.{job}.{instance}"
promq.web.10_1_129_175_9901	1	1575941187
promq.web.10_1_130_170_9901	1	1575941187
promq.web.10_1_131_53_9901	1	1575941187
promq.prometheus.localhost_9090	1	1575941187
```

This example is equivalent to results of promtool as below.

```
$ promtool query instant http://localhost:9090 up
up{instance="10.1.129.175:9901", job="web"} => 1 @[1575941187.73]
up{instance="10.1.130.170:9901", job="web"} => 1 @[1575941187.73]
up{instance="10.1.131.53:9901", job="web"} => 1 @[1575941187.73]
up{instance="localhost:9090", job="prometheus"} => 1 @[1575941187.73]
```

## Format metric keys

A plugin uses `-metric-key-format` option to make metric key.

`{foo}` is replaced as metric value for label "foo".

## LICENSE

MIT
