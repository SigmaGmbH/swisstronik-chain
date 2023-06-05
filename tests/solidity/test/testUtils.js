require('dotenv').config()

const {ethers} = require('hardhat')
const {getNodePublicKey, encryptECDH, decryptECDH, stringToU8a, deriveEncryptionKey, USER_KEY_PREFIX, hexToU8a, u8aToHex} = require('swisstronik.js')

var nodePublicKey

module.exports.getProvider = () => {
    return new ethers.providers.JsonRpcProvider(process.env.NODE_RPC || 'http://localhost:8545')
}

module.exports.sendShieldedTransaction = async (provider, privateKey, destination, data, value) => {
    // Construct signer from private key
    const wallet = new ethers.Wallet(privateKey)
    const signer = wallet.connect(provider)

    // Create encryption key
    const encryptionPrivateKey = deriveEncryptionKey(privateKey, stringToU8a(USER_KEY_PREFIX))

    // Obtain node public key if not presents
    if (!nodePublicKey) {
        const nodePublicKeyResponse = await getNodePublicKey(this.getProvider().connection.url)
        if (!nodePublicKeyResponse.publicKey) {
            throw new Error(`Cannot obtain node public key. Reason: ${nodePublicKeyResponse.error}`)
        }
        nodePublicKey = nodePublicKeyResponse.publicKey
    }

    // Encrypt data
    const encryptionResult = encryptECDH(encryptionPrivateKey, hexToU8a(nodePublicKey), hexToU8a(data))
    if (!encryptionResult.result) {
        throw new Error(`Encryption error. Reason: ${encryptedData.error}`)
    }
    const encryptedData = encryptionResult.result

    // Construct and sign transaction with encrypted data
    return await signer.sendTransaction({
        from: signer.address,
        to: destination,
        data: u8aToHex(encryptedData),
        value,
        gasLimit: 10000000000,
    })
}

module.exports.sendShieldedQuery = async (provider, privateKey, destination, data, value) => {
    // Create encryption key
    const encryptionPrivateKey = deriveEncryptionKey(privateKey, stringToU8a(USER_KEY_PREFIX))

    // Obtain node public key if not presents
    if (!nodePublicKey) {
        const nodePublicKeyResponse = await getNodePublicKey(this.getProvider().connection.url)
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