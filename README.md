## Overview

This is my solution to problem **3.4 User management: cache**. This is an API for user management. It has 9 endpoints, thier parameters listed in Architecture.md and on the endpoint `/swagger/index.html`. Cache was implemented using Redis on `GET` endpoints.

## How to run

Clone the repo: 

    git clone https://git.foxminded.ua/foxstudent106264/task-3.4.git

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

Run the app from cmd directory:

    go run main.go