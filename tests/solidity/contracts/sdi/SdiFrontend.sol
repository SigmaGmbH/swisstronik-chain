pragma solidity >=0.7.0 <0.9.0;

import "./SdiVerifier.sol";
import "../ComplianceBridge.sol";

error PrecompileError(bytes _data);

contract SdiFrontend {
    event Verified(address _user);

    address issuer = 0x2Fc0B35E41a9a2eA248a275269Af1c8B3a061167;

    PlonkVerifier public verifier;

    constructor (PlonkVerifier _verifier) {
        verifier = _verifier;
    }

    function verify(bytes memory proofData) public {
        uint256 issuanceRoot = getIssuanceRoot();
        uint256 revocationRoot = getRevocationRoot();

        emit Verified(msg.sender);
    }

    function getVerificationData(
        address userAddress
    ) public view returns (IComplianceBridge.VerificationData[] memory) {
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.getVerificationData,
            (userAddress, issuer)
        );
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        IComplianceBridge.VerificationData[] memory verificationData;
        if (success) {
            // Decode the bytes data into an array of structs
            verificationData = abi.decode(
                data,
                (IComplianceBridge.VerificationData[])
            );
        }
        return verificationData;
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