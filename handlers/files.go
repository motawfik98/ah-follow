package handlers

import (
	"../models"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (db *MyDB) validateFile(c echo.Context) error {

	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
	}
	fileGiven := form.File["upload"][0]
	file := models.File{}

	// Source
	src, err := fileGiven.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	reader, _ := fileGiven.Open()
	file.Bytes, _ = ioutil.ReadAll(reader)
	file.ContentType = http.DetectContentType(file.Bytes)
	fmt.Println(file.ContentType)

	db.GormDB.Create(&file)

	files := make([]models.File, 1)
	files[0] = file
	fileOutput := models.GenerateFilesObjectJson(files)

	return c.JSONPretty(http.StatusOK, fileOutput, " ")
}

func linkFiles(db *MyDB, c *echo.Context, taskID uint) {
	context := *c
	numberOfFiles, _ := strconv.Atoi(context.FormValue("data[files-many-count]"))
	var filesIDs []int
	for i := 0; i < numberOfFiles; i++ {
		fileID, _ := strconv.Atoi(context.FormValue(fmt.Sprintf("data[files][%d][id]", i)))
		filesIDs = append(filesIDs, fileID)

	}
	if filesIDs == nil {
		filesIDs = append(filesIDs, 0)
	}
	db.GormDB.Table("files").Where("id IN (?)", filesIDs).UpdateColumn("task_id", taskID)
	db.GormDB.Delete(models.File{}, "task_id = ? AND id NOT IN (?)", taskID, filesIDs)
}

func (db *MyDB) showFile(c echo.Context) error {
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
	c.Response().Header().Set("Content-Type", file.ContentType)
	c.Response().Header().Set("content-disposition", "inline;filename="+file.Hash)
	c.Response().Header().Set("Cache-control", "must-revalidate, post-check=0, pre-check=0")
	_, _ = c.Response().Write(file.Bytes)
	c.Response().Flush()
	return nil
}
