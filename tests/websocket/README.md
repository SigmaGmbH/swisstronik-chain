# Stress tester for JSON-RPC

This folder contains emitter of eth_calls and requests to various JSON-RPC endpoints to check
if our node can handle some load. This is not a final version, during the development it will be extended

## Build & run

Create `.env` file with private keys of funded accounts and address of node
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