# syshealth

This tool is intended to be used as a script health check, configured inside Consul (https://www.consul.io).

## Usage

Execute the binary (available in releases). Syshealth prints memory, load average and disk usage and exits with an appropriate return code according to warning/critical levels:

```
0 OK
1 Warning
2 Critical
```

This return code is interpreted by Consul for health checking.

## Roadmap

* Configurable warning/critical levels

## Credits

Thanks to the contributors of the `gopsutil` project https://github.com/shirou/gopsutil.
