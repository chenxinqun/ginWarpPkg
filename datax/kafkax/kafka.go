package kafkax

import (
	"github.com/Shopify/sarama"
	"github.com/chenxinqun/ginWarpPkg/convert"
	"strconv"
)

type Producer struct {
	Client sarama.SyncProducer
	info   Info
}

func (p *Producer) GetConfig() Info {
	return p.info
}

func newSyncProducer(cfg Info, optionHandlers ...OptionHandler) (sarama.SyncProducer, error) {
	opt := new(option)
	for _, handler := range optionHandlers {
		handler(opt)
	}
	// For the data collector, we are looking for strong consistency semantics.
	// Because we don't change the flush settings, sarama will try to produce messages
	// as fast as possible to keep latency low.
	if cfg.Producer.RetryMax <= 0 {
		cfg.Producer.RetryMax = DefaultRetryMax
	}
	var (
		requiredAcks sarama.RequiredAcks
	)
	if cfg.Producer.RequiredAcks != "" {
		r, err := strconv.Atoi(cfg.Producer.RequiredAcks)
		requiredAcks = sarama.RequiredAcks(r)
		if err != nil {
			return nil, err
		}
	} else {
		requiredAcks = sarama.WaitForAll
	}
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = requiredAcks       // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = cfg.Producer.RetryMax // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true
	tlsConfig, err := createTLSConfiguration(cfg)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.
	producer, err := sarama.NewSyncProducer(cfg.BrokerList, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func (p *Producer) Close() error {
	return p.Client.Close()
}

func (p *Producer) GetClient() *Producer {
	return p
}

// SendMessage 发送一条消息, 返回一个消息指针. 关注 msg.PartitionConsumer, msg.Offset 两个变量.
func (p *Producer) SendMessage(topic Topic, key string, value interface{}) (msg *ProducerMessage, err error) {
	var byteVal []byte
	byteVal, err = convert.StructToJSON(value)
	if err != nil {
		return
	}
	msg, err = p.SendMessageByte(topic, key, byteVal)
	return msg, nil
}

func (p *Producer) SendMessageByte(topic Topic, key string, value []byte) (msg *ProducerMessage, err error) {
	m := &sarama.ProducerMessage{
		Topic: topic.String(),
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	_, _, err = p.Client.SendMessage(m)
	if err != nil {
		return nil, err
	}
	msg = &ProducerMessage{*m}
	return msg, nil
}

// SendMessages 批量发送消息, 返回消息指针数组. 关注 msg.PartitionConsumer, msg.Offset 两个变量.
func (p *Producer) SendMessages(topic Topic, key string, values ...interface{}) (msgList []*ProducerMessage, err error) {
	mList := make([][]byte, len(values))
	for _, value := range values {
		var byteVal []byte
		byteVal, err = convert.StructToJSON(value)
		if err != nil {
			return nil, err
		}

		mList = append(mList, byteVal)
	}
	msgList, err = p.SendMessagesByte(topic, key, mList...)
	return msgList, err
}

func (p *Producer) SendMessagesByte(topic Topic, key string, values ...[]byte) (msgList []*ProducerMessage, err error) {
	mList := make([]*sarama.ProducerMessage, 0)
	for _, value := range values {
		msg := &sarama.ProducerMessage{
			Topic: topic.String(),
			Key:   sarama.StringEncoder(key),
			Value: sarama.ByteEncoder(value),
		}
		mList = append(mList, msg)
	}
	err = p.Client.SendMessages(mList)
	if err != nil {
		return nil, err
	}
	for _, value := range mList {
		if msgList == nil {
			msgList = make([]*ProducerMessage, 0)
		}
		msg := &ProducerMessage{*value}
		msgList = append(msgList, msg)
	}
	return msgList, err
}
