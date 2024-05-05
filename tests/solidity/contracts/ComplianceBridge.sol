pragma solidity ^0.8;

interface IComplianceBridge {
    struct VerificationData {
        // Verification issuer address
        string issuerAddress;
        // From which chain proof was transferred
        string originChain;
        // Original issuance timestamp
        uint32 issuanceTimestamp;
        // Original expiration timestamp
        uint32 expirationTimestamp;
        // Original proof data (ZK-proof)
        bytes originalData;
        // ZK-proof original schema
        string schema;
        // Verification id for checking(KYC/KYB/AML etc) from issuer side
        string issuerVerificationId;
        // Version
        uint32 version;
    }

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

    function getVerificationData(
        address userAddress,
        address issuerAddress
    ) external returns (bytes memory);
}

contract ComplianceProxy {
    event VerificationResponse(bool success, bytes data);
    event HasVerificationResponse(bool success, bytes data);
    event GetVerificationDataResponse(bool success, bytes data);

    uint32 public constant VERIFICATION_TYPE = 2;

    function markUserAsVerified(address userAddress) public {
        bytes memory proofData = new bytes(1);
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.addVerificationDetails,
            (
                userAddress,
                VERIFICATION_TYPE,
                uint32(block.timestamp % 2 ** 32),
                0,
                proofData
            )
        );

        (bool success, bytes memory data) = address(1028).call(payload);
        emit VerificationResponse(success, data);
    }

    function isUserVerified(
        address userAddress
    ) public view returns (bool isVerified) {
        address[] memory allowedIssuers;
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.hasVerification,
            (userAddress, VERIFICATION_TYPE, 0, allowedIssuers)
        );
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        if (success) {
            isVerified = abi.decode(data, (bool));
        } else {
            return false;
        }
    }

    function isUserVerifiedBy(
        address userAddress,
        address[] memory allowedIssuers
    ) public view returns (bool isVerified) {
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.hasVerification,
            (userAddress, VERIFICATION_TYPE, 0, allowedIssuers)
        );
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        if (success) {
            isVerified = abi.decode(data, (bool));
        } else {
            return false;
        }
    }

    function getVerificationData(
        address userAddress,
        address issuerAddress
    )
        public
        view
        returns (IComplianceBridge.VerificationData memory verificationData)
    {
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.getVerificationData,
            (userAddress, issuerAddress)
        );
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        if (success) {
            (
                string memory issuerAddress2,
                string memory originChain,
                uint32 issuanceTimestamp,
                uint32 expirationTimestamp,
                bytes memory originalData,
                string memory schema,
                string memory issuerVerificationId,
                uint32 version
            ) = abi.decode(
                    data,
                    (
                        string,
                        string,
                        uint32,
                        uint32,
                        bytes,
                        string,
                        string,
                        uint32
                    )
                );
            verificationData = IComplianceBridge.VerificationData(
                issuerAddress2,
                originChain,
                issuanceTimestamp,
                expirationTimestamp,
                originalData,
                schema,
                issuerVerificationId,
                version
            );
        }
    }
}
