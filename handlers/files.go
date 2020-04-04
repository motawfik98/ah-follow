package handlers

import (
	"ah-follow-modules/models"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

func (db *MyConfigurations) validateAndSaveFile(fileGiven *multipart.FileHeader, taskID, userID uint) uint {
	file := models.File{}

	src, err := fileGiven.Open()
	if err != nil {
		fmt.Println(err)
	} else {
		file.TaskID = taskID
		originalBytes, _ := ioutil.ReadAll(src)
		file.Bytes = originalBytes
		if err != nil {
			fmt.Println(err)
		}

		file.ContentType = http.DetectContentType(originalBytes)
		fullFileName := fileGiven.Filename
		file.FileName, file.Extension = getFileNameAndExtension(fullFileName)
		file.UserID = userID
		db.GormDB.Create(&file)
	}
	defer src.Close()
	return file.ID
}

func linkFiles(db *MyConfigurations, c *echo.Context, taskID uint) {
	userID, _ := getUserStatus(c)
	context := *c
	formFiles, _ := context.MultipartForm()
	files := formFiles.File["files[]"]

	for _, file := range files {
		db.validateAndSaveFile(file, taskID, userID)
	}

	totalDeletedFiles, _ := strconv.Atoi(context.FormValue("totalDeletedFiles"))
	deletedFilesHashes := make([]string, totalDeletedFiles)
	for i := 0; i < totalDeletedFiles; i++ {
		deletedFilesHashes = append(deletedFilesHashes, context.FormValue("deleted_file_"+strconv.Itoa(i)))
	}
	totalRenamedFiles, _ := strconv.Atoi(context.FormValue("totalRenamedFiles"))
	for i := 0; i < totalRenamedFiles; i++ {
		fileHash := context.FormValue("file_hash_" + strconv.Itoa(i))
		fileName := context.FormValue("file_name_" + strconv.Itoa(i))
		db.GormDB.Model(models.File{}).Where("hash = ?", fileHash).Update("file_name", fileName)
	}
	db.GormDB.Delete(models.File{}, "task_id = ? AND hash IN (?)", taskID, deletedFilesHashes)
}

func (db *MyConfigurations) showFile(c echo.Context) error {
	//userID, classification := getUserStatus(&c)
	hash := c.Param("hash")
	var file models.File
	//var fileTask models.Task
	db.GormDB.Find(&file, "hash = ?", hash)
	return displayFile(&c, file)

	//db.GormDB.Preload("Users").Find(&fileTask, file.TaskID)
	//if file.TaskID == 0 || classification == 1 {
	//	return displayFile(&c, file)
	//}
	//for _, user := range fileTask.Users {
	//	if user.UserID == userID {
	//		return displayFile(&c, file)
	//	}
	//}
	//return redirectWithFlashMessage("failure", "لم نتمكن من ايجاد الملف المطلوب", "/", &c)
}

func displayFile(context *echo.Context, file models.File) error {
	c := *context
	if checkIfRequestFromMobileDevice(c) {
		c.Response().Header().Set("Content-Type", file.ContentType)
		return c.Blob(http.StatusOK, file.ContentType, file.Bytes)
	}
	c.Response().Header().Set("Content-Type", file.ContentType)
	c.Response().Header().Set("content-disposition", "inline;filename="+file.FileName+"."+file.Extension)
	c.Response().Header().Set("Cache-control", "must-revalidate, post-check=0, pre-check=0")
	_, _ = c.Response().Write(file.Bytes)
	c.Response().Flush()
	return nil
}

func getFileNameAndExtension(fullFileName string) (string, string) {
	extension := fullFileName[strings.LastIndex(fullFileName, ".")+1:]
	fileName := fullFileName[:strings.LastIndex(fullFileName, ".")]
	return fileName, extension
}
