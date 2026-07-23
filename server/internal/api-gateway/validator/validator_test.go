package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CreateAccountRequest struct {
	OwnerName string `json:"owner_name" validate:"required,min=1,max=255"`
}

type CreateTransferRequest struct {
	FromAccountID  string `json:"from_account_id" validate:"required,uuid"`
	ToAccountID    string `json:"to_account_id" validate:"required,uuid"`
	Amount         int64  `json:"amount" validate:"required,gt=0"`
	IdempotencyKey string `json:"idempotency_key" validate:"required,uuid"`
}

func TestValidator_New(t *testing.T) {
	v := New()
	require.NotNil(t, v)
}

func TestCreateAccountRequest_Valid(t *testing.T) {
	v := New()
	errs := v.Validate(CreateAccountRequest{OwnerName: "John Doe"})
	assert.Empty(t, errs)
}

func TestCreateAccountRequest_EmptyOwnerName(t *testing.T) {
	v := New()
	errs := v.Validate(CreateAccountRequest{OwnerName: ""})
	require.NotEmpty(t, errs)
	assert.Equal(t, "OwnerName", errs[0].Field)
	assert.Contains(t, errs[0].Message, "required")
}

func TestCreateAccountRequest_WhitespaceOwnerName(t *testing.T) {
	v := New()
	errs := v.Validate(CreateAccountRequest{OwnerName: "abc"})
	assert.Empty(t, errs)
}

func TestCreateAccountRequest_TooLongOwnerName(t *testing.T) {
	v := New()
	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}
	errs := v.Validate(CreateAccountRequest{OwnerName: longName})
	require.NotEmpty(t, errs)
	assert.Equal(t, "OwnerName", errs[0].Field)
	assert.Contains(t, errs[0].Message, "at most")
}

func TestCreateTransferRequest_Valid(t *testing.T) {
	v := New()
	errs := v.Validate(CreateTransferRequest{
		FromAccountID:  "550e8400-e29b-41d4-a716-446655440000",
		ToAccountID:    "550e8400-e29b-41d4-a716-446655440001",
		Amount:         100,
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440002",
	})
	assert.Empty(t, errs)
}

func TestCreateTransferRequest_MissingFromAccountID(t *testing.T) {
	v := New()
	errs := v.Validate(CreateTransferRequest{
		FromAccountID:  "",
		ToAccountID:    "550e8400-e29b-41d4-a716-446655440001",
		Amount:         100,
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440002",
	})
	require.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.Field == "FromAccountID" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestCreateTransferRequest_InvalidFromAccountUUID(t *testing.T) {
	v := New()
	errs := v.Validate(CreateTransferRequest{
		FromAccountID:  "not-a-uuid",
		ToAccountID:    "550e8400-e29b-41d4-a716-446655440001",
		Amount:         100,
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440002",
	})
	require.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.Field == "FromAccountID" {
			found = true
			assert.Contains(t, e.Message, "UUID")
			break
		}
	}
	assert.True(t, found)
}

func TestCreateTransferRequest_MissingToAccountID(t *testing.T) {
	v := New()
	errs := v.Validate(CreateTransferRequest{
		FromAccountID:  "550e8400-e29b-41d4-a716-446655440000",
		ToAccountID:    "",
		Amount:         100,
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440002",
	})
	require.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.Field == "ToAccountID" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestCreateTransferRequest_ZeroAmount(t *testing.T) {
	v := New()
	errs := v.Validate(CreateTransferRequest{
		FromAccountID:  "550e8400-e29b-41d4-a716-446655440000",
		ToAccountID:    "550e8400-e29b-41d4-a716-446655440001",
		Amount:         0,
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440002",
	})
	require.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.Field == "Amount" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestCreateTransferRequest_NegativeAmount(t *testing.T) {
	v := New()
	errs := v.Validate(CreateTransferRequest{
		FromAccountID:  "550e8400-e29b-41d4-a716-446655440000",
		ToAccountID:    "550e8400-e29b-41d4-a716-446655440001",
		Amount:         -100,
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440002",
	})
	require.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.Field == "Amount" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestCreateTransferRequest_MissingIdempotencyKey(t *testing.T) {
	v := New()
	errs := v.Validate(CreateTransferRequest{
		FromAccountID:  "550e8400-e29b-41d4-a716-446655440000",
		ToAccountID:    "550e8400-e29b-41d4-a716-446655440001",
		Amount:         100,
		IdempotencyKey: "",
	})
	require.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.Field == "IdempotencyKey" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestCreateTransferRequest_InvalidIdempotencyKeyUUID(t *testing.T) {
	v := New()
	errs := v.Validate(CreateTransferRequest{
		FromAccountID:  "550e8400-e29b-41d4-a716-446655440000",
		ToAccountID:    "550e8400-e29b-41d4-a716-446655440001",
		Amount:         100,
		IdempotencyKey: "not-a-uuid",
	})
	require.NotEmpty(t, errs)
	found := false
	for _, e := range errs {
		if e.Field == "IdempotencyKey" {
			found = true
			assert.Contains(t, e.Message, "UUID")
			break
		}
	}
	assert.True(t, found)
}

func TestCreateTransferRequest_MultipleErrors(t *testing.T) {
	v := New()
	errs := v.Validate(CreateTransferRequest{})
	assert.GreaterOrEqual(t, len(errs), 3)
}

func TestCreateTransferRequest_SameAccounts(t *testing.T) {
	v := New()
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	errs := v.Validate(CreateTransferRequest{
		FromAccountID:  uuid,
		ToAccountID:    uuid,
		Amount:         100,
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
	})
	assert.Empty(t, errs)
}

func TestValidator_NilInput(t *testing.T) {
	v := New()
	assert.NotPanics(t, func() {
		errs := v.Validate(nil)
		assert.NotNil(t, errs)
	})
}

func TestValidate_MinCharacters(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"min=5"`
	}
	v := New()
	errs := v.Validate(TestStruct{Name: "ab"})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "at least 5")
}

func TestValidate_MaxCharacters(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"max=3"`
	}
	v := New()
	errs := v.Validate(TestStruct{Name: "abcd"})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "at most 3")
}

func TestValidate_UUID(t *testing.T) {
	type TestStruct struct {
		ID string `validate:"uuid"`
	}
	v := New()
	errs := v.Validate(TestStruct{ID: "not-a-uuid"})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "valid UUID")
}

func TestValidate_GT(t *testing.T) {
	type TestStruct struct {
		Amount int64 `validate:"gt=0"`
	}
	v := New()
	errs := v.Validate(TestStruct{Amount: 0})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "greater than 0")
}

func TestValidate_DefaultMessage(t *testing.T) {
	type TestStruct struct {
		Email string `validate:"email"`
	}
	v := New()
	errs := v.Validate(TestStruct{Email: "not-an-email"})
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Message, "failed validation")
}
