const { expect } = require("chai");
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('ComplianceBridge', () => {
    let contract

    before(async () => {
        const Counter = await ethers.getContractFactory('ComplianceProxy')
        contract = await Counter.deploy()
        await contract.deployed()
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

        const isVerifiedResponse = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerified", [signer.address])
        );
        const result = contract.interface.decodeFunctionResult("isUserVerified", isVerifiedResponse)
        expect(result.isVerified).to.be.true
    })

    it('Should be able to check for specific issuer of verification', async () => {
        const [signer] = await ethers.getSigners()

        const allowedIssuers = [contract.address]
        const isVerifiedResponse = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerifiedBy", [signer.address, allowedIssuers])
        );
        const result = contract.interface.decodeFunctionResult("isUserVerifiedBy", isVerifiedResponse)
        expect(result.isVerified).to.be.true
    })

    it('Should be able to get verification data', async () => {
        const [signer] = await ethers.getSigners()

        const verificationData = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("getVerificationData", [signer.address, contract.address])
        );
        const result = contract.interface.decodeFunctionResult("getVerificationData", verificationData)
        expect(result.issuerAddress).to.be.equal(contract.address);
    })
})