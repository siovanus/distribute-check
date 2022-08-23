package migrations

import (
	"github.com/pkg/errors"
	"github.com/polynetwork/distribute-check/store/models"
	"gorm.io/gorm"
)

// Migrate iterates through available migrations, running and tracking
// migrations that have not been run.
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.TrackHeight{})
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate TrackHeight")
	}

	err = db.AutoMigrate(&models.Validator{})
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Validator")
	}

	err = db.AutoMigrate(&models.EpochInfo{})
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate EpochInfo")
	}

	err = db.AutoMigrate(&models.StakeInfo{})
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate StakeInfo")
	}

	err = db.AutoMigrate(&models.DoneTx{})
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate DoneTx")
	}

	err = db.AutoMigrate(&models.TotalGas{})
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate TotalGas")
	}

	err = db.AutoMigrate(&models.Rewards{})
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Rewards")
	}

	return nil
}
