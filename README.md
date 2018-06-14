# syshealth

This tool provides a server binary providing an API and an UI allowing to visualise some basic metrics (CPU, load, memory & disk usage) sent by agents. 

## Why another monitoring tool?

We use several server providers, in a context of a lot of mixed server configurations. We need some isolation between servers, and we cannot have a VPN providing a private network to communicate with each one. Most of well-known monitoring tools provides too many features for our needs, with hard steps to setup correctly.

Agents in syshealth are authenticated using a JWT token. This token is created when registering the monitored server on syshealth API. To be secure, the API needs to be served with TLS.
This architecture is very simple but allows to setup monitoring with ease.

The API can notify on a Slack channel when some metrics go over thresholds.

The server also provides a private API to perform maintenance tasks (i.e DB backups).

## Setup

The configuration is done using command flags or environment variables.

### Server configuration

| Environment variable | Description |
| --- | --- |
| SYSHEALTH_LISTEN_IP | The public API will listen to this IP (i.e. 0.0.0.0) |
| SYSHEALTH_LISTEN_PORT | The public API will listen to this port (i.e. 1323) |
| SYSHEALTH_LISTEN_PRIVATE_PORT | The private API will listen to this port (i.e. 1324) |
| SYSHEALTH_AGENT_JWT_SECRET | Secret used to generate JWT tokens for agent authentication |
| SYSHEALTH_CLIENT_JWT_SECRET | Secret used to generate JWT tokens for API/UI clients |
| SYSHEALTH_SLACK_WEBHOOK_URL | (optional) Slack webhook URL to notify threshold overtaking |
| SYSHEALTH_DATABASE_DIRECTORY | Path where the DB will be stored |

### Agent configuration

| Environment variable | Description |
| --- | --- |
| SYSHEALTH_AGENT_JWT | The token generated by the server to authenticate the agent |
| SYSHEALTH_AGENT_SERVER_URL | Public URL of the API |

## Credits

Thanks to the contributors of the `gopsutil` project https://github.com/shirou/gopsutil.
