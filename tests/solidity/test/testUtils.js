const {ethers} = require('hardhat')
const {getNodePublicKey, encryptECDH, stringToU8a, deriveEncryptionKey, USER_KEY_PREFIX, hexToU8a, u8aToHex} = require('swisstronik.js')

var nodePublicKey

module.exports.sendShieldedTransaction = async (provider, privateKey, destination, data, value) => {
    // Construct signer from private key
    const wallet = new ethers.Wallet(privateKey)
    const signer = wallet.connect(provider)

    // Create encryption key
    const encryptionPrivateKey = deriveEncryptionKey(privateKey, stringToU8a(USER_KEY_PREFIX))

    // Obtain node public key if not presents
    if (!nodePublicKey) {
        const nodePublicKeyResponse = await getNodePublicKey('http://127.0.0.1:8535')
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
        value
    })
}

module.exports.sendShieldedQuery = async (provider, privateKey, destination, data, value) => {
    // Create encryption key
    const encryptionPrivateKey = deriveEncryptionKey(privateKey, stringToU8a(USER_KEY_PREFIX))

    // Obtain node public key if not presents
    if (!nodePublicKey) {
        const nodePublicKeyResponse = await getNodePublicKey('http://127.0.0.1:8535')
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

    return provider.call({
        to: destination,
        data: u8aToHex(encryptedData),
        value
    })
}