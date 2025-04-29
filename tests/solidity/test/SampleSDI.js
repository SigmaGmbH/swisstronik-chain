const {ethers} = require("hardhat");
const {expect} = require("chai");
const {randomBytes} = require("@noble/hashes/utils");
const {deriveSecretScalar, derivePublicKey} = require("@zk-kit/eddsa-poseidon");
const {packPoint, inCurve} = require("@zk-kit/baby-jubjub")
const snarkjs = require('snarkjs')
const {buildEddsa} = require('circomlibjs')
const path = require('path');

const DEFAULT_ISSUER_ADDRESS = '0x2fc0b35e41a9a2ea248a275269af1c8b3a061167'
// WARNING: This private key is publicly available
const DEFAULT_PK = "D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB";

const createKeypair = () => {
    const seed = randomBytes(30);
    const privateKey = deriveSecretScalar(seed);
    const publicKey = derivePublicKey(seed);

    if (!inCurve(publicKey)) {
        throw Error('public key is not on curve')
    }

    const compressedKey = packPoint(publicKey);

    return {
        seed, privateKey,
        publicKey: publicKey[0],
        compressedKey,
    };
}

const recoverCredentialHash = async (provider, verificationId) => {
    return await provider.send("eth_credentialHash", [verificationId]);
}

const getIssuanceProofInput = async (provider, credentialHash) => {
    const response = await provider.send("eth_issuanceProof", [credentialHash]);
    const issuanceProof = JSON.parse(response);

    return {
        issuanceRoot: issuanceProof.root,
        issuanceSiblings: fillWithZeroes(issuanceProof.siblings, 33),
        issuanceOldKey: issuanceProof.oldKey,
        issuanceOldValue: issuanceProof.oldValue,
        issuanceIsOld0: issuanceProof.isOld0,
    }
}

const getNonRevocationProofInput = async (provider, credentialHash) => {
    const response = await provider.send("eth_nonRevocationProof", [credentialHash]);
    const revocationProof = JSON.parse(response);

    return {
        revocationRoot: revocationProof.root,
        revocationSiblings: fillWithZeroes(revocationProof.siblings, 33),
        revocationOldKey: revocationProof.oldKey,
        revocationOldValue: revocationProof.oldValue,
        revocationIsOld0: revocationProof.isOld0,
    }
}

const fillWithZeroes = (input, size) => {
    if (input.length >= size) return input;

    const res = new Array(size).fill("0")
    for (let i=0; i<input.length; i++) {
        res[i] = input[i];
    }

    return res;
}

const getProofFiles = () => {
    return {
        sdi: {
            zkey: path.join(process.cwd(), 'test', 'misc', 'sdi.zkey'),
            wasm: path.join(process.cwd(), 'test', 'misc', 'sdi.wasm'),
        }
    }
}

const signMiMC = async (privateKey, message) => {
    const eddsa = await buildEddsa();
    const privKey = Buffer.from(privateKey);
    const msgBuffer = eddsa.F.e(message)
    const signature = eddsa.signMiMCSponge(privKey, msgBuffer);
    return {
        S: signature.S,
        R8: [eddsa.F.toObject(signature.R8[0]), eddsa.F.toObject(signature.R8[1])]
    };
}

const getIssuanceRoot = async (signer) => {
    const contract = await ethers.getContractAt("ComplianceProxy", DEFAULT_ISSUER_ADDRESS, signer)
    return await contract.callStatic.getIssuanceRoot();
}

const getRevocationRoot = async (signer) => {
    const contract = await ethers.getContractAt("ComplianceProxy", DEFAULT_ISSUER_ADDRESS, signer)
    return await contract.callStatic.getRevocationRoot();
}

const getVerificationData = async (signer, address) => {
    const contract = await ethers.getContractAt("ComplianceProxy", DEFAULT_ISSUER_ADDRESS, signer)
    return await contract.callStatic.getVerificationData(address);
}

