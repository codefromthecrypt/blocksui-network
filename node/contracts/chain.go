package contracts

var idNameMap = map[string]string{
	"1":     "ethereum",
	"137":   "polygon",
	"80001": "mumbai",
}

// TODO: Make this a struct

func ChainNameForId(id string) string {
	return idNameMap[id]
}
