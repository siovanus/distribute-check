// Package store encapsulates all database interaction.
package store

import (
	"fmt"
	"github.com/polynetwork/distribute-check/store/migrations"
	"github.com/polynetwork/distribute-check/store/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math/big"
)

// Client holds a connection to the database.
type Client struct {
	db *gorm.DB
}

// ConnectToDB attempts to connect to the database URI provided,
// and returns a new Client instance if successful.
func ConnectToDb(uri string) (*Client, error) {
	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to open %s for gorm DB: %+v", uri, err)
	}
	if err = migrations.Migrate(db); err != nil {
		return nil, fmt.Errorf("newDBStore#Migrate: %s", err)
	}
	store := &Client{
		db: db,
	}
	return store, nil
}

func (client Client) LoadTrackHeight() (uint64, error) {
	trackHeight := &models.TrackHeight{
		Height: 1,
	}
	err := client.db.Where(&models.TrackHeight{Name: "height"}).FirstOrCreate(&trackHeight).Error
	return trackHeight.Height, err
}

func (client Client) SaveTrackHeight(height uint64) error {
	trackHeight := &models.TrackHeight{
		Name:   "height",
		Height: height,
	}
	return client.db.Save(trackHeight).Error
}

func (client Client) LoadValidator(consensusAddress string) (*models.Validator, error) {
	validator := new(models.Validator)
	err := client.db.Where(&models.Validator{ConsensusAddress: consensusAddress}).First(validator).Error
	return validator, err
}

func (client Client) LoadValidatorNum() (uint64, error) {
	var num uint64
	err := client.db.Raw("SELECT COUNT(*) FROM validators").Scan(&num).Error
	return num, err
}

func (client Client) SaveValidator(validator *models.Validator) error {
	return client.db.Save(validator).Error
}

func (client Client) AddValidatorStake(stakeAddress, consensusAddress string, amount *big.Int) error {
	validator, err := client.LoadValidator(consensusAddress)
	if err != nil {
		return fmt.Errorf("AddValidatorStake, client.LoadValidator error: %s", err)
	}
	validator.TotalStake = models.NewBigInt(new(big.Int).Add(&validator.TotalStake.Int, amount))
	if validator.StakeAddress == stakeAddress {
		validator.SelfStake = models.NewBigInt(new(big.Int).Add(&validator.SelfStake.Int, amount))
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
	validator.TotalStake = models.NewBigInt(new(big.Int).Sub(&validator.TotalStake.Int, amount))
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
	stakeInfo := &models.StakeInfo{
		Amount: models.NewBigInt(new(big.Int)),
	}
	err := client.db.Where(&models.StakeInfo{StakeAddress: stakeAddress, ConsensusAddress: consensusAddr}).FirstOrCreate(stakeInfo).Error
	return stakeInfo, err
}

func (client Client) LoadAllStakeAddress(consensusAddr string) ([]string, error) {
	r := make([]string, 0)
	err := client.db.Model(&models.StakeInfo{}).Select("stake_address").Where("consensus_address = ?", consensusAddr).Find(&r).Error
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
	stakeInfo.Amount = models.NewBigInt(new(big.Int).Add(&stakeInfo.Amount.Int, amount))
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
	stakeInfo.Amount = models.NewBigInt(new(big.Int).Sub(&stakeInfo.Amount.Int, amount))
	err = client.SaveStakeInfo(stakeInfo)
	if err != nil {
		return fmt.Errorf("SubStakeInfo, client.SaveStakeInfo error: %s", err)
	}
	return nil
}

func (client Client) LoadDoneTx(hash string) ([]models.DoneTx, error) {
	doneTx := make([]models.DoneTx, 0)
	err := client.db.Where(&models.DoneTx{Hash: hash}).Find(&doneTx).Error
	return doneTx, err
}

func (client Client) SaveDoneTx(doneTx *models.DoneTx) error {
	return client.db.Save(doneTx).Error
}

func (client Client) CleanDoneTx() error {
	return client.db.Where("1 = 1").Delete(&models.DoneTx{}).Error
}

func (client Client) LoadTotalGas(height uint64) (*models.TotalGas, error) {
	totalGas := new(models.TotalGas)
	err := client.db.Where(&models.TotalGas{Height: height}).First(totalGas).Error
	return totalGas, err
}

func (client Client) SaveTotalGas(totalGas *models.TotalGas) error {
	return client.db.Save(totalGas).Error
}

func (client Client) LoadAccumulateGasFee(address string, height uint64) (*big.Int, error) {
	r := make([]models.GasFee, 0)
	err := client.db.Where("address = ? AND height <= ?", address, height).Find(&r).Error
	agf := new(big.Int)
	for _, v := range r {
		agf = new(big.Int).Add(agf, &v.GasFee.Int)
	}
	return agf, err
}

func (client Client) SaveGasFee(gasFee *models.GasFee) error {
	return client.db.Save(gasFee).Error
}

func (client Client) LoadAccumulateRewards(address string, height uint64) (*big.Int, error) {
	r := make([]models.Rewards, 0)
	err := client.db.Where("address = ? AND height <= ?", address, height).Find(&r).Error
	ar := new(big.Int)
	for _, v := range r {
		ar = new(big.Int).Add(ar, &v.Amount.Int)
	}
	return ar, err
}

func (client Client) SaveRewards(rewards *models.Rewards) error {
	return client.db.Save(rewards).Error
}

func (client Client) LoadRewards(address string, height uint64) (*big.Int, error) {
	rewards := &models.Rewards{
		Address: address,
		Height:  height,
		Amount:  models.NewBigInt(new(big.Int)),
	}
	err := client.db.Where(&models.Rewards{Address: address, Height: height}).FirstOrCreate(rewards).Error
	return &rewards.Amount.Int, err
}

func (client Client) LoadAccumulatedRewards() (*big.Int, error) {
	accumulatedRewards := models.AccumulatedRewards{
		Amount: models.NewBigInt(new(big.Int)),
	}
	err := client.db.Where(&models.AccumulatedRewards{Name: "accumulatedRewards"}).FirstOrCreate(&accumulatedRewards).Error
	return &accumulatedRewards.Amount.Int, err
}

func (client Client) SaveAccumulatedRewards(amount *big.Int) error {
	accumulatedRewards := &models.AccumulatedRewards{
		Name:   "accumulatedRewards",
		Amount: models.NewBigInt(amount),
	}
	return client.db.Save(accumulatedRewards).Error
}

func (client Client) LoadCommunityRate() (*big.Int, error) {
	communityRate := new(models.CommunityRate)
	err := client.db.Where(&models.CommunityRate{Name: "communityRate"}).First(communityRate).Error
	return &communityRate.Amount.Int, err
}

func (client Client) SaveCommunityRate(amount *big.Int) error {
	communityRate := &models.CommunityRate{
		Name:   "communityRate",
		Amount: models.NewBigInt(amount),
	}
	return client.db.Save(communityRate).Error
}
