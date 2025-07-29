package dao

import (
	"context"
	"database/sql"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// SystemSettingsDAO implements SystemSettingsDAOInterface
type SystemSettingsDAO struct {
	db bob.Executor
}

// NewSystemSettingsDAO creates a new SystemSettingsDAO
func NewSystemSettingsDAO(db bob.Executor) *SystemSettingsDAO {
	return &SystemSettingsDAO{db: db}
}

// GetSetting retrieves a system setting by key
func (dao *SystemSettingsDAO) GetSetting(ctx context.Context, settingKey string) (*models.SystemSetting, error) {
	settings, err := models.SystemSettings.Query(
		models.SelectWhere.SystemSettings.SettingKey.EQ(settingKey),
	).All(ctx, dao.db)
	if err != nil {
		return nil, err
	}

	if len(settings) == 0 {
		return nil, nil
	}

	return settings[0], nil
}

// SetSetting creates or updates a system setting
func (dao *SystemSettingsDAO) SetSetting(ctx context.Context, settingKey, settingValue, settingType string, updatedBy int64) error {
	// Check if setting exists
	existing, err := dao.GetSetting(ctx, settingKey)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing setting
		updateSetter := &models.SystemSettingSetter{
			SettingValue: &settingValue,
			SettingType:  &settingType,
			UpdatedBy:    &sql.Null[int64]{V: updatedBy, Valid: true},
		}
		return existing.Update(ctx, dao.db, updateSetter)
	} else {
		// Create new setting
		insertSetter := &models.SystemSettingSetter{
			SettingKey:   &settingKey,
			SettingValue: &settingValue,
			SettingType:  &settingType,
			UpdatedBy:    &sql.Null[int64]{V: updatedBy, Valid: true},
		}
		_, err := models.SystemSettings.Insert(insertSetter).One(ctx, dao.db)
		return err
	}
}

// GetAllSettings retrieves all system settings
func (dao *SystemSettingsDAO) GetAllSettings(ctx context.Context) ([]*models.SystemSetting, error) {
	settings, err := models.SystemSettings.Query().All(ctx, dao.db)
	if err != nil {
		return nil, err
	}
	return settings, nil
}
