# cordctl

`cordctl` is a Discord bot that lets you run shell commands on your server, remotely and securely, just by sending Discord messages. You define what commands it should respond to using simple YAML files. Handy for automation, server management, or just geeking out.

## Why?

I was picking up Golang and also wanted an easy way to run some shell commands on my VM without logging in every time. So I hacked together a Discord bot that could handle a few commands. But as the need grew, I kept editing the code to support more stuff. That got tiring. So I thought—why not make this configurable? Now, I can just drop in new YAML files with command definitions, no code changes needed. It scratched my itch *and* helped me learn Go in the process.

## Goals

- Make remote command execution via Discord flexible and easy
- Let anyone (even those who don’t code) add commands using YAML
- Keep things safe, auditable, and under control

## Requirements

- Go 1.18 or later
- A Discord bot token

## Getting Started

### 1. Clone and Build

```bash
git clone https://github.com/swarupsengupta2007/cordctl
cd cordctl
go build -o cordctl
```

### 2. Create a Discord Bot

- Head over to [Discord Developer Portal](https://discord.com/developers/applications)
- Make a new application and add a bot to it
- Copy the bot token—you’ll need it

### 3. Set Up Your `.env`

Create a `.env` file in your project directory with:

```
DISCORD_BOT_TOKEN=your-bot-token-here
```

This is how the bot knows how to talk to Discord.

### 4. Define Commands in YAML

Drop your YAML command files inside the `commands/` folder. Each file = one command.

## Writing Command YAMLs

A command YAML defines how the bot should react to a slash command. Here's the basic structure:

```yaml
name: <command-name>
description: <description>
command: <shell-command>
args:
  - ...
```

### How args Work

- `{$token}` → Mandatory value. User must provide this option.
- `{?token}` → Optional value. If provided, value is inserted.
- `{!token}` → Optional presence. If present (boolean), inserts pre+post (no value).
- `{=token}` → Optional presence token. If present (boolean), inserts pre+token+post (token name as literal).
- Anything else is passed as-is.
- Everything in passed in sequence it appreared in the args array.

**Example:**

```yaml
name: build
description: Build a project with optional flags
command: make
args:
  - "-B"
  - "{$target}"
  - "-j {?threads}"
  - "-d {!debug}"
  - "{=--no-print-directory}"
  - "--warn-undefined-variables"
```

If user sends `/build target:app threads:10 debug:true -k:true`, the bot will run:

```
make -B app -j 10 -d --no-print-directory --warn-undefined-variables
```
And return the output as response

## Running the Bot

Once you’ve got your `.env` file and your YAML commands ready, run:

```bash
./cordctl /path/to/commands/folder/
```

And that’s it—your bot’s live and listening for slash commands.

## Credits

Built with help from:

- [`gopkg.in/yaml.v3`](https://pkg.go.dev/gopkg.in/yaml.v3) – for reading the YAML files
- [`discordgo`](https://github.com/bwmarrin/discordgo) – the Go lib that talks to Discord

## License

See [LICENSE](LICENSE) for the legal bits.
