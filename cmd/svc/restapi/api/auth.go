package api

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/device"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/dbutil"
	"github.com/sprucehealth/backend/libs/golog"
)

// Authentication errors
var (
	ErrInvalidPassword    = errors.New("api: invalid password")
	ErrInvalidRoleType    = errors.New("api: invalid role type")
	ErrLoginAlreadyExists = errors.New("api: login already exists")
	ErrLoginDoesNotExist  = errors.New("api: login does not exist")
	ErrTokenDoesNotExist  = errors.New("api: token does not exist")
	ErrTokenExpired       = errors.New("api: token expired")
)

type authTokenConfig struct {
	expireDuration time.Duration
	renewDuration  time.Duration // When validating, if the time left on the token is less than this duration than the token is extended
}

type auth struct {
	regularAuth  *authTokenConfig
	extendedAuth *authTokenConfig
	db           *sql.DB
	hasher       PasswordHasher
	perms        map[int64]string
	permNames    map[string]int64
}

func normalizeEmail(email string) string {
	return strings.ToLower(email)
}

func NewAuthAPI(db *sql.DB, expireDuration, renewDuration, extendedAuthExpireDuration, extendedAuthRenewDuration time.Duration, hasher PasswordHasher) (AuthAPI, error) {
	ap := &auth{
		db: db,
		regularAuth: &authTokenConfig{
			renewDuration:  renewDuration,
			expireDuration: expireDuration,
		},
		extendedAuth: &authTokenConfig{
			renewDuration:  extendedAuthRenewDuration,
			expireDuration: extendedAuthExpireDuration,
		},
		hasher: hasher,
	}
	var err error
	ap.perms, err = ap.availableAccountPermissions()
	if err != nil {
		return nil, err
	}
	ap.permNames = make(map[string]int64, len(ap.perms))
	for id, name := range ap.perms {
		ap.permNames[name] = id
	}
	return ap, nil
}

