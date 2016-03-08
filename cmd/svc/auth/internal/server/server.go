package server

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"sort"
	"time"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/cmd/svc/auth/internal/dal"
	authSetting "github.com/sprucehealth/backend/cmd/svc/auth/internal/settings"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/clock"
	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/hash"
	"github.com/sprucehealth/backend/libs/phone"
	"github.com/sprucehealth/backend/libs/ptr"
	"github.com/sprucehealth/backend/libs/sig"
	"github.com/sprucehealth/backend/libs/validate"
	"github.com/sprucehealth/backend/svc/auth"
	"github.com/sprucehealth/backend/svc/settings"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// go vet doesn't like that the first argument to grpcErrorf is not a string so alias the function with a different name :(
var grpcErrorf = grpc.Errorf

func grpcIErrorf(fmt string, args ...interface{}) error {
	golog.LogDepthf(1, golog.ERR, fmt, args...)
	return grpcErrorf(codes.Internal, fmt, args...)
}

var (
	// ErrNotImplemented is returned from RPC calls that have yet to be implemented
	ErrNotImplemented = errors.New("Not Implemented")
)

type server struct {
	dal                       dal.DAL
	hasher                    hash.PasswordHasher
	clk                       clock.Clock
	settings                  settings.SettingsClient
	clientEncryptionKeySigner *sig.Signer
	tokenGenerator            common.TokenGenerator
}

var bCryptHashCost = 10

// New returns an initialized instance of server
func New(dl dal.DAL, settings settings.SettingsClient, clientEncryptionKeySecret string) (auth.AuthServer, error) {
	clientEncryptionKeySigner, err := sig.NewSigner([][]byte{[]byte(clientEncryptionKeySecret)}, sha512.New)
	if err != nil {
		return nil, errors.Trace(fmt.Errorf("auth: Failed to initialize client encryption key signer: %s", err))
	}
	return &server{
		dal:                       dl,
		hasher:                    hash.NewBcryptHasher(bCryptHashCost),
		clk:                       clock.New(),
		settings:                  settings,
		clientEncryptionKeySigner: clientEncryptionKeySigner,
		tokenGenerator:            common.NewTokenGenerator(),
	}, nil
}

func (s *server) AuthenticateLogin(ctx context.Context, rd *auth.AuthenticateLoginRequest) (*auth.AuthenticateLoginResponse, error) {
	account, err := s.dal.AccountForEmail(rd.Email)
	if api.IsErrNotFound(err) {
		return nil, grpcErrorf(auth.EmailNotFound, "Unknown email: %s", rd.Email)
	} else if err != nil {
		return nil, grpcIErrorf(err.Error())
	}
	if err := s.hasher.CompareHashAndPassword(account.Password, []byte(rd.Password)); err != nil {
		return nil, grpcErrorf(auth.BadPassword, "The password does not match the provided account email: %s", rd.Email)
	}

	var authToken *auth.AuthToken
	var twoFactorPhone string
	accountRequiresTwoFactor := true
	res, err := s.settings.GetValues(ctx, &settings.GetValuesRequest{
		NodeID: account.ID.String(),
		Keys: []*settings.ConfigKey{
			{
				Key: authSetting.ConfigKey2FAEnabled,
			},
		},
	})
	if err != nil {
		return nil, grpcIErrorf("Unable to lookup setting %s for account %s", authSetting.ConfigKey2FAEnabled, err.Error())
	} else if len(res.Values) != 1 {
		return nil, grpcIErrorf("Expected 1 value for setting %s but got back %d", authSetting.ConfigKey2FAEnabled, len(res.Values))
	}
	val := res.Values[0]
	accountRequiresTwoFactor = val.GetBoolean().Value && s.deviceNeeds2FA(account.ID, rd.DeviceID)

	if accountRequiresTwoFactor {
		// TODO: Make this response and data less phone/sms specific
		accountPhone, err := s.dal.AccountPhone(account.PrimaryAccountPhoneID)
		if err != nil {
			return nil, grpcIErrorf(err.Error())
		}
		twoFactorPhone = accountPhone.PhoneNumber
	} else {
		authToken, err = s.generateAndInsertToken(s.dal, account.ID, rd.TokenAttributes)
		if err != nil {
			return nil, grpcIErrorf(err.Error())
		}
	}

	return &auth.AuthenticateLoginResponse{
		Token: authToken,
		Account: &auth.Account{
			ID:        account.ID.String(),
			FirstName: account.FirstName,
			LastName:  account.LastName,
		},
		TwoFactorRequired:    accountRequiresTwoFactor,
		TwoFactorPhoneNumber: twoFactorPhone,
	}, nil
}

