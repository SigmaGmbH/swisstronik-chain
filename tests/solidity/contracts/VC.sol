pragma solidity ^0.8;

contract VC {
    mapping (address => bool) private _isAuthorized;

    function authorize(bytes calldata credential) public {
        (bool passed, bytes memory data) = address(1027).staticcall(credential);
        require(passed, "Cannot verify credential");

        address credentialSubject = bytesToAddress(data);
        _isAuthorized[credentialSubject] = true;
    }

    function isAuthorized(address user) public view returns (bool) {
        return _isAuthorized[user];
    }

    function bytesToAddress(bytes memory b) private pure returns (address) {
        return address(uint160(bytes20(b)));
    }  
}