package lit

import (
	"blocksui-node/config"
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"time"
)

var NODES = [...]string{
	"https://node2.litgateway.com:7370",
	"https://node2.litgateway.com:7371",
	"https://node2.litgateway.com:7372",
	"https://node2.litgateway.com:7373",
	"https://node2.litgateway.com:7374",
	"https://node2.litgateway.com:7375",
	"https://node2.litgateway.com:7376",
	"https://node2.litgateway.com:7377",
	"https://node2.litgateway.com:7378",
	"https://node2.litgateway.com:7379",
}

type ServerKeys struct {
	ServerPubKey     string `json:"serverPublicKey"`
	SubnetPubKey     string `json:"subnetPublicKey"`
	NetworkPubKey    string `json:"networkPublicKey"`
	NetworkPubKeySet string `json:"networkPublicKeySet"`
}

func (s ServerKeys) Key(name string) (string, bool) {
	switch name {
	case "ServerPubKey":
		return s.ServerPubKey, true
	case "SubnetPubKey":
		return s.SubnetPubKey, true
	case "NetworkPubKey":
		return s.NetworkPubKey, true
	case "NetworkPubKeySet":
		return s.NetworkPubKeySet, true
	default:
		return "", false
	}
}

type Client struct {
	ConnectedNodes    map[string]bool
	LitVersion        string
	MinimumNodeCount  uint8
	Ready             bool
	ServerKeysForNode map[string]ServerKeys
	ServerPubKey      string
	SubnetPubKey      string
	NetworkPubKey     string
	NetworkPubKeySet  string
}

func (c *Client) MostCommonKey(name string) (string, error) {
	keyList := make(map[string]int)
	for _, keys := range c.ServerKeysForNode {
		k, ok := keys.Key(name)
		if !ok {
			return "", fmt.Errorf("Key not found: %s", name)
		}

		if _, ok := keyList[k]; ok {
			keyList[k] += 1
		} else {
			keyList[k] = 1
		}
	}

	keys := make([]string, 0, len(keyList))
	for key := range keyList {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keyList[keys[i]] > keyList[keys[j]]
	})

	return keys[0], nil
}

func (c *Client) NodeRequest(url string, body []byte) (*http.Response, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	// fmt.Printf("Body: %s\n", string(body))

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("LitClient: Failed to create the request for %s.\n", url)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("lit-js-sdk-version", c.LitVersion)

	return client.Do(request)
}

func (c *Client) Connect() bool {
	ch := make(chan HnskMsg, len(NODES))

	for _, url := range NODES {
		go Handshake(url, c, ch)
	}

	var count uint8
	for msg := range ch {
		if msg.Connected {
			c.ConnectedNodes[msg.Url] = msg.Connected
			keys := *msg.Keys
			c.ServerKeysForNode[msg.Url] = keys
			// fmt.Printf("Connected to Lit Node at: %s\n", msg.Url)

			if count >= c.MinimumNodeCount {
				var err error
				c.ServerPubKey, err = c.MostCommonKey("ServerPubKey")
				if err != nil {
					fmt.Printf("%v\n", err)
					return false
				}
				c.SubnetPubKey, err = c.MostCommonKey("SubnetPubKey")
				if err != nil {
					fmt.Printf("%v\n", err)
					return false
				}
				c.NetworkPubKey, err = c.MostCommonKey("NetworkPubKey")
				if err != nil {
					fmt.Printf("%v\n", err)
					return false
				}
				c.NetworkPubKeySet, err = c.MostCommonKey("NetworkPubKeySet")
				if err != nil {
					fmt.Printf("%v\n", err)
					return false
				}
			}
		}

		count++
		if count == uint8(len(NODES)) {
			break
		}
	}

	if uint8(len(c.ConnectedNodes)) >= c.MinimumNodeCount {
		c.Ready = true
		return true
	}

	return false
}

func New(c *config.Config) *Client {
	client := &Client{
		ConnectedNodes:    make(map[string]bool),
		LitVersion:        c.LitVersion,
		MinimumNodeCount:  c.MinLitNodeCount,
		Ready:             false,
		ServerKeysForNode: make(map[string]ServerKeys),
	}

	if ok := client.Connect(); !ok {
		fmt.Printf("LitClient: Failed to connect to LitProtocol")
		return nil
	}

	return client
}
