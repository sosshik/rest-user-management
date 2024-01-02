## Overview

This is an API for user management. It has 9 endpoints, thier parameters listed in Architecture.md and on the endpoint `/swagger/index.html`. Cache was implemented using Redis on `GET` endpoints. Rating system implemented using ClickHouse.

## How to run

Clone the repo: 

    git clone https://github.com/sosshik/rest-user-management.git

Create `.env` file in cmd directory with parameters: 
- `PORT` - port where you wish to start the bot
- `DATABASE_URL` - your MongoDB connection string
- `CONN_CHECK` - use true or false to enable connection check
- `RECONN_TIME` - time before next connection check
- `LOG_LEVEL` - used to set log level
- `RECONN_TRIES` - used to set amount of reconnections in a row
- `JWT_KEY` - your JWT secret key
- `REDIS_ADDR` - address for Redis
- `REDIS_EXP_TIME` - cache expiration time 
- `CH_ADDR` = ClickHouse address
- `CH_DB` = ClickHouse database name
- `CH_USER` = ClickHouse username
- `CH_PASS` = ClickHouse password

Run the app from cmd directory:

    go run main.go
