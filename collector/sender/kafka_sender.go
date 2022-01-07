package sender

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
	"github.com/yihongzhi/log-kit/config"
	"github.com/yihongzhi/log-kit/kafka"
)

type KafkaSender struct {
	Producer  *kafka.Producer
	TopicName string
}

func NewKafkaSender(config *config.KafkaConfig) (*KafkaSender, error) {
	producer, err := kafka.NewKafkaProducer(config)
	if err != nil {
		log.Error("SyncProduce create failed !", err)
		return nil, err
	}
	return &KafkaSender{
		Producer:  producer,
		TopicName: config.TopicName,
	}, nil
}

// SendMessage 发送日志消息
func (d *KafkaSender) SendMessage(message *LogMessage) error {
	text, err := json.Marshal(&message)
	if err != nil {
		log.Errorln("serialization msg failed ", err)
		return err
	}
	msg := sarama.ProducerMessage{
		Topic: d.TopicName,
		Key:   sarama.StringEncoder(message.AppId),
		Value: sarama.StringEncoder(text),
	}
	partition, offset, err := d.Producer.SendMessage(&msg)
	if err != nil {
		log.Errorln("send kafka msg failed", err)
		return err
	}
	log.Debugf("send to kafka appId:[%s],toppic:[%s],partition:[%d],offset:[%d]", message.AppId, d.TopicName, partition, offset)
	return nil
}
