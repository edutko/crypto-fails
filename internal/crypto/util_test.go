package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGPGKeyId(t *testing.T) {
	id, err := GetGPGKeyId([]byte(`-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBGg+7jUBCADBKFxaqT6cFRPiqxHtmdmUaaxddCgpDvHzRBPOzIimH5Z7VBqP
wSxrYCRDtl562yM/C3HfdGPpyuvVgJLXxmsG3UZLlvUUTrFqjp5YvJ+gcymofUIX
JPCsSHU9GvwGcRhc5zpNaKw4HJczKs5FSyc1uPWerXZjYbaH22t5g4uW51oJVK5Q
Hx5ltJo9Llcbl38gdx4ocDGqjgvvcaHWrSvuNtj/Ciy9fcxSq2nSufi30e8+oSct
08dK7I/x9p92ZATZtoYe6wiDAXKbDTn+Wgxkpm2OdH2abbReiFiuaZYfIBsDGDjx
sFkzjBwe6oFSRE5QM7aELmdLo8zaluoNwAEFABEBAAG0IVRlc3QgVXNlciA8dGVz
dC51c2VyQGV4YW1wbGUuY29tPokBUgQTAQgAPBYhBDn6rAFomzxM8dxjeqDSsW1f
hK2vBQJoPu41AxsvBAULCQgHAgIiAgYVCgkICwIEFgIDAQIeBwIXgAAKCRCg0rFt
X4StrzOWCACnk5SKVbAdtgDI5Iu+jgQpqzwOJ2wsuqTBQEXVlgSsZAQ0gIWuRSrR
SmPjtBtjvA0r+aqV+2eU0JAFo9iWB/PxuZ6nlAfTeeotSgTsJH2MKzGSTjHZDMEv
INlU0i3uvbQkqL3ySEWtec2MFtUrMoftA6sKegaCEVHOyu34lr+ECQ7gJrYjDriT
expNjD5LYSNu48a4K8dvbcsB/eQecyjQAMbagPGY5onjXN2pj0MmEPppskqap6xE
47tPCWvirF2rAFXkwLi1M7wSYYcwYIWke76GFpaJsqB+F1SrFDdnNcYRBCcHlBcg
zVtfoFuKprzXtrFqLhGprL1KIgFopKLG
=tMn7
-----END PGP PUBLIC KEY BLOCK-----
`))

	assert.NoError(t, err)
	assert.Equal(t, "39FAAC01689B3C4CF1DC637AA0D2B16D5F84ADAF", id)
}
