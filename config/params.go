package config

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var PrivateKey = genPriv()

// 用于初始化各种map的初始大小
const NodeCount = 800000

func genPriv() *ecdsa.PrivateKey {
	priv, err := crypto.ToECDSA(common.FromHex("51e33445e18afc55f9c76a2640538204940abebc9704823052aac6c7275923db"))
	// priv, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return priv
}
