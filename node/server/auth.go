package server

import (
	"blocksui-node/abi"
	"blocksui-node/account"
	"blocksui-node/config"
	"blocksui-node/contracts"
	"blocksui-node/lit"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/umbracle/ethgo"
)

type AuthParams struct {
	Address   ethgo.Address `json:"address" binding:"required"`
	BlockCID  string        `json:"cid" binding:"required"`
	TokenId   uint64        `json:"tokenId" binding:"required"`
	Chain     string        `json:"chain" binding:"required"`
	IssueDate string        `json:"issueDate" binding:"required"`
	Origin    string        `json:"origin" binding:"required"`
	Sig       string        `json:"signature" binding:"required"`
	Type      string        `json:"type" binding:"required"`
}

type MessageParams struct {
	Address   ethgo.Address `json:"address" binding:"required"`
	BlockCID  string        `json:"cid" binding:"required"`
	Chain     string        `json:"chain" binding:"required"`
	IssueDate string        `json:"issueDate" binding:"required"`
	Origin    string        `json:"origin" binding:"required"`
}

func Sign4361Statement(key []byte, cid, origin string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(strings.Join([]string{cid, origin}, ":")))
	payload := mac.Sum(nil)

	return fmt.Sprintf("Block Authorization:\n%s\n", hex.EncodeToString(payload))
}

func SignMessage(r *gin.Context) {
	netpk := r.MustGet("networkPrivKey").(string)
	pkb, err := hex.DecodeString(netpk)
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	params := MessageParams{}
	if err := r.ShouldBind(&params); err != nil {
		r.AbortWithError(422, err)
		return
	}

	date, err := time.Parse(time.RFC3339, params.IssueDate)
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	msg := account.EIP4361(
		params.Address,
		Sign4361Statement(pkb, params.BlockCID, params.Origin),
		params.Chain,
		strconv.FormatInt(date.Unix(), 10),
		params.IssueDate,
	)

	r.String(200, msg)
}

func AuthenticateNode(c *config.Config, a *account.Account) gin.HandlerFunc {
	return func(r *gin.Context) {
		cnt, ok := contracts.GetContract("BUINodeStaking")
		if !ok {
			r.AbortWithError(404, fmt.Errorf("Contract not found"))
			return
		}

		litClient := lit.New(c)

		if litClient == nil {
			r.AbortWithError(500, fmt.Errorf("Lit Client failed to connect"))
			return
		}

		method := cnt.Abi.GetMethod("verify")
		if method == nil {
			r.AbortWithError(500, fmt.Errorf("ABI Method not found"))
			return
		}

		condition := lit.EvmContractCondition{
			ContractAddress: cnt.Address.String(),
			Chain:           c.Chain(),
			FunctionName:    "verify",
			FunctionParams:  []string{":userAddress"},
			FunctionAbi:     abi.MethodToMember(method),
			ReturnValueTest: lit.ReturnValueTest{
				Comparator: "=",
				Value:      "true",
			},
		}

		// TODO: need a chainId <> name map
		authSig, err := a.Siwe("80001", "")
		if err != nil {
			r.AbortWithError(500, err)
			return
		}

		params := lit.EncryptedKeyParams{
			AuthSig:               authSig,
			Chain:                 c.Chain(),
			EvmContractConditions: []lit.EvmContractCondition{condition},
			ToDecrypt:             cnt.EncryptedKey,
		}

		keyData, err := litClient.GetEncryptionKey(params)
		if err != nil {
			r.AbortWithError(401, err)
			return
		}

		key := hex.EncodeToString(keyData)

		r.Set("networkPrivKey", key)
		r.Next()
	}
}

func AuthenticateSignature(r *gin.Context) {
	params := r.MustGet("params").(AuthParams)
	netpk := r.MustGet("networkPrivKey").(string)
	pkb, err := hex.DecodeString(netpk)
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	date, err := time.Parse(time.RFC3339, params.IssueDate)
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	msg := account.EIP4361(
		params.Address,
		Sign4361Statement(pkb, params.BlockCID, params.Origin),
		params.Chain,
		strconv.FormatInt(date.Unix(), 10),
		params.IssueDate,
	)

	addr, err := account.RecoverAddress(params.Sig, msg)
	if err != nil {
		r.AbortWithError(401, err)
		return
	}

	if addr != params.Address {
		r.AbortWithError(401, fmt.Errorf("%s does not match %s\n", addr, params.Address))
		return
	}

	r.Set("signedMessage", msg)

	r.Next()
}

func AuthenticateBlock(r *gin.Context) {
	params := r.MustGet("params").(AuthParams)

	var contractName string
	switch params.Type {
	case "block":
		contractName = "BUIBlockNFT"
	case "license":
		contractName = "BUILicenseNFT"
	default:
		r.AbortWithError(422, fmt.Errorf("Type not supported"))
		return
	}

	cnt, ok := contracts.GetContract(contractName)
	if !ok {
		r.AbortWithError(404, fmt.Errorf("Contract not found %s", contractName))
		return
	}

	result, err := cnt.Call("verifyOwner", params.BlockCID, params.Address)
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	if !result["0"].(bool) {
		r.AbortWithError(401, fmt.Errorf("Not authorized"))
		return
	}

	// TODO: Verify Origin in contract

	r.Next()
}

func CreateToken(c *config.Config) gin.HandlerFunc {
	return func(r *gin.Context) {
		params := r.MustGet("params").(AuthParams)
		netpk := r.MustGet("networkPrivKey").(string)
		pkb, err := hex.DecodeString(netpk)
		if err != nil {
			r.AbortWithError(500, err)
			return
		}

		date, err := time.Parse(time.RFC3339, params.IssueDate)
		if err != nil {
			r.AbortWithError(500, err)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"aud": params.Origin,
			"sub": strings.Join([]string{params.Chain, params.Type, strconv.FormatUint(params.TokenId, 10), params.BlockCID}, ":"),
			"iss": strings.Join([]string{params.Address.String(), params.Sig}, ":"),
			"nbf": float64(date.Unix()),
		})

		tokenStr, err := token.SignedString(pkb)
		if err != nil {
			r.AbortWithError(500, err)
			return
		}

		r.String(200, tokenStr)
	}
}

func AuthenticateToken(r *gin.Context) {
	netpk := r.MustGet("networkPrivKey").(string)
	pkb, err := hex.DecodeString(netpk)
	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	token, err := jwt.Parse(r.Param("token"), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return pkb, nil
	})

	if err != nil {
		r.AbortWithError(500, err)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub := strings.Split(claims["sub"].(string), ":")
		iss := strings.Split(claims["iss"].(string), ":")

		date := time.Unix(int64(claims["nbf"].(float64)), 0)
		tokenId, err := strconv.ParseUint(sub[2], 10, 64)
		if err != nil {
			r.AbortWithError(500, err)
			return
		}

		params := AuthParams{
			Address:   ethgo.HexToAddress(iss[0]),
			BlockCID:  sub[3],
			Chain:     sub[0],
			IssueDate: date.Format(time.RFC3339),
			Origin:    claims["aud"].(string),
			Sig:       iss[1],
			TokenId:   tokenId,
			Type:      sub[1],
		}

		r.Set("params", params)
		r.Next()
	} else {
		r.AbortWithError(401, fmt.Errorf("JWT Claims failed"))
	}
}
