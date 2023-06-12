package postgres

import (
	"github.com/jinzhu/gorm"
	"oasisTracker/common/helpers"
	"oasisTracker/dmodels"
	"time"
)

func (d *Postgres) SaveBlocks(blocks []dmodels.Block) error {
	err := d.db.Transaction(func(tx *gorm.DB) error {
		b := new(dmodels.BlockInfo)
		bd := new(dmodels.BlockDayInfo)

		vs := new(dmodels.ValidatorInfo)
		vds := new(dmodels.ValidatorDayInfo)
		for i := range blocks {
			if err := tx.Select("*").
				Table(dmodels.BlocksPostgresTable).
				Order("id desc").
				Limit(1).
				Scan(&b).Error; err != nil {
				if gorm.IsRecordNotFoundError(err) {
					b.ID = 1
					b.TotalBlocks = 0
					b.LastLvl = 0
					if err = tx.Table(dmodels.BlocksPostgresTable).
						Create(b).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			if err := tx.Table(dmodels.BlocksPostgresTable).
				Where("id = ?", b.ID).
				Updates(map[string]interface{}{
					"total_count": gorm.Expr("total_count + 1"),
					"last_lvl":    blocks[i].Height,
				}).
				Error; err != nil {
				return err
			}

			if err := tx.Select("*").
				Table(dmodels.BlocksDayPostgresTable).
				Where("day = ?", helpers.TruncateToDay(time.Now())).
				Scan(&bd).Error; err != nil {
				if gorm.IsRecordNotFoundError(err) {
					lastId := new(dmodels.BlockDayInfo)
					if err = tx.Table(dmodels.BlocksDayPostgresTable).
						Select("id").
						Order("id desc").
						First(&lastId).Error; err != nil {
						if !gorm.IsRecordNotFoundError(err) {
							return err
						}
					}
					bd.ID = lastId.ID + 1
					bd.TotalDayBlocks = 0
					bd.Day = helpers.TruncateToDay(time.Now())
					if err = tx.Table(dmodels.BlocksDayPostgresTable).
						Create(bd).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			if err := tx.Table(dmodels.BlocksDayPostgresTable).
				Where("id = ?", bd.ID).
				Update("day_total_count", gorm.Expr("day_total_count + 1")).
				Error; err != nil {
				return err
			}

			if err := tx.Select("*").
				Table(dmodels.ValidatorsPostgresTable).
				Where("address = ?", blocks[i].ProposerAddress).
				Scan(&vs).Error; err != nil {
				if gorm.IsRecordNotFoundError(err) {
					lastId := new(dmodels.ValidatorInfo)
					if err = tx.Table(dmodels.ValidatorsPostgresTable).
						Select("id").
						Order("id desc").
						First(&lastId).Error; err != nil {
						if !gorm.IsRecordNotFoundError(err) {
							return err
						}
					}
					vs.ID = lastId.ID + 1
					vs.Address = blocks[i].ProposerAddress
					vs.TotalBlocks = 0
					vs.TotalSigs = 0
					vs.LastBlkTime = time.Time{}
					vs.LastSigTime = time.Time{}
					if err = tx.Table(dmodels.ValidatorsPostgresTable).
						Create(vs).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			if err := tx.Table(dmodels.ValidatorsPostgresTable).
				Where("id = ?", vs.ID).
				Updates(map[string]interface{}{
					"total_blk_count": gorm.Expr("total_blk_count + 1"),
					"last_blk_time":   time.Now(),
				}).
				Error; err != nil {
				return err
			}

			if err := tx.Select("*").
				Table(dmodels.ValidatorsDayStatsPostgresTable).
				Where("val_id = ? and day = ?", vs.ID, helpers.TruncateToDay(time.Now())).
				Scan(&vds).Error; err != nil {
				if gorm.IsRecordNotFoundError(err) {
					lastId := new(dmodels.ValidatorDayInfo)
					if err = tx.Table(dmodels.ValidatorsDayStatsPostgresTable).
						Select("id").
						Order("id desc").
						First(&lastId).Error; err != nil {
						if !gorm.IsRecordNotFoundError(err) {
							return err
						}
					}
					vds.ID = lastId.ID + 1
					vds.ValidatorID = vs.ID
					vds.DayBlocks = 0
					vds.DaySigs = 0
					vds.Day = helpers.TruncateToDay(time.Now())
					if err = tx.Table(dmodels.ValidatorsDayStatsPostgresTable).
						Create(vds).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			if err := tx.Table(dmodels.ValidatorsDayStatsPostgresTable).
				Where("id = ?", vds.ID).
				Update("day_blk_count", gorm.Expr("day_blk_count + 1")).
				Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
