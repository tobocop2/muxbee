> Warning: This project is unstable and in active development and code isn't the best and it will be rebuilt soon

# muxbee

[![CI](https://github.com/tobocop2/muxbee/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/tobocop2/muxbee/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/tobocop2/muxbee/graph/badge.svg)](https://codecov.io/gh/tobocop2/muxbee)

> **Early Development:** muxbee is experimental and bridge failures are expected. When bridges stop working, use `muxbee update` (or `u` in TUI) to pull latest images and restart. For local  use only

A batteries-included Matrix starter kit, chat bridge stack, and orchestrator. Built with Go and Docker, a single binary that facilitates putting all your messages in one place. 
The only manual configuration is authentication with each service which doesn't have to be done often and should take a few minutes at most.

![muxbee demo](assets/demo.gif)

```bash
muxbee
```

Run one command. The TUI walks you through setup, starts services, and you're done. Toggle bridges, message the bot to link your account, and your chats appear.

## What It Does

muxbee is a single binary that sets up a self-hosted Matrix server with messaging bridges. All your chats from different platforms in one place. Config files are generated and complexity is handled for you.
Essentially, muxbee is a locally hosted matrix starter kit that also happens to support bridges out of the box. Existing synapse users can use muxbee as a bridge configuration generator and then adapt their own configs with
what muxbee generates


- **Synapse** (Matrix homeserver)
- **Element Web** (bundled chat interface — disable via TUI or `muxbee init --no-element`)
- **Messaging bridges** (WhatsApp, Signal, Discord, Telegram, etc.) — ships with most [mautrix](https://mau.fi/bridges/) bridges by default, but not limited to them. See [Adding a New Bridge](ARCHITECTURE.md#adding-a-new-bridge) for alternatives. PRs welcome!

Like [Bitlbee](https://www.bitlbee.org/), you interact with bridge bots to link accounts (e.g., message `@whatsappbot` and follow the prompts). Unlike Bitlbee, messages sync in real-time, you don't miss messages when offline, and modern features like reactions, threads, and encryption work.

> muxbee runs its own Synapse — ideal for a turnkey setup, but doesn't integrate with existing homeservers yet. 
## Requirements

- **Docker 20.10+** with Compose V2 built-in (`docker compose`, not `docker-compose`) — verify you can run `docker compose`
- 2GB RAM (4GB recommended)
- 10GB disk space

> All bridges and Element enabled pulls ~3GB of Docker images. Toggle only the bridges you need, and disable Element if you prefer your own client.

## Install

Make sure you have the [requirements](#requirements) above.

If you have Go installed:

```bash
go install github.com/tobocop2/muxbee@latest
```

Installs the `muxbee` executable to `$GOPATH/bin` (or `$HOME/go/bin` by default).

Otherwise, download the binary for your platform:

```bash
curl -fsSL https://raw.githubusercontent.com/tobocop2/muxbee/main/scripts/install.sh | sh
```

Move it to your PATH.

Or build from source:
```bash
git clone https://github.com/tobocop2/muxbee.git
cd muxbee && go build -o muxbee .
```

See all releases: https://github.com/tobocop2/muxbee/releases

## Usage

Run `muxbee` for the TUI, or `muxbee --help` for CLI commands.

`muxbee` by default enables element and the TUI will show a link to the element web app. If you are new to the Matrix ecosystem, you can get your feet wet with element for web and configure all your bridges and enjoy watching 
all of your messages appear in one place once you authenticate with each service bot

See [USAGE.md](USAGE.md) for detailed documentation on bridges, connectivity modes, troubleshooting, and more.

## Supported Bridges

| Bridge | Login | Notes |
|--------|-------|-------|
| **WhatsApp** | QR code | Via linked device |
| **Signal** | QR code | No history sync (Signal limitation), needs testing |
| **Discord** | QR code or token | May violate ToS |
| **Telegram** | Phone + code | Requires API credentials from my.telegram.org |
| **Google Messages** | QR code | SMS/RCS |
| **Google Chat** | Cookies | Workspace accounts |
| **Google Voice** | Cookies | SMS/calls |
| **Slack** | Token + cookie | May violate ToS |
| **Bluesky** | Username/password | Needs testing |
| **Meta** | Cookies | Facebook/Instagram, needs testing |
| **Twitter** | Cookies | Needs testing |
| **LinkedIn** | Cookies | Needs testing |
| **IRC** | SASL | |

**Needs testing:** Signal, Bluesky, Meta, Twitter, and LinkedIn have only been verified for bot communication — full functionality needs testing. See [help wanted issues](https://github.com/tobocop2/muxbee/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22).

**Not supported:** iMessage (requires macOS or jailbroken iOS)

See [USAGE.md](USAGE.md) for login details and helper scripts for cookie extraction.

## Linking Your Accounts

muxbee hosts Element Web out of the box. Run `muxbee open` or visit the URL shown after `muxbee up`. Prefer your own client? Disable Element via TUI settings or `muxbee init --no-element`.

Log in with the admin credentials shown during setup. If you missed them:

```bash
muxbee config show --show-secrets
```

Look for the `admin:` section in the output. On Linux/Mac, you can extract it directly:

```bash
muxbee config show --show-secrets | grep -A2 "^admin:"
```

After login, you'll see bridge bots ready to message. Each bot walks you through authentication.

<img src="assets/chats.png" alt="Bridge bots on first Element login" width="1000">

<img src="assets/element.gif" alt="Bridge bot welcome messages" width="1000">

## Helper Scripts

Some bridges require cookie/token extraction from your browser. See [`scripts/cookie-helpers/`](scripts/cookie-helpers/) for Node.js scripts that automate this. Alternative: [mautrix-manager](https://github.com/mautrix/manager).

## Background

Matrix is difficult to set up. Synapse alone has hundreds of configuration options. Add bridges and you're dealing with: appservice registration files with cryptographic tokens, database configuration for each bridge, Docker networking, rate limit tuning, and double-puppeting setup. Each bridge has its own config format. Getting it all working together is difficult and easy to mess up.

[Bitlbee](https://www.bitlbee.org/) and [Pidgin](https://pidgin.im/) are a massive inspiratios. muxbee is similar in spirit: an orchestrator for matrix (for now mostly [mautrix](https://github.com/mautrix)) bridges, all accessible via Matrix. Previous solution limitations: bridges poll (delayed messages), you miss messages when disconnected, no modern E2EE support, no reactions/threads/edits, and formatting gets mangled. Matrix handles all of this natively.

[Beeper](https://beeper.com) also solves this problem with a polished app and cloud-hosted bridges. muxbee is not nearly as polished and is for tinkerers who want full control — no app installs, no cloud dependencies, runs on your hardware. There's some manual setup (messaging bridge bots), but it's simple for QR code bridges like WhatsApp, Discord, and Google Messages. You can also point Beeper or any Matrix client at the Synapse server muxbee sets up but this project bundles an Element web app. Beeper automatically will correct any issues with its cloud hosted bridges while locally hosting will require periodically updating bridges but this is an expected pain point around locally hosting things. With element enabled, muxbee is essentially a fully locally hosted no frills, buggier alternative to beeper (for those that don't need imessage) with some light manual curation. The responsibility for authentication is delegated to bots and following bot instructions.

This tool is meant for folks that loved tools like [Pidgin](https://pidgin.im/) and bitlbee but found them increasingly difficult to maintain as modern protocols evolved. While those ecosystems are awesome, they often require manual compilation and lack support for modern features. muxbee has an even smaller learning curve.

Unlike [Pidgin](https://pidgin.im/) and bitlbee, there is no need to build plugins from source. Everything is containerized right now and synchronizing bridges with the latest versions is all contained within muxbee. If you are like me and feel that the bitlbee/[Pidgin](https://pidgin.im/) bridge experiences were more than satisfactory, then this tool is definitely for you

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for design decisions and configuration details.

## Issues

Found a bug or have a feature request? Check [existing issues](https://github.com/tobocop2/muxbee/issues) first, then [open a new one](https://github.com/tobocop2/muxbee/issues/new) if it doesn't exist.

Some bridges need testing — look for issues labeled [help wanted](https://github.com/tobocop2/muxbee/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22).

## Credits

muxbee wouldn't exist without the incredible [mautrix bridge ecosystem](https://github.com/mautrix) by [Tulir Asokan](https://github.com/tulir). Massive thanks for making many amazing Matrix bridges available to everyone.

## Plans


Some bridges have quirks that need to be addressed but I have been able to send and receive messages reliably since the first version has been released. See issues for the behavior I have observed.

Next, I'd like to add more useful bridges.

Lastly, a streamlined self hosted solution is being prioritized but it's not ready yet. Some preliminary self hosting configs/code exist but these are not trustworthy and I do not recommend running this anywhere but your local machine for the time being.

Ideally, I imagine a docker image and a docker-compose.yml file will be supplied with placeholders for the enabled bridges. At this moment everything is self contained to the executable, but I will be likely removing that functionality as I want this tool to have a single responsibility
and that is to  streamline the orchestration and configuration of synapse with bridges.


## License

MIT
