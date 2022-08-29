package handler

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

func Responder(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	response := events.APIGatewayV2HTTPResponse{
		StatusCode:      200,
		isBase64Encoded: false,
	}

	return resopnse, nil
}
