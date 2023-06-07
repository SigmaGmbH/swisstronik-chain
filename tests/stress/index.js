require('dotenv').config()
const ethers = require('ethers') 
const fs = require('fs')
const {getNodePublicKey, encryptECDH, decryptECDH, stringToU8a, deriveEncryptionKey, USER_KEY_PREFIX, hexToU8a, u8aToHex} = require('swisstronik.js')

let nodePublicKey, provider

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
        console.log('eth_getBalance success')
    } catch {
        console.log('eth_getBalance failed')
    }
}

// Initial wallet sends some funds to random wallet, then
// random wallet sends it back
async function sendFundsBetweenWallets() {
    const senderWallet = new Wallet(process.env.FIRST_PRIVATE_KEY, getProvider())
    const receiverWallet = Wallet.createRandom(getProvider()) 
    console.log(`Sending 100 uswtr from ${senderWallet.address} to ${receiverWallet.address}`)
    const tx = await senderWallet.sendTransaction({
        to: receiverWallet.address,
        value: 100
    })
    await tx.wait(1)

    const backTx = await receiverWallet.sendTransaction({
        to: senderWallet.address,
        value: 100
    })
    await backTx.wait(1)
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

async function sendShieldedQuery(provider, privateKey, destination, data, value) {
    // Create encryption key
    const encryptionPrivateKey = deriveEncryptionKey(privateKey, stringToU8a(USER_KEY_PREFIX))

    // Obtain node public key if not presents
    if (!nodePublicKey) {
        const nodePublicKeyResponse = await getNodePublicKey(getProvider().connection.url)
        if (!nodePublicKeyResponse.publicKey) {
            throw new Error(`Cannot obtain node public key. Reason: ${nodePublicKeyResponse.error}`)
        }
        nodePublicKey = nodePublicKeyResponse.publicKey
    }

    // Encrypt data
    const encryptionResult = encryptECDH(encryptionPrivateKey, hexToU8a(nodePublicKey), hexToU8a(data))
    if (!encryptionResult.result) {
        throw new Error(`Encryption error. Reason: ${encryptionResult.error}`)
    }
    const encryptedData = encryptionResult.result

    const response = await provider.call({
        to: destination,
        data: u8aToHex(encryptedData),
        value
    })

    const decryptionResult = decryptECDH(encryptionPrivateKey, hexToU8a(nodePublicKey), hexToU8a(response))
    if (!decryptionResult.result) {
        throw new Error(`Decryption error. Reason: ${decryptionResult.error}`)
    }

    return decryptionResult.result
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

async function startFundsSendingLoop() {
    while (true) {
        await new Promise(r => setTimeout(r, 200));
        await sendFundsBetweenWallets()
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

    startERC20BalanceRequestLoop(contract)
    startERC20BalanceRequestLoop(contract)

    // Basic eth_requests

    // startFundsSendingLoop()
}

main()

