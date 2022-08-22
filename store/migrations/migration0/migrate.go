package migration0

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/polynetwork/distribute-check/store/models"
)

// Migrate runs the initial migration
func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&models.TrackHeight{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate TrackHeight")
	}

	err = tx.AutoMigrate(&models.Validator{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Validator")
	}

	err = tx.AutoMigrate(&models.EpochInfo{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate EpochInfo")
	}

	err = tx.AutoMigrate(&models.StakeInfo{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate StakeInfo")
	}

	err = tx.AutoMigrate(&models.DoneTx{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate DoneTx")
	}

	err = tx.AutoMigrate(&models.TotalGas{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate TotalGas")
	}

	err = tx.AutoMigrate(&models.Rewards{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Rewards")
	}

	return nil
}
