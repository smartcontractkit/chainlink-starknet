package kms

import (
	"bytes"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// SPKI represents the SubjectPublicKeyInfo structure as defined in [RFC 5280] in ASN.1 format.
//
// The public key that AWS KMS returns is a DER-encoded X.509 public key, also known as
// SubjectPublicKeyInfo (SPKI). This structure is used to unpack the public key returned by the KMS
// GetPublicKey API call.
//
// For more details: see the AWS KMS documentation on [GetPublicKey response syntax].
//
// [RFC 5280]: https://datatracker.ietf.org/doc/html/rfc5280
// [GetPublicKey response syntax]: https://docs.aws.amazon.com/kms/latest/APIReference/API_GetPublicKey.html#API_GetPublicKey_ResponseSyntax
type SPKI struct {
	AlgorithmIdentifier SPKIAlgorithmIdentifier
	SubjectPublicKey    asn1.BitString
}

// SPKIAlgorithmIdentifier represents the AlgorithmIdentifier structure for the
// SubjectPublicKeyInfo (SPKI) in ASN.1 format.
type SPKIAlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.ObjectIdentifier
}

// ECDSASig represents the ECDSA signature structure as defined in [RFC 3279] in ASN.1 format.
// This structure is used to unpack the ECDSA signature returned by AWS KMS when signing data.
//
// [RFC 3279] https://datatracker.ietf.org/doc/html/rfc3279#section-2.2.3
type ECDSASig struct {
	R asn1.RawValue
	S asn1.RawValue
}

var (
	// secp256k1N is the N value of the secp256k1 curve, used to adjust the S value in signatures.
	secp256k1N = crypto.S256().Params().N
	// secp256k1HalfN is half of the secp256k1 N value, used to adjust the S value in signatures.
	secp256k1HalfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

// ASN1ToSEC1PublicKey converts a KMS public key (ASN.1 DER-encoded SPKI format) to SEC1 format
// (uncompressed: 0x04 || X || Y, 65 bytes).
//
// KMS returns public keys in ASN.1 DER-encoded SubjectPublicKeyInfo (SPKI) format as defined in
// RFC 5280. This function extracts the public key and converts it to SEC1 uncompressed format.
//
// This matches the implementation in chainlink-deployments-framework:
// https://github.com/smartcontractkit/chainlink-deployments-framework/blob/main/chain/evm/provider/kms_signer.go#L78
func ASN1ToSEC1PublicKey(asn1PublicKey []byte) ([]byte, error) {
	var spki SPKI
	if _, err := asn1.Unmarshal(asn1PublicKey, &spki); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ASN.1 public key: %w", err)
	}

	// Unmarshal the KMS public key bytes into an ECDSA public key.
	// KMS includes the 0x04 prefix in the BitString (65 bytes: 0x04 || X || Y).
	// If the BitString is 64 bytes (X || Y), we need to prepend 0x04.
	pubKeyBytes := spki.SubjectPublicKey.Bytes
	if len(pubKeyBytes) == 64 {
		// BitString is 64 bytes (X || Y), prepend 0x04
		pubKeyBytes = append([]byte{0x04}, pubKeyBytes...)
	} else if len(pubKeyBytes) != 65 || pubKeyBytes[0] != 0x04 {
		return nil, fmt.Errorf("invalid public key length in BitString: expected 64 or 65 bytes, got %d", len(pubKeyBytes))
	}
	// BitString already has 0x04 prefix (KMS format), use as-is

	pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	// SEC1 uncompressed format: 0x04 || X || Y (65 bytes)
	sec1PubKeyBytes := secp256k1.S256().Marshal(pubKey.X, pubKey.Y)
	return sec1PubKeyBytes, nil
}

// SEC1ToASN1PublicKey converts a SEC1 uncompressed public key (0x04 || X || Y, 65 bytes)
// to ASN.1 DER-encoded SubjectPublicKeyInfo format.
//
// This is the reverse operation of ASN1ToSEC1PublicKey.
func SEC1ToASN1PublicKey(sec1PubKey []byte) ([]byte, error) {
	if len(sec1PubKey) != 65 || sec1PubKey[0] != 0x04 {
		return nil, fmt.Errorf("invalid SEC1 public key format: expected 65 bytes starting with 0x04, got %d bytes", len(sec1PubKey))
	}

	// Unmarshal the SEC1 public key to get the ECDSA public key
	// This ensures proper handling and validates the key
	pubKey, err := crypto.UnmarshalPubkey(sec1PubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SEC1 public key: %w", err)
	}

	x := make([]byte, 32)
	y := make([]byte, 32)
	pubKey.X.FillBytes(x)
	pubKey.Y.FillBytes(y)

	// Create the public key as a BitString
	// KMS includes the 0x04 prefix in the BitString (65 bytes: 0x04 || X || Y)
	// This matches the SEC1 uncompressed format
	pubKeyBytes := slices.Concat([]byte{0x04}, x, y)

	// OID for ecPublicKey: 1.2.840.10045.2.1
	ecPublicKeyOID := asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	// OID for secp256k1: 1.3.132.0.10
	secp256k1OID := asn1.ObjectIdentifier{1, 3, 132, 0, 10}

	// Create SPKI structure
	spki := SPKI{
		AlgorithmIdentifier: SPKIAlgorithmIdentifier{
			Algorithm:  ecPublicKeyOID,
			Parameters: secp256k1OID,
		},
		SubjectPublicKey: asn1.BitString{
			Bytes:     pubKeyBytes,
			BitLength: len(pubKeyBytes) * 8,
		},
	}

	// Marshal to ASN.1 DER
	return asn1.Marshal(spki)
}

