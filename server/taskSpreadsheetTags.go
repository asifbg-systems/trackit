//   Copyright 2017 MSolution.IO
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"context"
	"database/sql"
	"flag"
	"time"

	"github.com/asifbg-systems/jsonlog"

	"github.com/asifbg-systems/trackit/aws"
	"github.com/asifbg-systems/trackit/aws/usageReports/history"
	"github.com/asifbg-systems/trackit/db"
	"github.com/asifbg-systems/trackit/models"
	"github.com/asifbg-systems/trackit/reports"
)

// taskSpreadsheet generates Spreadsheet with reports for a given AwsAccount.
func taskTagsSpreadsheet(ctx context.Context) error {
	args := flag.Args()
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	logger.Debug("Running task 'Spreadsheet Tags'.", map[string]interface{}{
		"args": args,
	})
	aaId, date, err := checkArguments(args)
	if err != nil {
		return err
	} else {
		return generateTagsReport(ctx, aaId, date)
	}
}

func generateTagsReport(ctx context.Context, aaId int, date time.Time) (err error) {
	var tx *sql.Tx
	var aa aws.AwsAccount
	var updateId int64
	var generation bool
	forceGeneration := !date.IsZero()
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	defer func() {
		if tx != nil {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}
	}()
	if tx, err = db.Db.BeginTx(ctx, nil); err != nil {
	} else if aa, err = aws.GetAwsAccountWithId(aaId, tx); err != nil {
	} else if generation, err = checkTagsReportGeneration(ctx, db.Db, aa, forceGeneration); err != nil || !generation {
	} else if updateId, err = registerAccountTagsReportGeneration(db.Db, aa); err != nil {
	} else {
		errs := reports.GenerateTagsReport(ctx, aa, nil, date)
		updateAccountTagsReportGenerationCompletion(ctx, aaId, db.Db, updateId, nil, errs, forceGeneration)
	}
	if err != nil {
		logger.Error("Error while generating spreadsheet tags report.", map[string]interface{}{
			"awsAccountId": aaId,
			"error":        err.Error(),
		})
		updateAccountTagsReportGenerationCompletion(ctx, aaId, db.Db, updateId, err, nil, forceGeneration)
	}
	return
}

func checkTagsReportGeneration(ctx context.Context, db *sql.DB, aa aws.AwsAccount, forceGeneration bool) (bool, error) {
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	if forceGeneration {
		return true, nil
	}
	startDate, endDate := history.GetHistoryDate()
	logger.Info("Checking report generation conditions", map[string]interface{}{
		"awsAccountId": aa.Id,
		"startDate":    startDate.Format("2006-01-02T15:04:05Z"),
		"endDate":      endDate.Format("2006-01-02T15:04:05Z"),
	})
	complete, err := history.CheckBillingDataCompleted(ctx, startDate, endDate, aa)
	if err != nil {
		logger.Info("Error while checking if billing data are completed", map[string]interface{}{
			"awsAccountId": aa.Id,
			"error":        err,
		})
		return false, err
	} else if !complete {
		logger.Info("Billing data are not completed", map[string]interface{}{
			"awsAccountId": aa.Id,
		})
		return false, nil
	}
	return true, nil
}

func registerAccountTagsReportGeneration(db *sql.DB, aa aws.AwsAccount) (int64, error) {
	dbReportGeneration := models.AwsAccountTagsReportsJob{
		AwsAccountID: aa.Id,
		WorkerID:     backendId,
	}
	err := dbReportGeneration.Insert(db)
	if err != nil {
		return 0, err
	}
	return int64(dbReportGeneration.ID), err
}

func updateAccountTagsReportGenerationCompletion(ctx context.Context, aaId int, db *sql.DB, updateId int64, jobErr error, errs map[string]error, forceGeneration bool) {
	rErr := registerAccountTagsReportGenerationCompletion(db, aaId, updateId, jobErr, errs, forceGeneration)
	if rErr != nil {
		logger := jsonlog.LoggerFromContextOrDefault(ctx)
		logger.Error("Failed to register account processing completion.", map[string]interface{}{
			"awsAccountId": aaId,
			"error":        rErr.Error(),
			"updateId":     updateId,
		})
	}
}

func registerAccountTagsReportGenerationCompletion(db *sql.DB, aaId int, updateId int64, jobErr error, errs map[string]error, forceGeneration bool) error {
	dbAccountReports, err := models.AwsAccountTagsReportsJobByID(db, int(updateId))
	if err != nil {
		return err
	}
	date := time.Now()
	dbAccountReports.Completed = date
	dbAccountReports.Joberror = errToStr(jobErr)
	dbAccountReports.Spreadsheeterror = errToStr(errs["speadsheetError"])
	dbAccountReports.Tagsreporterror = errToStr(errs["tagsError"])
	err = dbAccountReports.Update(db)
	if err != nil {
		return err
	}
	if !forceGeneration {
		dbAccount, err := models.AwsAccountByID(db, aaId)
		if err != nil {
			return err
		}
		dbAccount.LastTagsSpreadsheetReportGeneration = date
		err = dbAccount.Update(db)
	}
	return err
}
