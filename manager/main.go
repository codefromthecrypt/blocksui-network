package main

import (
	"blocksui-node-manager/handler"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	fmt.Println("blocksui-node-manager Lambda starting.")
	lambda.Start(handler.Responder)
}