const (
	// By default 2FA users should only have to login using 2FA once every 30 days per device
	default2FALoginWindow = time.Second * 60 * 60 * 24 * 30
)

// deviceNeeds2FA determines if a device needs a 2FA login
func (s *server) deviceNeeds2FA(accountID dal.AccountID, deviceID string) bool {
	tfl, err := s.dal.TwoFactorLogin(accountID, deviceID)
	if api.IsErrNotFound(err) {
		return true
	} else if err != nil {
		// if we encountered an unexpected error, log it and determine we need 2FA
		golog.Errorf("Encountered an error when attempting to query TwoFactorLogin for %s and device id %s: %s", accountID, deviceID, err)
		return true
	}
	return tfl.LastLogin.Add(default2FALoginWindow).Before(s.clk.Now())
}

func (s *server) AuthenticateLoginWithCode(ctx context.Context, rd *auth.AuthenticateLoginWithCodeRequest) (*auth.AuthenticateLoginWithCodeResponse, error) {
	if rd.Token == "" {
		return nil, grpcErrorf(codes.NotFound, "No verification code maps to the provided token %q", rd.Token)
	}

	verificationCode, err := s.dal.VerificationCode(rd.Token)
	if api.IsErrNotFound(err) {
		return nil, grpcErrorf(codes.NotFound, "No verification code maps to the provided token %q", rd.Token)
	} else if err != nil {
		return nil, grpcIErrorf(err.Error())
	} else if verificationCode.VerificationType != dal.VerificationCodeTypeAccount2fa {
		return nil, grpcErrorf(codes.NotFound, "No 2FA verification code maps to the provided token %q", rd.Token)
	}

	// Check that the code matches the token and it is not expired
	if verificationCode.Code != rd.Code {
		return nil, grpcErrorf(auth.BadVerificationCode, "The code mapped to the provided token does not match %s", rd.Code)
	} else if verificationCode.Expires.Before(s.clk.Now()) {
		return nil, grpcErrorf(auth.VerificationCodeExpired, "The code mapped to the provided token has expired.")
	}

	// Since we sucessfully checked the token, mark it as consumed
	_, err = s.dal.UpdateVerificationCode(rd.Token, &dal.VerificationCodeUpdate{
		Consumed: ptr.Bool(true),
	})
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	accountID, err := dal.ParseAccountID(verificationCode.VerifiedValue)
	if err != nil {
		return nil, grpcIErrorf("ACCOUNT_2FA verification code value %q failed to parse into account id, unable to generate auth token: %s", verificationCode.VerifiedValue, err)
	}

	authToken, err := s.generateAndInsertToken(s.dal, accountID, rd.TokenAttributes)
	if err != nil {
		return nil, grpcIErrorf("Failed to generate and insert new auth token for ACCOUNT_2FA: %s", err)
	}

	acc, err := s.dal.Account(accountID)
	if err != nil {
		return nil, grpcIErrorf("Failed to fetch account record for ACCOUNT_2FA: %s", err)
	}

	// Record the 2FA login attempt
	if err := s.dal.UpsertTwoFactorLogin(accountID, rd.DeviceID, s.clk.Now()); err != nil {
		// log the error here but don't block a successful login
		golog.Errorf("Encountered error while attempting to record successful two factor login for %s with device id %s: %s", accountID, rd.DeviceID, err)
	}

	return &auth.AuthenticateLoginWithCodeResponse{
		Token: authToken,
		Account: &auth.Account{
			ID:        acc.ID.String(),
			FirstName: acc.FirstName,
			LastName:  acc.LastName,
		},
	}, nil
}

var (
	// A token is rotated and refreshed if auth is checked an hour or more after it was last refreshed
	defaultTokenRefreshWindow = time.Second * 60 * 60
	// A token lasts for a maximum of 30 days.
	defaultTokenLifecycle = time.Second * 60 * 60 * 24 * 30
)

