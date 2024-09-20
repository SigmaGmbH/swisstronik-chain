const { expect } = require("chai");
const { ethers } = require("hardhat");

const unencryptedProvider = new ethers.providers.JsonRpcProvider('http://localhost:8547')
const unencryptedSigner = new ethers.Wallet("DBE7E6AE8303E055B68CEFBF01DEC07E76957FF605E5333FA21B6A8022EA7B55", unencryptedProvider)

describe("MsgInfo", function () {
    describe('Unencrypted', () => {
        let msgInfoContract

        before(async () => {
            const factory = await ethers.getContractFactory('MsgInfo')
            msgInfoContract = await factory.connect(unencryptedSigner).deploy()
            await msgInfoContract.connect(unencryptedProvider).deployed()
        })

        it('Should be initialized properly',  async() => {
            const storedValue = await msgInfoContract.connect(unencryptedProvider).storedValue()
            expect(storedValue).to.be.equal(0)
        })

        it("Should update storedValue with msg.value", async () => {
            const sendValue = ethers.utils.parseEther("0.1")

            const tx = await msgInfoContract.connect(unencryptedSigner).updateValue({ value: sendValue })
            await tx.wait()

            expect(await msgInfoContract.connect(unencryptedProvider).storedValue()).to.equal(sendValue)
        })

        it('Should return correct msg.sig', async () => {
            const functionSignature = ethers.utils.id("getMsgSig()").slice(0, 10)
            const msgSig = await msgInfoContract.connect(unencryptedProvider).getMsgSig()
            expect(msgSig).to.equal(functionSignature)
        })

        it('Should return correct msg.data', async () => {
            const param = Math.floor(Math.random() * 10000)
            const encodedData = msgInfoContract.interface.encodeFunctionData("getMsgData(uint256)", [param])
            const msgData = await msgInfoContract.connect(unencryptedProvider).getMsgData(param)
            expect(msgData).to.equal(encodedData)
        })
    })
});