package repository

const (
	queryCreateAccount = `
		INSERT INTO accounts (id, owner_name, balance, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	queryGetAccountByID = `
		SELECT id, owner_name, balance, status, created_at, updated_at
		FROM accounts WHERE id = $1`

	queryUpdateAccountStatus = `
		UPDATE accounts SET status = $1, updated_at = NOW() WHERE id = $2`

	queryReserve = `
		UPDATE accounts SET balance = balance - $1, updated_at = NOW()
		WHERE id = $2 AND balance >= $1 AND status = 'ACTIVE'`

	queryCredit = `
		UPDATE accounts SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2 AND status = 'ACTIVE'`

	queryDebit = `
		UPDATE accounts SET balance = balance - $1, updated_at = NOW()
		WHERE id = $2 AND balance >= $1 AND status = 'ACTIVE'`

	queryCommitReservation = `
		UPDATE accounts SET updated_at = NOW() WHERE id = $1 AND status = 'ACTIVE'`

	queryCancelReservation = `
		UPDATE accounts SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2 AND status = 'ACTIVE'`

	queryBlockAtomically = `
		UPDATE accounts SET status = 'BLOCKED', updated_at = NOW()
		WHERE id = $1 AND status != 'CLOSED'`

	queryUnblockAtomically = `
		UPDATE accounts SET status = 'ACTIVE', updated_at = NOW()
		WHERE id = $1 AND status != 'CLOSED'`

	queryCloseAtomically = `
		UPDATE accounts SET status = 'CLOSED', updated_at = NOW()
		WHERE id = $1 AND balance = 0 AND status != 'CLOSED'`
)
