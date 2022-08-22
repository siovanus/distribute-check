// Package store encapsulates all database interaction.
package store

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/polynetwork/distribute-check/store/migrations"
	"github.com/polynetwork/distribute-check/store/models"
	"math/big"
)

const (
	sqlDialect = "postgres"
)

// Client holds a connection to the database.
type Client struct {
	db *gorm.DB
}

// ConnectToDB attempts to connect to the database URI provided,
// and returns a new Client instance if successful.
func ConnectToDb(uri string) (*Client, error) {
	db, err := gorm.Open(sqlDialect, uri)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s for gorm DB: %+v", uri, err)
	}
	if err = migrations.Migrate(db); err != nil {
		return nil, fmt.Errorf("newDBStore#Migrate: %s", err)
	}
	store := &Client{
		db: db.Set("gorm:auto_preload", true),
	}
	return store, nil
}

// Close will close the connection to the database.
func (client Client) Close() error {
	return client.db.Close()
}

func (client Client) LoadTrackHeight() (uint64, error) {
	var trackHeight models.TrackHeight
	err := client.db.Where(&models.TrackHeight{Name: "TrackHeight"}).FirstOrCreate(&trackHeight).Error
	return trackHeight.Height, err
}

func (client Client) SaveTrackHeight(height uint64) error {
	trackHeight := &models.TrackHeight{
		Name:   "TrackHeight",
		Height: height,
	}
	return client.db.Save(trackHeight).Error
}

func (client Client) LoadValidator(consensusAddress string) (*models.Validator, error) {
	validator := new(models.Validator)
	err := client.db.Where(&models.Validator{ConsensusAddress: consensusAddress}).First(validator).Error
	return validator, err
}

func (client Client) SaveValidator(validator *models.Validator) error {
	return client.db.Save(validator).Error
}

func (client Client) AddValidatorStake(stakeAddress, consensusAddress string, amount *big.Int) error {
	validator, err := client.LoadValidator(consensusAddress)
	if err != nil {
		return fmt.Errorf("AddValidatorStake, client.LoadValidator error: %s", err)
	}
	validator.TotalStake = models.NewBigInt(new(big.Int).Add(validator.TotalStake.Int, amount))
	if validator.StakeAddress == stakeAddress {
		validator.SelfStake = models.NewBigInt(new(big.Int).Add(validator.SelfStake.Int, amount))
	}
	err = client.SaveValidator(validator)
	if err != nil {
		return fmt.Errorf("AddValidatorStake, client.SaveValidator error: %s", err)
	}
	return nil
}

func (client Client) SubValidatorStake(consensusAddress string, amount *big.Int) error {
	validator, err := client.LoadValidator(consensusAddress)
	if err != nil {
		return fmt.Errorf("SubValidatorStake, client.LoadValidator error: %s", err)
	}
	validator.TotalStake = models.NewBigInt(new(big.Int).Sub(validator.TotalStake.Int, amount))
	err = client.SaveValidator(validator)
	if err != nil {
		return fmt.Errorf("SubValidatorStake, client.SaveValidator error: %s", err)
	}
	return nil
}

func (client Client) LoadLatestEpochInfo() (*models.EpochInfo, error) {
	epochInfo := new(models.EpochInfo)
	err := client.db.Where(&models.EpochInfo{}).Last(epochInfo).Error
	return epochInfo, err
}

func (client Client) SaveEpochInfo(epochInfo *models.EpochInfo) error {
	return client.db.Save(epochInfo).Error
}

func (client Client) LoadStakeInfo(stakeAddress, consensusAddr string) (*models.StakeInfo, error) {
	stakeInfo := new(models.StakeInfo)
	err := client.db.Where(&models.StakeInfo{StakeAddress: stakeAddress, ConsensusAddr: consensusAddr}).FirstOrCreate(stakeInfo).Error
	return stakeInfo, err
}

func (client Client) LoadAllStakeAddress(consensusAddr string) ([]string, error) {
	r := make([]string, 0)
	err := client.db.Select("stake_address").Where("consensus_address == ?", consensusAddr).Find(&r).Error
	return r, err
}

func (client Client) SaveStakeInfo(stakeInfo *models.StakeInfo) error {
	return client.db.Save(stakeInfo).Error
}

func (client Client) AddStakeInfo(stakeAddress, consensusAddress string, amount *big.Int) error {
	stakeInfo, err := client.LoadStakeInfo(stakeAddress, consensusAddress)
	if err != nil {
		return fmt.Errorf("AddStakeInfo, client.LoadStakeInfo error: %s", err)
	}
	stakeInfo.Amount = models.NewBigInt(new(big.Int).Add(stakeInfo.Amount.Int, amount))
	err = client.SaveStakeInfo(stakeInfo)
	if err != nil {
		return fmt.Errorf("AddStakeInfo, client.SaveStakeInfo error: %s", err)
	}
	return nil
}

func (client Client) SubStakeInfo(stakeAddress, consensusAddress string, amount *big.Int) error {
	stakeInfo, err := client.LoadStakeInfo(stakeAddress, consensusAddress)
	if err != nil {
		return fmt.Errorf("SubStakeInfo, client.LoadStakeInfo error: %s", err)
	}
	stakeInfo.Amount = models.NewBigInt(new(big.Int).Sub(stakeInfo.Amount.Int, amount))
	err = client.SaveStakeInfo(stakeInfo)
	if err != nil {
		return fmt.Errorf("SubStakeInfo, client.SaveStakeInfo error: %s", err)
	}
	return nil
}

func (client Client) LoadDoneTx(hash string) (*models.DoneTx, error) {
	doneTx := new(models.DoneTx)
	err := client.db.Where(&models.DoneTx{Hash: hash}).First(doneTx).Error
	return doneTx, err
}

func (client Client) SaveDoneTx(doneTx *models.DoneTx) error {
	return client.db.Save(doneTx).Error
}

func (client Client) CleanDoneTx() error {
	return client.db.Where(&models.DoneTx{}).Delete(&models.DoneTx{}).Error
}

func (client Client) LoadTotalGas(height uint64) (*models.TotalGas, error) {
	totalGas := new(models.TotalGas)
	err := client.db.Where(&models.TotalGas{Height: height}).First(totalGas).Error
	return totalGas, err
}

func (client Client) SaveTotalGas(totalGas *models.TotalGas) error {
	return client.db.Save(totalGas).Error
}

func (client Client) LoadAccumulateRewards(address string, height uint64) (*big.Int, error) {
	r := make([]*models.BigInt, 0)
	err := client.db.Select("amount").Where("address == ? AND height <= ", address, height).Find(&r).Error
	var ar *big.Int
	for _, v := range r {
		ar = new(big.Int).Add(ar, v.Int)
	}
	return ar, err
}

func (client Client) SaveRewards(rewards *models.Rewards) error {
	return client.db.Save(rewards).Error
}
