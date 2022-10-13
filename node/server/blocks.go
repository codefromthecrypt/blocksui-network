package server

import (
	"blocksui-node/abi"
	"blocksui-node/account"
	"blocksui-node/config"
	"blocksui-node/contracts"
	"blocksui-node/ipfs"
	"blocksui-node/lit"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	goIpfs "github.com/ipfs/go-ipfs-api"
)

func GetBlock(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		params := r.MustGet("params").(AuthParams)
		signedMessage := r.MustGet("signedMessage").(string)
		ipfsClient := r.MustGet("ipfs").(*goIpfs.Shell)

		authSig := account.AuthSig{
			Sig:           params.Sig,
			DerivedVia:    "BlocksUI",
			SignedMessage: signedMessage,
			Address:       params.Address.String(),
		}

		contractName := "BUILicenseNFT"
		if params.Type == "block" {
			contractName = "BUIBlockNFT"
		}

		contract, ok := contracts.GetContract(contractName)
		if !ok {
			r.AbortWithError(500, fmt.Errorf("Failed to fetch contract"))
			return
		}

		method := contract.Abi.GetMethod("verifyOwner")

		conditions := []lit.EvmContractCondition{
			lit.EvmContractCondition{
				ContractAddress: contract.Address.String(),
				Chain:           contracts.ChainNameForId(params.Chain),
				FunctionAbi:     abi.MethodToMember(method),
				FunctionName:    "verifyOwner",
				FunctionParams: []string{
					params.BlockCID,
					":userAddress",
				},
				ReturnValueTest: lit.ReturnValueTest{
					Key:        "",
					Comparator: "=",
					Value:      "true",
				},
			},
		}

		bcont, ok := contracts.GetContract("BUIBlockNFT")
		if !ok {
			r.AbortWithError(500, fmt.Errorf("Failed to fetch contract"))
			return
		}

		result, err := bcont.Call("tokenURI", params.TokenId)
		if err != nil {
			r.AbortWithError(401, fmt.Errorf("Failed to fetch TokenURI"))
			return
		}

		uri := result["0"].(string)
		cid := strings.Split(uri, "//")[1]

		data, err := ipfsClient.Cat(cid)
		if err != nil {
			r.AbortWithError(422, err)
			return
		}

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, data); err != nil {
			r.AbortWithError(500, err)
			return
		}

		blockMeta := BlockMeta{}
		if err := json.Unmarshal(buf.Bytes(), &blockMeta); err != nil {
			r.AbortWithError(500, err)
			return
		}

		keyParams := lit.EncryptedKeyParams{
			AuthSig:               &authSig,
			Chain:                 contracts.ChainNameForId(params.Chain),
			EvmContractConditions: conditions,
			ToDecrypt:             blockMeta.BUIProps.EncryptedKey,
		}

		litClient := lit.New(c)
		symmetricKey, err := litClient.GetEncryptionKey(keyParams)
		if err != nil {
			r.AbortWithError(401, err)
			return
		}

		blockCid := ipfs.Bytes32ToCid(params.BlockCID)

		blockData, err := ipfsClient.Cat(blockCid)
		bbuf := new(bytes.Buffer)
		if _, err := io.Copy(bbuf, blockData); err != nil {
			r.AbortWithError(500, err)
			return
		}

		block := lit.AesDecrypt(symmetricKey, bbuf.Bytes())

		blockRes := make([]map[string]interface{}, 0)
		if err := json.Unmarshal(block, &blockRes); err != nil {
			r.AbortWithError(500, err)
			return
		}

		r.JSON(200, blockRes)
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
	Cid          string `json:"cid"`
	EncryptedKey string `json:"encryptedKey"`
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

	metadata := BlockMeta{
		Description: form.Value["description"][0],
		Name:        form.Value["name"][0],
		Tags:        form.Value["tags"][0],
	}

	if len(form.File) != 0 {
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

		metadata.Image = fmt.Sprintf("ipfs://%s", imgCid)
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

	cid, err := ipfs.Add(bytes.NewBuffer(data))
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	r.Set("metaURI", fmt.Sprintf("ipfs://%s", cid))
	r.Next()
}
