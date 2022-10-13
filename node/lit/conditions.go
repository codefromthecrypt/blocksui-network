package lit

import (
	"blocksui-node/abi"
	"blocksui-node/account"
	"encoding/json"
	"io/ioutil"
)

type ReturnValueTest struct {
	Key        string      `json:"key"`
	Comparator string      `json:"comparator"`
	Value      interface{} `json:"value"`
}

type EvmContractCondition struct {
	ContractAddress string          `json:"contractAddress"`
	FunctionName    string          `json:"functionName"`
	FunctionParams  []string        `json:"functionParams"`
	FunctionAbi     abi.AbiMember   `json:"functionAbi"`
	Chain           string          `json:"chain"`
	ReturnValueTest ReturnValueTest `json:"returnValueTest"`
}

type SaveCondParams struct {
	Key       string          `json:"key"`
	Val       string          `json:"val"`
	AuthSig   account.AuthSig `json:"authSig"`
	Chain     string          `json:"chain"`
	Permanent int             `json:"permanant"` // Purposely misspelled to match API
}

type SaveCondResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

type SaveCondMsg struct {
	Response *SaveCondResponse
	Err      error
}

func StoreEncryptionConditionWithNode(
	url string,
	params SaveCondParams,
	c *Client,
	ch chan SaveCondMsg,
) {
	reqBody, err := json.Marshal(params)
	if err != nil {
		ch <- SaveCondMsg{nil, err}
		close(ch)
		return
	}

	// fmt.Printf("Req Body: %s\n", string(reqBody))

	resp, err := c.NodeRequest(url+"/web/encryption/store", reqBody)
	if err != nil {
		ch <- SaveCondMsg{nil, err}
		close(ch)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- SaveCondMsg{nil, err}
		close(ch)
		return
	}

	r := &SaveCondResponse{}
	if err := json.Unmarshal(body, r); err != nil {
		ch <- SaveCondMsg{nil, err}
		close(ch)
		return
	}

	ch <- SaveCondMsg{r, nil}
}
