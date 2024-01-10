package password

import (
	"testing"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/randstring"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := randstring.RandString(6)

	hashedPassword1, err := HashPassword(password)
	if err != nil {
		t.Error(err)
	}
	if len(hashedPassword1) == 0 {
		t.Error("Got empty hash")
	}

	err = CheckPassword(password, hashedPassword1)
	if err != nil {
		t.Error(err)
	}

	wrongPassword := randstring.RandString(6)
	err = CheckPassword(wrongPassword, hashedPassword1)
	if err != bcrypt.ErrMismatchedHashAndPassword {
		t.Error("Matched with wrong password")
	}

	hashedPassword2, err := HashPassword(password)
	if err != nil {
		t.Error(err)
	}
	if len(hashedPassword2) == 0 {
		t.Error("Got empty hash")
	}
	if hashedPassword1 == hashedPassword2 {
		t.Error("Got equal hash for different password")
	}

	emptyPassword, err := HashPassword("-1")
	if err != nil {
		t.Error(err)
	}
	if len(emptyPassword) != 0 {
		t.Error("Got hash for empty password", emptyPassword)
	}
}
