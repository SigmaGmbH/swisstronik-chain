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
        await tx.wait()
    })
})