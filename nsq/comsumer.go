package nsq

import (
	"github.com/bitly/go-simplejson"
	"github.com/nsqio/go-nsq"
	"log"
	"net/url"
	"svf-project/config"
	"svf-project/controllers"
	"svf-project/models"
)

var video controllers.Video

func Start() {
	cfg := nsq.NewConfig()

	consumer, err := nsq.NewConsumer(config.Get().Nsq.Topic, config.Get().Nsq.Channel, cfg)
	if err != nil {
		log.Fatal(err)
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		res, err := simplejson.NewJson(message.Body)
		if err != nil {
			log.Println("Failed to init simpleJson")
			return nil
		}

		eventName := res.Get("EventName").MustString()
		if eventName == "s3:ObjectCreated:Put" {
			objectName := res.Get("Records").GetIndex(0).Get("s3").Get("object").Get("key").MustString()
			objectName, err = url.QueryUnescape(objectName)
			if err != nil {
				log.Println("Failed to unescape string")
				return err
			}

			v := models.Video{ObjectName: objectName}
			video.Uploaded(&v)

			log.Println(objectName + " is uploaded")
		}

		return nil
	}))

	if err := consumer.ConnectToNSQD(config.Get().Nsq.Endpoint); err != nil {
		log.Fatal(err)
	}

	<-consumer.StopChan
}
