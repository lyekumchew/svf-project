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
	"strings"
	"svf-project/config"
	"svf-project/database"
	"svf-project/minio"
	"svf-project/models"
	"time"
)

type Video struct {
	Basic
}

func (v *Video) Index(c *gin.Context) {
	var video models.Video

	var needPasswd bool

	if database.DB.Where(&models.Video{VideoId: c.Param("video_id")}).First(&video).Error != nil {
		v.JsonFail(c, http.StatusNotFound, "Data does not exist.")
		return
	}

	// if not uploaded yet
	if video.IsUploaded.Bool == false {
		v.JsonSuccess(c, http.StatusOK, gin.H{"is_uploaded": false})
		return
	}

	// if need password
	if video.Password.Valid == true {
		needPasswd = true
	}

	v.JsonSuccess(c, http.StatusOK, gin.H{"video_name": video.VideoName, "is_uploaded": true, "need_passwd": needPasswd})
}

func (v *Video) Show(c *gin.Context) {
	var video models.Video
	var request QueryRequest

	if err := c.ShouldBind(&request); err == nil {
		if database.DB.Where(&models.Video{VideoId: c.Param("video_id")}).First(&video).Error != nil {
			v.JsonFail(c, http.StatusNotFound, "Data does not exist.")
			return
		}

		// Not Uploaded
		if video.IsUploaded.Bool == false {
			v.JsonFail(c, http.StatusNotFound, "This video is not uploaded yet.")
			return
		}

		// if not need to provide the password
		if video.Password.Valid == false {
			url, err := minio.GetObjectUrl(video.ObjectName)

			if err != nil {
				log.Println("Failed to get object url")
				v.JsonFail(c, http.StatusInternalServerError, "Unable to get object url.")
				return
			}

			v.JsonSuccess(c, http.StatusOK, gin.H{"video_name": video.VideoName, "url": url})

			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(video.Password.String), []byte(request.Password)); err != nil {
			v.JsonFail(c, http.StatusUnauthorized, "Password is wrong.")
			return
		}

		url, err := minio.GetObjectUrl(video.ObjectName)
		if err != nil {
			log.Println("Failed to get object url")
			v.JsonFail(c, http.StatusInternalServerError, "Unable to get object url.")
			return
		}

		v.JsonSuccess(c, http.StatusOK, gin.H{"url": strings.Replace(url, "http://"+config.Get().Minio.Endpoint, config.Get().Minio.ExternalEndPoint, 1)})

	} else {
		v.JsonFail(c, http.StatusBadRequest, "Please check your json is ok.")
	}
}

func (v *Video) Store(c *gin.Context) {
	var request CreateRequest

	if err := c.ShouldBind(&request); err == nil {
		if request.VideoSuffix != "mp4" {
			v.JsonFail(c, http.StatusBadRequest, "This video format does not allow uploads")
			return
		}

		timeLocal, _ := time.LoadLocation("Asia/Shanghai")
		time.Local = timeLocal
		t := time.Now().Local()
		date := fmt.Sprintf("%4d/%2d/%2d/", t.Year(), t.Month(), t.Day())

		videoId := xid.New().String()
		url, formData, err := minio.CreatePostUrl(date + videoId + "." + request.VideoSuffix)
		if err != nil {
			v.JsonFail(c, http.StatusBadRequest, err.Error())
			return
		}

		var password sql.NullString
		if request.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Println(err)
			}
			password = sql.NullString{String: string(hash), Valid: true}
		} else {
			password = sql.NullString{Valid: false}
		}

		video := models.Video{
			VideoId:    videoId,
			VideoName:  request.VideoName,
			ObjectName: date + videoId + "." + request.VideoSuffix,
			Password:   password,
			DeleteId:   ksuid.New().String(),
		}

		if err := database.DB.Create(&video).Error; err != nil {
			v.JsonFail(c, http.StatusBadRequest, err.Error())
			return
		}

		// Test for curl
		//fmt.Printf("curl ")
		//for k, v := range formData {
		//	fmt.Printf("-F %s=%s ", k, v)
		//}
		//fmt.Printf("-F file=@test.jpg ")
		//fmt.Printf("%s\n", url)

		v.JsonSuccess(c, http.StatusCreated, gin.H{"upload_url": strings.Replace(url, "http://"+config.Get().Minio.Endpoint, config.Get().Minio.ExternalEndPoint, 1), "formData": formData, "video_id": video.VideoId, "delete_id": video.DeleteId})

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

// when the file is uploaded, this function will work
func (v *Video) Uploaded(objectName string) error {
	var video models.Video

	if objectName == "" {
		log.Println("Can't not found the object")
	}

	if err := database.DB.Where("object_name = ?", objectName).First(&video).Error; err != nil {
		log.Printf("Could not found object_name %v.\n", objectName)
		return err
	} else {
		database.DB.Model(&video).Where(video).Update("is_uploaded", sql.NullBool{Valid: true, Bool: true})
		return nil
	}
}

type CreateRequest struct {
	VideoName   string `form:"video_name" json:"video_name" binding:"required,max=128"`
	Password    string `form:"password" json:"password" binding:"max=64"`
	VideoSuffix string `form:"format" json:"format" binding:"required,min=3,max=4"`
}

type QueryRequest struct {
	Password string `form:"password" json:"password" binding:"max=64"` // discard "required,min=1,"
}