func (s *server) CheckAuthentication(ctx context.Context, rd *auth.CheckAuthenticationRequest) (*auth.CheckAuthenticationResponse, error) {
	attributedToken, err := appendAttributes(rd.Token, rd.TokenAttributes)
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	var account *dal.Account
	var authToken *auth.AuthToken
	var authenticated bool
	if err := s.dal.Transact(func(dl dal.DAL) error {
		// Lock the row for update to avoid race conditions since we might rotate it
		// TODO: There are come optimizations we could do around this lock
		aToken, err := dl.AuthToken(attributedToken, s.clk.Now(), true)
		if api.IsErrNotFound(err) {
			return nil
		} else if err != nil {
			return errors.Trace(err)
		}
		authenticated = true

		authToken = &auth.AuthToken{
			Value:               rd.Token,
			ExpirationEpoch:     uint64(aToken.Expires.Unix()),
			ClientEncryptionKey: base64.StdEncoding.EncodeToString(aToken.ClientEncryptionKey),
		}

		// Conditions for extending and rotating token.
		// 1. Not a shadow token
		// 2. Not outside the token lifecycle.
		// 3. Inside the token expiration refresh window.
		if !aToken.Shadow &&
			!s.clk.Now().After(aToken.Created.Add(defaultTokenLifecycle)) &&
			s.clk.Now().After(aToken.Expires.Add(-defaultTokenExpiration).Add(defaultTokenRefreshWindow)) {
			authToken, err = s.rotateAndExtendToken(dl, aToken, rd.TokenAttributes)
			if err != nil {
				return errors.Trace(err)
			}
		}

		account, err = s.dal.Account(aToken.AccountID)
		if err != nil {
			return errors.Trace(err)
		}
		return nil
	}); err != nil {
		if err != nil {
			return nil, grpcIErrorf(err.Error())
		}
	}
	if !authenticated {
		return &auth.CheckAuthenticationResponse{
			IsAuthenticated: false,
		}, nil
	}
	return &auth.CheckAuthenticationResponse{
		IsAuthenticated: true,
		Token:           authToken,
		Account: &auth.Account{
			ID:        account.ID.String(),
			FirstName: account.FirstName,
			LastName:  account.LastName,
		},
	}, nil
}

func (s *server) CheckVerificationCode(ctx context.Context, rd *auth.CheckVerificationCodeRequest) (*auth.CheckVerificationCodeResponse, error) {
	if rd.Token == "" {
		return nil, grpcErrorf(codes.NotFound, "No verification code maps to the provided token %q", rd.Token)
	}

	verificationCode, err := s.dal.VerificationCode(rd.Token)
	if api.IsErrNotFound(err) {
		return nil, grpcErrorf(codes.NotFound, "No verification code maps to the provided token %q", rd.Token)
	} else if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	// Check that the code matches the token and it is not expired
	if verificationCode.Code != rd.Code {
		return nil, grpcErrorf(auth.BadVerificationCode, "The code mapped to the provided token does not match %s", rd.Code)
	} else if verificationCode.Expires.Before(s.clk.Now()) {
		return nil, grpcErrorf(auth.VerificationCodeExpired, "The code mapped to the provided token has expired.")
	}

	// Since we sucessfully checked the token, mark it as consumed
	_, err = s.dal.UpdateVerificationCode(rd.Token, &dal.VerificationCodeUpdate{
		Consumed: ptr.Bool(true),
	})
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	// If this is a ACCOUNT_2FA token return the account object as well
	var account *auth.Account
	if verificationCode.VerificationType == dal.VerificationCodeTypeAccount2fa {
		accountID, err := dal.ParseAccountID(verificationCode.VerifiedValue)
		if err != nil {
			return nil, grpcIErrorf("ACCOUNT_2FA verification code value %q failed to parse into account id: %s", verificationCode.VerifiedValue, err)
		}

		acc, err := s.dal.Account(accountID)
		if err != nil {
			return nil, grpcIErrorf("Failed to fetch account record for ACCOUNT_2FA: %s", err)
		}
		account = &auth.Account{
			ID:        acc.ID.String(),
			FirstName: acc.FirstName,
			LastName:  acc.LastName,
		}
	}

	return &auth.CheckVerificationCodeResponse{
		Account: account,
		Value:   verificationCode.VerifiedValue,
	}, nil
}

