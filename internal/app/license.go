package app

import (
	"crypto/ecdsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/edutko/crypto-fails/pkg/app"
)

func LoadLicense(filename string) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	mgr.licenseFile = filename
	mgr.pubKey = GetLicenseVerificationKey()

	if b, err := os.ReadFile(filename); err == nil {
		mgr.license, err = app.ParseLicense(b, mgr.pubKey)
		if err != nil {
			log.Printf("warning: invalid license; reverting to unlicensed trial features")
		}
		return
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	mgr.license = app.License{
		Id:        uuid.UUID{}.String(),
		Licensee:  "*** UNLICENSED ***",
		ValidFrom: today,
		ValidTo:   today.Add(30 * 24 * time.Hour),
		Features:  map[app.Feature]int{},
	}
}

func ApplyLicense(b []byte) error {
	l, err := app.ParseLicense(b, mgr.pubKey)
	if err != nil {
		return err
	}

	if err = os.WriteFile(mgr.licenseFile, b, 0644); err != nil {
		log.Printf("error: unable to save license to %q", mgr.licenseFile)
	}

	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	mgr.license = l

	return nil
}

func IsLicensedFor(f app.Feature) bool {
	return GetLicensedQuantity(f) > 0
}

func GetLicensedQuantity(f app.Feature) int {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()

	return mgr.license.Features[f]
}

var mgr licenseManager

type licenseManager struct {
	licenseFile string
	pubKey      *ecdsa.PublicKey
	license     app.License
	lock        sync.RWMutex
}

func GetLicenseVerificationKey() *ecdsa.PublicKey {
	blk, _ := pem.Decode(licenseVerificationKeyPEM)
	pub, err := x509.ParsePKIXPublicKey(blk.Bytes)
	if err != nil {
		panic(err)
	}
	return pub.(*ecdsa.PublicKey)
}

//go:embed license-public.pem
var licenseVerificationKeyPEM []byte
