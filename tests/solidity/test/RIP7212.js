const { expect } = require("chai");
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('RIP7212', () => {
    let contract

    before(async () => {
        const RIP7212 = await ethers.getContractFactory('RIP7212')
        contract = await RIP7212.deploy()
        await contract.deployed()
    })

    it('Should be able to execute RIP7212 precompile', async () => {
        const [signer] = await ethers.getSigners()

        const isAvailableResponse = await sendShieldedQuery(
            signer.provider,
            contract.address,
            contract.interface.encodeFunctionData("isPreCompiledP256Available", [])
        );
        const result = contract.interface.decodeFunctionResult("isPreCompiledP256Available", isAvailableResponse)[0]
        expect(result).to.be.true
    })
})