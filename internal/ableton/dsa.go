package ableton

import (
	"crypto/dsa"
	cryptorand "crypto/rand"
	"crypto/sha1"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

type DSAParametersASN1 struct {
	P *big.Int
	Q *big.Int
	G *big.Int
}
type AlgorithmIdentifierASN1 struct {
	OID        asn1.ObjectIdentifier
	Parameters DSAParametersASN1
}
type DSAPublicKeyASN1 struct {
	Algorithm        AlgorithmIdentifierASN1
	SubjectPublicKey asn1.BitString
}

func signDSA(key dsa.PrivateKey, message string) (string, error) {
	if key.X.BitLen() > 1024 {
		return "", fmt.Errorf("key size must be 1024 bits")
	}

	hasher := sha1.New()
	hasher.Write([]byte(message))
	hash := hasher.Sum(nil)

	var r *big.Int
	var s *big.Int
	var err error

	r, s, err = dsa.Sign(cryptorand.Reader, &key, hash)
	if err != nil {
		return "", fmt.Errorf("sign %s: %w", s, err)
	}

	rBytes := fmt.Sprintf("%040X", r)
	sBytes := fmt.Sprintf("%040X", s)

	return rBytes + sBytes, nil
}

type DSAPrivateKeyASN1 struct {
	Version int
	P       *big.Int
	Q       *big.Int
	G       *big.Int
	Y       *big.Int
	X       *big.Int
}

func PrivateDSAToHex(key *dsa.PrivateKey) (string, error) {
	k := DSAPrivateKeyASN1{
		Version: 0,
		P:       key.P,
		Q:       key.Q,
		G:       key.G,
		Y:       key.Y,
		X:       key.X,
	}

	keyBytes, err := asn1.Marshal(k)
	if err != nil {
		return "", errors.New("failed to marshal dsa key: " + err.Error())
	}

	hexString := hex.EncodeToString(keyBytes)
	hexString = strings.ToUpper(hexString)

	return hexString, nil
}

func PublicDSAToHex(key *dsa.PublicKey) (string, error) {
	encodedPubKey, err := asn1.Marshal(key.Y)
	if err != nil {
		return "", fmt.Errorf("marshal public key: %w", err)
	}

	asn1Key := DSAPublicKeyASN1{
		Algorithm: AlgorithmIdentifierASN1{
			OID: asn1.ObjectIdentifier{1, 2, 840, 10040, 4, 1},
			Parameters: DSAParametersASN1{
				P: key.P,
				Q: key.Q,
				G: key.G,
			},
		},
		SubjectPublicKey: asn1.BitString{
			Bytes: encodedPubKey,
		},
	}

	derBytes, err := asn1.Marshal(asn1Key)
	if err != nil {
		return "", err
	}

	hexString := hex.EncodeToString(derBytes)
	hexString = strings.ToUpper(hexString)

	return hexString, nil
}

func HexToPrivateDSA(hexString string) (*dsa.PrivateKey, error) {
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, err
	}

	k := DSAPrivateKeyASN1{}
	rest, err := asn1.Unmarshal(bytes, &k)
	if err != nil {
		return nil, fmt.Errorf("unmarshal private key: %w", err)
	}
	if len(rest) > 0 {
		return nil, errors.New("garbage after key")
	}

	return &dsa.PrivateKey{
		PublicKey: dsa.PublicKey{
			Parameters: dsa.Parameters{
				P: k.P,
				Q: k.Q,
				G: k.G,
			},
			Y: k.Y,
		},
		X: k.X,
	}, nil
}
