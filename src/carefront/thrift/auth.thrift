namespace go carefront.thriftapi

include "common.thrift"

struct AuthResponse {
	1: required string token
	2: required i64 account_id
}

struct TokenValidationResponse {
	1: required bool is_valid
	2: optional i64 account_id
}

exception NoSuchLogin {
}

exception LoginAlreadyExists {
	1: required i64 account_id
}

service Auth {
	AuthResponse signup(
		1: required string login,
		2: required string password
	) throws (
		1: common.InternalServerError error,
		2: common.AccessDenied access_denied,
		3: common.OverCapacity over_capacity,
		4: LoginAlreadyExists already_exists)

	AuthResponse login(
		1: required string login,
		2: required string password
	) throws (
		1: common.InternalServerError error,
		2: common.AccessDenied access_denied,
		3: common.OverCapacity over_capacity,
		4: NoSuchLogin no_such_login)

	void logout(
		1: required string token,
	) throws (
		1: common.InternalServerError error,
		2: common.AccessDenied access_denied,
		3: common.OverCapacity over_capacity)

	TokenValidationResponse validate_token(
		1: required string token,
	) throws (
		1: common.InternalServerError error,
		2: common.AccessDenied access_denied,
		3: common.OverCapacity over_capacity)
}
