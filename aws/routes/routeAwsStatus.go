package routes

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/asifbg-systems/jsonlog"

	"github.com/asifbg-systems/trackit/aws"
	"github.com/asifbg-systems/trackit/aws/s3"
	"github.com/asifbg-systems/trackit/db"
	"github.com/asifbg-systems/trackit/routes"
	"github.com/asifbg-systems/trackit/users"
)

func getAwsAccountsStatus(r *http.Request, a routes.Arguments) (int, interface{}) {
	var awsAccounts []aws.AwsAccount
	var awsAccountsWithBillRepositories []s3.AwsAccountWithBillRepositoriesWithPending
	u := a[users.AuthenticatedUser].(users.User)
	tx := a[db.Transaction].(*sql.Tx)
	l := jsonlog.LoggerFromContextOrDefault(r.Context())
	awsAccounts, err := aws.GetAwsAccountsFromUser(u, tx)
	if err != nil {
		l.Error("failed to get user's AWS accounts", err.Error())
		return 500, errors.New("failed to retrieve AWS accounts")
	}
	awsAccountsWithBillRepositories, err = s3.WrapAwsAccountsWithBillRepositories(awsAccounts, tx)
	if err != nil {
		l.Error("failed to get AWS accounts' bill repositories", err.Error())
		return 500, errors.New("failed to retrieve bill repositories")
	}
	billRepositoriesIds := make([]int, 0)
	for _, awsAccount := range awsAccountsWithBillRepositories {
		for _, billRepository := range awsAccount.BillRepositories {
			billRepositoriesIds = append(billRepositoriesIds, billRepository.Id)
		}
	}
	result := s3.WrapAwsAccountsWithBillRepositoriesWithPendingWithStatus(awsAccountsWithBillRepositories, tx)
	return 200, result
}
