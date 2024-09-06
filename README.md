# G-Chat
<img width="1043" alt="G-chat example image" src="https://github.com/user-attachments/assets/9f0e3aa4-e6ea-4a17-adce-2e3d8fcdd604">

G-Chat is a web-based chat client for Habbo hotel origins.
It is an extension for [G-Earth](https://github.com/sirjonasxx/G-Earth)

Special thanks to [b7c](https://github.com/b7c)/[xabbo]() for [goearth](https://xabbo.b7c.io/nx) and [nx](https://xabbo.b7c.io/nx), both of which are used heavily in this project.

## Running

[air](https://github.com/air-verse/air) is used for hot reloading during development, ran by running `air`

## Assets

JS/ CSS can be built to `static/` by running `npm run build` / `npm run dev` (for auto reloading of css)

## Database

G-Chat uses [goose](https://github.com/pressly/goose) and [sqlc](https://github.com/sqlc-dev/sqlc) as tools for managing db migrations/ safe Go code generation from SQL queries

A database can be generated at `./db/app.db` by running migrations from `./db/migrations` with `./goose.sh up`, Go code can be generated for queries in `./db/queries` with `sqlc generate`

