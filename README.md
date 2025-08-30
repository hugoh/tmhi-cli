# tmhi-cli - CLI to control T-Mobile Home Internet gateway

## Usage

<!-- markdownlint-disable-next-line fenced-code-language -->
```
NAME:
   tmhi-cli - Utility to interact with T-Mobile Home Internet gateway

USAGE:
   tmhi-cli [global options] [command [command options]]

COMMANDS:
   login    Verify that the credentials can log the tool in
   reboot   Reboot the router
   info     Get gateway information
   status   Check gateway status
   req      Make a custom HTTP request to the gateway
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config string, -c string  use the specified TOML configuration file (default: "/Users/hugoh/.tmhi-cli.toml")
   --debug, -d                 display debugging output in the console (default: false)
   --dry-run, -D               do not perform any change to the gateway (default: false)
   --gateway.model string      gateway model: options: ARCADYAN, NOK5G21
   --gateway.ip string         gateway IP (default: "192.168.12.1")
   --login.username string     admin username (default: "admin")
   --login.password string     admin password
   --retries int               number of retries (default: 0)
   --timeout duration          request timeout in seconds (default: 5s)
   --help, -h                  show help
   --version, -v               print the version
```

## See also

- [hugoh/hubitat-tmo-gateway: Hubitat T-Mobile Internet Gateway Driver](https://github.com/hugoh/hubitat-tmo-gateway)
- [cloud-unpacked/tmhi: TMHI is a CLI to manage your local T-Mobile Home Internet 5G Gateway/Router.](https://github.com/cloud-unpacked/tmhi)
- [highvolt-dev/tmo-monitor: A lightweight, cross-platform Python 3 script that can monitor the T-Mobile Home Internet Nokia, Arcadyan, and Sagecom 5G Gateways for 4G/5G bands, cellular site (tower), and internet connectivity and reboots as needed or on-demand.](https://github.com/highvolt-dev/tmo-monitor)
- [highvolt-dev/tmo-live-graph: A simpe react app that plots a live view of the T-Mobile Home Internet Nokia 5G Gateway signal stats, helpful for optimizing signal.](https://github.com/highvolt-dev/tmo-live-graph)