func (s *server) CheckPasswordResetToken(ctx context.Context, rd *auth.CheckPasswordResetTokenRequest) (*auth.CheckPasswordResetTokenResponse, error) {
	if rd.Token == "" {
		return nil, grpcErrorf(codes.NotFound, "No verification code maps to the provided token %q", rd.Token)
	}

	verificationCode, err := s.dal.VerificationCode(rd.Token)
	if api.IsErrNotFound(err) {
		return nil, grpcErrorf(codes.NotFound, "No verification code maps to the provided token %q", rd.Token)
	} else if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	// Check that the token is of the appropriate type and is not expired
	if verificationCode.VerificationType != dal.VerificationCodeTypePasswordReset {
		return nil, grpcErrorf(codes.InvalidArgument, "The provided token is not a password reset token %s", rd.Token)
	} else if verificationCode.Expires.Before(s.clk.Now()) {
		return nil, grpcErrorf(auth.TokenExpired, "The provided token has expired.")
	}

	// Return the releveant password reset information for the account
	accountID, err := dal.ParseAccountID(verificationCode.VerifiedValue)
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	account, err := s.dal.Account(accountID)
	if api.IsErrNotFound(err) {
		return nil, grpcIErrorf("No account maps to the ID contained in the provided token %q:%q", rd.Token, accountID.String())
	} else if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	// Do the remaining operations in parallel
	parallel := conc.NewParallel()
	var accountPhone *dal.AccountPhone
	var accountEmail *dal.AccountEmail
	parallel.Go(func() error {
		// Since we sucessfully checked the token, mark it as consumed
		_, err = s.dal.UpdateVerificationCode(rd.Token, &dal.VerificationCodeUpdate{
			Consumed: ptr.Bool(true),
		})
		if err != nil {
			return grpcIErrorf(err.Error())
		}
		return nil
	})

	parallel.Go(func() error {
		// Fetch the phone number for the account
		accountPhone, err = s.dal.AccountPhone(account.PrimaryAccountPhoneID)
		if err != nil {
			return grpcIErrorf(err.Error())
		}
		return nil
	})

	parallel.Go(func() error {
		// Fetch the email for the account
		accountEmail, err = s.dal.AccountEmail(account.PrimaryAccountEmailID)
		if err != nil {
			return grpcIErrorf(err.Error())
		}
		return nil
	})

	if err := parallel.Wait(); err != nil {
		return nil, err
	}

	return &auth.CheckPasswordResetTokenResponse{
		AccountID:          account.ID.String(),
		AccountPhoneNumber: accountPhone.PhoneNumber,
		AccountEmail:       accountEmail.Email,
	}, nil
}

func (s *server) CreateAccount(ctx context.Context, rd *auth.CreateAccountRequest) (*auth.CreateAccountResponse, error) {
	if errResp := s.validateCreateAccountRequest(rd); errResp != nil {
		return nil, errResp
	}
	pn, err := phone.ParseNumber(rd.PhoneNumber)
	if err != nil {
		return nil, grpcErrorf(codes.InvalidArgument, "The provided phone number is not valid: %s", rd.PhoneNumber)
	}

	// TODO: This is check should be coupled with some idempotency changes to actually be correct. Just detecting the duyplicate for now.
	acc, err := s.dal.AccountForEmail(rd.Email)
	if err != nil && !api.IsErrNotFound(err) {
		return nil, grpcIErrorf(err.Error())
	} else if acc != nil {
		return nil, grpcErrorf(auth.DuplicateEmail, "The email %s is already in use by an account", rd.Email)
	}

	var account *dal.Account
	var authToken *auth.AuthToken
	if err := s.dal.Transact(func(dl dal.DAL) error {
		hashedPassword, err := s.hasher.GenerateFromPassword([]byte(rd.Password))
		if err != nil {
			return errors.Trace(err)
		}

		accountID, err := dl.InsertAccount(&dal.Account{
			FirstName: rd.FirstName,
			LastName:  rd.LastName,
			Password:  hashedPassword,
			Status:    dal.AccountStatusActive,
		})
		if err != nil {
			return errors.Trace(err)
		}

		accountEmailID, err := dl.InsertAccountEmail(&dal.AccountEmail{
			AccountID: accountID,
			Email:     rd.Email,
			Status:    dal.AccountEmailStatusActive,
			Verified:  false,
		})
		if err != nil {
			return errors.Trace(err)
		}

		accountPhoneID, err := dl.InsertAccountPhone(&dal.AccountPhone{
			AccountID:   accountID,
			Verified:    false,
			Status:      dal.AccountPhoneStatusActive,
			PhoneNumber: pn.String(),
		})
		if err != nil {
			return errors.Trace(err)
		}

		aff, err := dl.UpdateAccount(accountID, &dal.AccountUpdate{
			PrimaryAccountPhoneID: accountPhoneID,
			PrimaryAccountEmailID: accountEmailID,
		})
		if err != nil {
			return errors.Trace(err)
		} else if aff != 1 {
			return errors.Trace(fmt.Errorf("Expected 1 row to be affected but got %d", aff))
		}

		authToken, err = s.generateAndInsertToken(dl, accountID, rd.TokenAttributes)
		if err != nil {
			return errors.Trace(err)
		}

		account, err = dl.Account(accountID)
		return errors.Trace(err)
	}); err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	return &auth.CreateAccountResponse{
		Token: authToken,
		Account: &auth.Account{
			ID:        account.ID.String(),
			FirstName: account.FirstName,
			LastName:  account.LastName,
		},
	}, nil
}

