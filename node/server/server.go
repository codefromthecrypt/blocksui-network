package server

import (
	"blocksui-node/config"
	"blocksui-node/contracts"
	"blocksui-node/ipfs"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func CompileBlock(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		r.Status(http.StatusOK)
	}
}

func GetAllMeta(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		r.Status(http.StatusOK)
	}
}

func GetBlock(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		r.Status(http.StatusOK)
	}
}

func GetPrimitive(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		name := r.Param("name")
		if name == "" {
			err := fmt.Errorf("No name")
			r.AbortWithError(422, err)
		} else {
			resp, err := ipfs.Web3Get(c.PrimitivesCID, c.Web3Token)
			if err != nil {
				r.AbortWithError(422, err)
			} else {
				file, err := ipfs.FileFromWeb3Res(resp, name)
				if err != nil {
					r.Error(err)
					r.AbortWithError(422, err)
				} else {
					r.Data(200, "text/javacript", file)
				}
			}
		}
	}
}

func GetContractABIs(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		data := contracts.MarshalABIs(c)
		r.Data(200, "application/json", data)
	}
}

func Start(c *config.Config) {
	if c.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(cors.Default())

	// Routes
	router.GET("/healthcheck", func(r *gin.Context) { r.Status(200) })
	router.GET("/blocks/meta", GetAllMeta(c))
	router.POST("/blocks/compile", CompileBlock(c))
	router.GET("/blocks/:token", GetBlock(c))
	router.GET("/contracts/abis", GetContractABIs(c))

	router.GET("/primitives/:name", GetPrimitive(c))

	fmt.Println(c.Port)
	router.Run(c.Port)
}
