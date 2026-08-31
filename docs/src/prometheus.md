# prometheus

This directive enables prometheus metrics for Casket.

**[Full documentation](https://github.com/tmpim/casket/blob/master/caskethttp/prometheus/README.md)**

## Examples

### Enable metrics

``` casketfile
prometheus
```

For each virtual host that you want to see metrics for. It optionally takes an address where the metrics are exported,
the default is localhost:9180. The metrics path is fixed to /metrics. You&#39;ll need to put this module early in the
chain, so that the duration histogram actually makes sense. I&#39;ve put it at number 0.
