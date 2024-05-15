pragma solidity ^0.8;

interface IComplianceBridge {
    struct VerificationData {
        // Verification issuer address
        address issuerAddress;
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
        bytes memory proofData,
        string memory schema,
        string memory issuerVerificationId,
        uint32 version
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
        // Use empty payload data for proof, schema, issuer's verification id and version for testing
        bytes memory proofData = new bytes(1);
        string memory schema = "schema";
        string memory issuerVerificationId = "issuerVerificationId";
        uint32 version;
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.addVerificationDetails,
            (
                userAddress, // user address
                VERIFICATION_TYPE, // verification type
                uint32(block.timestamp % 2 ** 32), // issuance timestamp
                0, // expiration timestamp
                proofData, // proof data
                schema, // schema
                issuerVerificationId, // issuer verification id
                version // version
            )
        );

        (bool success, bytes memory data) = address(1028).call(payload);
        emit VerificationResponse(success, data);
    }

    function isUserVerified(
        address userAddress
    ) public view returns (bool) {
        address[] memory allowedIssuers;
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.hasVerification,
            (userAddress, VERIFICATION_TYPE, 0, allowedIssuers)
        );
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        if (success) {
            return abi.decode(data, (bool));
        } else {
            return false;
        }
    }

    function isUserVerifiedBy(
        address userAddress,
        address[] memory allowedIssuers
    ) public view returns (bool) {
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.hasVerification,
            (userAddress, VERIFICATION_TYPE, 0, allowedIssuers)
        );
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        if (success) {
            return abi.decode(data, (bool));
        } else {
            return false;
        }
    }

    function getVerificationData(
        address userAddress
    )
        public
        view
        returns (IComplianceBridge.VerificationData[] memory verificationData)
    {
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.getVerificationData,
            (userAddress, address(this))
        );
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        if (success) {
            // Decode the bytes data into an array of structs
            verificationData = abi.decode(
                data,
                (IComplianceBridge.VerificationData[])
            );
        }
    }
}
