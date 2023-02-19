package kafkax

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/httpx/trace"
	"os" //nolint:nolintlint,gci
	"sync"

	"github.com/Shopify/sarama"
)

const (
	// DefaultRetryMax 默认重试时间, 单位秒.
	DefaultRetryMax = 10
	// DefaultFlushFrequency 默认刷新频率, 500毫秒.
	DefaultFlushFrequency = 500
	// DefaultVersion 默认卡夫卡版本
	DefaultVersion = "2.8.0"
)

type ProducerConfigInfo struct {
	Enable bool `toml:"Enable" json:"Enable"`
	// RequiredAcks 两种模式都要配置.要求填数字. 因为有0值, 所以不用int型, 以免产生歧义. 不填有默认值.
	RequiredAcks string `toml:"RequiredAcks" json:"RequiredAcks"`

	// RetryMax 同步模式时的重试次数. 要求填数字, 不填有默认值
	RetryMax int `toml:"RetryMax" json:"RetryMax"`

	// Compression 异步模式时配置 要求填数字. 因为有0值, 所以不用int型, 以免产生歧义. 不填有默认值.
	Compression string `toml:"Compression" json:"Compression"`
	// 异步模式时刷新频率, 单位毫秒. 不填有默认值
	FlushFrequency int64 `toml:"FlushFrequency" json:"FlushFrequency"`
}

type GroupConfigInfo struct {
	GroupID  string `toml:"GroupID" json:"GroupID"`
	Assignor string `toml:"Assignor" json:"Assignor"`
	Oldest   bool   `toml:"Oldest" json:"Oldest"`
}

type TLSConfigInfo struct {
	CertFile  string `toml:"CertFile" json:"CertFile"`
	KeyFile   string `toml:"KeyFile" json:"KeyFile"`
	CaFile    string `toml:"CaFile" json:"CaFile"`
	VerifySsl bool   `toml:"VerifySsl" json:"VerifySsl"`
}

type Info struct {
	// 连接地址, 集群则配置多个. 必填.
	BrokerList []string `toml:"BrokerList" json:"BrokerList"`
	// 生产者相关配置.
	Producer      ProducerConfigInfo `toml:"Producer" json:"Producer"`
	AsyncProducer ProducerConfigInfo `toml:"AsyncProducer" json:"AsyncProducer"`
	TLS           TLSConfigInfo      `toml:"TLS" json:"TLS"`
	Group         GroupConfigInfo    `toml:"Group" json:"Group"`
	Version       string             `toml:"Version" json:"Version"`
}

// Topic 卡夫卡的主题名称, 都用这个结构体封装
type Topic interface {
	String() (topic string)
}

type CloseFunc func() error

type OptionHandler func(*option)

type Trace = trace.T

type option struct {
	Trace *trace.Trace
	Kafka *trace.Kafka
}

type ProducerMessage struct {
	sarama.ProducerMessage
}
type ProducerError struct {
	sarama.ProducerError
}
type ConsumerMessage struct {
	*sarama.ConsumerMessage
}

type ConsumerError struct {
	sarama.ConsumerError
}

type MarkMessageFunc func(msg *ConsumerMessage, metadata string)

type ConsumeHandler func(msgChan <-chan *ConsumerMessage,
	errChan <-chan *ConsumerError, markFunc MarkMessageFunc) error

type ConsumeGroupHandler func(msg *ConsumerMessage) (error, bool)

type ConsumerGroupRepo interface {
	Close() error
	GetClient() *ConsumerGroup
	// Consume 消费者, 这个是一个阻塞的动作. 应该包裹在一个for循环中.
	Consume(topics []Topic, handler ConsumeGroupHandler) error
	Errors() <-chan error
	GetConfig() Info
}

type AsyncProducerRepo interface {
	GetClient() *AsyncProducer

	// AsyncClose triggers a shutdown of the producer. The shutdown has completed
	// when both the Errors and Successes channels have been closed. When calling
	// AsyncClose, you *must* continue to read from those channels in order to
	// drain the results of any messages in flight.
	AsyncClose()

	// Close shuts down the producer and waits for any buffered messages to be
	// flushed. You must call this function before a producer object passes out of
	// scope, as it may otherwise leak memory. You must call this before process
	// shutting down, or you may lose messages. You must call this before calling
	// Close on the underlying client.
	Close() error
	// SendMessage 发送一条消息, 返回一个消息指针. 关注 msg.PartitionConsumer, msg.Offset 两个变量.
	SendMessage(topic Topic, key string, value interface{}) (msg *ProducerMessage, err error)
	SendMessageByte(topic Topic, key string, value []byte) (msg *ProducerMessage, err error)
	// SendMessages 批量发送消息, 返回消息指针数组. 关注 msg.PartitionConsumer, msg.Offset 两个变量.
	SendMessages(topic Topic, key string, values ...interface{}) (msgList []*ProducerMessage, err error)
	SendMessagesByte(topic Topic, key string, values ...[]byte) (msgList []*ProducerMessage, err error)
	// Successes is the success output channel back to the user when Return.Successes is
	// enabled. If Return.Successes is true, you MUST read from this channel or the
	// Producer will deadlock. It is suggested that you send and read messages
	// together in a single select statement.
	Successes() <-chan *ProducerMessage

	// Errors is the error output channel back to the user. You MUST read from this
	// channel or the Producer will deadlock when the channel is full. Alternatively,
	// you can set Producer.Return.Errors in your config to false, which prevents
	// errors to be returned.
	Errors() <-chan *ProducerError
}

