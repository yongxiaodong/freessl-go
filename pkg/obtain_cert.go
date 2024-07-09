package pkg

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"log"
	"os"
	"time"
)

type CertInstance interface {
	NewObtain() bool
	Renew() bool
	Valid() bool
}

type ProviderApiKey struct {
	AccessKey string
	SecretKey string
}

type CertParams struct {
	Email          string
	Registration   *registration.Resource
	Key            crypto.PrivateKey
	Domains        []string
	CertPath       string
	KeyPath        string
	RenewBeforeDay int
}

type CertKeyParams struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// Renew 暂时和新颁发证书是一样的逻辑
func (myCert *CertParams) Renew() {
	// No nothing to do
}

func (myCert *CertParams) Log(msg string, args ...interface{}) {
	mess := fmt.Sprintf(msg, args)
	log.Printf("%s=== %s", myCert.Domains[0], mess)
}

func (myCert *CertParams) Valid() bool {
	base := CertKeyParams{}
	if fileExists(myCert.CertPath) && fileExists(myCert.KeyPath) {
		// 读取证书文件
		certBytes, err := os.ReadFile(myCert.CertPath)
		if err != nil {
			myCert.Log("Failed to read certificate file: %v", err)
			return false
		}

		// 解码 PEM 格式的证书
		block, _ := pem.Decode(certBytes)
		if block == nil || block.Type != "CERTIFICATE" {
			myCert.Log("Failed to decode certificate PEM")
			return false
		}

		// 解析证书
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			myCert.Log("Failed to parse certificate: %v", err)
			return false
		}
		base.PublicKey = cert.PublicKey.(*rsa.PublicKey)

		// 打印证书有效期
		myCert.Log("  - 证书过期时间: %v\n", cert.NotAfter)
		currentTime := time.Now().UTC()
		daysDifference := currentTime.Sub(cert.NotAfter).Hours() / 24

		// 读取私钥文件
		keyBytes, err := os.ReadFile(myCert.KeyPath)
		if err != nil {
			log.Fatalf("Failed to read private key file: %v", err)
		}

		// 解码 PEM 格式的私钥
		block, _ = pem.Decode(keyBytes)
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			myCert.Log("Failed to decode private key PEM")
			return false
		}
		base.PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)

		// 尝试解析私钥（不会验证其有效性，仅检查格式）
		_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			myCert.Log("Failed to parse private key: %v", err)
			return false
		}

		myCert.Log("Private key 有效")
		// 使用公钥加密
		originalData := "data to be encrypted"
		data := []byte(originalData)
		ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, base.PublicKey, data, nil)
		if err != nil {
			myCert.Log("Failed to encrypt data: %v", err)
			return false
		}
		// 使用私钥解密
		plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, base.PrivateKey, ciphertext, nil)
		if err != nil {
			myCert.Log("Failed to decrypt data: %v", err)
			return false
		}
		if data[0] == plaintext[0] {
			myCert.Log("证书加解密成功")
		}

		if daysDifference > -float64(myCert.RenewBeforeDay) {
			myCert.Log("证书需要续签，剩余天数：%.2f ", daysDifference)
			return false
		}
		myCert.Log("证书有效， 剩余天数： %.2f, 证书可用. ", daysDifference)
		return true
	}
	return false
}

type DnsProviderParams struct {
	ProviderApiKey
	Provider challenge.Provider
}

func (myCert *CertParams) GetEmail() string {
	return myCert.Email
}
func (myCert *CertParams) GetRegistration() *registration.Resource {
	return myCert.Registration
}
func (myCert *CertParams) GetPrivateKey() crypto.PrivateKey {
	return myCert.Key
}

func (myCert *CertParams) NewObtain(dnsProviderParams DnsProviderParams) bool {
	log.Printf("开始申请新证书. domain: %+v\n", myCert.Domains)
	config := lego.NewConfig(myCert)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return false
	}

	// 禁cname支持
	_ = os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", "true")

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
		return false
	}
	myCert.Registration = reg

	request := certificate.ObtainRequest{
		Domains: myCert.Domains,
		Bundle:  true,
	}

	err = client.Challenge.SetDNS01Provider(dnsProviderParams.Provider, dns01.AddDNSTimeout(10*time.Minute))
	if err != nil {
		log.Fatal(err)
		return false
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(myCert.CertPath, certificates.Certificate, 0755); err != nil {
		log.Fatal(err)
		return false
	}
	if err := os.WriteFile(myCert.KeyPath, certificates.PrivateKey, 0755); err != nil {
		log.Fatal(err)
		return false
	}
	return true
}
