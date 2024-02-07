package repository

import (
	"github.com/MangataL/BangumiBuddy/internal/transfer"
)

type fileTransferredSchema struct {
	ID                uint   `gorm:"type:int;primaryKey;autoIncrement"`
	OriginFile        string `gorm:"type:varchar(512);uniqueIndex"`
	NewFileID         string `gorm:"type:varchar(512);uniqueIndex"`
	NewFile           string `gorm:"type:varchar(512);index"`
	SubscriptionID    string `gorm:"type:varchar(36);index"`
	BangumiName       string `gorm:"type:varchar(255);not null;index:idx_bangumi,priority:1"`
	Season            int    `gorm:"type:int;not null;index:idx_bangumi,priority:2"`
}

func (fileTransferredSchema) TableName() string {
	return "file_transferred"
}

func toFileTransferred(schema fileTransferredSchema) transfer.FileTransferred {
	return transfer.FileTransferred{
		OriginFile:     schema.OriginFile,
		NewFileID:      schema.NewFileID,
		BangumiName:    schema.BangumiName,
		Season:         schema.Season,
		SubscriptionID: schema.SubscriptionID,
		NewFile:        schema.NewFile,
	}
}

func fromFileTransferred(fileTransferred transfer.FileTransferred) fileTransferredSchema {
	return fileTransferredSchema{
		OriginFile:     fileTransferred.OriginFile,
		NewFileID:      fileTransferred.NewFileID,
		BangumiName:    fileTransferred.BangumiName,
		Season:         fileTransferred.Season,
		SubscriptionID: fileTransferred.SubscriptionID,
		NewFile:        fileTransferred.NewFile,
	}
}