const constructProof = async (
    signer, 
    verificationId,
    userKeypair
) => {
    const provider = signer.provider;

    const allowedIssuers = [BigInt(DEFAULT_ISSUER_ADDRESS).toString(), "0", "0", "0", "0"];
    const currentTimestamp = Date.now();

    const credentialHash = await recoverCredentialHash(provider, verificationId);
    const issuanceProof = await getIssuanceProofInput(provider, credentialHash);
    const nonRevocationProof = await getNonRevocationProofInput(provider, credentialHash);

    const verificationData = await getVerificationData(signer, signer.address);
    const index = verificationData.length - 1;
    const encodedIssuer = BigInt(verificationData[index].issuerAddress);

    const credentialElements = [
        `${verificationData[index].verificationType}`,
        encodedIssuer.toString(),
        `${verificationData[index].expirationTimestamp}`,
        `${verificationData[index].issuanceTimestamp}`,
    ];

    const holderSignature = await signMiMC(userKeypair.seed, BigInt(credentialHash));

    const input = {
        holderPrivateKey: userKeypair.privateKey,
        ...issuanceProof,
        ...nonRevocationProof,
        credentialElements,
        allowedIssuers,
        currentTimestamp,
        S: holderSignature.S,
        Rx: holderSignature.R8[0],
        Ry: holderSignature.R8[1],
    };

    const proofFiles = getProofFiles();
    const {proof, publicSignals} = await snarkjs.plonk.fullProve(input, proofFiles.sdi.wasm, proofFiles.sdi.zkey);

    const calldata = await snarkjs.plonk.exportSolidityCallData(proof, publicSignals);
    const [encodedProof] = calldata.split(',')
    const proofBytes = encodedProof.trim()

    return [proofBytes, publicSignals]
}

describe('SDI tests', () => {
    let contract
    let userKeypair
    let userSigner;
    let verificationId;

    let provider;

    let verifierContract;

    before(async () => {
        provider = new ethers.providers.JsonRpcProvider('http://localhost:8547'); // Unencrypted rpc url
        const signer = new ethers.Wallet(DEFAULT_PK, provider);
        contract = await ethers.getContractAt('ComplianceProxy', DEFAULT_ISSUER_ADDRESS, signer);

        // Construct user signer
        userSigner = ethers.Wallet.createRandom().connect(provider);

        // Generate user keypair
        userKeypair = createKeypair();

        // Verify user
        const encodedPublicKey = ethers.utils.hexlify(userKeypair.compressedKey)
        const tx = await contract.markUserAsVerifiedV2(userSigner.address, encodedPublicKey, {gasLimit: 500_000});
        const res = await tx.wait();

        expect(res.events[0].args.success).to.be.true
        verificationId = res.events[0].args.data;

        // Deploy verifier
        const verifierFactory = await ethers.getContractFactory("PlonkVerifier", signer);
        verifierContract = await verifierFactory.deploy();
        await verifierContract.deployed();

        // Deploy AirdropSDI contract
        const frontendFactory = await ethers.getContractFactory("AirdropSDI", signer);
        frontendContract = await frontendFactory.deploy(verifierContract.address, [DEFAULT_ISSUER_ADDRESS]);
        await frontendContract.deployed();

        // Initialize contract with roots
        const issuanceRoot = await getIssuanceRoot(signer)
        const revocationRoot = await getRevocationRoot(signer)
        const updateRootsTx = await frontendContract.connect(signer).updateRoots(issuanceRoot, revocationRoot)
        await updateRootsTx.wait()

        // Fund user
        const fundTx = await signer.sendTransaction({to: userSigner.address, value: ethers.utils.parseEther("1.0")}) 
        await fundTx.wait()
    });

    it('Should be able to claim airdrop if proof is valid', async () => {
        const [proofBytes, publicSignals] = await constructProof(userSigner, verificationId, userKeypair);

        const tx = await frontendContract.connect(userSigner).claimAirdrop(proofBytes, publicSignals);
        await tx.wait();
        
        const isEligible = await frontendContract.isUserEligible(userSigner.address)
        expect(isEligible).to.be.true
    });
})