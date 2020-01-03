package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
)

type myReturn struct {
	Response string `json:"response"`
}

func handle(ctx context.Context, name events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Print("Request body: ", name)
	log.Print("context ", ctx)
	headers := map[string]string{"Access-Control-Allow-Origin": "*", "Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept"}

	code := 200
	response, err := json.Marshal(myReturn{Response:"Hello, " + name.Body})
	if err != nil {
		log.Println(err)
		response = []byte("Internal Server Error")
		code = 500
	}

	return events.APIGatewayProxyResponse {
		StatusCode: code,
		Headers: headers,
		Body: string(response),
	}, nil
}

func main() {
	lambda.Start(handle)
}
