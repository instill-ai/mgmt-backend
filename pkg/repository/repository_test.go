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

const decimal float64 = 10.86850005

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

func TestRepository_AddCredit(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	t0 := time.Now().UTC().Add(-1 * time.Minute)

	cache, _ := redismock.NewClientMock()

	tx := db.Begin()
	c.Cleanup(func() { tx.Rollback() })

	repo := NewRepository(tx, cache)

	credit := datamodel.Credit{
		OwnerUID: uuid.Must(uuid.NewV4()),
		Amount:   10.86850000,
	}
	err := repo.AddCredit(ctx, credit)
	c.Check(err, qt.IsNil)

	got := new(datamodel.Credit)
	err = tx.Model(datamodel.Credit{}).Where("owner_uid = ?", credit.OwnerUID).First(got).Error
	c.Check(err, qt.IsNil)
	c.Check(got.UID, qt.Not(qt.Equals), uuid.UUID{})
	c.Check(got.CreateTime.After(t0), qt.IsTrue)
	c.Check(got.UpdateTime.After(t0), qt.IsTrue)
}

func TestRepository_GetRemainingCredit(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	cache, _ := redismock.NewClientMock()
	ownerUID := uuid.Must(uuid.NewV4())

	c.Run("ok - no credit records", func(c *qt.C) {
		tx := db.Begin()
		c.Cleanup(func() { tx.Rollback() })
		repo := NewRepository(tx, cache)

		credit, err := repo.GetRemainingCredit(ctx, ownerUID)
		c.Check(err, qt.IsNil)
		c.Check(credit, qt.Equals, float64(0))
	})

	c.Run("ok - count only non-expired, nonzero amounts", func(c *qt.C) {
		tx := db.Begin()
		c.Cleanup(func() { tx.Rollback() })
		repo := NewRepository(tx, cache)

		now := time.Now().UTC()
		onDB := []datamodel.Credit{
			{
				OwnerUID: uuid.Must(uuid.NewV4()),
				Amount:   10,
			},
			{
				OwnerUID: ownerUID,
				Amount:   20,
				ExpireTime: sql.NullTime{
					Time:  now.Add(-10 * time.Hour),
					Valid: true,
				},
			},
			{OwnerUID: ownerUID},
			{
				OwnerUID: ownerUID,
				Amount:   decimal,
				ExpireTime: sql.NullTime{
					Time:  now.Add(10 * time.Hour),
					Valid: true,
				},
			},
			{
				OwnerUID: ownerUID,
				Amount:   100,
			},
		}

		for _, record := range onDB {
			err := repo.AddCredit(ctx, record)
			c.Assert(err, qt.IsNil)
		}

		credit, err := repo.GetRemainingCredit(ctx, ownerUID)
		c.Check(err, qt.IsNil)
		c.Check(credit, qt.Equals, decimal+100)
	})
}

func TestRepository_SubtractCredit(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	now := time.Now()

	cache, _ := redismock.NewClientMock()

	ownerUID := uuid.Must(uuid.NewV4())

	c.Run("nok - no records", func(c *qt.C) {
		tx := db.Begin()
		c.Cleanup(func() { tx.Rollback() })
		repo := NewRepository(tx, cache)

		err := repo.SubtractCredit(ctx, ownerUID, 100)
		c.Check(errors.Is(err, ErrNotEnoughCredit), qt.IsTrue)
	})

	existingCredit := []datamodel.Credit{
		{ // different user
			OwnerUID: uuid.Must(uuid.NewV4()),
			Amount:   10,
		},
		{ // expired
			OwnerUID: ownerUID,
			Amount:   20,
			ExpireTime: sql.NullTime{
				Time:  now.Add(-10 * time.Hour),
				Valid: true,
			},
		},
		{ // used up
			OwnerUID: ownerUID,
		},
		{ // with expiration
			OwnerUID: ownerUID,
			Amount:   10,
			ExpireTime: sql.NullTime{
				Time:  now.Add(10 * time.Hour),
				Valid: true,
			},
		},
		{ // without expiration
			OwnerUID: ownerUID,
			Amount:   20,
		},
	}

	c.Run("nok - not enough credit", func(c *qt.C) {
		tx := db.Begin()
		c.Cleanup(func() { tx.Rollback() })
		repo := NewRepository(tx, cache)

		for _, record := range existingCredit {
			err := repo.AddCredit(ctx, record)
			c.Assert(err, qt.IsNil)
		}
		err := repo.SubtractCredit(ctx, ownerUID, 100)
		c.Check(errors.Is(err, ErrNotEnoughCredit), qt.IsTrue)

		credit, err := repo.GetRemainingCredit(ctx, ownerUID)
		c.Check(err, qt.IsNil)
		c.Check(credit, qt.Equals, float64(0))
	})

	c.Run("ok - subtract first from credit with expiration date", func(c *qt.C) {
		tx := db.Begin()
		c.Cleanup(func() { tx.Rollback() })
		repo := NewRepository(tx, cache)

		for _, record := range existingCredit {
			err := repo.AddCredit(ctx, record)
			c.Assert(err, qt.IsNil)
		}
		err := repo.SubtractCredit(ctx, ownerUID, 25)
		c.Check(err, qt.IsNil)

		credit, err := repo.GetRemainingCredit(ctx, ownerUID)
		c.Check(err, qt.IsNil)
		c.Check(credit, qt.Equals, float64(5))

		// Check credit with expiration was used first.
		q := tx.Model(datamodel.Credit{}).Where("owner_uid = ?", ownerUID).
			Where("amount > 0").
			Where("expire_time is null or expire_time > ?", time.Now())

		count := int64(0)
		err = q.Count(&count).Error
		c.Check(err, qt.IsNil)
		c.Check(count, qt.Equals, int64(1))

		got := new(datamodel.Credit)
		err = q.First(got).Error
		c.Check(err, qt.IsNil)
		c.Check(got.ExpireTime.Valid, qt.IsFalse)
	})
}
