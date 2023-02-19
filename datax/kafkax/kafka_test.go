package kafkax

func getCfg() Info {
	return Info{ //nolint:exhaustivestruct
		BrokerList:    []string{"192.168.110.147:9092", "192.168.110.127:9092", "192.168.110.53:9092"},
		Producer:      ProducerConfigInfo{Enable: true}, //nolint:exhaustivestruct
		AsyncProducer: ProducerConfigInfo{Enable: true}, //nolint:exhaustivestruct
		Group:         GroupConfigInfo{GroupID: "test"},
		Version:       "2.8.0",
	}
}

func CreateTestAsyncProducer(cfgs ...Info) AsyncProducerRepo {
	cfg := getCfg()
	if len(cfgs) == 1 {
		cfg = cfgs[0]
	}
	p, _ := NewAsyncProducer(cfg)

	return p
}

func CreateTestProducer(cfgs ...Info) ProducerRepo {
	cfg := getCfg()
	if len(cfgs) == 1 {
		cfg = cfgs[0]
	}
	p, _ := NewProducer(cfg)

	return p
}
func CreateTestConsumerGroup(cfgs ...Info) ConsumerGroupRepo {
	cfg := getCfg()
	if len(cfgs) == 1 {
		cfg = cfgs[0]
	}
	p, _ := NewConsumerGroup(cfg)

	return p
}