func (m *auth) CreateAccount(email, password, roleType string) (int64, error) {
	email = normalizeEmail(email)

	// ensure to check that the email does not already exist in the database
	var id int64
	if err := m.db.QueryRow("SELECT id FROM account WHERE email = ?", email).Scan(&id); err == nil {
		return 0, ErrLoginAlreadyExists
	} else if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	// Hash the password. If one was not provided then use an empty hash which makes
	// it impossible to authenticate the account.
	var hashedPassword string
	if password != "" {
		hash, err := m.hasher.GenerateFromPassword([]byte(password))
		if err != nil {
			return 0, err
		}
		hashedPassword = string(hash)
	}

	var roleTypeID int64
	if err := m.db.QueryRow("SELECT id FROM role_type WHERE role_type_tag = ?", roleType).Scan(&roleTypeID); err == sql.ErrNoRows {
		return 0, ErrInvalidRoleType
	}

	// begin transaction to create an account
	tx, err := m.db.Begin()
	if err != nil {
		return 0, err
	}

	// create a new account since the user does not exist on the platform
	res, err := tx.Exec("INSERT INTO account (email, password, role_type_id) VALUES (?, ?, ?)", email, hashedPassword, roleTypeID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return lastID, tx.Commit()
}

func (m *auth) UpdateAccount(accountID int64, email *string, twoFactorEnabled *bool) error {
	args := dbutil.MySQLVarArgs()
	if email != nil {
		args.Append("email", *email)
	}
	if twoFactorEnabled != nil {
		args.Append("two_factor_enabled", *twoFactorEnabled)
	}
	if args.IsEmpty() {
		return nil
	}
	_, err := m.db.Exec(`UPDATE account SET `+args.ColumnsForUpdate()+` WHERE id = ?`, append(args.Values(), accountID)...)
	return err
}

func (m *auth) Authenticate(email, password string) (*common.Account, error) {
	email = normalizeEmail(email)

	var account common.Account
	var hashedPassword string

	// use the email address to lookup the Account from the table
	if err := m.db.QueryRow(`
		SELECT account.id, role_type_tag, password, email, registration_date, two_factor_enabled
		FROM account
		INNER JOIN role_type ON role_type_id = role_type.id
		WHERE email = ?`, email,
	).Scan(
		&account.ID, &account.Role, &hashedPassword, &account.Email,
		&account.Registered, &account.TwoFactorEnabled,
	); err == sql.ErrNoRows {
		return nil, ErrLoginDoesNotExist
	} else if err != nil {
		return nil, err
	}

	// compare the hashed password value to that stored in the database to authenticate the user
	if err := m.hasher.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return &account, nil
}

type CreateTokenOption int

const (
	CreateTokenExtended CreateTokenOption = 1 << iota
	CreateTokenAllowMany
)

func (cto CreateTokenOption) has(opt CreateTokenOption) bool {
	return (cto & opt) != 0
}

func (m *auth) CreateToken(accountID int64, platform Platform, options CreateTokenOption) (string, error) {
	token, err := common.GenerateToken()
	if err != nil {
		return "", err
	}

	// delete any existing token and create a new one
	tx, err := m.db.Begin()
	if err != nil {
		return "", err
	}

	if !options.has(CreateTokenAllowMany) {
		// delete the token that exists (if one exists)
		_, err = tx.Exec("DELETE FROM auth_token WHERE account_id = ? AND platform = ?", accountID, string(platform))
		if err != nil {
			tx.Rollback()
			return "", err
		}
	}

	// insert new token
	now := time.Now().UTC()
	expires := now.Add(m.regularAuth.expireDuration)
	if options.has(CreateTokenExtended) {
		expires = now.Add(m.extendedAuth.expireDuration)
	}

	_, err = tx.Exec("INSERT INTO auth_token (token, account_id, platform, created, expires, extended) VALUES (?, ?, ?, ?, ?, ?)",
		token, accountID, string(platform), now, expires, options.has(CreateTokenExtended))
	if err != nil {
		tx.Rollback()
		return "", err
	}

	return token, tx.Commit()
}

func (m *auth) DeleteToken(token string) error {
	_, err := m.db.Exec("DELETE FROM auth_token WHERE token = ?", token)
	return err
}

func (m *auth) ValidateToken(token string, platform Platform) (*common.Account, error) {
	var account common.Account
	var expires time.Time
	var extended bool
	var tokenPlatform string
	if err := m.db.QueryRow(`
		SELECT account_id, role_type_tag, expires, email, registration_date, two_factor_enabled, platform, extended
		FROM auth_token
		INNER JOIN account ON account.id = account_id
		INNER JOIN role_type ON role_type_id = role_type.id
		WHERE token = ?`, token,
	).Scan(&account.ID, &account.Role, &expires, &account.Email, &account.Registered,
		&account.TwoFactorEnabled, &tokenPlatform, &extended); err == sql.ErrNoRows {
		return nil, ErrTokenDoesNotExist
	} else if err != nil {
		return nil, err
	}

	if tokenPlatform != string(platform) {
		golog.Warningf("Platform does not match while validating token (expected %s got %+v)", tokenPlatform, platform)
		return nil, ErrTokenDoesNotExist
	}

	// Check the expiration to ensure that it is valid
	now := time.Now().UTC()
	left := expires.Sub(now)
	if left <= 0 {
		return nil, ErrTokenExpired
	}
	// Extend token if necessary
	authConfig := m.regularAuth
	if extended {
		authConfig = m.extendedAuth
	}
	if authConfig.renewDuration > 0 && left < authConfig.renewDuration {
		if _, err := m.db.Exec("UPDATE auth_token SET expires = ? WHERE token = ?", now.Add(authConfig.expireDuration), token); err != nil {
			golog.Errorf("services/auth: failed to extend token expiration: %s", err.Error())
			// Don't return an error response because this doesn't prevent anything else from working
		}
	}

	return &account, nil
}

func (m *auth) GetToken(accountID int64) (string, error) {
	var token string
	err := m.db.QueryRow(`select token from auth_token where account_id = ?`, accountID).Scan(&token)
	if err == sql.ErrNoRows {
		return "", ErrNotFound("auth_token")
	} else if err != nil {
		return "", err
	}

	return token, err
}

func (m *auth) SetPassword(accountID int64, password string) error {
	if password == "" {
		return ErrInvalidPassword
	}
	hashedPassword, err := m.hasher.GenerateFromPassword([]byte(password))
	if err != nil {
		return err
	}
	if res, err := m.db.Exec("UPDATE account SET password = ? WHERE id = ?", string(hashedPassword), accountID); err != nil {
		return err
	} else if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrNotFound("account")
	}
	// Log out any existing tokens for the account
	if _, err := m.db.Exec("DELETE FROM auth_token WHERE account_id = ?", accountID); err != nil {
		return err
	}
	return nil
}

