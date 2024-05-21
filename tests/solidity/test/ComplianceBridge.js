const { expect } = require("chai");
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('ComplianceBridge', () => {
    let contract

    before(async () => {
        const [signer] = await ethers.getSigners()
        const ComplianceProxyFactory = await ethers.getContractFactory('ComplianceProxy')
        contract = new ethers.Contract("0x30252afe8c1683fd184c99a3c44aa5d547d59dd4", ComplianceProxyFactory.interface, signer);
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
        expect(parsedLog.args.data.length).to.be.greaterThan(0);

        const isVerifiedResponse = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isUserVerified", [signer.address])
        )
        const result = contract.interface.decodeFunctionResult("isUserVerified", isVerifiedResponse)
        expect(result[0]).to.be.true
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
        expect(result[0].length).to.be.greaterThan(0);
        for (const details of result[0]) {
            expect(details.issuerAddress.length).to.be.greaterThan(0)
            expect(details.issuerAddress.toLowerCase()).to.be.equal(contract.address.toLowerCase())
        }
    })
})