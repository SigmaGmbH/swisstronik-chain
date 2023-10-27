// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8;

import "solidity-rlp/contracts/RLPReader.sol";

contract VC {
    using RLPReader for RLPReader.RLPItem;
    using RLPReader for RLPReader.Iterator;
    using RLPReader for bytes;
    
    struct IssuerAuthorized {
        string  issuer;
        bool    isAuthorized;
    }

    mapping (address => IssuerAuthorized) private _isAuthorized;

    function authorize(bytes calldata credential) public {
        (bool passed, bytes memory rlpBytes) = address(1027).staticcall(credential);
        require(passed, "Cannot verify credential");
        RLPReader.RLPItem[] memory ls = rlpBytes.toRlpItem().toList();
        require(ls.length == 2, "Invalid credential");

        RLPReader.RLPItem memory credentialSubjectItem = ls[0];
        address credentialSubject = bytesToAddress(credentialSubjectItem.toBytes());
        RLPReader.RLPItem memory issuerItem = ls[1];
        string memory issuer = string(issuerItem.toBytes());

        IssuerAuthorized memory authorized = IssuerAuthorized(issuer, true);
        _isAuthorized[credentialSubject] = authorized;
    }

    function isAuthorized(address user) public view returns (bool) {
        return _isAuthorized[user].isAuthorized;
    }

    function getIssuer(address user) public view returns (string memory) {
        return _isAuthorized[user].issuer;
    }

    function bytesToAddress(bytes memory b) private pure returns (address) {
        return address(uint160(bytes20(b)));
    }  
}