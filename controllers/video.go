package controllers

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"svf-project/database"
	"svf-project/minio"
	"svf-project/models"
	"time"
)

type Video struct {
	Basic
}

func (v *Video) Index(c *gin.Context) {
	var videos []models.Video

	database.DB.Select("id, name, username, created_at, updated_at").Order("id").Find(&videos)

	v.JsonSuccess(c, http.StatusOK, gin.H{"data": videos})
}

func (v *Video) Show(c *gin.Context) {
	var video models.Video

	if database.DB.Where(&models.Video{VideoId: c.Param("video_id")}).First(&video).Error != nil {
		v.JsonFail(c, http.StatusNotFound, "Data does not exist.")
		return
	}

	// need password
	if video.Password.Valid != false {
		v.JsonFail(c, http.StatusUnauthorized, "Please provide the password")
		return
	}

	// Not Uploaded
	if video.IsUploaded.Bool == false {
		v.JsonSuccess(c, http.StatusOK, gin.H{"is_uploaded": 0})
		return
	}

	url, err := minio.GetObjectUrl(video.ObjectName)
	if err != nil {
		log.Println("Failed to get object url")
		return
	}

	v.JsonSuccess(c, http.StatusOK, gin.H{"url": url})
}

func (v *Video) ShowWithPassword(c *gin.Context) {
	var video models.Video
	var request QueryRequest

	if err := c.ShouldBind(&request); err == nil {
		if database.DB.Where(&models.Video{VideoId: c.Param("video_id")}).First(&video).Error != nil {
			v.JsonFail(c, http.StatusNotFound, "Data does not exist.")
			return
		}

		// Not Uploaded
		if video.IsUploaded.Bool == false {
			v.JsonSuccess(c, http.StatusOK, gin.H{"is_uploaded": 0})
			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(video.Password.String), []byte(request.Password)); err != nil {
			v.JsonFail(c, http.StatusUnauthorized, "Password is wrong")
			return
		}

		url, err := minio.GetObjectUrl(video.ObjectName)
		if err != nil {
			log.Println("Failed to get object url")
			v.JsonFail(c, http.StatusInternalServerError, "Unable to get object url")
			return
		}

		v.JsonSuccess(c, http.StatusOK, gin.H{"url": url})
	}
}

func (v *Video) Store(c *gin.Context) {
	var request CreateRequest
	if err := c.ShouldBind(&request); err == nil {
		timeLocal, _ := time.LoadLocation("Asia/Shanghai")
		time.Local = timeLocal
		t := time.Now().Local()
		date := fmt.Sprintf("%4d/%2d/%2d/", t.Year(), t.Month(), t.Day())

		videoId := xid.New().String()
		url, fromData, err := minio.CreatePostUrl(date + videoId)
		if err != nil {
			v.JsonFail(c, http.StatusBadRequest, err.Error())
			return
		}

		var pw sql.NullString
		if request.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Println(err)
			}
			pw = sql.NullString{String: string(hash), Valid: true}
		} else {
			pw = sql.NullString{Valid: false}
		}

		video := models.Video{
			VideoId:    videoId,
			ObjectName: date + videoId,
			Password:   pw,
			DeleteId:   ksuid.New().String(),
		}

		if err := database.DB.Create(&video).Error; err != nil {
			v.JsonFail(c, http.StatusBadRequest, err.Error())
			return
		}

		v.JsonSuccess(c, http.StatusCreated, gin.H{"url": url, "fromData": fromData})
	} else {
		v.JsonFail(c, http.StatusBadRequest, err.Error())
	}
}

func (v *Video) Destroy(c *gin.Context) {
	var video models.Video

	if database.DB.Where(&models.Video{DeleteId: c.Param("delete_id")}).First(&video).Error != nil {
		v.JsonFail(c, http.StatusNotFound, "Data does not exist")
		return
	}

	if err := database.DB.Delete(&video).Error; err != nil {
		v.JsonFail(c, http.StatusBadRequest, err.Error())
		return
	}

	v.JsonSuccess(c, http.StatusCreated, gin.H{})
}

func (v *Video) Uploaded(video *models.Video) {
	if video.ObjectName == "" {
		log.Println("Can't not found this object")
	}

	database.DB.Model(video).Where(video).Update("is_uploaded", sql.NullBool{Valid: true, Bool: true})
}

type CreateRequest struct {
	Password string `form:"password" json:"password" binding:"max=64"`
}

type QueryRequest struct {
	Password string `form:"password" json:"password" binding:"required,min=1,max=64"`
}
