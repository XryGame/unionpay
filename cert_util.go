package unionpay

import (
	"io/ioutil"
	"golang.org/x/crypto/pkcs12"
	"crypto/rsa"
	"crypto/rand"
	"crypto"
	"encoding/base64"
	"fmt"
	"crypto/sha256"
	"encoding/pem"
	"errors"
	"crypto/x509"
)

func PfxSign(pfxpath, pfxpassword string, signsrc string) (string, error) {
	var pfxData []byte
	pfxData, err := ioutil.ReadFile(pfxpath)
	if err != nil {
		return "", err
	}

	//解析证书
	priv, _, err := pkcs12.Decode(pfxData, pfxpassword)
	if err != nil {
		return"", err
	}
	private := priv.(*rsa.PrivateKey)

	rng := rand.Reader

	hashed := sha256.Sum256([]byte(fmt.Sprintf("%x", sha256.Sum256([]byte(signsrc)))))

	signer, err := rsa.SignPKCS1v15(rng, private , crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	s := base64.StdEncoding.EncodeToString(signer)
	return s, nil
}


func GetCertId(pfxpath, pfxpassword string) string {
	var pfxData []byte
	pfxData, err := ioutil.ReadFile(pfxpath)
	if err != nil {
		return ""
	}
	//解析证书
	_, cert, err := pkcs12.Decode(pfxData, pfxpassword)
	if err != nil {
		return ""
	}

	return cert.SerialNumber.String()
}



func VerifyPKCS1v15(src, sig, key []byte, hash crypto.Hash) error {

	hashed := sha256.Sum256([]byte(fmt.Sprintf("%x", sha256.Sum256(src))))

	var err error
	var block *pem.Block
	block, _ = pem.Decode(key)
	if block == nil {
		return errors.New("public key error")
	}

	var pubInterface = new(x509.Certificate)
	pubInterface, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	var pub = pubInterface.PublicKey.(*rsa.PublicKey)

	return rsa.VerifyPKCS1v15(pub, hash, hashed[:], sig)
}