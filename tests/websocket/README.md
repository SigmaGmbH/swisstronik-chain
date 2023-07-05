# Websocket tester for JSON-RPC

This folder contains a WebSocket event listener for the new block with transactions using ether.js Websocketprovider. Its purpose is to fetch a new block with transactions and output it to the console to verify if our node is running and its websocket connection is stable. Please note that this is not the final version; it will be further extended during development.

## Build & run

Create `.env` file with websocket endpoint.
```sh
cp .env.example .env
```

Install all necessary packages by running:
```sh
npm install
```

After that run script:
```sh
node index.js
```