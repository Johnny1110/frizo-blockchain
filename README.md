# Frizo Blockchain

<br>

---

<br>

A blockchain implements by golang (academic research purpose)

<br>

Project Structure:

```
frizo-blockchain/
├── cmd/
│   └── frizo/              # Main application entry point
│       └── main.go         # Only main package here
├── internal/               # Private packages (can't be imported by other projects)
│   ├── core/         # Core blockchain logic
│   │   ├── block.go
│   │   ├── chain.go
│   │   └── transaction.go
│   ├── consensus/          # Consensus algorithms
│   │   ├── pos.go
│   │   └── pow.go
│   ├── network/           # P2P networking
│   │   ├── peer.go
│   │   └── protocol.go
│   ├── wallet/            # Wallet functionality
│   │   ├── wallet.go
│   │   └── keys.go
│   └── storage/           # Database/storage layer
│       ├── leveldb.go
│       └── memory.go
├── pkg/                   # Public packages (can be imported by other projects)
│   ├── crypto/            # Cryptographic utilities
│   │   ├── hash.go
│   │   └── signature.go
│   └── utils/             # General utilities
│       ├── logger.go
│       └── config.go
├── api/                   # API definitions (REST, gRPC, etc.)
│   ├── rest/
│   └── rpc/
├── config/                # Configuration files
├── scripts/               # Build and deployment scripts
├── docs/                  # Documentation
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Logs

* [EP1](https://www.youtube.com/watch?v=oCm46sUILcs&list=PL0xRBLFXXsP6-hxQmCDcl_BHJMm0mhxx7&index=1&t=176s&ab_channel=AnthonyGG)
* [EP2](https://youtu.be/_f6SNxI2mog?list=PL0xRBLFXXsP6-hxQmCDcl_BHJMm0mhxx7&t=1924)