package kafkax

import (
	"context"
	"fmt"
	"sync"

	"github.com/chenxinqun/ginWarpPkg/loggerx"
	"go.uber.org/zap"

	"github.com/Shopify/sarama"
)

type ConsumerGroup struct {
	Client         sarama.ConsumerGroup
	info           Info
	optionHandlers []OptionHandler
	cancelList     []context.CancelFunc
	wgList         []*sync.WaitGroup
	handlerList    []*consumerGroupHandler
	clientEctype   []sarama.ConsumerGroup
}

func (p *ConsumerGroup) GetConfig() Info {
	return p.info
}

func newConsumerGroup(cfg Info, optionHandlers ...OptionHandler) (sarama.ConsumerGroup, error) {
	opt := new(option)
	for _, handler := range optionHandlers {
		handler(opt)
	}
	// 使用默认版本
	if cfg.Version == "" {
		cfg.Version = DefaultVersion
	}
	// 解析版本为程序可用类型
	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, err
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the Consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = version

	switch cfg.Group.Assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	}
	// 默认从旧的开始消费
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true
	client, err := sarama.NewConsumerGroup(cfg.BrokerList, cfg.Group.GroupID, config)
	go func() {
		for err := range client.Errors() {
			if loggerx.Default() != nil {
				loggerx.Default().Error("kafkax consumer group 客户端异常", zap.Error(err))
			} else {
				fmt.Printf("kafkax consumer group 客户端异常 err: %v \n", err)
			}
		}
	}()
	if err != nil {
		return nil, err
	}
	return client, err
}

func (g *ConsumerGroup) resetClient() {
	client, err := newConsumerGroup(g.info, g.optionHandlers...)
	if err != nil {
		loggerx.Default().Fatal("卡夫卡连接异常, 并且重置失败, 退出程序")
	}
	g.Client = client
	loggerx.Default().Warn("kafkax 客户端连接 重置成功")
}

func (g *ConsumerGroup) copyClient(suffix string) sarama.ConsumerGroup {
	client, err := newConsumerGroup(g.info, g.optionHandlers...)
	if err != nil {
		loggerx.Default().Error("copy 卡夫卡客户端, 异常")
	}
	return client
}

func (g *ConsumerGroup) Close() (err error) {
	for _, h := range g.handlerList {
		h.Stop = true
	}
	for _, c := range g.cancelList {
		c()
	}
	for _, w := range g.wgList {
		w.Wait()
	}
	for _, client := range g.clientEctype {
		_ = client.Close()
	}
	err = g.Client.Close()
	return
}

func (g *ConsumerGroup) GetClient() *ConsumerGroup {
	return g
}

func (g *ConsumerGroup) initHandler(t []string, handler ConsumeGroupHandler) *consumerGroupHandler {
	h := newConsumerGroupHandler(handler)
	h.topics = t
	h.version = g.info.Version
	h.channameBufferSize = 1000
	h.ready = make(chan bool)
	return h
}

// Consume 消费者, 这个是一个阻塞的动作. 应该包裹在一个for循环中. for循环结束记得调用cancel.
func (g *ConsumerGroup) Consume(topics []Topic, handler ConsumeGroupHandler) error {
	var client sarama.ConsumerGroup
	t := make([]string, len(topics))
	for i := 0; i < len(t); i++ {
		t[i] = topics[i].String()
	}
	if len(g.handlerList) > 0 {
		client = g.copyClient("")
		g.clientEctype = append(g.clientEctype, client)
	} else {
		client = g.Client
	}
	var (
		err       error
		needBreak bool
	)
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	h := g.initHandler(t, handler)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				needBreak = true
				loggerx.Default().Warn("kafkax context 关闭, 退出消费", zap.String(" consumer group", g.info.Group.GroupID))
				break
			default:
				err = client.Consume(ctx, t, h)
				if err != nil {
					switch err {
					case sarama.ErrClosedClient, sarama.ErrClosedConsumerGroup:
						// 退出
						loggerx.Default().Info("kafkax 连接关闭, 退出消费", zap.String(" consumer group", g.info.Group.GroupID))
						needBreak = true
					case sarama.ErrOutOfBrokers:
						loggerx.Default().Error("kafkax 消费者崩溃, 重新连接", zap.String(" consumer group", g.info.Group.GroupID))
					default:
						loggerx.Default().Error("kafkax 消费者出现异常", zap.Error(err))
					}
				}
			}
			if e := ctx.Err(); e != nil {
				cancel()
				loggerx.Default().Error("kafkax context 使用过程中发生异常, 退出循环", zap.Error(e))
			}
			if needBreak {
				break
			}
			h.ready = make(chan bool)
		}
	}()
	<-h.ready
	g.cancelList = append(g.cancelList, cancel)
	g.wgList = append(g.wgList, wg)
	g.handlerList = append(g.handlerList, h)
	loggerx.Default().Info("kafkax 消费者已经创立", zap.Any("topics", t))
	return err
}

// Errors returns a read channel of errors that occurred during the Consumer life-cycle.
// By default, errors are logged and not returned over this channel.
// If you want to implement any custom error handling, set your config's
// Consumer.Return.Errors setting to true, and read from this channel.
func (g *ConsumerGroup) Errors() <-chan error {
	return g.Client.Errors()
}

type ConsumerGroupHandler interface {
	sarama.ConsumerGroupHandler
}

type consumerGroupHandler struct {
	handler            ConsumeGroupHandler
	topics             []string
	channameBufferSize int
	version            string
	assignor           string
	ready              chan bool
	Stop               bool
}

func newConsumerGroupHandler(handler ConsumeGroupHandler) *consumerGroupHandler {
	return &consumerGroupHandler{handler: handler}
}

func (h *consumerGroupHandler) Setup(g sarama.ConsumerGroupSession) error {
	loggerx.Default().Info("kafkax 消费者初始化", zap.Any("Topics", h.topics), zap.String("MemberID", g.MemberID()))
	//h.ready <- true
	close(h.ready)
	return nil
}
func (h *consumerGroupHandler) Cleanup(g sarama.ConsumerGroupSession) error {
	loggerx.Default().Warn("kafkax 消费者退出", zap.Any("Topics", h.topics), zap.String("MemberID", g.MemberID()))
	return nil
}
func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var err error
	defer func() {
		defer func() {
			e, ok := recover().(error)
			if ok {
				err = e
			}
		}()
	}()
	loggerx.Default().Info("kafkax 消费者开始监听", zap.String("topic", claim.Topic()), zap.Any("partition", claim.Partition()))
	for m := range claim.Messages() {
		if h.Stop {
			break
		}
		var (
			e    error
			mark bool
		)
		msg := &ConsumerMessage{m}
		e, mark = h.handler(msg)
		if mark {
			sess.MarkMessage(m, "")
		}
		if e != nil {
			loggerx.Default().Error("卡夫卡消费业务报错", zap.Error(e))
		}
	}
	return err
}
