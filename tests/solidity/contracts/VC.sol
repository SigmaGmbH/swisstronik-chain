pragma solidity ^0.8;

/**
 Sample contract to test if precompile for verifiable credentials
*/
contract VC {
    /**
        Verifies provided JWT Proof for Verifiable Credential.
        Returns `address` of credential subject and string with DID URL of issuer
     */
    function verifyJWT(bytes calldata credential) public view returns (address, string memory) {
        (bool success, bytes memory data) = address(1027).staticcall(credential);
        require(success, "Cannot verify credential");

        (address subject, string memory issuer) = decodeResult(data);
        return (subject, issuer);
    }

    /**
        Decodes precompile result. Precompile returns credential subject address and DID URL of issuer
     */
    function decodeResult(bytes memory output) public pure returns (address, string memory) {
        (address subject, string memory issuer) = abi.decode(output, (address, string));
        return (subject, issuer);
    }
}