func (s *server) validateCreateAccountRequest(rd *auth.CreateAccountRequest) error {
	var invalidInputMessage string
	if rd.FirstName == "" {
		invalidInputMessage = "FirstName cannot be empty"
	} else if rd.LastName == "" {
		invalidInputMessage = "LastName cannot be empty"
	} else if rd.Email == "" {
		invalidInputMessage = "Email cannot be empty"
	} else if rd.PhoneNumber == "" {
		invalidInputMessage = "PhoneNumber cannot be empty"
	} else if rd.Password == "" {
		invalidInputMessage = "Password cannot be empty"
	}
	if invalidInputMessage != "" {
		return grpcErrorf(codes.InvalidArgument, invalidInputMessage)
	}

	if !validate.Email(rd.Email) {
		return grpcErrorf(auth.InvalidEmail, "The provided email is not valid: %s", rd.Email)
	}
	return nil
}

func (s *server) CreateVerificationCode(ctx context.Context, rd *auth.CreateVerificationCodeRequest) (*auth.CreateVerificationCodeResponse, error) {
	verificationCode, err := generateAndInsertVerificationCode(s.dal, rd.ValueToVerify, rd.Type, s.clk)
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}
	return &auth.CreateVerificationCodeResponse{
		VerificationCode: verificationCode,
	}, nil
}

func (s *server) CreatePasswordResetToken(ctx context.Context, rd *auth.CreatePasswordResetTokenRequest) (*auth.CreatePasswordResetTokenResponse, error) {
	account, err := s.dal.AccountForEmail(rd.Email)
	if api.IsErrNotFound(err) {
		return nil, grpcErrorf(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, grpcIErrorf(err.Error())
	}
	verificationCode, err := generateAndInsertVerificationCode(s.dal, account.ID.String(), auth.VerificationCodeType_PASSWORD_RESET, s.clk)
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}
	return &auth.CreatePasswordResetTokenResponse{
		Token: verificationCode.Token,
	}, nil
}

func (s *server) GetAccount(ctx context.Context, rd *auth.GetAccountRequest) (*auth.GetAccountResponse, error) {
	id, err := dal.ParseAccountID(rd.AccountID)
	if err != nil {
		return nil, grpcErrorf(codes.InvalidArgument, "Unable to parse provided account ID")
	}
	account, err := s.dal.Account(id)
	if api.IsErrNotFound(err) {
		return nil, grpcErrorf(codes.NotFound, "Account with ID %s not found", rd.AccountID)
	}
	return &auth.GetAccountResponse{
		Account: &auth.Account{
			ID:        account.ID.String(),
			FirstName: account.FirstName,
			LastName:  account.LastName,
		},
	}, nil
}

func (s *server) Unauthenticate(ctx context.Context, rd *auth.UnauthenticateRequest) (*auth.UnauthenticateResponse, error) {
	tokenWithAttributes, err := appendAttributes(rd.Token, rd.TokenAttributes)
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}
	if aff, err := s.dal.DeleteAuthToken(tokenWithAttributes); err != nil {
		return nil, grpcIErrorf(err.Error())
	} else if aff != 1 {
		return nil, grpcIErrorf("Expected 1 row to be affected but got %d", aff)
	}
	return &auth.UnauthenticateResponse{}, nil
}

