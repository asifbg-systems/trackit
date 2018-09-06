// Package models contains the types for schema 'trackit'.
package models

// Code generated by xo. DO NOT EDIT.

import (
	"errors"
)

// SharedAccount represents a row from 'trackit.shared_account'.
type SharedAccount struct {
	ID              int  `json:"id"`               // id
	AccountID       int  `json:"account_id"`       // account_id
	UserID          int  `json:"user_id"`          // user_id
	UserPermission  int  `json:"user_permission"`  // user_permission
	SharingAccepted bool `json:"sharing_accepted"` // sharing_accepted

	// xo fields
	_exists, _deleted bool
}

// Exists determines if the SharedAccount exists in the database.
func (sa *SharedAccount) Exists() bool {
	return sa._exists
}

// Deleted provides information if the SharedAccount has been deleted from the database.
func (sa *SharedAccount) Deleted() bool {
	return sa._deleted
}

// Insert inserts the SharedAccount to the database.
func (sa *SharedAccount) Insert(db XODB) error {
	var err error

	// if already exist, bail
	if sa._exists {
		return errors.New("insert failed: already exists")
	}

	// sql insert query, primary key provided by autoincrement
	const sqlstr = `INSERT INTO trackit.shared_account (` +
		`account_id, user_id, user_permission, sharing_accepted` +
		`) VALUES (` +
		`?, ?, ?, ?` +
		`)`

	// run query
	XOLog(sqlstr, sa.AccountID, sa.UserID, sa.UserPermission, sa.SharingAccepted)
	res, err := db.Exec(sqlstr, sa.AccountID, sa.UserID, sa.UserPermission, sa.SharingAccepted)
	if err != nil {
		return err
	}

	// retrieve id
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// set primary key and existence
	sa.ID = int(id)
	sa._exists = true

	return nil
}

// Update updates the SharedAccount in the database.
func (sa *SharedAccount) Update(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !sa._exists {
		return errors.New("update failed: does not exist")
	}

	// if deleted, bail
	if sa._deleted {
		return errors.New("update failed: marked for deletion")
	}

	// sql query
	const sqlstr = `UPDATE trackit.shared_account SET ` +
		`account_id = ?, user_id = ?, user_permission = ?, sharing_accepted = ?` +
		` WHERE id = ?`

	// run query
	XOLog(sqlstr, sa.AccountID, sa.UserID, sa.UserPermission, sa.SharingAccepted, sa.ID)
	_, err = db.Exec(sqlstr, sa.AccountID, sa.UserID, sa.UserPermission, sa.SharingAccepted, sa.ID)
	return err
}

// Save saves the SharedAccount to the database.
func (sa *SharedAccount) Save(db XODB) error {
	if sa.Exists() {
		return sa.Update(db)
	}

	return sa.Insert(db)
}

// Delete deletes the SharedAccount from the database.
func (sa *SharedAccount) Delete(db XODB) error {
	var err error

	// if doesn't exist, bail
	if !sa._exists {
		return nil
	}

	// if deleted, bail
	if sa._deleted {
		return nil
	}

	// sql query
	const sqlstr = `DELETE FROM trackit.shared_account WHERE id = ?`

	// run query
	XOLog(sqlstr, sa.ID)
	_, err = db.Exec(sqlstr, sa.ID)
	if err != nil {
		return err
	}

	// set deleted
	sa._deleted = true

	return nil
}

// AwsAccount returns the AwsAccount associated with the SharedAccount's AccountID (account_id).
//
// Generated from foreign key 'shared_account_ibfk_1'.
func (sa *SharedAccount) AwsAccount(db XODB) (*AwsAccount, error) {
	return AwsAccountByID(db, sa.AccountID)
}

// User returns the User associated with the SharedAccount's UserID (user_id).
//
// Generated from foreign key 'shared_account_ibfk_2'.
func (sa *SharedAccount) User(db XODB) (*User, error) {
	return UserByID(db, sa.UserID)
}

// SharedAccountsByAccountID retrieves a row from 'trackit.shared_account' as a SharedAccount.
//
// Generated from index 'foreign_aws_account'.
func SharedAccountsByAccountID(db XODB, accountID int) ([]*SharedAccount, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, account_id, user_id, user_permission, sharing_accepted ` +
		`FROM trackit.shared_account ` +
		`WHERE account_id = ?`

	// run query
	XOLog(sqlstr, accountID)
	q, err := db.Query(sqlstr, accountID)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*SharedAccount{}
	for q.Next() {
		sa := SharedAccount{
			_exists: true,
		}

		// scan
		err = q.Scan(&sa.ID, &sa.AccountID, &sa.UserID, &sa.UserPermission, &sa.SharingAccepted)
		if err != nil {
			return nil, err
		}

		res = append(res, &sa)
	}

	return res, nil
}

// SharedAccountsByUserID retrieves a row from 'trackit.shared_account' as a SharedAccount.
//
// Generated from index 'foreign_user_id'.
func SharedAccountsByUserID(db XODB, userID int) ([]*SharedAccount, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, account_id, user_id, user_permission, sharing_accepted ` +
		`FROM trackit.shared_account ` +
		`WHERE user_id = ?`

	// run query
	XOLog(sqlstr, userID)
	q, err := db.Query(sqlstr, userID)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	// load results
	res := []*SharedAccount{}
	for q.Next() {
		sa := SharedAccount{
			_exists: true,
		}

		// scan
		err = q.Scan(&sa.ID, &sa.AccountID, &sa.UserID, &sa.UserPermission, &sa.SharingAccepted)
		if err != nil {
			return nil, err
		}

		res = append(res, &sa)
	}

	return res, nil
}

// SharedAccountByID retrieves a row from 'trackit.shared_account' as a SharedAccount.
//
// Generated from index 'shared_account_id_pkey'.
func SharedAccountByID(db XODB, id int) (*SharedAccount, error) {
	var err error

	// sql query
	const sqlstr = `SELECT ` +
		`id, account_id, user_id, user_permission, sharing_accepted ` +
		`FROM trackit.shared_account ` +
		`WHERE id = ?`

	// run query
	XOLog(sqlstr, id)
	sa := SharedAccount{
		_exists: true,
	}

	err = db.QueryRow(sqlstr, id).Scan(&sa.ID, &sa.AccountID, &sa.UserID, &sa.UserPermission, &sa.SharingAccepted)
	if err != nil {
		return nil, err
	}

	return &sa, nil
}
