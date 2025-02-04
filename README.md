# My Telegram Bot

This project is a simple Telegram bot built using Go. It demonstrates how to interact with the Telegram Bot API and handle incoming messages.

## Project Structure

```
go_binance_bot
├── database 
│ ├── db.sqlite3 # SQLite database file 
│ └── .gitkeep # Keep the database directory in the repository 
├── src 
│ ├── apis 
│ │ └── binance.go # Binance API interaction 
│ ├── bot 
│ │ └── bot.go # Bot struct and methods for API interaction 
│ ├── config 
│ │ ├── bot.go # Bot configuration struct 
│ │ └── setting.go # Load and manage settings 
│ ├── db 
│ │ ├── db.go # Database initialization and connection 
│ │ └── user.go # User-related database operations 
│ ├── handlers 
│ │ └── message_handler.go # Message handling logic 
│ ├── helpers 
│ │ ├── msg_gen 
│ │ │ └── config_mes.go # Message generation for configuration 
│ │ └── priority_queue 
│ │   └── queue.go # Priority queue implementation 
│ ├── jobs 
│ │ ├── job.go # Job struct and methods for job processing 
│ │ └── manager.go # Job manager for handling multiple jobs 
│ ├── utils 
│ │ └── utils.go # Utility functions 
│ └── main.go # Entry point of the application 
├── .env # Environment variables 
├── .gitignore # Git ignore file 
├── Dockerfile # Dockerfile for containerizing the application 
├── go.mod # Module definition and dependencies 
├── go.sum # Dependency checksums 
└── README.md # Project documentation
```

## Setup Instructions

1. Clone the repository:
   ```
   git clone https://github.com/agayen/go_binance_bot.git
   cd go_binance_bot
   ```

2. Install the necessary dependencies:
   ```
   go mod tidy
   ```

3. Set up your Telegram bot by talking to [BotFather](https://t.me/botfather) and obtain your bot token.

4. Update the bot token in `src/main.go`.

## Env

Need to create a `.env` file in root to the dir

```env
TELEGRAM_BOT_TOKEN=8163178289:AAHENuiewaGeXp1mNES0utUGpJb023dk1JgjJQ
AUTHORIZED_USERS=2085862918,6426250048
DATABASE_PATH=database/db.sqlite3
BINANCE_URL=https://api.binance.com
PROCESS_QUEUE_INTERVAL=1
CALL_JOB_INTERVAL=1
```

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
docker run --env-file .env go_binance_bot
```

### Mount the db also

```bash
docker run -d --env-file .env -v $(pwd)/database:/database --name telegram_bot_my go_binance_bot
```

## TODO:
1. Need to fix docker support to delpoy app
2. Some Bugs fixes
   - admin flow test and fix some part.
   - put the tracker properly.
   - minimum ammount filter not working
   - `system error` why coming in api call 
   - totalamount and investment amount is not working
3. Find out the way to make a parallel request
   