package kafkax

import (
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/chenxinqun/ginWarpPkg/convert"
	"github.com/chenxinqun/ginWarpPkg/loggerx"

	"github.com/Shopify/sarama"
)

type AsyncProducer struct {
	Client    sarama.AsyncProducer
	successes chan *ProducerMessage
	errors    chan *ProducerError
}

func newAsyncProducer(cfg Info, optionHandlers ...OptionHandler) (sarama.AsyncProducer, error) { //nolint:nolintlint,cyclop
	// For the access log, we are looking for AP semantics, with high throughput.
	// By creating batches of compressed messages, we reduce network I/O at a cost of more latency.
	opt := new(option)
	for _, handler := range optionHandlers {
		handler(opt)
	}
	config := sarama.NewConfig()
	tlsConfig, err := createTLSConfiguration(cfg)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	var (
		requiredAcks sarama.RequiredAcks
		compression  sarama.CompressionCodec
	)
	if cfg.Producer.RequiredAcks != "" {
		r, err := strconv.Atoi(cfg.Producer.RequiredAcks) //nolint:govet
		requiredAcks = sarama.RequiredAcks(r)
		if err != nil {
			return nil, err
		}
	} else {
		requiredAcks = sarama.WaitForLocal
	}
	if cfg.Producer.Compression != "" {
		r, err := strconv.Atoi(cfg.Producer.Compression) //nolint:govet
		compression = sarama.CompressionCodec(r)
		if err != nil {
			return nil, err
		}
	} else {
		compression = sarama.CompressionSnappy
	}
	if cfg.Producer.FlushFrequency <= 0 {
		cfg.Producer.FlushFrequency = DefaultFlushFrequency
	}
	config.Producer.RequiredAcks = requiredAcks
	config.Producer.Compression = compression
	config.Producer.Flush.Frequency = time.Duration(cfg.Producer.FlushFrequency) * time.Millisecond
	config.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(cfg.BrokerList, config)
	if err != nil {
		return nil, err
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for err := range producer.Errors() {
			loggerx.Default().Error("kafkax async producer 客户端异常", zap.Error(err))
		}
	}()

	return producer, nil
}

func (p *AsyncProducer) AsyncClose() {
	p.Client.AsyncClose()
}

func (p *AsyncProducer) GetClient() *AsyncProducer {
	return p
}

// Close shuts down the producer and waits for any buffered messages to be
// flushed. You must call this function before a producer object passes out of
// scope, as it may otherwise leak memory. You must call this before process
// shutting down, or you may lose messages. You must call this before calling
// Close on the underlying client.
func (p *AsyncProducer) Close() error {
	return p.Client.Close()
}

func (p *AsyncProducer) SendMessage(topic Topic, key string, value interface{}) (msg *ProducerMessage, err error) {
	var byteVal []byte
	byteVal, err = convert.StructToJSON(value)
	if err != nil {
		return nil, err
	}
	msg, err = p.SendMessageByte(topic, key, byteVal)
	return msg, nil
}

func (p *AsyncProducer) SendMessageByte(topic Topic, key string, value []byte) (msg *ProducerMessage, err error) {
	m := &sarama.ProducerMessage{
		Topic: topic.String(),
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	// TODO ctx, close的情况.
	p.Client.Input() <- m
	msg = &ProducerMessage{*m}
	return msg, nil
}

func (p *AsyncProducer) SendMessages(topic Topic,
	key string, values ...interface{}) (msgList []*ProducerMessage, err error) {
	mList := make([][]byte, 0)
	for _, value := range values {
		var byteVal []byte
		byteVal, err = convert.StructToJSON(value)
		if err != nil {
			return nil, err
		}
		mList = append(mList, byteVal)
	}
	msgList, err = p.SendMessagesByte(topic, key, mList...)

	return msgList, nil
}

func (p *AsyncProducer) SendMessagesByte(topic Topic, key string, values ...[]byte) (msgList []*ProducerMessage, err error) {
	for _, value := range values {
		if msgList == nil {
			msgList = make([]*ProducerMessage, 0)
		}
		m := &sarama.ProducerMessage{
			Topic: topic.String(),
			Key:   sarama.StringEncoder(key),
			Value: sarama.ByteEncoder(value),
		}
		p.Client.Input() <- m
		msg := &ProducerMessage{*m}
		msgList = append(msgList, msg)
	}
	return msgList, nil
}

// Successes is the success output channel back to the user when Return.Successes is
// enabled. If Return.Successes is true, you MUST read from this channel or the
// Producer will deadlock. It is suggested that you send and read messages
// together in a single select statement.
func (p *AsyncProducer) Successes() <-chan *ProducerMessage {
	if p.successes == nil {
		p.successes = make(chan *ProducerMessage)
	}
	go func() {
		for succ := range p.Client.Successes() {
			s := &ProducerMessage{*succ}
			p.successes <- s
		}
	}()

	return p.successes
}

// Errors is the error output channel back to the user. You MUST read from this
// channel or the Producer will deadlock when the channel is full. Alternatively,
// you can set Producer.Return.Errors in your config to false, which prevents
// errors to be returned.
func (p *AsyncProducer) Errors() <-chan *ProducerError {
	if p.errors == nil {
		p.errors = make(chan *ProducerError)
	}
	go func() {
		for succ := range p.Client.Errors() {
			s := &ProducerError{*succ}
			p.errors <- s
		}
	}()

	return p.errors
}
