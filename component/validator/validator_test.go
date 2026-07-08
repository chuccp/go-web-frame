package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestValidator(t *testing.T) *Validator {
	t.Helper()
	v := &Validator{}
	err := v.Init(nil)
	assert.NoError(t, err)
	return v
}

type mobileStruct struct {
	Phone string `validate:"mobile"`
}

type passwordStruct struct {
	Password string `validate:"password"`
}

type requiredStruct struct {
	Name string `validate:"required"`
}

func TestValidator_Init(t *testing.T) {
	v := &Validator{}
	err := v.Init(nil)
	assert.NoError(t, err)
	assert.NotNil(t, v.validate)
}

func TestValidator_Mobile_Valid(t *testing.T) {
	v := newTestValidator(t)

	validNumbers := []string{
		"13812345678",
		"15912345678",
		"18612345678",
		"+8613812345678",
		"138 1234 5678",
		"138-1234-5678",
	}
	for _, phone := range validNumbers {
		err := v.Validate(&mobileStruct{Phone: phone})
		assert.NoError(t, err, "expected valid for: %s", phone)
	}
}

func TestValidator_Mobile_Invalid(t *testing.T) {
	v := newTestValidator(t)

	invalidNumbers := []string{
		"123456",
		"abc12345678",
		"23812345678", // doesn't start with 1
		"",
		"1381234567",   // 10 digits, not 11
		"138123456789", // 12 digits
	}
	for _, phone := range invalidNumbers {
		err := v.Validate(&mobileStruct{Phone: phone})
		assert.Error(t, err, "expected invalid for: %s", phone)
	}
}

func TestValidator_Password_Valid(t *testing.T) {
	v := newTestValidator(t)

	validPasswords := []string{
		"Abcdef1x",
		"Passw0rd",
		"Hello123World",
		"  Abcdef1x  ", // trimmed
	}
	for _, pw := range validPasswords {
		err := v.Validate(&passwordStruct{Password: pw})
		assert.NoError(t, err, "expected valid for: %s", pw)
	}
}

func TestValidator_Password_Invalid(t *testing.T) {
	v := newTestValidator(t)

	invalidPasswords := []string{
		"short1",        // too short
		"alllowercase1", // no upper
		"ALLUPPERCASE1", // no lower
		"NoDigitsHere",  // no digit
		"Abcdefg",       // no digit
	}
	for _, pw := range invalidPasswords {
		err := v.Validate(&passwordStruct{Password: pw})
		assert.Error(t, err, "expected invalid for: %s", pw)
	}
}

func TestValidator_Required(t *testing.T) {
	v := newTestValidator(t)

	err := v.Validate(&requiredStruct{Name: "hello"})
	assert.NoError(t, err)

	err = v.Validate(&requiredStruct{Name: ""})
	assert.Error(t, err)
}

func TestValidator_ValidStruct(t *testing.T) {
	v := newTestValidator(t)

	err := v.Validate(&requiredStruct{Name: "ok"})
	assert.NoError(t, err)
}

func TestValidator_NonStruct(t *testing.T) {
	v := newTestValidator(t)

	// Passing a non-struct should return an error from go-playground/validator
	err := v.Validate("just a string")
	assert.Error(t, err)
}
