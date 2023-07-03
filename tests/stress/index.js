require('dotenv').config()
const ethers = require('ethers') 
const fs = require('fs')
const {decryptNodeResponse} = require('@swisstronik/swisstronik.js')

let provider

const NUM_PARALLEL = 3

function getProvider() {
    if (!provider) {
        provider = new ethers.providers.JsonRpcProvider(process.env.NODE_RPC || 'http://localhost:8545')
    }
    return provider
}

// Creates promise for eth_balance request
async function requestBalance() {
    try {
        const wallet = ethers.Wallet.createRandom(getProvider())
        console.log(`request eth_balance for ${wallet.address}`)
        const balance = await getProvider().getBalance(wallet.address)
        console.log('eth_getBalance success. Balance ', balance)
    } catch {
        console.log('eth_getBalance failed')
    }
}

// Deploys sample ERC20 contract
async function deployERC20() {
    const metadata = JSON.parse(fs.readFileSync('contracts/ERC20Token.json'))
    const senderWallet = new ethers.Wallet(process.env.FIRST_PRIVATE_KEY, getProvider())
    const factory = new ethers.ContractFactory(metadata.abi, metadata.bytecode, senderWallet);

    const contract = await factory.deploy('test token', 'TT', 10000000);
    await contract.deployed()

    return contract
}

const sendShieldedQuery = async (provider, destination, data) => {
    // Encrypt call data
    const [encryptedData, usedEncryptedKey] = await encryptDataField(
        provider.connection.url,
        data
    )

    // Do call
    const response = await provider.call({
        to: destination,
        data: encryptedData,
    })

    // Decrypt call result
    return await decryptNodeResponse(provider.connection.url, response, usedEncryptedKey)
}

async function requestERC20Balance(contract) {
    try {
        const wallet = ethers.Wallet.createRandom(getProvider())
        console.log(`request erc20 balance for ${wallet.address}`)
        const balanceResponse = await sendShieldedQuery(
            getProvider(),
            wallet.privateKey,
            contract.address,
            contract.interface.encodeFunctionData("balanceOf", [wallet.address])
        );
        const balance = contract.interface.decodeFunctionResult("balanceOf", balanceResponse)[0]
        console.log('erc20_balance success')
    } catch {
        console.log('erc20_balance failed')
    }
}

async function startERC20BalanceRequestLoop(contract) {
    while (true) {
        await new Promise(r => setTimeout(r, 100));
        requestERC20Balance(contract)
    }
}

async function startBalanceRequestLoop() {
    while (true) {
        await new Promise(r => setTimeout(r, 100));
        requestBalance()
    }
}

// Deploys test ERC20 contract and starts to send different eth_calls to it
async function main() {
    console.log('Deploying test ERC20')
    const contract = await deployERC20()
    console.log(`ERC20 deployed with address: ${contract.address}`)


    for (let i=0; i<NUM_PARALLEL; i++) {
        startERC20BalanceRequestLoop(contract)
        startBalanceRequestLoop()
    }
}

main()

