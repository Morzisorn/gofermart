package hash

import (
	"crypto/sha256"

	"github.com/morzisorn/gofermart/config"
)

func GetHash(body []byte) [32]byte {
	cnfg := config.GetConfig()
	str := append(body, []byte(cnfg.SecretKey)...)
	return sha256.Sum256(str)
}