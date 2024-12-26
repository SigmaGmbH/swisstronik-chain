const {ethers} = require("hardhat");
const {sendShieldedQuery, sendShieldedTransaction} = require("./testUtils");
const {expect} = require("chai");
const {randomBytes} = require("@noble/hashes/utils");
const {deriveSecretScalar, derivePublicKey, packPublicKey} = require("@zk-kit/eddsa-poseidon");
const {packPoint, inCurve} = require("@zk-kit/baby-jubjub")

const DEFAULT_PROXY_CONTRACT_ADDRESS = '0x2fc0b35e41a9a2ea248a275269af1c8b3a061167'
// WARNING: This private key is publicly available
const DEFAULT_PK = "D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB";

const createKeypair = () => {
    const seed = randomBytes(30);
    const privateKey = deriveSecretScalar(seed);
    const publicKey = derivePublicKey(seed);

    console.log('created public key: ', publicKey)

    if (!inCurve(publicKey)) {
        throw Error('public key is not on curve')
    }

    const compressedKey = packPoint(publicKey);
    console.log('compressed key: ', compressedKey.toString())

    return {
        seed, privateKey,
        publicKey: publicKey[0],
        compressedKey,
    };
}

const recoverCredentialHash = async (provider, verificationId) => {
    const res = await provider.send("eth_credentialHash", [verificationId]);
    console.log('DEBUG credential hash: ', res);
    return res;
}

describe('SDI tests', () => {
    let contract
    let userKeypair
    let userSigner;
    let verificationId;

    let provider

    before(async () => {
        provider = new ethers.providers.JsonRpcProvider('http://localhost:8547'); // Unencrypted rpc url
        const signer = new ethers.Wallet(DEFAULT_PK, provider);
        contract = await ethers.getContractAt('ComplianceProxy', DEFAULT_PROXY_CONTRACT_ADDRESS, signer);

        // Construct user signer
        userSigner = ethers.Wallet.createRandom().connect(provider);

        // Generate user keypair
        userKeypair = createKeypair();

        // Verify user
        const encodedPublicKey = ethers.utils.hexlify(userKeypair.compressedKey)
        console.log(encodedPublicKey)
        const tx = await contract.markUserAsVerifiedV2(userSigner.address, encodedPublicKey);
        const res = await tx.wait();

        expect(res.events[0].args.success).to.be.true
        verificationId = res.events[0].args.data;
    });

    it('Should be able to verify correct proof', async () => {
        const credentialHash = await recoverCredentialHash(provider, verificationId)
        const issuanceProof = await provider.send("eth_issuanceProof", [credentialHash]);
        const revocationProof = await provider.send("eth_nonRevocationProof", [credentialHash]);

        console.log('debug issuance proof: ', issuanceProof);
        console.log('debug revocation proof: ', revocationProof);
    })
})