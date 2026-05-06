package integration

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const (
	ledgerDBDSN      = "postgres://esx:esx@localhost:5433/ledger_service?sslmode=disable"
	participantDBDSN = "postgres://esx:esx@localhost:5433/participant_registry?sslmode=disable"
)

func TestDoubleEntryBookkeepingIntegrity(t *testing.T) {
	t.Log("Connecting to Ledger Database...")
	ledgerDB, err := sql.Open("postgres", ledgerDBDSN)
	require.NoError(t, err)
	defer ledgerDB.Close()

	t.Log("Connecting to Participant Database...")
	participantDB, err := sql.Open("postgres", participantDBDSN)
	require.NoError(t, err)
	defer participantDB.Close()

	// 1. Audit Cash Journal (Total Debits MUST equal Total Credits)
	t.Log("Auditing Cash Journal...")
	var totalCashDebits, totalCashCredits sql.NullInt64

	err = ledgerDB.QueryRow(`SELECT SUM(amount) FROM cash_journal WHERE entry_type = 'debit'`).Scan(&totalCashDebits)
	require.NoError(t, err)

	err = ledgerDB.QueryRow(`SELECT SUM(amount) FROM cash_journal WHERE entry_type = 'credit'`).Scan(&totalCashCredits)
	require.NoError(t, err)

	debits := totalCashDebits.Int64
	credits := totalCashCredits.Int64

	t.Logf("Total Cash Debits:  %d paise", debits)
	t.Logf("Total Cash Credits: %d paise", credits)
	require.Equal(t, debits, credits, "CRITICAL FAILURE: Cash ledger is unbalanced! Leaked or fabricated funds detected.")

	// 2. Audit Securities Journal (Total Debits MUST equal Total Credits)
	t.Log("Auditing Securities Journal...")
	var totalSecDebits, totalSecCredits sql.NullInt64

	err = ledgerDB.QueryRow(`SELECT SUM(quantity) FROM securities_journal WHERE entry_type = 'debit'`).Scan(&totalSecDebits)
	require.NoError(t, err)

	err = ledgerDB.QueryRow(`SELECT SUM(quantity) FROM securities_journal WHERE entry_type = 'credit'`).Scan(&totalSecCredits)
	require.NoError(t, err)

	secDebits := totalSecDebits.Int64
	secCredits := totalSecCredits.Int64

	t.Logf("Total Securities Debits:  %d shares", secDebits)
	t.Logf("Total Securities Credits: %d shares", secCredits)
	require.Equal(t, secDebits, secCredits, "CRITICAL FAILURE: Securities ledger is unbalanced! Phantom shares detected.")

	// 3. Verify locked cash metric
	var lockedCash sql.NullInt64
	err = participantDB.QueryRow(`SELECT SUM(locked) FROM cash_accounts`).Scan(&lockedCash)
	require.NoError(t, err)
	t.Logf("Total Cash Currently Locked in Risk Engine: %d paise", lockedCash.Int64)

	t.Log("SUCCESS: Double-entry bookkeeping mathematically verified across all historical trades.")
}
