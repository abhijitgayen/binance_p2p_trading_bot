# My Telegram Bot

This project is a simple Telegram bot built using Go. It demonstrates how to interact with the Telegram Bot API and handle incoming messages.

## Project Structure

```
my-telegram-bot
├── src
│   ├── main.go          # Entry point of the application
│   ├── bot
│   │   └── bot.go      # Bot struct and methods for API interaction
│   ├── handlers
│   │   └── message_handler.go # Message handling logic
│   └── utils
│       └── utils.go    # Utility functions
├── go.mod               # Module definition and dependencies
└── README.md            # Project documentation
```

## Setup Instructions

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/my-telegram-bot.git
   cd my-telegram-bot
   ```

2. Install the necessary dependencies:
   ```
   go mod tidy
   ```

3. Set up your Telegram bot by talking to [BotFather](https://t.me/botfather) and obtain your bot token.

4. Update the bot token in `src/main.go`.

## Usage

To run the bot, execute the following command:

```
go run src/main.go
```

## build

```
go build -o bin/go_binance_bot src/main.go
```

## Docker

### Build Docker Image

Builds the Docker image and tags it as go_binance_bot.

```bash
docker build -t go_binance_bot .
```
### Run Docker Container

```bash
docker run --env-file .env -p go_binance_bot
```