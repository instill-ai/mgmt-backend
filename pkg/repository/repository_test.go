//go:build dbtest
// +build dbtest

package repository

import (
	"context"
	"database/sql"
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

func TestRepository_SubtractCredit(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()
	cache, _ := redismock.NewClientMock()
	ownerUID := uuid.Must(uuid.NewV4())

	c.Run("nok - negative subtraction", func(c *qt.C) {
		tx := db.Begin()
		c.Cleanup(func() { tx.Rollback() })
		repo := NewRepository(tx, cache)

		err := repo.SubtractCredit(ctx, ownerUID, -10)
		c.Check(err, qt.ErrorMatches, "only positive amounts are allowed")
	})

	c.Run("ok - subtract with lock", func(c *qt.C) {
		// We need one committed record to assert the lock mechanism.
		c.Cleanup(func() {
			err := db.Exec("DELETE FROM credit WHERE owner_uid = ?", ownerUID).Error
			c.Assert(err, qt.IsNil)
		})
		repo := NewRepository(db, cache)
		err := repo.AddCredit(ctx, datamodel.Credit{
			OwnerUID: ownerUID,
			Amount:   100,
		})
		c.Assert(err, qt.IsNil)

		tx1, tx2 := db.Begin(), db.Begin()
		r1, r2 := NewRepository(tx1, cache), NewRepository(tx2, cache)
		endOfTx1, endOfTx2 := tx1.Rollback, tx2.Rollback

		err = r1.SubtractCredit(ctx, ownerUID, 80)
		c.Check(err, qt.IsNil)

		ch := make(chan struct{})
		go func() {
			_ = r2.SubtractCredit(ctx, ownerUID, 80)
			c.Check(err, qt.IsNil)
			endOfTx2()
			close(ch)
		}()

		// until tx1 is closed, tx2 cannot acquire the lock
		select {
		case <-ch:
			c.Fatal("cannot read from ch as tx1 is not over")
		case <-time.After(100 * time.Millisecond):
			c.Log("lock blocked until tx1 is over")
		}

		endOfTx1() // lock is released

		select {
		case _, ok := <-ch:
			c.Check(ok, qt.IsFalse) // channel is closed
			c.Log("tx1 is over, next lock can be acquired")
		case <-time.After(100 * time.Millisecond):
			c.Fatal("tx2 should be able to acquire the lock")
		}
	})
}

func TestRepository_CreditLedger(t *testing.T) {
	c := qt.New(t)
	ctx := context.Background()

	now := time.Now().UTC()
	ownerUID := uuid.Must(uuid.NewV4())

	cache, _ := redismock.NewClientMock()

	// Add credit for different user.
	tx := db.Begin()
	c.Cleanup(func() { tx.Rollback() })
	repo := NewRepository(tx, cache)

	err := repo.AddCredit(ctx, datamodel.Credit{
		OwnerUID: uuid.Must(uuid.NewV4()),
		Amount:   100,
	})
	c.Assert(err, qt.IsNil)

	// These test cases build upon each other. They add a new credit entry and
	// / or subtract a quantity and check the remaining credit at a given
	// moment.
	testcases := []struct {
		name                string
		addAmount           float64
		addExpiration       time.Time
		subtractAmount      float64
		wantRemainingCredit float64
		wantErr             error

		// shiftExpiration will update the expiration times of the existing
		// records to the previous day. Since the remaining credit is compared
		// against `time.Now()`, this is a way we can build a ledger history
		// and check the remaining credit at different points in time.
		shiftExpiration bool
	}{
		// | amount | expiration |
		// |--------|------------|
		// |        |            |
		{
			name:                "nok - no records",
			subtractAmount:      10,
			wantRemainingCredit: 0,
			wantErr:             ErrNotEnoughCredit,
		},
		// | amount | expiration |
		// |--------|------------|
		// | +100   | tomorrow   |
		{
			name:                "ok - add monthly credit",
			addAmount:           100,
			addExpiration:       now.Add(24 * time.Hour),
			wantRemainingCredit: 100,
		},
		// | amount | expiration |
		// |--------|------------|
		// | +100   | tomorrow   |
		// | -80    | tomorrow   |
		{
			name:                "ok - subtract from expiring credit",
			subtractAmount:      80,
			wantRemainingCredit: 20,
		},
		// | amount | expiration |
		// |--------|------------|
		// | +100   | tomorrow   |
		// | -80    | tomorrow   |
		// | -20    | tomorrow   |
		{
			name:                "nok - insufficient, removes remaining credit",
			subtractAmount:      80,
			wantErr:             ErrNotEnoughCredit,
			wantRemainingCredit: 0,
		},
		// | amount | expiration |
		// |--------|------------|
		// | +100   | tomorrow   |
		// | -80    | tomorrow   |
		// | -20    | tomorrow   |
		// | +500   |            |
		{
			name:                "ok - with monthly credit",
			addAmount:           500,
			wantRemainingCredit: 500,
		},
		// | amount | expiration |
		// |--------|------------|
		// | +100   | tomorrow   |
		// | -80    | tomorrow   |
		// | -20    | tomorrow   |
		// | +500   |            |
		// | -80    |            |
		{
			name:                "ok - subtract from non-expiring credit",
			subtractAmount:      80,
			wantRemainingCredit: 420,
		},
		// | amount | expiration |
		// |--------|------------|
		// | +100   | yesterday  |
		// | -80    | yesterday  |
		// | -20    | yesterday  |
		// | +500   |            |
		// | -80    |            |
		{shiftExpiration: true},
		// | amount | expiration |
		// |--------|------------|
		// | +100   | yesterday  |
		// | -80    | yesterday  |
		// | -20    | yesterday  |
		// | +500   |            |
		// | -80    |            |
		// | +100   | tomorrow   |
		{
			name:                "ok - add new monthly credit",
			addAmount:           100,
			addExpiration:       now.Add(24 * time.Hour),
			wantRemainingCredit: 520,
		},
		// | amount | expiration |
		// |--------|------------|
		// | +100   | yesterday  |
		// | -80    | yesterday  |
		// | -20    | yesterday  |
		// | +500   |            |
		// | -80    |            |
		// | +100   | tomorrow   |
		// | -80    | tomorrow   |
		// | -20    | tomorrow   |
		// | -60    |            |
		{
			name:                "ok - subtract from new monthly credit",
			subtractAmount:      80,
			wantRemainingCredit: 440,
		},
		{
			name:                "ok - mixed subtraction",
			subtractAmount:      80,
			wantRemainingCredit: 360,
		},
		// | amount | expiration |
		// |--------|------------|
		// | +100   | 3 days ago |
		// | -80    | 3 days ago |
		// | -20    | 3 days ago |
		// | +500   |            |
		// | -80    |            |
		// | +100   | yesterday  |
		// | -80    | yesterday  |
		// | -20    | yesterday  |
		// | -60    |            |
		{shiftExpiration: true},
		{
			name:                "ok - expiring credit is subtracted first",
			wantRemainingCredit: 360,
		},
	}

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			if tc.shiftExpiration {
				q := "UPDATE credit SET expire_time = expire_time - INTERVAL '2 day' WHERE expire_time IS NOT NULL AND owner_uid = ?"
				err := tx.Exec(q, ownerUID).Error
				c.Check(err, qt.IsNil)
				return
			}

			if tc.addAmount > 0 {
				newEntry := datamodel.Credit{
					OwnerUID: ownerUID,
					Amount:   tc.addAmount,
					ExpireTime: sql.NullTime{
						Time:  tc.addExpiration,
						Valid: !tc.addExpiration.IsZero(),
					},
				}

				err := repo.AddCredit(ctx, newEntry)
				c.Check(err, qt.IsNil)
			}

			var err error
			if tc.subtractAmount > 0 {
				err = repo.SubtractCredit(ctx, ownerUID, tc.subtractAmount)
			}
			c.Check(err, qt.Equals, tc.wantErr)

			got, err := repo.GetRemainingCredit(ctx, ownerUID)
			c.Check(err, qt.IsNil)
			c.Check(got, qt.Equals, tc.wantRemainingCredit)
		})
	}
}