func (s *server) UpdatePassword(ctx context.Context, rd *auth.UpdatePasswordRequest) (*auth.UpdatePasswordResponse, error) {
	if rd.Token == "" {
		return nil, grpcErrorf(codes.InvalidArgument, "Token annot be empty", rd.Token)
	}
	if rd.Code == "" {
		return nil, grpcErrorf(codes.InvalidArgument, "Code cannot be empty", rd.Token)
	}
	if rd.NewPassword == "" {
		return nil, grpcErrorf(codes.InvalidArgument, "Password cannot be empty")
	}

	verificationCode, err := s.dal.VerificationCode(rd.Token)
	if api.IsErrNotFound(err) {
		return nil, grpcErrorf(codes.NotFound, "No verification code maps to the provided token %q", rd.Token)
	} else if err != nil {
		return nil, grpcIErrorf(err.Error())
	} else if verificationCode.Expires.Before(s.clk.Now()) {
		return nil, grpcErrorf(auth.VerificationCodeExpired, "The code mapped to the provided token has expired.")
	} else if verificationCode.Code != rd.Code {
		return nil, grpcErrorf(auth.BadVerificationCode, "The provided code %q does not match", rd.Code)
	}

	hashedPassword, err := s.hasher.GenerateFromPassword([]byte(rd.NewPassword))
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	accountID, err := dal.ParseAccountID(verificationCode.VerifiedValue)
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	_, err = s.dal.UpdateAccount(accountID, &dal.AccountUpdate{
		Password: &hashedPassword,
	})
	if err != nil {
		return nil, grpcIErrorf(err.Error())
	}

	// Since we sucessfully checked the token, mark it as consumed
	_, err = s.dal.UpdateVerificationCode(rd.Token, &dal.VerificationCodeUpdate{
		Consumed: ptr.Bool(true),
	})
	if err != nil {
		golog.Errorf("Error while marking password reset token as consumed: %s", err)
	}

	// Delete any active auth tokens for the account
	_, err = s.dal.DeleteAuthTokens(accountID)
	if err != nil {
		golog.Errorf("Error while deleting existing auth tokens for password reset of account %s: %s", accountID, err)
	}

	return &auth.UpdatePasswordResponse{}, nil
}

func (s *server) VerifiedValue(ctx context.Context, rd *auth.VerifiedValueRequest) (*auth.VerifiedValueResponse, error) {
	if rd.Token == "" {
		return nil, grpcErrorf(codes.NotFound, "No verification code maps to the provided token %q", rd.Token)
	}

	verificationCode, err := s.dal.VerificationCode(rd.Token)
	if api.IsErrNotFound(err) {
		return nil, grpcErrorf(codes.NotFound, "No verification code maps to the provided token %q", rd.Token)
	} else if err != nil {
		return nil, grpcIErrorf(err.Error())
	} else if !verificationCode.Consumed {
		return nil, grpcErrorf(auth.ValueNotYetVerified, "The value mapped to this token %q has not yet been verified", rd.Token)
	}

	return &auth.VerifiedValueResponse{
		Value: verificationCode.VerifiedValue,
	}, nil
}

// generateAndInsertToken generates and inserts an auth token for the provided account and information
func (s *server) generateAndInsertToken(dl dal.DAL, accountID dal.AccountID, tokenAttributes map[string]string) (*auth.AuthToken, error) {
	token, err := s.tokenGenerator.GenerateToken()
	if err != nil {
		return nil, errors.Trace(err)
	}
	tokenWithAttributes, err := appendAttributes(token, tokenAttributes)
	if err != nil {
		return nil, errors.Trace(err)
	}
	tokenExpiration := s.clk.Now().Add(defaultTokenExpiration)

	// Utilize the auth token to generate a client encryption key
	key, err := s.clientEncryptionKeySigner.Sign([]byte(token))
	if err != nil {
		return nil, errors.Trace(err)
	}

	if err := dl.InsertAuthToken(&dal.AuthToken{
		AccountID:           accountID,
		Expires:             tokenExpiration,
		Token:               []byte(tokenWithAttributes),
		ClientEncryptionKey: key,
	}); err != nil {
		return nil, errors.Trace(err)
	}
	return &auth.AuthToken{
		Value:               token,
		ExpirationEpoch:     uint64(tokenExpiration.Unix()),
		ClientEncryptionKey: base64.StdEncoding.EncodeToString(key),
	}, nil
}

