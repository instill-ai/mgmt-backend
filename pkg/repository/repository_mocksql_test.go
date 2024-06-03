package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	qt "github.com/frankban/quicktest"
	"github.com/go-redis/redismock/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func mockDBRepository() (sqlmock.Sqlmock, *sql.DB, Repository, error) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, nil, err
	}

	gormdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqldb,
	}))

	if err != nil {
		return nil, nil, nil, err
	}

	redisClient, _ := redismock.NewClientMock()
	repository := NewRepository(gormdb, redisClient)

	return mock, sqldb, repository, err
}

func TestRepository_UpdateTokenLastUseTime(t *testing.T) {
	c := qt.New(t)
	tokenAccess := "fakeTokenAccess"

	mock, sqldb, repository, err := mockDBRepository()
	c.Assert(err, qt.IsNil)
	defer sqldb.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tokens" SET "last_use_time"=$1,"update_time"=$2 WHERE access_token = $3`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), tokenAccess).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repository.UpdateTokenLastUseTime(context.Background(), tokenAccess)
	c.Assert(err, qt.IsNil)
}
