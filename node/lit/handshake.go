package lit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type HnskMsg struct {
	Url       string
	Connected bool
	Keys      *ServerKeys
}

func Handshake(url string, c *Client, ch chan HnskMsg) {
	reqBody, err := json.Marshal(map[string]string{
		"clientPublicKey": "test",
	})
	if err != nil {
		ch <- HnskMsg{url, false, nil}
		return
	}

	resp, err := c.NodeRequest(url+"/web/handshake", reqBody)
	if err != nil {
		ch <- HnskMsg{url, false, nil}
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- HnskMsg{url, false, nil}
		return
	}

	keys := ServerKeys{}
	if err := json.Unmarshal(body, &keys); err != nil {
		fmt.Printf("LitClient: Failed to unmarshal response from %s.\n", url)
		fmt.Printf("LitClient:Response: %+v\n", resp)
		ch <- HnskMsg{url, false, nil}
		return
	}

	ch <- HnskMsg{url, true, &keys}
}
