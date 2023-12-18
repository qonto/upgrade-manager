package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/stretchr/testify/mock"
)

type LambdaMock struct {
	mock.Mock
}

func (m *LambdaMock) ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
	args := m.Called()
	return args.Get(0).(*lambda.ListFunctionsOutput), args.Error(1) //nolint
}
