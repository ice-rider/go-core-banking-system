package repository

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryConstant_CreateAccount(t *testing.T) {
	assert.Contains(t, queryCreateAccount, "INSERT INTO accounts")
	assert.Contains(t, queryCreateAccount, "$1")
	assert.Contains(t, queryCreateAccount, "$2")
	assert.Contains(t, queryCreateAccount, "$3")
	assert.Contains(t, queryCreateAccount, "$4")
	assert.Contains(t, queryCreateAccount, "$5")
	assert.Contains(t, queryCreateAccount, "$6")
}

func TestQueryConstant_GetAccountByID(t *testing.T) {
	assert.Contains(t, queryGetAccountByID, "SELECT")
	assert.Contains(t, queryGetAccountByID, "FROM accounts")
	assert.Contains(t, queryGetAccountByID, "$1")
}

func TestQueryConstant_UpdateAccountStatus(t *testing.T) {
	assert.Contains(t, queryUpdateAccountStatus, "UPDATE accounts")
	assert.Contains(t, queryUpdateAccountStatus, "SET status")
	assert.Contains(t, queryUpdateAccountStatus, "$1")
	assert.Contains(t, queryUpdateAccountStatus, "$2")
}

func TestQueryConstant_Reserve(t *testing.T) {
	assert.Contains(t, queryReserve, "UPDATE accounts")
	assert.Contains(t, queryReserve, "balance - $1")
	assert.Contains(t, queryReserve, "$2")
	assert.Contains(t, queryReserve, "balance >= $1")
}

func TestQueryConstant_Credit(t *testing.T) {
	assert.Contains(t, queryCredit, "UPDATE accounts")
	assert.Contains(t, queryCredit, "balance + $1")
	assert.Contains(t, queryCredit, "$2")
}

func TestQueryConstant_Debit(t *testing.T) {
	assert.Contains(t, queryDebit, "UPDATE accounts")
	assert.Contains(t, queryDebit, "balance - $1")
	assert.Contains(t, queryDebit, "$2")
	assert.Contains(t, queryDebit, "balance >= $1")
}

func TestQueryConstant_CommitReservation(t *testing.T) {
	assert.Contains(t, queryCommitReservation, "UPDATE accounts")
	assert.Contains(t, queryCommitReservation, "$1")
}

func TestQueryConstant_CancelReservation(t *testing.T) {
	assert.Contains(t, queryCancelReservation, "UPDATE accounts")
	assert.Contains(t, queryCancelReservation, "balance + $1")
	assert.Contains(t, queryCancelReservation, "$2")
}

func TestQueryConstant_BlockAtomically(t *testing.T) {
	assert.Contains(t, queryBlockAtomically, "UPDATE accounts")
	assert.Contains(t, queryBlockAtomically, "status = 'BLOCKED'")
	assert.Contains(t, queryBlockAtomically, "$1")
}

func TestQueryConstant_UnblockAtomically(t *testing.T) {
	assert.Contains(t, queryUnblockAtomically, "UPDATE accounts")
	assert.Contains(t, queryUnblockAtomically, "status = 'ACTIVE'")
	assert.Contains(t, queryUnblockAtomically, "$1")
}

func TestQueryConstant_CloseAtomically(t *testing.T) {
	assert.Contains(t, queryCloseAtomically, "UPDATE accounts")
	assert.Contains(t, queryCloseAtomically, "status = 'CLOSED'")
	assert.Contains(t, queryCloseAtomically, "$1")
	assert.Contains(t, queryCloseAtomically, "balance = 0")
}

func TestAllQueryConstantsAreDefined(t *testing.T) {
	queries := map[string]string{
		"queryCreateAccount":       queryCreateAccount,
		"queryGetAccountByID":      queryGetAccountByID,
		"queryUpdateAccountStatus": queryUpdateAccountStatus,
		"queryReserve":             queryReserve,
		"queryCredit":              queryCredit,
		"queryDebit":               queryDebit,
		"queryCommitReservation":   queryCommitReservation,
		"queryCancelReservation":   queryCancelReservation,
		"queryBlockAtomically":     queryBlockAtomically,
		"queryUnblockAtomically":   queryUnblockAtomically,
		"queryCloseAtomically":     queryCloseAtomically,
	}

	for name, query := range queries {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, query, "query constant %s should not be empty", name)
		})
	}
}

func TestQueryParameterOrdering(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		maxParam int
	}{
		{"CreateAccount has 6 params", queryCreateAccount, 6},
		{"GetAccountByID has 1 param", queryGetAccountByID, 1},
		{"UpdateAccountStatus has 2 params", queryUpdateAccountStatus, 2},
		{"Reserve has 2 params", queryReserve, 2},
		{"Credit has 2 params", queryCredit, 2},
		{"Debit has 2 params", queryDebit, 2},
		{"CommitReservation has 1 param", queryCommitReservation, 1},
		{"CancelReservation has 2 params", queryCancelReservation, 2},
		{"BlockAtomically has 1 param", queryBlockAtomically, 1},
		{"UnblockAtomically has 1 param", queryUnblockAtomically, 1},
		{"CloseAtomically has 1 param", queryCloseAtomically, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 1; i <= tt.maxParam; i++ {
				assert.Contains(t, tt.query, fmt.Sprintf("$%d", i), "query should contain $%d", i)
			}
			assert.NotContains(t, tt.query, fmt.Sprintf("$%d", tt.maxParam+1), "query should not contain $%d", tt.maxParam+1)
		})
	}
}
