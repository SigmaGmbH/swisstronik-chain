const { expect } = require("chai")

const getTokenBalance = async (contract, address) => {
    const balanceResponse = await sendShieldedQuery(
        ethers.provider,
        contract.address,
        contract.interface.encodeFunctionData("balanceOf", [address])
    );
    return contract.interface.decodeFunctionResult("balanceOf", balanceResponse)[0]
}

describe('VC', () => {
    let tokenContract

    before(async () => {
        const ERC20 = await ethers.getContractFactory('VC')
        tokenContract = await ERC20.deploy()
        await tokenContract.deployed()
    })

    it('Should return correct address and issuer', async () => {
        const [sender] = await ethers.getSigners()
        const jwtCrential = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ2YyI6eyJAY29udGV4dCI6WyJodHRwczovL3d3dy53My5vcmcvMjAxOC9jcmVkZW50aWFscy92MSJdLCJ0eXBlIjpbIlZlcmlmaWFibGVDcmVkZW50aWFsIl0sImNyZWRlbnRpYWxTdWJqZWN0Ijp7Imt5Y1Bhc3NlZCI6dHJ1ZSwiZGF0ZU9mQmlydGgiOiIwMTAxMTk5MCIsImFkZHJlc3MiOiIweDk1RjY4NjhBNzJBZDNFODJhMzBDRWYwYWNBOWFiREFDYTE5MzRBOTcifX0sInN1YiI6ImRpZDpzd3RyOjJIUzNoa1VKNjQ5WUxRSmZkQUwya3UiLCJuYmYiOjE2OTUzOTMzNDgsImlzcyI6ImRpZDpzd3RyOjJIUzNoa1VKNjQ5WUxRSmZkQUwya3UifQ.IIUxG5-xbZWJlSSbAQHe-LD9v5VqDU2-eR3bkFTjFbI"

        await tokenContract.connect(sender).authorize(jwtCrential)

        const address = "0x95F6868A72Ad3E82a30CEf0acA9abDACa1934A97"
        const authorized = await tokenContract.isAuthorized(address)
        expect(authorized).to.be.equal(true)
        const issuer = await tokenContract.getIssuer(address)
        expect(authorized).to.be.equal("did:swtr:2HS3hkUJ649YLQJfdAL2ku")

    })
})