func (m *auth) UpdateLastOpenedDate(accountID int64) error {
	if res, err := m.db.Exec(`update account set last_opened_date = now(6) where id = ?`, accountID); err != nil {
		return err
	} else if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrNotFound("account")
	}
	return nil
}

func (m *auth) GetPhoneNumbersForAccount(accountID int64) ([]*common.PhoneNumber, error) {
	rows, err := m.db.Query(`
		SELECT phone, phone_type, status, verified
		FROM account_phone
		WHERE account_id = ? AND status = ?`, accountID, StatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var numbers []*common.PhoneNumber
	for rows.Next() {
		num := &common.PhoneNumber{}
		if err := rows.Scan(&num.Phone, &num.Type, &num.Status, &num.Verified); err != nil {
			return nil, err
		}
		numbers = append(numbers, num)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return numbers, nil
}

func (m *auth) UpdateAppDevice(accountID int64, appVersion *encoding.Version, p device.Platform, platformVersion, device, deviceModel, build string) error {
	if appVersion == nil {
		return nil
	}

	_, err := m.db.Exec(`
		REPLACE INTO account_app_version (account_id, major, minor, patch, platform, platform_version, device, device_model, build)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, accountID, appVersion.Major, appVersion.Minor, appVersion.Patch,
		p.String(), platformVersion, device, deviceModel, build)
	return err
}

func (m *auth) LatestAppInfo(accountID int64) (*AppInfo, error) {
	var aInfo AppInfo
	var major, minor, patch int
	err := m.db.QueryRow(`
		SELECT major, minor, patch, build, platform, platform_version,
		device, device_model, last_modified_date
		FROM account_app_version
		WHERE account_id = ?
		ORDER BY last_modified_date DESC LIMIT 1`, accountID).Scan(
		&major,
		&minor,
		&patch,
		&aInfo.Build,
		&aInfo.Platform,
		&aInfo.PlatformVersion,
		&aInfo.Device,
		&aInfo.DeviceModel,
		&aInfo.LastSeen)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound("account_app_version")
	}

	aInfo.Version = &encoding.Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}

	return &aInfo, err
}

func (m *auth) ReplacePhoneNumbersForAccount(accountID int64, numbers []*common.PhoneNumber) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM account_phone WHERE account_id = ?`, accountID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, n := range numbers {
		_, err = tx.Exec(`
			INSERT INTO account_phone (account_id, phone, phone_type, status, verified)
			VALUES (?, ?, ?, ?, ?)`, accountID, n.Phone.String(), n.Type.String(), n.Status, n.Verified)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (m *auth) AccountForEmail(email string) (*common.Account, error) {
	email = normalizeEmail(email)
	var account common.Account
	if err := m.db.QueryRow(`
		SELECT account.id, role_type_tag, email, registration_date, two_factor_enabled, account_code
		FROM account
		INNER JOIN role_type ON role_type_id = role_type.id
		WHERE email = ?`, email,
	).Scan(
		&account.ID, &account.Role, &account.Email, &account.Registered, &account.TwoFactorEnabled, &account.AccountCode,
	); err == sql.ErrNoRows {
		return nil, ErrLoginDoesNotExist
	} else if err != nil {
		return nil, err
	}
	return &account, nil
}

func (m *auth) GetAccount(id int64) (*common.Account, error) {
	account := &common.Account{
		ID: id,
	}
	if err := m.db.QueryRow(`
		SELECT account.id, role_type_tag, email, registration_date, two_factor_enabled, account_code
		FROM account
		INNER JOIN role_type ON role_type_id = role_type.id
		WHERE account.id = ?`, id,
	).Scan(
		&account.ID, &account.Role, &account.Email, &account.Registered, &account.TwoFactorEnabled, &account.AccountCode,
	); err == sql.ErrNoRows {
		return nil, ErrNotFound("account")
	} else if err != nil {
		return nil, err
	}
	return account, nil
}

func (m *auth) CreateTempToken(accountID int64, expireSec int, purpose AuthTokenPurpose, token string) (string, error) {
	if token == "" {
		var err error
		token, err = common.GenerateToken()
		if err != nil {
			return "", err
		}
	}
	expires := time.Now().Add(time.Duration(expireSec) * time.Second)
	_, err := m.db.Exec(`INSERT INTO temp_auth_token (token, purpose, account_id, expires) VALUES (?, ?, ?, ?)`,
		token, purpose.String(), accountID, expires)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (m *auth) ValidateTempToken(purpose AuthTokenPurpose, token string) (*common.Account, error) {
	row := m.db.QueryRow(`
		SELECT expires, account_id, role_type_tag, email, registration_date, two_factor_enabled
		FROM temp_auth_token
		LEFT JOIN account ON account.id = account_id
		LEFT JOIN role_type ON role_type.id = account.role_type_id
		WHERE purpose = ? AND token = ?`, purpose.String(), token)
	var expires time.Time
	var account common.Account
	if err := row.Scan(
		&expires, &account.ID, &account.Role, &account.Email, &account.Registered, &account.TwoFactorEnabled,
	); err == sql.ErrNoRows {
		return nil, ErrTokenDoesNotExist
	} else if err != nil {
		return nil, err
	}
	if time.Now().After(expires) {
		return nil, ErrTokenExpired
	}
	return &account, nil
}

func (m *auth) DeleteTempToken(purpose AuthTokenPurpose, token string) error {
	_, err := m.db.Exec(`DELETE FROM temp_auth_token WHERE token = ? AND purpose = ?`, token, purpose.String())
	return err
}

func (m *auth) DeleteTempTokensForAccount(accountID int64) error {
	_, err := m.db.Exec(`DELETE FROM temp_auth_token WHERE account_id = ?`, accountID)
	return err
}

func (m *auth) GetAccountDevice(accountID int64, deviceID string) (*common.AccountDevice, error) {
	if deviceID == "" {
		return nil, errors.New("no device ID provided")
	}

	row := m.db.QueryRow(`
		SELECT verified, verified_tstamp, created
		FROM account_device
		WHERE account_id = ? AND device_id = ?`, accountID, deviceID)

	device := &common.AccountDevice{
		AccountID: accountID,
		DeviceID:  deviceID,
	}
	if err := row.Scan(&device.Verified, &device.VerifiedTime, &device.Created); err == sql.ErrNoRows {
		return nil, ErrNotFound("account_device")
	} else if err != nil {
		return nil, err
	}
	return device, nil
}

func (m *auth) TimezoneForAccount(id int64) (string, error) {
	var name string
	err := m.db.QueryRow(`SELECT tz_name FROM account_timezone WHERE account_id = ?`, id).Scan(&name)
	if err == sql.ErrNoRows {
		return "", ErrNotFound("account_timezone")
	}
	return name, err
}

func (m *auth) UpdateAccountDeviceVerification(accountID int64, deviceID string, verified bool) error {
	if deviceID == "" {
		return errors.New("no device ID provided")
	}

	var tstamp *time.Time
	if verified {
		now := time.Now()
		tstamp = &now
	}
	_, err := m.db.Exec(`
		INSERT INTO account_device (account_id, device_id, verified, verified_tstamp)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE verified = ?, verified_tstamp = ?`,
		accountID, deviceID, verified, tstamp, verified, tstamp)
	return err
}