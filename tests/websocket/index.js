require('dotenv').config()
const ethers = require('ethers') 
const fs = require('fs')

let provider

function getProvider() {
    if (!provider) {
        provider = new ethers.providers.WebSocketProvider(process.env.NODE_RPC || 'http://localhost:8546')
    }
    return provider
}

// Subscribe new block data
async function subscribeNewBlockData() {
    const provider = getProvider()
    provider.on("block", async (blockNumber) => {
        const block = await provider.getBlock(blockNumber);
        console.log("New block:", block);
        console.log("Transactions:", block.transactions);
      });
}

// Fetch new block using websocket endpoint.
async function main() {
    subscribeNewBlockData()
}

main()

