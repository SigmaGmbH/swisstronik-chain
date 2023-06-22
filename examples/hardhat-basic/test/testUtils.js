require('dotenv').config()

const {ethers} = require('hardhat')
const {getNodePublicKey, encryptECDH, decryptECDH, stringToU8a, deriveEncryptionKey, USER_KEY_PREFIX, hexToU8a, u8aToHex} = require('swisstronik.js')

var nodePublicKey

module.exports.getProvider = () => {
    return new ethers.providers.JsonRpcProvider(process.env.NODE_RPC || 'http://localhost:8545')
}

// Sends shielded query encrypted by random one time key
module.exports.sendShieldedQuery = async (signer, destination, data) => {
    // Create encryption key
    const userEncryptionKey = ethers.utils.randomBytes(32)
    const encryptionPrivateKey = deriveEncryptionKey(userEncryptionKey, stringToU8a(USER_KEY_PREFIX))

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

    const response = await signer.call({
        to: destination,
        data: u8aToHex(encryptedData),
    })

    const decryptionResult = decryptECDH(encryptionPrivateKey, hexToU8a(nodePublicKey), hexToU8a(response))
    if (!decryptionResult.result) {
        throw new Error(`Decryption error. Reason: ${decryptionResult.error}`)
    }

    return decryptionResult.result
}