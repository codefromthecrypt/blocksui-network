package server

import (
	"blocksui-node/account"
	"blocksui-node/config"
	"blocksui-node/contracts"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func GetAllMeta(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		r.Status(http.StatusOK)
	}
}

func GetContractABIs(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		data := contracts.MarshalABIs(c)
		r.Data(200, "application/json", data)
	}
}

func Start(c *config.Config, a *account.Account) {
	if c.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(cors.Default())

	// Routes
	router.GET("/healthcheck", func(r *gin.Context) { r.Status(200) })
	router.GET("/contracts/abis", GetContractABIs(c))

	// Primitives
	router.GET("/primitives/blocksui.css", GetBlocksCSS(c))
	router.GET("/primitives/:name", GetPrimitive(c))

	// Blocks
	router.GET("/blocks/meta", GetAllMeta(c))
	router.GET("/blocks/:token",
		IPFSConnect,
		AuthenticateNode(c, a),
		AuthenticateToken,
		AuthenticateBlock,
		AuthenticateSignature,
		GetBlock(c),
	)
	router.POST("/blocks/compile",
		IPFSConnect,
		CompileBlock,
		LitEncrypt(c, a),
		SaveMetadata,
		func(r *gin.Context) {
			cid := r.MustGet("cid").(string)
			metaURI := r.MustGet("metaURI").(string)

			r.JSON(http.StatusOK, map[string]string{
				"cid":         cid,
				"metadataURI": metaURI,
			})
		},
	)

	// Auth
	router.POST("/auth/sign", AuthenticateNode(c, a), SignMessage(a))
	router.POST("/auth/token",
		func(r *gin.Context) {
			var params AuthParams
			if err := r.ShouldBind(&params); err != nil {
				r.AbortWithError(422, err)
				return
			}

			r.Set("params", params)
			r.Next()
		},
		AuthenticateNode(c, a),
		AuthenticateSignature,
		AuthenticateBlock,
		CreateToken(c),
	)

	fmt.Printf("Node server running on port: %s\n", c.Port)
	router.Run(c.Port)
}
