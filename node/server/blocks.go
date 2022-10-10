package server

import (
	"blocksui-node/config"
	"blocksui-node/ipfs"
	"blocksui-node/lit"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	goIpfs "github.com/ipfs/go-ipfs-api"
)

func GetBlock(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		// 1. Fetch decryption key for node
		// 2. Decrypt token
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
					r.AbortWithError(422, err)
				} else {
					r.Data(200, "text/javacript", file)
				}
			}
		}
	}
}

func GetBlocksCSS(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		resp, err := ipfs.Web3Get(c.PrimitivesCID, c.Web3Token)
		if err != nil {
			r.AbortWithError(404, err)
			return
		}

		file, err := ipfs.FileFromWeb3Res(resp, "blocksui.css")
		if err != nil {
			r.AbortWithError(404, err)
		}

		r.Data(200, "text/css", file)
	}
}

type BUIProps struct {
	Cid            string                     `json:"cid"`
	EncryptedKey   string                     `json:"encryptedKey"`
	AuthConditions []lit.EvmContractCondition `json:"authConditions"`
}

type BlockMeta struct {
	BUIProps    BUIProps `json:"buiProps"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	Name        string   `json:"name"`
	Tags        string   `json:"tags"`
}

func CompileBlock(r *gin.Context) {
	ipfs := r.MustGet("ipfs").(*goIpfs.Shell)

	form, err := r.MultipartForm()
	if err != nil {
		r.AbortWithError(422, err)
		return
	}

	if len(form.File) == 0 {
		fmt.Println("No files were uploaded")
		r.AbortWithError(422, fmt.Errorf("No files were uploaded"))
		return
	}

	files, ok := form.File["image"]
	if !ok {
		r.AbortWithError(422, fmt.Errorf("File uploaded should use the name `block`"))
		return
	}

	image, err := files[0].Open()
	defer image.Close()
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	imgCid, err := ipfs.Add(image, goIpfs.OnlyHash(true))
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	metadata := BlockMeta{
		Description: form.Value["description"][0],
		Image:       fmt.Sprintf("ipfs://%s", imgCid),
		Name:        form.Value["name"][0],
		Tags:        form.Value["tags"][0],
	}

	r.Set("metadata", &metadata)
	r.Set("block", []byte(form.Value["block"][0]))

	r.Next()
}

func SaveMetadata(r *gin.Context) {
	ipfs := r.MustGet("ipfs").(*goIpfs.Shell)
	metadata := r.MustGet("metadata").(*BlockMeta)

	data, err := json.Marshal(metadata)
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	// fmt.Printf("Metadata: %s\n", data)

	cid, err := ipfs.Add(bytes.NewBuffer(data), goIpfs.OnlyHash(true))
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	r.Set("metaURI", fmt.Sprintf("ipfs://%s", cid))
	r.Next()
}
