package keys

import (
	"bytes"
	cryptorand "crypto/rand"
	"io"
	"math/big"

	"github.com/NethermindEth/juno/pkg/crypto/pedersen"
	starksig "github.com/NethermindEth/juno/pkg/crypto/signature"
	"github.com/NethermindEth/juno/pkg/crypto/weierstrass"
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/medianreport"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/chains/evmutil"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ ocrtypes.OnchainKeyring = &OCR2Key{}

type OCR2Key struct {
	privateKey starksig.PrivateKey
}

func NewOCR2Key(material io.Reader) (*OCR2Key, error) {
	privKey, err := starksig.GenerateKey(curve, material)
	if err != nil {
		return nil, err
	}
	return &OCR2Key{privateKey: *privKey}, err
}

func (sk *OCR2Key) PublicKey() ocrtypes.OnchainPublicKey {
	return PubKeyToStarkKey(sk.privateKey.PublicKey)
}

func (sk *OCR2Key) reportToSigData(reportCtx ocrtypes.ReportContext, report ocrtypes.Report) ([]byte, error) {
	var dataArray []*big.Int
	rawReportContext := evmutil.RawReportContext(reportCtx)
	dataArray = append(dataArray, new(big.Int).SetBytes(rawReportContext[0][:]))
	dataArray = append(dataArray, new(big.Int).SetBytes(rawReportContext[1][:]))
	dataArray = append(dataArray, new(big.Int).SetBytes(starknet.EnsureFelt(rawReportContext[2]))) // convert 32 byte extraHash to 31 bytes

	// split report into seperate felts for hashing
	splitReport, err := medianreport.SplitReport(report)
	if err != nil {
		return []byte{}, err
	}
	for i := 0; i < len(splitReport); i++ {
		dataArray = append(dataArray, new(big.Int).SetBytes(splitReport[i]))
	}

	hash := pedersen.ArrayDigest(dataArray...)
	return hash.Bytes(), nil
}

func (sk *OCR2Key) Sign(reportCtx ocrtypes.ReportContext, report ocrtypes.Report) ([]byte, error) {
	hash, err := sk.reportToSigData(reportCtx, report)
	if err != nil {
		return []byte{}, err
	}

	r, s, err := starksig.Sign(cryptorand.Reader, &sk.privateKey, hash)
	if err != nil {
		return []byte{}, err
	}

	// encoding: public key (32 bytes) + r (32 bytes) + s (32 bytes)
	buff := bytes.NewBuffer([]byte(sk.PublicKey()))
	if _, err := buff.Write(starknet.PadBytesBigInt(r, byteLen)); err != nil {
		return []byte{}, err
	}
	if _, err := buff.Write(starknet.PadBytesBigInt(s, byteLen)); err != nil {
		return []byte{}, err
	}

	out := buff.Bytes()
	if len(out) != sk.MaxSignatureLength() {
		return []byte{}, errors.Errorf("unexpected signature size, got %d want %d", len(out), sk.MaxSignatureLength())
	}
	return out, nil
}

func (sk *OCR2Key) Verify(publicKey ocrtypes.OnchainPublicKey, reportCtx ocrtypes.ReportContext, report ocrtypes.Report, signature []byte) bool {
	// check valid signature length
	if len(signature) != sk.MaxSignatureLength() {
		return false
	}

	// convert OnchainPublicKey (starkkey) into ecdsa public keys (prepend 2 or 3 to indicate +/- Y coord)
	var keys [2]starksig.PublicKey
	prefix := []byte{2, 3}
	for i := 0; i < len(prefix); i++ {
		keys[i] = starksig.PublicKey{Curve: curve}

		// prepend sign byte
		compressedKey := append([]byte{prefix[i]}, publicKey...)
		keys[i].X, keys[i].Y = weierstrass.UnmarshalCompressed(curve, compressedKey)

		// handle invalid publicKey
		if keys[i].X == nil || keys[i].Y == nil {
			return false
		}
	}

	hash, err := sk.reportToSigData(reportCtx, report)
	if err != nil {
		return false
	}

	r := new(big.Int).SetBytes(signature[32:64])
	s := new(big.Int).SetBytes(signature[64:])

	return starksig.Verify(&keys[0], hash, r, s) || starksig.Verify(&keys[1], hash, r, s)
}

func (sk *OCR2Key) MaxSignatureLength() int {
	return 32 + 32 + 32 // publickey + r + s
}

func (sk *OCR2Key) Marshal() ([]byte, error) {
	// https://github.com/ethereum/go-ethereum/blob/07508ac0e9695df347b9dd00d418c25151fbb213/crypto/crypto.go#L159
	return starknet.PadBytesBigInt(sk.privateKey.D, sk.privateKeyLen()), nil
}

func (sk *OCR2Key) privateKeyLen() int {
	// https://github.com/NethermindEth/juno/blob/3e71279632d82689e5af03e26693ca5c58a2376e/pkg/crypto/weierstrass/weierstrass.go#L377
	N := curve.Params().N
	bitSize := N.BitLen()
	return (bitSize + 7) / 8 // 32
}

func (sk *OCR2Key) Unmarshal(in []byte) error {
	// enforce byte length
	if len(in) != sk.privateKeyLen() {
		return errors.Errorf("unexpected seed size, got %d want %d", len(in), sk.privateKeyLen())
	}

	sk.privateKey.D = new(big.Int).SetBytes(in)
	sk.privateKey.PublicKey.Curve = curve
	sk.privateKey.PublicKey.X, sk.privateKey.PublicKey.Y = curve.ScalarBaseMult(in)
	return nil
}
