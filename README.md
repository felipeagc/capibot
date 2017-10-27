# Capibot
## Features
 - Plays audio from youtube
 - Playlist is saved in a database
 - Automatically joins a voice channel with a suggestive name
 - Skip-voting system


## How to run
Create a `.env` file at the root of this repo and add the following environment variables:
 - `DISCORD_TOKEN`
 - `GOOGLE_KEY`
 - `POSTGRES_USER`
 - `POSTGRES_PASSWORD`
 - `POSTGRES_DB`
 
Use the following command to run the bot:
```
docker-compose up --build
```
