const { expect } = require("chai")
const { sendShieldedQuery } = require("./testUtils")

const getTokenBalance = async (signer, contract) => {
    const messageHash = ethers.utils.solidityKeccak256(["address"], [signer.address])
    const messageHashBinary = ethers.utils.arrayify(messageHash)
    const signature = await signer.signMessage(messageHashBinary)

    const balanceResponse = await sendShieldedQuery(
        signer.provider,
        contract.address,
        contract.interface.encodeFunctionData("balanceOfWithSignature", [signer.address, signature])
    );
    return contract.interface.decodeFunctionResult("balanceOfWithSignature", balanceResponse)[0]
}

describe('Wrapped SWTR', () => {
    let tokenContract

    before(async () => {
        const ERC20 = await ethers.getContractFactory('PrivateSWTR')
        tokenContract = await ERC20.deploy()
        await tokenContract.deployed()
    })

    it('Send SWTR to get PSWTR', async () => {
        const [signer] = await ethers.getSigners()
        const value = 100

        const balanceBefore = await getTokenBalance(signer, tokenContract)

        // Send 100 uswtr to convert them to PSWTR
        const tx = await signer.sendTransaction({
            from: signer.address,
            to: tokenContract.address,
            value
        })
        await tx.wait()

        // Check if balance was updated
        const balanceAfter = await getTokenBalance(signer, tokenContract)
        expect(balanceAfter).to.be.equal(balanceBefore + value)
    })
})