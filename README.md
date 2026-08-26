<p align="center">
  <img src="images/logo.svg" alt="Chatdoor Logo" width="400">
</p>

[![Build CI](https://github.com/SuperDaveAU/chatdoor/actions/workflows/build.yml/badge.svg)](https://github.com/SuperDaveAU/chatdoor/actions/workflows/build.yml)
[![Go Mod Vuln Scan](https://github.com/SuperDaveAU/chatdoor/actions/workflows/security.yml/badge.svg)](https://github.com/SuperDaveAU/chatdoor/actions/workflows/security.yml) 
[![Dependabot](https://img.shields.io/badge/Dependabot-active-brightgreen?logo=dependabot&logoColor=white)](https://github.com/SuperDaveAU/chatdoor/network/updates)
![Go version](https://img.shields.io/github/go-mod/go-version/:user/:repo)


**Covert peer-to-peer chat disguised as IoT traffic**

ChatDoor is a privacy-focused, end-to-end encrypted chat application that disguises messages as normal IoT device telemetry. To network observers, it looks like a simple home automation dashboard, but it's actually a secure communication channel.

## Features

- 🔐 **End-to-end encryption** using Curve25519 + NaCl box
- 🎭 **Traffic obfuscation** - Messages hidden in fake IoT sensor data
- 🔑 **Simple key management** - Share a public key to connect
- 💬 **Direct P2P** - No central server, just two peers
- 🏠 **Realistic cover** - Simulates temperature sensors, motion detectors, door/window sensors
- 🚫 **No metadata logging** - Ephemeral connections, no history stored

## How It Works

ChatDoor operates in two modes:

### Server Mode
1. Generates a Curve25519 keypair
2. Listens on a WebSocket port (default: 8080)
3. Displays public key for sharing
4. Waits for client to connect
5. Exchanges keys and establishes encrypted channel
6. Generates fake IoT traffic as cover

### Client Mode
1. Takes server's public key (via file or command line)
2. Generates ephemeral session keypair
3. Connects to server via WebSocket
4. Verifies server's public key
5. Establishes encrypted channel
6. Generates fake IoT traffic as cover

### Cover Traffic

Both sides continuously generate realistic fake IoT messages:
- Temperature sensors (gradual drift between 15-28°C)
- Humidity sensors (30-80%)
- Motion detectors (random triggers)
- Door/window sensors (occasional state changes)

Real messages are disguised as door/window events with encrypted payloads embedded in the JSON.

## Quick Start

### Build

```bash
make build
```

### Start Server

```bash
./bin/chatdoor server
```

This will:
1. Generate a new keypair (saved to `chatdoor_private.key` and `chatdoor_public.key`)
2. Display the public key to share with your peer
3. Start listening on port 8080

### Connect as Client

```bash
# Using public key file
./bin/chatdoor client -addr localhost:8080 -pubkey chatdoor_public.key

# Using public key string
./bin/chatdoor client -addr example.com:8080 -key "base64key..."

# Interactive prompt
./bin/chatdoor client -addr localhost:8080
```

### Chat

Once connected, simply type messages and press Enter. They'll be encrypted and disguised as IoT events.

Type `/quit` to exit.

## Usage Examples

### Local Testing

**Terminal 1 (Server):**
```bash
$ chatdoor server

 ██████╗██╗  ██╗ █████╗ ████████╗██████╗  ██████╗  ██████╗ ██████╗ 
██╔════╝██║  ██║██╔══██╗╚══██╔══╝██╔══██╗██╔═══██╗██╔═══██╗██╔══██╗
██║     ███████║███████║   ██║   ██║  ██║██║   ██║██║   ██║██████╔╝
██║     ██╔══██║██╔══██║   ██║   ██║  ██║██║   ██║██║   ██║██╔══██╗
╚██████╗██║  ██║██║  ██║   ██║   ██████╔╝╚██████╔╝╚██████╔╝██║  ██║
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝  ╚═╝

🔑 Generating new keypair...
✓ Private key saved to: chatdoor_private.key
✓ Public key saved to: chatdoor_public.key

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📤 Share this public key with your peer:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
YourBase64PublicKeyHere...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🏠 IoT Dashboard running on ws://localhost:8080/ws
💬 Waiting for connection...

✓ Client connected!
🔐 Secure channel established
📊 Generating fake IoT traffic...

> Hello from server!
[Them]: Hey, I got your message!
```

**Terminal 2 (Client):**
```bash
$ chatdoor client -addr localhost:8080 -pubkey chatdoor_public.key
📂 Loading public key from: chatdoor_public.key
✓ Valid public key loaded
🔌 Connecting to ws://localhost:8080/ws...
✓ Connected!
🔐 Secure channel established
🏠 Observing IoT devices...

[Them]: Hello from server!
> Hey, I got your message!
```

### Remote Connection

**Server (on a VPS or home server):**
```bash
chatdoor server -port 8080
# Share the public key via Signal, email, QR code, etc.
```

**Client:**
```bash
chatdoor client -addr yourserver.com:8080 -key "base64_public_key..."
```

## Security

### Cryptography

- **Key Exchange**: Out-of-band public key verification prevents MITM
- **Encryption**: NaCl box (Curve25519 + XSalsa20-Poly1305)
- **Forward Secrecy**: Client generates ephemeral session keys
- **Authentication**: Public key cryptography ensures only intended recipient can decrypt

### Threat Model

**What ChatDoor protects against:**
- ✅ Passive network surveillance (messages are encrypted)
- ✅ Traffic analysis (disguised as IoT telemetry)
- ✅ Man-in-the-middle (public key verification)
- ✅ Metadata collection (no central server, no logs)

**What ChatDoor does NOT protect against:**
- ❌ Compromised endpoints (keyloggers, malware)
- ❌ Traffic correlation over long periods (timing analysis)
- ❌ Active adversaries with access to both endpoints
- ❌ Quantum computer attacks (uses elliptic curve crypto)

### Best Practices

1. **Share keys securely** - Use a trusted channel (in person, Signal, PGP-signed email)
2. **Verify fingerprints** - Compare public key hashes out-of-band
3. **Use Tor** - Route traffic through Tor for additional anonymity
4. **Clean up** - Delete key files after sessions
5. **Don't reuse keys** - Generate new keypairs for different conversations

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                          Server Mode                        │
│                                                             │
│  ┌─────────────┐      ┌──────────────┐                      │
│  │   Keypair   │─────▶│  WebSocket   │                      │
│  │ Generation  │      │   Server     │                      │
│  └─────────────┘      └──────┬───────┘                      │
│                              │                              │
│                              ▼                              │
│                       ┌──────────────┐                      │
│                       │ Key Exchange │                      │
│                       └──────┬───────┘                      │
│                              │                              │
│         ┌────────────────────┼────────────────────┐         │
│         ▼                    ▼                    ▼         │
│  ┌──────────────┐     ┌──────────────┐    ┌──────────────┐  │
│  │ Fake Traffic │     │   Encrypt    │    │   Decrypt    │  │
│  │  Generator   │     │   & Send     │    │  & Display   │  │
│  └──────────────┘     └──────────────┘    └──────────────┘  │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                          Client Mode                        │
│                                                             │
│  ┌─────────────┐      ┌──────────────┐                      │
│  │   Server    │─────▶│  WebSocket   │                      │
│  │  Public Key │      │   Client     │                      │
│  └─────────────┘      └──────┬───────┘                      │
│                              │                              │
│                              ▼                              │
│                       ┌──────────────┐                      │
│                       │ Key Exchange │                      │
│                       └──────┬───────┘                      │
│                              │                              │
│          ┌───────────────────┼────────────────────┐         │
│          ▼                   ▼                    ▼         │
│  ┌──────────────┐     ┌──────────────┐    ┌──────────────┐  │
│  │ Fake Traffic │     │   Encrypt    │    │   Decrypt    │  │
│  │  Generator   │     │   & Send     │    │  & Display   │  │
│  └──────────────┘     └──────────────┘    └──────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
chatdoor/
├── cmd/
│   └── chatdoor/
│       └── main.go              # Entry point, CLI handling
├── internal/
│   ├── crypto/
│   │   └── crypto.go            # Curve25519 key management & encryption
│   ├── iot/
│   │   └── traffic.go           # Fake IoT message generation
│   ├── server/
│   │   └── server.go            # WebSocket server
│   └── client/
│       └── client.go            # WebSocket client
├── go.mod
├── Makefile
└── README.md
```

## Development

### Prerequisites

- Go 1.25.6+
- Make (optional)

### Building

```bash
# Download dependencies
go mod download

# Build binary
make build

# Run tests
make test

# Install globally
make install
```

### Testing Locally

Use two terminals:

```bash
# Terminal 1
make server

# Terminal 2  
make client
```

## Advanced Usage

### Custom Port

```bash
chatdoor server -port 9000
```

### Reuse Existing Key

```bash
chatdoor server -key chatdoor_private.key
```

### Behind Nginx Reverse Proxy

```nginx
location /ws {
    proxy_pass http://localhost:8080/ws;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
    proxy_set_header Host $host;
}
```

### Through Tor

**Server:**
```bash
# Configure hidden service in torrc
HiddenServiceDir /var/lib/tor/chatdoor/
HiddenServicePort 80 127.0.0.1:8080

# Start server
chatdoor server
```

**Client:**
```bash
# Connect via Tor SOCKS proxy
torsocks chatdoor client -addr your-onion.onion:80 -pubkey server.pub
```

## FAQ

**Q: Is this secure?**  
A: The encryption is solid (NaCl box), but security depends on operational security. Share keys securely, verify them, and understand the threat model.

**Q: Can this be detected?**  
A: Sophisticated traffic analysis might detect patterns. The fake IoT traffic is realistic but not perfect. Use Tor for additional protection.

**Q: Why not just use Signal/WhatsApp?**  
A: ChatDoor is for scenarios where you want direct P2P communication without intermediaries, or when you need plausible deniability about what kind of traffic you're generating.

**Q: Can I use this for group chat?**  
A: No, it's designed for 1-on-1 communication only.

**Q: What happens if connection drops?**  
A: You'll need to reconnect manually. There's no automatic reconnection or message queuing.

## License

GPL3 License - See LICENSE file


