const { expect } = require("chai");
const { ethers } = require("hardhat")
const { sendShieldedQuery } = require("./testUtils")

describe('ED25519VerifyPrecompile', () => {
    let contract

    before(async () => {
        const Ed25519Verify = await ethers.getContractFactory('ED25519VerifyPrecompile')
        contract = await Ed25519Verify.deploy()
        await contract.deployed()
    })

    it('Should be able to execute ED25519Verify precompile', async () => {
        const [signer] = await ethers.getSigners()

        const isAvailableResponse = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("checkPrecompile", [])
        );
        const result = contract.interface.decodeFunctionResult("checkPrecompile", isAvailableResponse)[0]
        expect(result).to.be.true
    })
})