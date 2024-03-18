pragma solidity ^0.8;

interface IComplianceBridge {
    function addVerificationDetails(
        address userAddress,
        uint32 verificationType,
        uint32 issuanceTimestamp,
        uint32 expirationTimestamp,
        bytes memory proofData
    ) external;

    function hasVerification(
        address userAddress,
        uint32 verificationType,
        uint32 expirationTimestamp,
        address[] memory allowedIssuers
    ) external returns (bool);
}

contract ComplianceProxy {
    event VerificationResponse(bool success, bytes data);
    event HasVerificationResponse(bool success, bytes data);

    uint32 constant public VERIFICATION_TYPE = 2;

    function markUserAsVerified(address userAddress) public {
        bytes memory proofData = new bytes(1);
        bytes memory payload = abi.encodeCall(IComplianceBridge.addVerificationDetails, (
            userAddress,
            VERIFICATION_TYPE,
            uint32(block.timestamp % 2**32),
            0,
            proofData
        ));

        (bool success, bytes memory data) = address(1028).call(payload);
        emit VerificationResponse(success, data);
    }

    function isUserVerified(address userAddress) public view returns (bool isVerified) {
        address[] memory allowedIssuers;
        bytes memory payload = abi.encodeCall(IComplianceBridge.hasVerification, (
            userAddress,
            VERIFICATION_TYPE,
            0,
            allowedIssuers
        ));
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        if (success) {
            isVerified = abi.decode(data, (bool));
        } else {
            return false;
        }
    }

    function isUserVerifiedBy(address userAddress, address[] memory allowedIssuers) public view returns (bool isVerified) {
        bytes memory payload = abi.encodeCall(IComplianceBridge.hasVerification, (
            userAddress,
            VERIFICATION_TYPE,
            0,
            allowedIssuers
        ));
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        if (success) {
            isVerified = abi.decode(data, (bool));
        } else {
            return false;
        }
    }
}
