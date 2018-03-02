package Trixie

import "errors"

var (
	E_LOGIN_FAILED            = errors.New("login failed")
	E_RENEW_FAILED            = errors.New("renewal failed")
	E_USERNAME_PASSWORD_BLANK = errors.New("username or password blank")
	E_HTTP_REQUEST_FAILED     = errors.New("HTTP request failed")
	E_JSON_MARSHAL            = errors.New("cannot (un)marshal JSON")
	E_FILE_RW                 = errors.New("cannot read/write file")
	E_UNKNOWN                 = errors.New("unknown command")
)
