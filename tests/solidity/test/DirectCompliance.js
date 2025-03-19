const { expect } = require('chai')
const { ethers } = require('hardhat')

const compliancePrecompile = "0x0000000000000000000000000000000000000404";

describe('ComplianceBridge', () => {
    it('Should be able to add verification details', async () => {
        const provider = new ethers.providers.JsonRpcProvider("http://localhost:8547")
        const signer = new ethers.Wallet("D5DA6D43250C8EB630C1AB8A80F19C673267A6B210C10C41065D5C34FC369DCB", provider);

        const contract = await ethers.getContractAt("IComplianceBridge", compliancePrecompile)

        const user = ethers.Wallet.createRandom();
        const userAddress = user.address;
        const originChain = "Ethereum"; 
        const verificationType = 9; // BIOMETRIC type
        const issuanceTimestamp = Math.floor(Date.now() / 1000); 
        const expirationTimestamp = issuanceTimestamp + 31536000; 
        const proofData = ethers.utils.hexlify(ethers.utils.toUtf8Bytes("Some proof data"));
        const schema = "ExampleSchema"; 
        const issuerVerificationId = "Issuer123"; 
        const version = 1;

        const encodedData = contract.interface.encodeFunctionData("addVerificationDetails", [
            userAddress,
            originChain,
            verificationType,
            issuanceTimestamp,
            expirationTimestamp,
            proofData,
            schema,
            issuerVerificationId,
            version
        ]);

        const tx = await signer.sendTransaction({
            to: compliancePrecompile,
            data: encodedData,
            gasLimit: 1_000_000,
        })
        const res = await tx.wait();
        expect(res.status).to.be.equal(1)
    })
})
