pragma solidity ^0.8;

interface IComplianceBridge {
    struct VerificationData {
        // Verification type
        uint32 verificationType;
        // Verification Id
        bytes verificationId;
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
        string memory originChain,
        uint32 verificationType,
        uint32 issuanceTimestamp,
        uint32 expirationTimestamp,
        bytes memory proofData,
        string memory schema,
        string memory issuerVerificationId,
        uint32 version
    ) external returns (bytes memory);

    function addVerificationDetailsV2(
        address userAddress,
        string memory originChain,
        uint32 verificationType,
        uint32 issuanceTimestamp,
        uint32 expirationTimestamp,
        bytes memory proofData,
        string memory schema,
        string memory issuerVerificationId,
        uint32 version,
        bytes32 publicKey
    ) external returns (bytes memory);

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

    function getRevocationTreeRoot() external returns (bytes memory);

    function getIssuanceTreeRoot() external returns (bytes memory);

    function revokeVerification(bytes memory verificationId) external;
}

contract ComplianceProxy {
    event VerificationResponse(bool success, bytes data);
    event HasVerificationResponse(bool success, bytes data);
    event GetVerificationDataResponse(bool success, bytes data);
    event RevocationResponse(bool success, bytes data);

    uint32 public constant VERIFICATION_TYPE = 2;

    function markUserAsVerified(
        address userAddress
    ) public returns (bytes memory) {
        // Use empty payload data for proof, schema, issuer's verification id and version for testing
        bytes memory proofData = new bytes(1);
        string memory originChain = "chain_1291-1";
        string memory schema = "schema";
        string memory issuerVerificationId = "issuerVerificationId";
        uint32 version;
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.addVerificationDetails,
            (
                userAddress, // user address
                originChain,
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
        bytes memory verificationId = abi.decode(data, (bytes));
        emit VerificationResponse(success, verificationId);
        return verificationId;
    }

    function markUserAsVerifiedV2(
        address userAddress,
        bytes32 userPublicKey
    ) public returns (bytes memory) {
        // Use empty payload data for proof, schema, issuer's verification id and version for testing
        bytes memory proofData = new bytes(1);
        string memory issuerVerificationId = "issuerVerificationId";
        uint32 version;
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.addVerificationDetailsV2,
            (
                userAddress, // user address
                "chain_1291-1",
                VERIFICATION_TYPE, // verification type
                uint32(block.timestamp % 2 ** 32), // issuance timestamp
                0, // expiration timestamp
                proofData, // proof data
                "schema", // schema
                issuerVerificationId, // issuer verification id
                version, // version
                userPublicKey // user BJJ public key
            )
        );

        (bool success, bytes memory data) = address(1028).call(payload);
        bytes memory verificationId = abi.decode(data, (bytes));
        emit VerificationResponse(success, verificationId);
        return verificationId;
    }

    function isUserVerified(address userAddress) public view returns (bool) {
        address[] memory allowedIssuers;
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.hasVerification,
            (
                userAddress, // user address
                VERIFICATION_TYPE, // verification type
                0, // expiration timestamp, 0 for infinite period
                allowedIssuers // expected allowed issuers
            )
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
    ) public view returns (IComplianceBridge.VerificationData[] memory) {
        bytes memory payload = abi.encodeCall(
            IComplianceBridge.getVerificationData,
            (userAddress, address(this))
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

    function getIssuanceRoot() public view returns (bytes memory) {
        bytes memory payload = abi.encodeCall(IComplianceBridge.getIssuanceTreeRoot, ());
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        return data;
    }

    function getRevocationRoot() public view returns (bytes memory) {
        bytes memory payload = abi.encodeCall(IComplianceBridge.getRevocationTreeRoot, ());
        (bool success, bytes memory data) = address(1028).staticcall(payload);
        return data;
    }

    function revokeVerification(bytes memory verificationId) public{
        bytes memory payload = abi.encodeCall(IComplianceBridge.revokeVerification,(verificationId));
        (bool success, bytes memory data) = address(1028).call(payload);

        emit RevocationResponse(success, data);
    }
}
