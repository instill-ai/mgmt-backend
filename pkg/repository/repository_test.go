//go:build dbtest
// +build dbtest

package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/go-redis/redismock/v9"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"

	"github.com/instill-ai/mgmt-backend/config"
	"github.com/instill-ai/mgmt-backend/pkg/datamodel"
	database "github.com/instill-ai/mgmt-backend/pkg/db"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	if err := config.Init("../../config/config.yaml"); err != nil {
		panic(err)
	}

	db = database.GetConnection(&config.Config.Database)
	defer database.Close(db)

	os.Exit(m.Run())
}

func TestRepository_CreateUser(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	cache, _ := redismock.NewClientMock()
	tx := db.Begin()
	c.Cleanup(func() { tx.Rollback() })

	repo := NewRepository(tx, cache)

	id := "homer-wombat"
	email := "homer@wombats.com"
	user := &datamodel.Owner{
		ID:    id,
		Email: email,
		OwnerType: sql.NullString{
			String: "user",
			Valid:  true,
		},
	}

	t0 := time.Now()
	err := repo.CreateUser(ctx, user)
	c.Check(err, qt.IsNil)

	got, err := repo.GetUser(ctx, id, false)
	c.Check(err, qt.IsNil)
	c.Check(got.CreateTime.After(t0), qt.IsTrue)
	c.Check(got.UpdateTime.After(t0), qt.IsTrue)
}

func TestRepository_GetOwner(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	cache, _ := redismock.NewClientMock()
	tx := db.Begin()
	c.Cleanup(func() { tx.Rollback() })

	repo := NewRepository(tx, cache)

	user := &datamodel.Owner{
		ID:    "piano-wombat",
		Email: "piano@wombats.com",
		OwnerType: sql.NullString{
			String: "user",
			Valid:  true,
		},
	}
	org := &datamodel.Owner{
		ID:    "the-wombies",
		Email: "org@wombats.com",
		OwnerType: sql.NullString{
			String: "organization",
			Valid:  true,
		},
	}

	user.UID, org.UID = uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())

	err := repo.CreateUser(ctx, user)
	c.Check(err, qt.IsNil)

	err = repo.CreateOrganization(ctx, org)
	c.Check(err, qt.IsNil)

	c.Run("nok - get with wrong type", func(c *qt.C) {
		_, err := repo.GetUser(ctx, org.ID, false)
		c.Check(errors.Is(err, gorm.ErrRecordNotFound), qt.IsTrue, qt.Commentf(err.Error()))

		_, err = repo.GetUserByUID(ctx, org.UID)
		c.Check(errors.Is(err, gorm.ErrRecordNotFound), qt.IsTrue, qt.Commentf(err.Error()))

		_, err = repo.GetOrganization(ctx, user.ID, false)
		c.Check(errors.Is(err, gorm.ErrRecordNotFound), qt.IsTrue, qt.Commentf(err.Error()))

		_, err = repo.GetOrganizationByUID(ctx, user.UID)
		c.Check(errors.Is(err, gorm.ErrRecordNotFound), qt.IsTrue, qt.Commentf(err.Error()))
	})

	c.Run("ok - get with type", func(c *qt.C) {
		gotUser, err := repo.GetUser(ctx, user.ID, false)
		c.Check(err, qt.IsNil)
		c.Check(gotUser.UID, qt.Equals, user.UID)

		gotUser, err = repo.GetUserByUID(ctx, user.UID)
		c.Check(err, qt.IsNil)
		c.Check(gotUser.ID, qt.Equals, user.ID)

		gotOrg, err := repo.GetOrganization(ctx, org.ID, false)
		c.Check(err, qt.IsNil)
		c.Check(gotOrg.UID, qt.Equals, org.UID)

		gotOrg, err = repo.GetOrganizationByUID(ctx, org.UID)
		c.Check(err, qt.IsNil)
		c.Check(gotOrg.ID, qt.Equals, org.ID)
	})

	c.Run("ok - get without type", func(c *qt.C) {
		gotUser, err := repo.GetOwner(ctx, user.ID, false)
		c.Check(err, qt.IsNil)
		c.Check(gotUser.UID, qt.Equals, user.UID)

		gotUser, err = repo.GetOwnerByUID(ctx, user.UID)
		c.Check(err, qt.IsNil)
		c.Check(gotUser.ID, qt.Equals, user.ID)

		gotOrg, err := repo.GetOwner(ctx, org.ID, false)
		c.Check(err, qt.IsNil)
		c.Check(gotOrg.UID, qt.Equals, org.UID)

		gotOrg, err = repo.GetOwnerByUID(ctx, org.UID)
		c.Check(err, qt.IsNil)
		c.Check(gotOrg.ID, qt.Equals, org.ID)
	})
}
