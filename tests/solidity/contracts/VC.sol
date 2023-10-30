pragma solidity ^0.8;

/**
 Sample contract to test if precompile for verifiable credentials
*/
contract VC {
    mapping (address => bool) private _isAuthorized;

    /**
        Verifies provided JWT Proof for Verifiable Credential.
        Returns `address` of credential subject 
     */
    function verifyJWT(bytes calldata credential) public view returns (address) {
        (bool success, bytes memory data) = address(1027).staticcall(credential);
        require(success, "Cannot verify credential");

        return bytesToAddress(data);
    }

    /**
        Converts provided bytes array to address
     */
    function bytesToAddress(bytes memory b) private pure returns (address) {
        return address(uint160(bytes20(b)));
    }  
}