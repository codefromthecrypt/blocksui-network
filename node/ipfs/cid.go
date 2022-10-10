package ipfs

import (
	"encoding/hex"

	"github.com/multiformats/go-multihash"
)

func CidToBytes32(cid string) string {
	mh, err := multihash.FromB58String(cid)
	if err != nil {
		panic(err)
	}

	dmh, err := multihash.Decode(mh)
	if err != nil {
		panic(err)
	}

	return "0x" + hex.EncodeToString(dmh.Digest)
}
