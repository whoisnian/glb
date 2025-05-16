package strutil_test

import (
	"bytes"
	"testing"

	"github.com/whoisnian/glb/util/strutil"
)

func TestShellEscape(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{``, `''`},
		{`/tmp/dir 1`, `'/tmp/dir 1'`},
		{`"/tmp/dir 1"`, `'"/tmp/dir 1"'`},
		{`'/tmp/dir 1'`, `''"'"'/tmp/dir 1'"'"''`},
		{`~/doc`, `'~/doc'`},
	}
	for _, test := range tests {
		if got := strutil.ShellEscape(test.input); got != test.want {
			t.Errorf("ShellEscape(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestShellEscapeExceptTilde(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{``, `''`},
		{`/tmp/dir 1`, `'/tmp/dir 1'`},
		{`"/tmp/dir 1"`, `'"/tmp/dir 1"'`},
		{`'/tmp/dir 1'`, `''"'"'/tmp/dir 1'"'"''`},
		{`~/doc`, `~/'doc'`},
	}
	for _, test := range tests {
		if got := strutil.ShellEscapeExceptTilde(test.input); got != test.want {
			t.Errorf("ShellEscapeExceptTilde(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestIsDigitString(t *testing.T) {
	var tests = []struct {
		input string
		want  bool
	}{
		{"", false},
		{"-1", false},
		{"0.1", false},
		{"1e5", false},
		{"0", true},
		{"001", true},
		{"12345", true},
	}
	for _, test := range tests {
		if got := strutil.IsDigitString(test.input); got != test.want {
			t.Errorf("IsDigitString(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}

func TestCamelize(t *testing.T) {
	var tests = []struct {
		input string
		upper bool
		want  string
	}{
		{"", true, ""},
		{",.?|", true, ""},
		{"user_id", true, "UserId"},
		{"user_id", false, "userId"},
		{"user.name", true, "UserName"},
		{"UserName", true, "Username"},
		{"UserName", false, "username"},
		{"_csrf_token", true, "CsrfToken"},
		{"_csrf_token", false, "csrfToken"},
		{"::Net::HTTP", true, "NetHttp"},
		{"::Net::HTTP", false, "netHttp"},
		{"accept-encoding", true, "AcceptEncoding"},
		{"SHA256 hash", true, "Sha256Hash"},
		{"s3 access key", true, "S3AccessKey"},
		{"9P protocol", true, "9PProtocol"},
		{"Regular 4G", true, "Regular4G"},
		{"Mysql DSN", true, "MysqlDsn"},
		{"hex2bin", true, "Hex2bin"},
		{"/var/log/nginx/", true, "VarLogNginx"},
		{"SCREAMING_SNAKE_CASE", true, "ScreamingSnakeCase"},
		{"Test with + sign", true, "TestWithSign"},
		{"!@#$bad %^&* characters()-=", true, "BadCharacters"},
	}
	for _, test := range tests {
		if got := strutil.Camelize(test.input, test.upper); got != test.want {
			t.Errorf("Camelize(%q, %t) = %q, want %q", test.input, test.upper, got, test.want)
		}
	}
}

func TestUnderscore(t *testing.T) {
	var tests = []struct {
		input string
		upper bool
		want  string
	}{
		{"", false, ""},
		{",.?|", false, ""},
		{"userId", false, "user_id"},
		{"UserID", false, "user_id"},
		{"UserID", true, "USER_ID"},
		{"user.name", false, "user_name"},
		{"user_name", false, "user_name"},
		{"UserName", false, "user_name"},
		{"UserName", true, "USER_NAME"},
		{"_csrf_token", false, "csrf_token"},
		{"_csrf_token", true, "CSRF_TOKEN"},
		{"::Net::HTTP", false, "net_http"},
		{"::Net::HTTP", true, "NET_HTTP"},
		{"Accept-Encoding", false, "accept_encoding"},
		{"SHA256 hash", false, "sha256_hash"},
		{"S3AccessKey", false, "s3_access_key"},
		{"9PProtocol", false, "9p_protocol"},
		{"Regular 4G", false, "regular_4g"},
		{"MysqlDSN", false, "mysql_dsn"},
		{"Hex2bin", false, "hex2bin"},
		{"/var/log/nginx/", false, "var_log_nginx"},
		{"SCREAMING_SNAKE_CASE", false, "screaming_snake_case"},
		{"Test with + sign", false, "test_with_sign"},
		{"!@#$bad %^&* characters()-=", false, "bad_characters"},
	}
	for _, test := range tests {
		if got := strutil.Underscore(test.input, test.upper); got != test.want {
			t.Errorf("Underscore(%q, %t) = %q, want %q", test.input, test.upper, got, test.want)
		}
	}
}

func TestUnsafeStringToBytes(t *testing.T) {
	var inputs = []string{
		"",
		" ",
		"	",
		"hello, world",
		"\a\b\\\\t\n\r\"'",
		"\x00\x01\x02\x03\x04",
	}
	for _, input := range inputs {
		if got := strutil.UnsafeStringToBytes(input); !bytes.Equal(got, []byte(input)) {
			t.Errorf("UnsafeStringToBytes(%q) = %v, want %v", input, got, []byte(input))
		}
	}
}

func TestUnsafeBytesToString(t *testing.T) {
	var inputs = [][]byte{
		nil,
		{},
		{'	'},
		{'h', 'e', 'l', 'l', 'o'},
		{0, 1, 2, 3, 4},
	}
	for _, input := range inputs {
		if got := strutil.UnsafeBytesToString(input); got != string(input) {
			t.Errorf("UnsafeBytesToString(%q) = %q, want %q", input, got, string(input))
		}
	}
}
