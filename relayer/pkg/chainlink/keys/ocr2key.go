package keys

import (
	"bytes"
	"io"
	"math/big"

	"github.com/NethermindEth/juno/pkg/crypto/pedersen"
	"github.com/dontpanicdao/caigo"
	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/ocr2/medianreport"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
	"github.com/smartcontractkit/libocr/offchainreporting2/chains/evmutil"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ ocrtypes.OnchainKeyring = &OCR2Key{}

type OCR2Key struct {
	Key
}

func NewOCR2Key(material io.Reader) (*OCR2Key, error) {
	k, err := GenerateKey(material)

	return &OCR2Key{k}, err
}

func (sk *OCR2Key) PublicKey() ocrtypes.OnchainPublicKey {
	return PubKeyToStarkKey(sk.pub)
}

func ReportToSigData(reportCtx ocrtypes.ReportContext, report ocrtypes.Report) (*big.Int, error) {
	var dataArray []*big.Int
	rawReportContext := evmutil.RawReportContext(reportCtx)
	dataArray = append(dataArray, new(big.Int).SetBytes(rawReportContext[0][:]))
	dataArray = append(dataArray, new(big.Int).SetBytes(rawReportContext[1][:]))
	dataArray = append(dataArray, new(big.Int).SetBytes(starknet.EnsureFelt(rawReportContext[2]))) // convert 32 byte extraHash to 31 bytes

	// split report into seperate felts for hashing
	splitReport, err := medianreport.SplitReport(report)
	if err != nil {
		return &big.Int{}, err
	}
	for i := 0; i < len(splitReport); i++ {
		dataArray = append(dataArray, new(big.Int).SetBytes(splitReport[i]))
	}

	hash := pedersen.ArrayDigest(dataArray...)
	return hash, nil
}

func (sk *OCR2Key) Sign(reportCtx ocrtypes.ReportContext, report ocrtypes.Report) ([]byte, error) {
	hash, err := ReportToSigData(reportCtx, report)
	if err != nil {
		return []byte{}, err
	}

	r, s, err := caigo.Curve.Sign(hash, sk.priv)
	if err != nil {
		return []byte{}, err
	}

	// encoding: public key (32 bytes) + r (32 bytes) + s (32 bytes)
	buff := bytes.NewBuffer([]byte(sk.PublicKey()))
	if _, err := buff.Write(starknet.PadBytes(r.Bytes(), byteLen)); err != nil {
		return []byte{}, err
	}
	if _, err := buff.Write(starknet.PadBytes(s.Bytes(), byteLen)); err != nil {
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
	var keys [2]PublicKey
	keys[0].X = new(big.Int).SetBytes(publicKey)
	keys[0].Y = caigo.Curve.GetYCoordinate(keys[0].X)
	keys[1].X = keys[0].X
	keys[1].Y = new(big.Int).Mul(keys[0].Y, big.NewInt(-1))

	hash, err := ReportToSigData(reportCtx, report)
	if err != nil {
		return false
	}

	r := new(big.Int).SetBytes(signature[32:64])
	s := new(big.Int).SetBytes(signature[64:])

	return caigo.Curve.Verify(hash, r, s, keys[0].X, keys[0].Y) || caigo.Curve.Verify(hash, r, s, keys[1].X, keys[1].Y)
}

func (sk *OCR2Key) MaxSignatureLength() int {
	return 32 + 32 + 32 // publickey + r + s
}

func (sk *OCR2Key) Marshal() ([]byte, error) {
	return starknet.PadBytes(sk.priv.Bytes(), sk.privateKeyLen()), nil
}

func (sk *OCR2Key) privateKeyLen() int {
	// https://github.com/NethermindEth/juno/blob/3e71279632d82689e5af03e26693ca5c58a2376e/pkg/crypto/weierstrass/weierstrass.go#L377
	return 32
}

func (sk *OCR2Key) Unmarshal(in []byte) error {
	// enforce byte length
	if len(in) != sk.privateKeyLen() {
		return errors.Errorf("unexpected seed size, got %d want %d", len(in), sk.privateKeyLen())
	}

	sk.Key = Raw(in).Key()
	return nil
}