type ProducerRepo interface {

	// Close shuts down the producer; you must call this function before a producer
	// object passes out of scope, as it may otherwise leak memory.
	// You must call this before calling Close on the underlying client.
	Close() error
	GetClient() *Producer
	// SendMessage 发送一条消息, 返回一个消息指针. 关注 msg.PartitionConsumer, msg.Offset 两个变量.
	SendMessage(topic Topic, key string, value interface{}) (msg *ProducerMessage, err error)
	SendMessageByte(topic Topic, key string, value []byte) (msg *ProducerMessage, err error)
	// SendMessages 批量发送消息, 返回消息指针数组. 关注 msg.PartitionConsumer, msg.Offset 两个变量.
	SendMessages(topic Topic, key string, values ...interface{}) (msgList []*ProducerMessage, err error)
	SendMessagesByte(topic Topic, key string, values ...[]byte) (msgList []*ProducerMessage, err error)
}

type PartitionConsumerRepo interface {
	// AsyncClose initiates a shutdown of the PartitionConsumer. This method will return immediately, after which you
	// should continue to service the 'Messages' and 'Errors' channels until they are empty. It is required to call this
	// function, or Close before a Consumer object passes out of scope, as it will otherwise leak memory. You must call
	// this before calling Close on the underlying client.
	AsyncClose()

	// Close stops the PartitionConsumer from fetching messages. It will initiate a shutdown just like AsyncClose, drain
	// the Messages channel, harvest any errors & return them to the caller. Note that if you are continuing to service
	// the Messages channel when this function is called, you will be competing with Close for messages; consider
	// calling AsyncClose, instead. It is required to call this function (or AsyncClose) before a Consumer object passes
	// out of scope, as it will otherwise leak memory. You must call this before calling Close on the underlying client.
	Close() error

	// Messages returns the read channel for the messages that are returned by
	// the broker.
	Messages() <-chan *ConsumerMessage

	// Errors returns a read channel of errors that occurred during consuming, if
	// enabled. By default, errors are logged and not returned over this channel.
	// If you want to implement any custom error handling, set your config's
	// Consumer.Return.Errors setting to true, and read from this channel.
	Errors() <-chan *ConsumerError

	// HighWaterMarkOffset returns the high water mark offset of the partition,
	// i.e. the offset that will be used for the next message that will be produced.
	// You can use this to determine how far behind the processing is.
	HighWaterMarkOffset() int64
}

// 消费组.
var defaultConsumerGroup ConsumerGroupRepo

// 异步生产者.
var defaultAsyncProducer AsyncProducerRepo

// 生产者.
var defaultProducer ProducerRepo

func DefaultConsumerGroup() ConsumerGroupRepo {
	return defaultConsumerGroup
}
func DefaultAsyncProducer() AsyncProducerRepo {
	return defaultAsyncProducer
}
func DefaultProducer() ProducerRepo {
	return defaultProducer
}

var (
	ProducerEnableErr      = errno.NewError("cfg.Producer.Enable is false")      //nolint:nolintlint,errname
	AsyncProducerEnableErr = errno.NewError("cfg.AsyncProducer.Enable is false") //nolint:nolintlint,errname
)

func NewProducer(cfg Info, optionHandlers ...OptionHandler) (ProducerRepo, error) {
	if !cfg.Producer.Enable {
		return nil, ProducerEnableErr
	}
	client, err := newSyncProducer(cfg, optionHandlers...)
	repo := &Producer{
		Client: client,
		info:   cfg,
	}
	if defaultProducer == nil {
		defaultProducer = repo
	}

	return repo, err
}

func NewAsyncProducer(cfg Info, optionHandlers ...OptionHandler) (AsyncProducerRepo, error) {
	if !cfg.AsyncProducer.Enable {
		return nil, AsyncProducerEnableErr
	}
	client, err := newAsyncProducer(cfg, optionHandlers...)
	repo := &AsyncProducer{Client: client}
	if defaultAsyncProducer == nil {
		defaultAsyncProducer = repo
	}

	return repo, err
}

func NewConsumerGroup(cfg Info, optionHandlers ...OptionHandler) (ConsumerGroupRepo, error) {
	client, err := newConsumerGroup(cfg, optionHandlers...)
	repo := &ConsumerGroup{
		Client:         client,
		info:           cfg,
		optionHandlers: optionHandlers,
		cancelList:     make([]context.CancelFunc, 0),
		wgList:         make([]*sync.WaitGroup, 0),
		handlerList:    make([]*consumerGroupHandler, 0),
	}
	if defaultConsumerGroup == nil {
		defaultConsumerGroup = repo
	}

	return repo, err
}

func createTLSConfiguration(cfg Info, optionHandlers ...OptionHandler) (*tls.Config, error) {
	opt := new(option)
	for _, handler := range optionHandlers {
		handler(opt)
	}
	if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" && cfg.TLS.CaFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			return nil, err
		}

		caCert, err := os.ReadFile(cfg.TLS.CaFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t := &tls.Config{
			Rand:                        nil,
			Time:                        nil,
			Certificates:                []tls.Certificate{cert},
			GetCertificate:              nil,
			GetClientCertificate:        nil,
			GetConfigForClient:          nil,
			VerifyPeerCertificate:       nil,
			VerifyConnection:            nil,
			RootCAs:                     caCertPool,
			NextProtos:                  nil,
			ServerName:                  "",
			ClientAuth:                  0,
			ClientCAs:                   nil,
			InsecureSkipVerify:          cfg.TLS.VerifySsl, //nolint:nolintlint,gosec
			CipherSuites:                nil,
			PreferServerCipherSuites:    false,
			SessionTicketsDisabled:      false,
			ClientSessionCache:          nil,
			MinVersion:                  0,
			MaxVersion:                  0,
			CurvePreferences:            nil,
			DynamicRecordSizingDisabled: false,
			Renegotiation:               0,
			KeyLogWriter:                nil,
		}

		return t, nil
	}
	// will be nil by default if nothing is provided
	return nil, nil
}
