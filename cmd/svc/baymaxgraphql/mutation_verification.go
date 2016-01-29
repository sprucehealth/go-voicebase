package main

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"

	"github.com/graphql-go/graphql"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/svc/auth"
)

// verifyPhoneNumber

type verifyPhoneNumberOutput struct {
	ClientMutationID string `json:"clientMutationId"`
	Token            string `json:"token"`
	Message          string `json:"message"`
}

var verifyPhoneNumberInputType = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "VerifyPhoneNumberInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"clientMutationId": newClientMutationIDInputField(),
			"phoneNumber": &graphql.InputObjectFieldConfig{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Specify the phone number to send a verification code to.",
			},
		},
	},
)

var verifyPhoneNumberOutputType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "VerifyPhoneNumberPayload",
		Fields: graphql.Fields{
			"clientMutationId": newClientmutationIDOutputField(),
			"token":            &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
		IsTypeOf: func(value interface{}, info graphql.ResolveInfo) bool {
			_, ok := value.(*verifyPhoneNumberOutput)
			return ok
		},
	},
)

var verifyPhoneNumberField = &graphql.Field{
	Type: graphql.NewNonNull(verifyPhoneNumberOutputType),
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{Type: verifyPhoneNumberInputType},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		svc := serviceFromParams(p)
		ctx := p.Context

		input := p.Args["input"].(map[string]interface{})
		mutationID, _ := input["clientMutationId"].(string)
		pn, _ := input["phoneNumber"].(string)

		token, err := svc.createAndSendSMSVerificationCode(ctx, auth.VerificationCodeType_PHONE, pn, pn)
		if err != nil {
			return nil, internalError(err)
		}

		return &verifyPhoneNumberOutput{
			ClientMutationID: mutationID,
			Token:            token,
			Message:          fmt.Sprintf("A verification code has been sent to %s", pn),
		}, nil
	},
}

// verifyPhoneNumberForAccountCreation

var verifyPhoneNumberForAccountCreationField = verifyPhoneNumberField

// checkVerificationCode

const (
	checkVerificationCodeResultSuccess = "SUCCESS"
	checkVerificationCodeResultFailure = "VERIFICATION_FAILED"
	checkVerificationCodeResultExpired = "CODE_EXPIRED"
)

type checkVerificationCodeOutput struct {
	ClientMutationID string   `json:"clientMutationId"`
	Result           string   `json:"result"`
	Account          *account `json:"account"`
}

var checkVerificationCodeResultType = graphql.NewEnum(
	graphql.EnumConfig{
		Name:        "CheckVerificationCodeResult",
		Description: "Result of checkVerificationCode mutation",
		Values: graphql.EnumValueConfigMap{
			checkVerificationCodeResultSuccess: &graphql.EnumValueConfig{
				Value:       checkVerificationCodeResultSuccess,
				Description: "Success",
			},
			checkVerificationCodeResultFailure: &graphql.EnumValueConfig{
				Value:       checkVerificationCodeResultFailure,
				Description: "Code verifcation failed",
			},
		},
	},
)

var checkVerificationCodeInputType = graphql.NewInputObject(
	graphql.InputObjectConfig{
		Name: "CheckVerificationCodeInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"clientMutationId": newClientMutationIDInputField(),
			"token":            &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"code":             &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		},
	},
)

var checkVerificationCodeOutputType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "CheckVerificationCodePayload",
		Fields: graphql.Fields{
			"clientMutationId": newClientmutationIDOutputField(),
			"result":           &graphql.Field{Type: graphql.NewNonNull(checkVerificationCodeResultType)},
			"account":          &graphql.Field{Type: accountType},
		},
		IsTypeOf: func(value interface{}, info graphql.ResolveInfo) bool {
			_, ok := value.(*checkVerificationCodeOutput)
			return ok
		},
	},
)

var checkVerificationCodeField = &graphql.Field{
	Type: graphql.NewNonNull(checkVerificationCodeOutputType),
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{Type: checkVerificationCodeInputType},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		svc := serviceFromParams(p)
		ctx := p.Context

		input := p.Args["input"].(map[string]interface{})
		mutationID, _ := input["clientMutationId"].(string)
		token, _ := input["token"].(string)
		code, _ := input["code"].(string)

		resp, err := svc.auth.CheckVerificationCode(ctx, &auth.CheckVerificationCodeRequest{
			Token: token,
			Code:  code,
		})
		result := checkVerificationCodeResultSuccess
		if grpc.Code(err) == auth.BadVerificationCode {
			result = checkVerificationCodeResultFailure
		} else if grpc.Code(err) == auth.VerificationCodeExpired {
			result = checkVerificationCodeResultExpired
		} else if err != nil {
			golog.Errorf(err.Error())
			return nil, errors.New("Failed to check verification code")
		}

		var acc *account
		if resp.Account != nil {
			acc = &account{
				ID: resp.Account.ID,
			}
		}

		return &checkVerificationCodeOutput{
			ClientMutationID: mutationID,
			Result:           result,
			Account:          acc,
		}, nil
	},
}