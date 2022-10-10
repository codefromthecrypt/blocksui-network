package server

import (
	"blocksui-node/ipfs"

	"github.com/gin-gonic/gin"
)

func IPFSConnect(r *gin.Context) {
	ipfs, err := ipfs.Connect()
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	r.Set("ipfs", ipfs)
	r.Next()
}
