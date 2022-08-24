package migrations

import (
	"fmt"
	"github.com/polynetwork/distribute-check/store/models"
	"gorm.io/gorm"
)

// Migrate iterates through available migrations, running and tracking
// migrations that have not been run.
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.TrackHeight{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate TrackHeight: %s", err)
	}

	err = db.AutoMigrate(&models.Validator{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate Validator: %s", err)
	}

	err = db.AutoMigrate(&models.EpochInfo{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate EpochInfo: %s", err)
	}

	err = db.AutoMigrate(&models.StakeInfo{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate StakeInfo: %s", err)
	}

	err = db.AutoMigrate(&models.DoneTx{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate DoneTx: %s", err)
	}

	err = db.AutoMigrate(&models.TotalGas{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate TotalGas: %s", err)
	}

	err = db.AutoMigrate(&models.Rewards{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate Rewards: %s", err)
	}

	err = db.AutoMigrate(&models.AccumulatedRewards{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate AccumulatedRewards: %s", err)
	}

	return nil
}
