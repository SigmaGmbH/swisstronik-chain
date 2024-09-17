const { expect } = require("chai");
const { ethers } = require("hardhat")
const { sendShieldedTransaction, sendShieldedQuery } = require("./testUtils")

describe('Counter', () => {
    let counterContract

    before(async () => {
        const Counter = await ethers.getContractFactory('Counter')
        counterContract = await Counter.deploy({gasLimit: 1_000_000})
        await counterContract.deployed()

        console.log('Counter deployed to: ', counterContract.address)
    })

    // it('DEBUG', async () => {
    //     const [signer] = await ethers.getSigners()
    //     const nonce = await signer.provider.getTransactionCount(signer.address)
    //
    //     console.log('Caller: ', signer.address)
    //     console.log('Caller nonce: ', nonce)
    //     const calculatedAddr = ethers.utils.getContractAddress({from: signer.address, nonce: nonce})
    //     console.log('Exp address: ', calculatedAddr) // 0xa079a9f7d09fb132e95aa627c2ff414a43787c69
    //
    //     // 1 - 0x56f078d59950b21251e32fd659c89f80cf669fcc js: 0x30252aFE8C1683fD184C99a3c44Aa5d547d59dd4
    //     // 0 -   js: 0x54014e46667907922c58Ac224dFb0848f43EA800
    //     // 2 -   js: 0x202B7E8CaF217a1bc54cBFcA986d17B268cdd9d7
    //     // 3 -   js: 0x56f078d59950b21251E32Fd659c89f80Cf669fcc
    //
    //     //
    //     const Counter = await ethers.getContractFactory('Counter')
    //     const counterContract = await Counter.deploy({gasLimit: 1_000_000})
    //     await counterContract.deployed()
    //
    //     console.log('Counter deployed to: ', counterContract.address)
    //
    //     const contractCode = await signer.provider.getCode(counterContract.address)
    //     console.log('Contract code: ', contractCode)
    // })



    it('Should add', async () => {
        const [signer] = await ethers.getSigners()

        const countBeforeResponse = await sendShieldedQuery(
            signer.provider,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countBefore = counterContract.interface.decodeFunctionResult("counter", countBeforeResponse)
        console.log('res: ', countBefore)

        const tx = await sendShieldedTransaction(
            signer,
            counterContract.address,
            counterContract.interface.encodeFunctionData("add", [])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => counterContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Added')).to.be.true
        expect(logs.some(log => log.name === 'Changed')).to.be.true

        const countAfterResponse = await sendShieldedQuery(
            signer.provider,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countAfter = counterContract.interface.decodeFunctionResult("counter", countAfterResponse)
        expect(countAfter[0].toNumber()).to.be.equal(countBefore[0].toNumber() + 1)
    })

    it('Should subtract', async () => {
        const [signer] = await ethers.getSigners()

        const countBeforeResponse = await sendShieldedQuery(
            signer.provider,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countBefore = counterContract.interface.decodeFunctionResult("counter", countBeforeResponse)

        const tx = await sendShieldedTransaction(
            signer,
            counterContract.address,
            counterContract.interface.encodeFunctionData("subtract", [])
        )
        const receipt = await tx.wait()
        const logs = receipt.logs.map(log => counterContract.interface.parseLog(log))
        expect(logs.some(log => log.name === 'Changed')).to.be.true

        const countAfterResponse = await sendShieldedQuery(
            signer.provider,
            counterContract.address,
            counterContract.interface.encodeFunctionData("counter", [])
        );
        const countAfter = counterContract.interface.decodeFunctionResult("counter", countAfterResponse)
        expect(countAfter[0].toNumber()).to.be.equal(countBefore[0].toNumber() - 1)
    })

    it('Should revert correctly', async () => {
        const [signer] = await ethers.getSigners()

        let failed = false
        try {
            const tx = await sendShieldedTransaction(
                signer,
                counterContract.address,
                counterContract.interface.encodeFunctionData("subtract", [])
            )
            await tx.wait()
        } catch (e) {
            failed = e.reason.indexOf('reverted') !== -1
        }

        expect(failed).to.be.true
    })
})