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
	"errors"
	"flag"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/asifbg-systems/jsonlog"

	"github.com/asifbg-systems/trackit/aws"
	"github.com/asifbg-systems/trackit/aws/s3"
	"github.com/asifbg-systems/trackit/cache"
	"github.com/asifbg-systems/trackit/db"
)

const iso8601DateFormat = "2006-01-02"

// taskIngest ingests billing data for a given BillRepository and AwsAccount.
func taskIngestLimit(ctx context.Context) error {
	args := flag.Args()
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	logger.Debug("Running task 'ingest-limit'.", map[string]interface{}{
		"args": args,
	})
	if len(args) != 3 {
		return errors.New("taskIngest requires three integer arguments")
	} else if aa, err := strconv.Atoi(args[0]); err != nil {
		return err
	} else if br, err := strconv.Atoi(args[1]); err != nil {
		return err
	} else if dateUpperLimit, err := time.Parse(iso8601DateFormat, args[2]); err != nil {
		logger.Debug("Error while decoding date", map[string]interface{}{
			"date": dateUpperLimit,
			"err":  err,
		})
		return err
	} else {
		logger.Debug("Launching ingest billing", map[string]interface{}{
			"date": dateUpperLimit,
		})
		return ingestBillingDataForBillRepositoryLimit(ctx, aa, br, dateUpperLimit)
	}
}

// ingestBillingDataForBillRepository ingests the billing data for a
// BillRepository.
func ingestBillingDataForBillRepositoryLimit(ctx context.Context, aaId, brId int, dateUpperLimit time.Time) (err error) {
	var tx *sql.Tx
	var aa aws.AwsAccount
	var br s3.BillRepository
	var updateId int64
	var latestManifest time.Time
	logger := jsonlog.LoggerFromContextOrDefault(ctx)
	logger.Info("In ingest billing 1", nil)
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
	} else if br, err = s3.GetBillRepositoryForAwsAccountById(aa, brId, tx); err != nil {
	} else if updateId, err = registerUpdate(db.Db, br); err != nil {
	} else if latestManifest, err = s3.UpdateReportLimit(ctx, aa, br, dateUpperLimit); err != nil {
		if billError, castok := err.(awserr.Error); castok {
			br.Error = billError.Message()
			s3.UpdateBillRepositoryWithoutContext(br, db.Db)
		}
	} else {
		logger.Info("In ingest billing else error", nil)
		br.Error = ""
		err = updateBillRepositoryForNextUpdate(ctx, tx, br, latestManifest)
	}
	if err != nil {
		logger.Error("Failed to ingest billing data.", map[string]interface{}{
			"awsAccountId":     aaId,
			"billRepositoryId": brId,
			"error":            err.Error(),
		})
	}
	updateCompletion(ctx, aaId, brId, db.Db, updateId, err)
	updateSubAccounts(ctx, aa)
	var affectedRoutes = []string{
		"/costs",
		"/costs/diff",
		"/costs/tags/keys",
		"/costs/tags/values",
		"/s3/costs",
	}
	_ = cache.RemoveMatchingCache(affectedRoutes, []string{aa.AwsIdentity}, logger)
	return
}
