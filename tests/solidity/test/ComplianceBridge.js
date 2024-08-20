const { expect } = require("chai")
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

const unencryptedProvider = new ethers.providers.JsonRpcProvider("http://localhost:8547") // Unencrypted rpc url

describe('ComplianceBridge', () => {
    const CONTRACT_ADDRESS = "0x30252afe8c1683fd184c99a3c44aa5d547d59dd4"
    let contract

    before(async () => {
        const [signer] = await ethers.getSigners()
        const ComplianceProxyFactory = await ethers.getContractFactory('ComplianceProxy')
        contract = new ethers.Contract(CONTRACT_ADDRESS, ComplianceProxyFactory.interface, signer)
    })

    it('Should be able to add verification details', async () => {
        const [signer] = await ethers.getSigners()

        const tx = await sendShieldedTransaction(
            signer,
            contract.address,
            contract.interface.encodeFunctionData("markUserAsVerified", [signer.address])
        )
        const res = await tx.wait()
        const parsedLog = contract.interface.parseLog(res.logs[0])
        expect(parsedLog.args.success).to.be.true
        expect(parsedLog.args.data.length).to.be.greaterThan(0)

        // Confirm that verified status was changed after tx confirmation
        const isVerifiedRespAfterTx = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerified", [signer.address])
        )
        const isVerifiedAfterTx = contract.interface.decodeFunctionResult("isUserVerified", isVerifiedRespAfterTx)
        expect(isVerifiedAfterTx[0]).to.be.true
    })

    it('Should not change state with eth_call', async () => {
        const [signer] = await ethers.getSigners()

        const wallet = ethers.Wallet.createRandom()
        if (signer.address === wallet.address) {
            return
        }

        // Check if random wallet was not verified yet
        const isVerifiedRespBeforeCall = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerified", [wallet.address])
        )
        const isVerifiedBeforeCall = contract.interface.decodeFunctionResult("isUserVerified", isVerifiedRespBeforeCall)
        if (isVerifiedBeforeCall[0]) {
            return
        }

        // Try to call `markUserAsVerified` by using `eth_call`
        const markUserAsVerifiedResp = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("markUserAsVerified", [wallet.address])
        )
        const verificationId = contract.interface.decodeFunctionResult("markUserAsVerified", markUserAsVerifiedResp)
        expect(verificationId.length).to.be.gt(0)

        // Confirm that verified status was not changed yet
        const isVerifiedRespAfterCall = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerified", [wallet.address])
        )
        const isVerifiedAfterCall = contract.interface.decodeFunctionResult("isUserVerified", isVerifiedRespAfterCall)
        expect(isVerifiedAfterCall[0]).to.be.false
    })

    it('Should not change state with eth_estimateGas', async () => {
        const [signer] = await ethers.getSigners()

        const ComplianceProxyFactory = await ethers.getContractFactory('ComplianceProxy')
        const contractUnencrypted = new ethers.Contract(CONTRACT_ADDRESS, ComplianceProxyFactory.interface, unencryptedProvider)

        const wallet = ethers.Wallet.createRandom()
        if (signer.address === wallet.address) {
            return
        }

        // Check if random wallet was not verified yet
        const isVerifiedRespBeforeEstimateGas = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerified", [wallet.address])
        )
        const isVerifiedBeforeEstimateGas = contract.interface.decodeFunctionResult("isUserVerified", isVerifiedRespBeforeEstimateGas)
        if (isVerifiedBeforeEstimateGas[0]) {
            return
        }

        const gas = await contractUnencrypted.estimateGas.markUserAsVerified(signer.address)
        expect(gas.gt(ethers.BigNumber.from(0))).to.be.true

        // Confirm that verified status was not changed yet
        const isVerifiedRespAfterEstimateGas = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerified", [wallet.address])
        )
        const isVerifiedAfterEstimateGas = contract.interface.decodeFunctionResult("isUserVerified", isVerifiedRespAfterEstimateGas)
        expect(isVerifiedAfterEstimateGas[0]).to.be.false

        const tx = await sendShieldedTransaction(
            signer,
            contract.address,
            contract.interface.encodeFunctionData("markUserAsVerified", [wallet.address])
        )
        const res = await tx.wait()
        const parsedLog = contract.interface.parseLog(res.logs[0])
        expect(parsedLog.args.success).to.be.true
        expect(parsedLog.args.data.length).to.be.greaterThan(0)

        // Confirm that verified status was changed after tx confirmation
        const isVerifiedRespAfterTx = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerified", [wallet.address])
        )
        const isVerifiedAfterTx = contract.interface.decodeFunctionResult("isUserVerified", isVerifiedRespAfterTx)
        expect(isVerifiedAfterTx[0]).to.be.true
    })

    it('Should be able to check for specific issuer of verification', async () => {
        const [signer] = await ethers.getSigners()

        const allowedIssuers = [contract.address]
        const isVerifiedResponse = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerifiedBy", [signer.address, allowedIssuers])
        )
        const result = contract.interface.decodeFunctionResult("isUserVerifiedBy", isVerifiedResponse)
        expect(result[0]).to.be.true
    })

    it('Should be able to check for verification without issuers', async () => {
        const [signer] = await ethers.getSigners()

        const isVerifiedResponse = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerified", [signer.address])
        )
        const result = contract.interface.decodeFunctionResult("isUserVerified", isVerifiedResponse)
        expect(result[0]).to.be.true
    })

    it('Should be able to get verification data', async () => {
        const [signer] = await ethers.getSigners()

        const verificationData = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("getVerificationData", [signer.address])
        )
        const result = contract.interface.decodeFunctionResult("getVerificationData", verificationData)
        expect(result[0].length).to.be.greaterThan(0)
        for (const details of result[0]) {
            expect(details.issuerAddress.length).to.be.greaterThan(0)
            expect(details.issuerAddress.toLowerCase()).to.be.equal(contract.address.toLowerCase())
        }
    })
})
