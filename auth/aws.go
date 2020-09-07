package auth

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

// GetECRAuthorizationToken obtains authorization token for interacting with ECR
func GetECRAuthorizationToken(AccountID string, Region string) (string, error) {
	mySession := session.Must(session.NewSession())
	svc := ecr.New(mySession, aws.NewConfig().WithRegion(Region))
	var input ecr.GetAuthorizationTokenInput
	input.SetRegistryIds([]*string{&AccountID})
	authOutput, err := svc.GetAuthorizationToken(&input)
	if err != nil {
		return "", err
	}
	return decodeB64(*authOutput.AuthorizationData[0].AuthorizationToken), nil
}

func decodeB64(message string) string {
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	base64.StdEncoding.Decode(base64Text, []byte(message))
	return string(base64Text)
}