// SEC1ToASN1Sig converts a SEC1 signature (R || S || V, 65 bytes) to ASN.1 DER format.
//
// The SEC1 signature format is: [32 bytes R][32 bytes S][1 byte V]
// The ASN.1 format is a SEQUENCE of two INTEGERs: { R, S }
// The recovery ID (V) is not included in ASN.1 format as it's only used for public key recovery.
//
// This is the reverse operation of KMSToSEC1Sig, but note that the recovery ID (V) is lost
// in the conversion since ASN.1 format doesn't include it.
func SEC1ToASN1Sig(sec1Sig []byte) ([]byte, error) {
	if len(sec1Sig) != 65 {
		return nil, fmt.Errorf("invalid SEC1 signature format: expected 65 bytes, got %d", len(sec1Sig))
	}

	// Extract R and S (first 32 bytes are R, next 32 bytes are S)
	rBytes := sec1Sig[0:32]
	sBytes := sec1Sig[32:64]
	// V (recovery ID) is in sec1Sig[64], but we don't need it for ASN.1

	// Convert bytes to big.Int, removing leading zeros
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	// ASN.1 ECDSA signature is a SEQUENCE of two INTEGERs (R and S)
	type ecdsaSignature struct {
		R *big.Int
		S *big.Int
	}

	sig := ecdsaSignature{
		R: r,
		S: s,
	}

	// Marshal to ASN.1 DER
	return asn1.Marshal(sig)
}

// ASN1ToSEC1Sig converts a ASN.1 signature (ASN.1 format) to SEC1 format (R || S || V). This follows this
// example provided by AWS Guides. Notably Ethereum and most blockchain systems use the SEC1 format for signatures.
//
// [AWS Guides]: https://aws.amazon.com/blogs/database/part2-use-aws-kms-to-securely-manage-ethereum-accounts/
func ASN1ToSEC1Sig(asn1Sig, ecdsaPubKeyBytes, hash []byte) ([]byte, error) {
	var ecdsaSig ECDSASig
	if _, err := asn1.Unmarshal(asn1Sig, &ecdsaSig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal KMS signature: %w", err)
	}

	rBytes := ecdsaSig.R.Bytes
	sBytes := ecdsaSig.S.Bytes

	// Adjust S value from signature to match SEC1 standard.
	//
	// After we extract r and s successfully, we have to test if the value of s is greater than
	// secp256k1n/2 as specified in EIP-2 and flip it if required.
	sBigInt := new(big.Int).SetBytes(sBytes)
	if sBigInt.Cmp(secp256k1HalfN) > 0 {
		sBytes = new(big.Int).Sub(secp256k1N, sBigInt).Bytes()
	}

	return recoverSEC1Signature(ecdsaPubKeyBytes, hash, rBytes, sBytes)
}

// recoverSEC1Signature attempts to reconstruct the SEC1 signature by trying both possible recovery
// IDs (v = 0 and v = 1). It compares the recovered public key with the expected public key bytes
// to determine the correct signature.
//
// Returns the valid SEC1 signature if successful, or an error if neither recovery ID matches.
func recoverSEC1Signature(expectedPublicKey, txHash, r, s []byte) ([]byte, error) {
	// SEC1 signatures require r and s to be exactly 32 bytes each.
	rsSig := append(padTo32Bytes(r), padTo32Bytes(s)...)
	// SEC1 signatures have a 65th byte called the recovery ID (v), which can be 0 or 1.
	// Here we append 0 to the signature to start with for the first recovery attempt.
	sec1Sig := append(rsSig, []byte{0}...)

	recoveredPublicKey, err := crypto.Ecrecover(txHash, sec1Sig)
	if err != nil {
		return nil, fmt.Errorf("failed to recover signature with v=0: %w", err)
	}

	if hex.EncodeToString(recoveredPublicKey) != hex.EncodeToString(expectedPublicKey) {
		// If the first recovery attempt failed, we try with v=1.
		sec1Sig = append(rsSig, []byte{1}...)
		recoveredPublicKey, err = crypto.Ecrecover(txHash, sec1Sig)
		if err != nil {
			return nil, fmt.Errorf("failed to recover signature with v=1: %w", err)
		}

		if hex.EncodeToString(recoveredPublicKey) != hex.EncodeToString(expectedPublicKey) {
			return nil, errors.New("cannot reconstruct public key from sig")
		}
	}

	return sec1Sig, nil
}

// padTo32Bytes pads the given byte slice to 32 bytes by trimming leading zeros and prepending
// zeros.
func padTo32Bytes(buffer []byte) []byte {
	buffer = bytes.TrimLeft(buffer, "\x00")
	for len(buffer) < 32 {
		zeroBuf := []byte{0}
		buffer = append(zeroBuf, buffer...)
	}

	return buffer
}