func (s *server) rotateAndExtendToken(dl dal.DAL, authToken *dal.AuthToken, tokenAttributes map[string]string) (*auth.AuthToken, error) {
	token, err := s.tokenGenerator.GenerateToken()
	if err != nil {
		return nil, errors.Trace(err)
	}
	tokenWithAttributes, err := appendAttributes(token, tokenAttributes)
	if err != nil {
		return nil, errors.Trace(err)
	}
	tokenExpiration := s.clk.Now().Add(defaultTokenExpiration)

	if err := dl.Transact(func(dl dal.DAL) error {
		// Update our existing token to preserve the Created information that we rely on in other parts of the system
		if _, err := dl.UpdateAuthToken(string(authToken.Token), &dal.AuthTokenUpdate{
			Expires: ptr.Time(s.clk.Now().Add(defaultTokenExpiration)),
			Token:   []byte(tokenWithAttributes),
		}); err != nil {
			return errors.Trace(err)
		}

		// Insert a shadow token so that in flight requests will continue to work. This token will expire in 5 minutes
		return errors.Trace(dl.InsertAuthToken(&dal.AuthToken{
			AccountID:           authToken.AccountID,
			Expires:             s.clk.Now().Add(time.Minute * 5),
			Token:               authToken.Token,
			ClientEncryptionKey: authToken.ClientEncryptionKey,
			Shadow:              true,
		}))
	}); err != nil {
		return nil, errors.Trace(err)
	}
	return &auth.AuthToken{
		Value:               token,
		ExpirationEpoch:     uint64(tokenExpiration.Unix()),
		ClientEncryptionKey: base64.StdEncoding.EncodeToString(authToken.ClientEncryptionKey),
	}, nil
}

const (
	maxTokenSize           = 250
	defaultTokenExpiration = time.Second * 60 * 60 * 24 * 4 // A token by default expires in 4 days
)

func appendAttributes(token string, tokenAttributes map[string]string) (string, error) {
	if len(tokenAttributes) > 0 {
		token += ":"
		// due to the non deterministic nature of maps we need to sort the keys and always apply in that order
		// note: to get around this for optimization purposes we could require the caller to provide a list instead
		var i int
		keys := make([]string, len(tokenAttributes))
		for k := range tokenAttributes {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		for _, k := range keys {
			token += (k + tokenAttributes[k])
		}
		if len(token) >= maxTokenSize {
			return "", errors.Trace(fmt.Errorf("Provided client data makes token too long..."))
		}
	}
	return token, nil
}

const (
	verificationCodeDigitCount        = 6
	verificationCodeMaxValue          = 999999
	defaultVerificationCodeExpiration = 15 * time.Minute
)

func generateAndInsertVerificationCode(dl dal.DAL, valueToVerify string, codeType auth.VerificationCodeType, clk clock.Clock) (*auth.VerificationCode, error) {
	token, err := common.GenerateToken()
	if err != nil {
		return nil, errors.Trace(err)
	}
	code, err := common.GenerateRandomNumber(verificationCodeMaxValue, verificationCodeDigitCount)
	if err != nil {
		return nil, errors.Trace(err)
	}
	tokenExpiration := clk.Now().Add(defaultVerificationCodeExpiration)

	vType, err := dal.ParseVerificationCodeType(auth.VerificationCodeType_name[int32(codeType)])
	if err != nil {
		return nil, errors.Trace(err)
	}

	// TODO: Remove logging of the code perhaps?
	golog.Debugf("Inserting verification code %s - with token %s - for value %s - expires %+v.", token, valueToVerify, tokenExpiration)
	if err := dl.InsertVerificationCode(&dal.VerificationCode{
		Token:            token,
		Code:             code,
		Expires:          tokenExpiration,
		VerificationType: vType,
		VerifiedValue:    valueToVerify,
	}); err != nil {
		return nil, errors.Trace(err)
	}
	return &auth.VerificationCode{
		Token:           token,
		Type:            codeType,
		Code:            code,
		ExpirationEpoch: uint64(tokenExpiration.Unix()),
	}, nil
}
