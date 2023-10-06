pragma solidity ^0.8;

contract VC {
    mapping (address => bool) private _isAuthorized;

    function authorize(bytes calldata credential) public {
        // For now, we ignore response from precompiled contract. 0x42 is address of precompiled contract for verification of credential
        // and it can be changed in future releases
        (bool passed, _) = address(1027).staticcall(credential);
        require(passed, "Cannot verify credential");
        _isAuthorized[msg.sender] = true;
    }

    function isAuthorized(address user) public view returns (bool) {
        return _isAuthorized[user];
    }
}