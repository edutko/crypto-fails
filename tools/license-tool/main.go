package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	appinternal "github.com/edutko/crypto-fails/internal/app"
	badecdsa "github.com/edutko/crypto-fails/internal/crypto/ecdsa"
	"github.com/edutko/crypto-fails/pkg/app"
)

func usage() {
	fmt.Println("usage: license keygen")
	fmt.Println("               create <licensee> <start> <end> [<feature>=<value> ...]")
	fmt.Println("               sign <file> [<private_key>]")
	fmt.Println("               check <file> [<public_key>]")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 5 {
			usage()
		}
		licensee := os.Args[2]
		start, err := time.Parse("2006-01-02", os.Args[3])
		if err != nil {
			log.Fatalf("invalid start date: %v", err)
		}
		end, err := time.Parse("2006-01-02", os.Args[4])
		if err != nil {
			log.Fatalf("invalid end date: %v", err)
		}
		newLicense(licensee, start, end, os.Args[5:])

	case "keygen":
		generateKeypair()

	case "sign":
		if len(os.Args) < 3 || len(os.Args) > 4 {
			usage()
		}
		licenseFile := os.Args[2]
		privateKeyFile := "license-private.pem"
		if len(os.Args) == 4 {
			privateKeyFile = os.Args[3]
		}
		sign(licenseFile, privateKeyFile)

	case "check":
		if len(os.Args) < 3 || len(os.Args) > 4 {
			usage()
		}
		licenseFile := os.Args[2]
		publicKeyFile := "license-public.pem"
		if len(os.Args) == 4 {
			publicKeyFile = os.Args[3]
		}
		checkSignature(licenseFile, publicKeyFile)

	default:
		usage()
	}
}

func newLicense(licensee string, start time.Time, end time.Time, featureList []string) {
	features := make(map[app.Feature]int)
	for _, f := range featureList {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) == 1 {
			features[app.Feature(parts[0])] = 1
		} else {
			if v, _ := strconv.Atoi(parts[1]); v != 0 {
				features[app.Feature(parts[0])] = v
			}
		}
	}

	l := app.NewLicense(licensee, start, end, features)
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		log.Fatalf("error: json.MarshalIndent: %v", err)
	}

	fmt.Println(string(b))
}

func generateKeypair() {
	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("error: ecdsa.GenerateKey: %v", err)
	}

	b, err := x509.MarshalPKCS8PrivateKey(sk)
	if err != nil {
		log.Fatalf("error: x509.MarshalPKCS8PrivateKey: %v", err)
	}

	pb := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	})

	err = os.WriteFile("license-private.pem", pb, 0600)
	if err != nil {
		log.Fatalf("error: os.WriteFile:  %v", err)
	}

	b, err = x509.MarshalPKIXPublicKey(&sk.PublicKey)
	if err != nil {
		log.Fatalf("error: x509.MarshalPKIXPublicKey: %v", err)
	}

	pb = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	})

	err = os.WriteFile("license-public.pem", pb, 0600)
	if err != nil {
		log.Fatalf("error: os.WriteFile: %v", err)
	}
}

func sign(licenseFile, privateKeyFile string) {
	b, err := os.ReadFile(licenseFile)
	if err != nil {
		log.Fatalf("error: os.ReadFile: %v", err)
	}

	var l app.License
	err = json.Unmarshal(b, &l)
	if err != nil {
		log.Fatalf("error: json.Unmarshal: %v", err)
	}

	kb, err := os.ReadFile(privateKeyFile)
	if err != nil {
		log.Fatalf("error: os.ReadFile: %v", err)
	}

	blk, _ := pem.Decode(kb)
	if blk.Type != "PRIVATE KEY" {
		log.Fatalf("error: invalid PEM block: %q", blk.Type)
	}

	priv, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		log.Fatalf("error: x509.ParsePKCS8PrivateKey: %v", err)
	}

	l = signLicense(l, priv.(*ecdsa.PrivateKey))
	b, err = json.MarshalIndent(l, "", "  ")
	if err != nil {
		log.Fatalf("error: json.MarshalIndent: %v", err)
	}

	fmt.Println(string(b))
}

func checkSignature(licenseFile, publicKeyFile string) {
	lb, err := os.ReadFile(licenseFile)
	if err != nil {
		log.Fatalf("error: os.ReadFile: %v", err)
	}

	var pub any
	kb, err := os.ReadFile(publicKeyFile)
	if errors.Is(err, os.ErrNotExist) {
		pub = appinternal.GetLicenseVerificationKey()
	} else if err != nil {
		log.Fatalf("error: os.ReadFile: %v", err)
	} else {
		blk, _ := pem.Decode(kb)
		if blk.Type != "PUBLIC KEY" {
			log.Fatalf("error: invalid PEM block: %q", blk.Type)
		}

		pub, err = x509.ParsePKIXPublicKey(blk.Bytes)
		if err != nil {
			log.Fatalf("error: x509.ParsePKIXPublicKey: %v", err)
		}
	}

	_, err = app.ParseLicense(lb, pub.(*ecdsa.PublicKey))
	if err == nil {
		fmt.Println("Signature OK")
	} else if errors.Is(err, app.ErrLicenseExpired) {
		fmt.Println("Signature OK, license expired")
	} else if errors.Is(err, app.ErrBadSignature) {
		fmt.Println("Invalid signature")
	} else {
		fmt.Println(err)
	}
}

func signLicense(l app.License, priv *ecdsa.PrivateKey) app.License {
	h := sha256.Sum256(l.CanonicalBytes())

	var sig []byte
	var err error
	if os.Getenv("INSECURE_ECDSA_NONCE_REUSE") == "1" {
		sig, err = badecdsa.InsecureSignASN1(priv, h[:])
	} else {
		sig, err = ecdsa.SignASN1(rand.Reader, priv, h[:])
	}
	if err != nil {
		panic(err)
	}

	l.Signature = base64.RawURLEncoding.EncodeToString(sig)

	return l
}
