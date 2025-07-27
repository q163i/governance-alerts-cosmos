# Governance Alerts Cosmos

A Go service that monitors governance proposals on Cosmos networks and sends notifications when voting is about to start or end.

## Features

- **Real-time monitoring** of governance proposals across multiple Cosmos networks
- **Smart notifications** for voting start/end with configurable time thresholds
- **Multiple notification channels**: Telegram and Slack
- **Startup notifications** to confirm service is running
- **Comprehensive logging** with structured output
- **Production-ready** with error handling and graceful shutdown

## Supported Networks (EXAMPLES)

- **Babylon Mainnet** (`bbn-1`) - via PublicNode REST API
- **ZetaChain Mainnet** (`zetachain_7000-1`) - via BlockPI REST API

## Quick Start

### Prerequisites

- Go 1.22+
- Telegram bot token (optional)
- Slack webhook URL (optional)

### Installation

```bash
# Clone and build
git clone <repository>
cd governance-alerts-cosmos
go build -o governance-alerts-cosmos .

# Run with default config
./governance-alerts-cosmos --config config/config.yaml
```

### Configuration

Edit `config/config.yaml`:

```yaml
# Alert settings
alerts:
  hours_before_start: 24    # Notify 24h before voting starts
  hours_before_end: 6       # Notify 6h before voting ends
  check_interval_minutes: 60 # Check every hour
  notify_on_startup: true   # Send notification when service starts

# Networks
networks:
  babylon-mainnet:
    name: "Babylon Mainnet"
    rest_endpoint: "https://babylon-rest.publicnode.com"
    chain_id: "bbn-1"
  
  zetachain-mainnet:
    name: "ZetaChain Mainnet"
    rest_endpoint: "https://zetachain-athens.blockpi.network/lcd/v1/public"
    chain_id: "zetachain_7000-1"

# Notifications
notifications:
  telegram:
    enabled: true
    bot_token: "YOUR_BOT_TOKEN"
    chat_id: 123456789
```

## Architecture

```
governance-alerts-cosmos/
├── cmd/                    # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── governance/        # Cosmos governance client
│   ├── notifications/     # Notification handlers
│   ├── service/           # Core service logic
│   └── types/             # Data structures
├── config/                # Configuration files
└── docs/                  # Documentation
```

## Development

### Building

```bash
go build -o governance-alerts-cosmos .
```

### Testing

```bash
go test ./...
```

### Running

```bash
# With custom config
./governance-alerts-cosmos --config /path/to/config.yaml

# With debug logging
./governance-alerts-cosmos --log-level debug
```

## Monitoring

### Health Checks

The service provides:
- Startup notifications
- Real-time proposal monitoring
- Error logging for network issues
- Graceful shutdown handling

### Logs

```bash
# View logs
tail -f governance-alerts-cosmos.log

# Check for errors
grep "Error" governance-alerts-cosmos.log
```

## Production Deployment

### Docker

```bash
# Build image
docker build -t governance-alerts-cosmos .

# Run container
docker run -d \
  --name governance-alerts-cosmos \
  -v $(pwd)/config:/app/config \
  governance-alerts-cosmos
```

### Systemd Service

Create `/etc/systemd/system/governance-alerts-cosmos.service`:

```ini
[Unit]
Description=Governance Alerts Cosmos Service
After=network.target

[Service]
Type=simple
User=governance
WorkingDirectory=/opt/governance-alerts-cosmos
ExecStart=/opt/governance-alerts-cosmos/governance-alerts-cosmos --config /opt/governance-alerts-cosmos/config/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Troubleshooting

### Common Issues

1. **Network connectivity errors**: Check REST endpoint availability
2. **Telegram bot errors**: Verify bot token and chat ID
3. **No proposals found**: Networks may not have active governance proposals

### Debug Mode

```bash
./governance-alerts-cosmos --log-level debug --config config/config.yaml
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License 