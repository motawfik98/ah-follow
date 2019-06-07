package models

import (
	"github.com/jinzhu/gorm"
	"strconv"
)

type File struct {
	gorm.Model
	Bytes       []byte `json:"bytes" gorm:"type:varbinary(max)"`
	TaskID      uint   `json:"task_id" gorm:"default:null"`
	ContentType string `json:"content_type"`
	FileCount   int    `json:"file_count"`
	Hash        string `json:"hash"`
}

func (file *File) AfterCreate(scope *gorm.Scope) error {
	ID := int(file.ID)
	hash := generateHash(ID)
	scope.DB().Model(file).Updates(File{Hash: hash})
	return nil
}

func GenerateFilesObjectJson(files []File) map[string]interface{} {
	fileOutput := make(map[string]interface{})
	uploadID := make(map[string]interface{})
	stringID := strconv.Itoa(int(files[0].ID))
	uploadID["id"] = stringID
	number := GenerateNumberObjectJson(files)
	fileOutput["files"] = map[string]interface{}{
		"files": number,
	}
	fileOutput["upload"] = uploadID
	return fileOutput
}

func GenerateNumberObjectJson(files []File) map[string]interface{} {
	number := make(map[string]interface{})
	fileIndexInTheSameTask := 0
	for fileNumber, file := range files {
		if fileNumber != 0 {
			if files[fileNumber].TaskID == files[fileNumber-1].TaskID {
				fileIndexInTheSameTask++
			} else {
				fileIndexInTheSameTask = 0
			}
		}
		stringID := strconv.Itoa(int(file.ID))
		number[stringID] = map[string]string{
			"filename":   file.Hash,
			"web_path":   "/tasks/file/" + file.Hash,
			"created_at": file.CreatedAt.String()[0:10] + "  رقم:  " + strconv.Itoa(fileIndexInTheSameTask+1),
		}
	}
	return number
}
