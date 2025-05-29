package user

import "regexp"

type User struct {
	Username     string   `json:"username"`
	Password     string   `json:"password,omitempty"`
	PasswordHash string   `json:"passwordHash,omitempty"`
	Email        string   `json:"email,omitempty"`
	RealName     string   `json:"realName,omitempty"`
	Roles        []string `json:"roles"`
}

func (u User) WithoutSecrets() User {
	u.Password = ""
	u.PasswordHash = ""
	return u
}

var UsernamePattern = regexp.MustCompile("^[-a-zA-Z0-9._@+]+$")

// https://learn.microsoft.com/en-us/windows/win32/fileio/naming-a-file
// Invalid filename characters on Windows:
// < (less than)
// > (greater than)
// : (colon)
// " (double quote)
// / (forward slash)
// \ (backslash)
// | (vertical bar or pipe)
// ? (question mark)
// * (asterisk)
// 0x00-0x1f

// Invalid Windows file names (with or without an extension):
// CON, PRN, AUX, NUL,
// COM1, COM2, COM3, COM4, COM5, COM6, COM7, COM8, COM9, COM¹, COM², COM³,
// LPT1, LPT2, LPT3, LPT4, LPT5, LPT6, LPT7, LPT8, LPT9, LPT¹, LPT², and LPT³
