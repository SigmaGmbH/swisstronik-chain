pragma solidity >=0.7.0 <0.9.0;

import "./SdiVerifier.sol";
import "../ComplianceBridge.sol";

error PrecompileError(bytes _data);

contract SdiFrontend {
    event Verified(address _user);

    PlonkVerifier public verifier;

    constructor (PlonkVerifier _verifier) {
        verifier = _verifier;
    }

    function verify(bytes memory proofData) public {
        uint256 issuanceRoot = getIssuanceRoot();
        uint256 revocationRoot = getRevocationRoot();

        emit Verified(msg.sender);
    }

    function getIssuanceRoot() internal view returns (uint256 _root) {
        bytes memory payload = abi.encodeCall(IComplianceBridge.getIssuanceTreeRoot, ());
        (bool success, bytes memory data) = address(1028).staticcall(payload);

        if (!success) {
            revert PrecompileError({_data: data});
        }

        (_root) = abi.decode(data, (uint256));
    }

    function getRevocationRoot() internal view returns (uint256 _root) {
        bytes memory payload = abi.encodeCall(IComplianceBridge.getRevocationTreeRoot, ());
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        if (!success) {
            revert PrecompileError({_data: data});
        }
        
        (_root) = abi.decode(data, (uint256));
    }
}