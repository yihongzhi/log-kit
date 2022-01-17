package analyzer

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	logs "github.com/sirupsen/logrus"
	"github.com/yihongzhi/log-kit/analyzer/parser"
	"github.com/yihongzhi/log-kit/collector/sender"
	"github.com/yihongzhi/log-kit/config"
	"github.com/yihongzhi/log-kit/elastic"
	"github.com/yihongzhi/log-kit/kafka"
	"sync"
)

// LogAnalyzer 日志解析服务
type LogAnalyzer struct {
	TopName       string
	EsClient      *elastic.ESClient
	KafkaConsumer *kafka.Consumer
	LogParser     parser.LogParser
}

func NewLogAnalyzer(config *config.AppConfig) (*LogAnalyzer, error) {
	esClient, err := elastic.NewESClient(config.Elastic)
	if err != nil {
		return nil, err
	}
	kafkaConsumer, err := kafka.NewKafkaConsumer(config.Kafka)
	if err != nil {
		return nil, err
	}
	return &LogAnalyzer{
		TopName:       config.Kafka.TopicName,
		EsClient:      esClient,
		KafkaConsumer: kafkaConsumer,
		LogParser:     nil,
	}, nil
}

func (a *LogAnalyzer) Start() error {
	h := &logMessageHandler{
		parser: parser.NewJavaLogParser(),
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := a.KafkaConsumer.Consume(context.Background(), []string{a.TopName}, h); err != nil {
				// 当setup失败的时候，error会返回到这里
				logs.Errorf("Error from consumer: %v", err)
				return
			}
		}
	}()
	logs.Infoln("Sarama consumer up and running!...")
	wg.Wait()
	return nil
}

type logMessageHandler struct {
	parser parser.LogParser
}

func (h *logMessageHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *logMessageHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *logMessageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		//fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		message := &sender.LogMessage{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			return err
		}
		h.parser.Parse(message)
		session.MarkMessage(msg, "")
	}
	claim.Messages()
	return nil
}
