# Binance p2p trading Bot

This project is a lightweight Telegram bot developed in Go (Golang) for automating Binance P2P trading operations. It uses the Telegram Bot API to process commands and deliver real-time updates, making it easy to manage and monitor P2P trading directly from Telegram.

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
│── main.go # Entry point of the application 
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
TELEGRAM_BOT_TOKEN=7621952856:AAH_KESBZzc1X0I00u4nRpimNw3EixfzyGU
AUTHORIZED_USERS=2085862928,6426250648
NOTIFY_USER_ID=2085862928
DATABASE_PATH=database/db.sqlite3
BINANCE_URL=https://api.binance.com
PROCESS_QUEUE_INTERVAL=100
CALL_JOB_INTERVAL=100
CREATE_ORDER_INTERVAL=9
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

- On Linux/Mac:

   ```bash
   go build -o go_binance_bot .
   ```

- On Windows:
   ```bash
   go build -o go_binance_bot.exe .
   ```
### Run Bot

```bash
ENV_FILE=.env ./go_binance_bot
```

## Docker Build

### Docker image build


```bash
GOOS=linux GOARCH=amd64 go build -o go_binance_bot .
```

```bash
docker build -t my-go-bot .
```

### Run docker container

```bash
docker run --rm --env-file .env my-go-bot
```


## TODO:
1. Need to fix docker support to delpoy app
2. bugs
  - `This Ad has an insufficient balance. Please choose another Ad.` 
         -> need to find some optimiz solution for this.
         -> this ads capture by another bot.
3. Find out the way to make a parallel request
4. Need to integrate the message properly.
