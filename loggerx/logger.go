package loggerx

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// DefaultLevel the default log level
	DefaultLevel = zapcore.InfoLevel

	// DefaultTimeLayout the default time layout;
	DefaultTimeLayout = "2006-01-02 15:04:05"
)

type LoggersInfo struct {
	// loggers.Level日志级别
	Level string `toml:"Level" json:"Level"`
	// loggers.File 项目日志存放文件
	File string `toml:"File" json:"File"`
	// loggers.ErrFile 项目错误日志存放文件
	ErrFile string `toml:"ErrFile" json:"ErrFile"`
	// 时间风格
	TimeLayout string `toml:"TimeLayout" json:"TimeLayout"`
	// MaxSize  单个文件最大尺寸，默认单位 M
	MaxSize int `toml:"MaxSize" json:"MaxSize"`
	// MaxBackups 最多保留 多少个备份
	MaxBackup int `toml:"MaxBackup" json:"MaxBackup"`
	// MaxAge 备份的最大天数
	MaxAge int `toml:"MaxAge" json:"MaxAge"`
	// 是否开启终端日志
	DisableConsole bool `toml:"DisableConsole" json:"DisableConsole"`
	// 日志域, 标明所属项目. 使用项目名称, 服务名称, 监听地址等参数合成.
	domain string
}

func (l *LoggersInfo) SetDomain(domain string) {
	l.domain = domain
}

func (l *LoggersInfo) GetDomain() string {
	return l.domain
}

// Option custom setup config
type Option func(*option)

type option struct { //nolint:maligned
	level          zapcore.Level
	fields         map[string]string
	logFile        io.Writer
	errorFile      io.Writer
	timeLayout     string
	disableConsole bool
}

// WithLevel 设置日志级别.
func WithLevel(level string) Option {
	var optCtl Option
	switch strings.ToLower(level) {
	case "debug":
		optCtl = func(opt *option) {
			opt.level = zapcore.DebugLevel
		}
	case "info":
		optCtl = func(opt *option) {
			opt.level = zapcore.InfoLevel
		}
	case "warn":
		optCtl = func(opt *option) {
			opt.level = zapcore.WarnLevel
		}
	case "error":
		optCtl = func(opt *option) {
			opt.level = zapcore.ErrorLevel
		}
	default:
		optCtl = func(opt *option) { opt.level = DefaultLevel }
	}
	return optCtl
}

// WithField 添加一些公共字段到日志.
func WithField(key, value string) Option {
	return func(opt *option) {
		opt.fields[key] = value
	}
}

// WithLogFile 使用日志文件切割.
func WithLogFile(file string, maxSize int, maxBackups int, maxAge int) Option {
	if file != "" {
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, 0766); err != nil {
			panic(err)
		}
	}
	return func(opt *option) {
		if file != "" {
			opt.logFile = &lumberjack.Logger{ // concurrent-safe
				Filename:   file,       // 文件路径
				MaxSize:    maxSize,    // 单个文件最大尺寸，默认单位 M
				MaxBackups: maxBackups, // 最多保留 多少个备份
				MaxAge:     maxAge,     // 备份的最大天数
				LocalTime:  true,       // 使用本地时间
				Compress:   true,       // 是否压缩 disabled by default
			}
		}
	}
}

// WithErrorFile 独立的error日志.
func WithErrorFile(errFile string, maxSize int, maxBackups int, maxAge int) Option {
	if errFile != "" {
		dir := filepath.Dir(errFile)
		if err := os.MkdirAll(dir, 0766); err != nil {
			panic(err)
		}
	}

	return func(opt *option) {
		if errFile != "" {
			opt.errorFile = &lumberjack.Logger{ // concurrent-safe
				Filename:   errFile,    // 文件路径
				MaxSize:    maxSize,    // 单个文件最大尺寸，默认单位 M
				MaxBackups: maxBackups, // 最多保留 多少个备份
				MaxAge:     maxAge,     // 备份的最大天数
				LocalTime:  true,       // 使用本地时间
				Compress:   true,       // 是否压缩 disabled by default
			}
		}
	}
}

// WithTimeLayout 时间格式化风格
func WithTimeLayout(timeLayout string) Option {
	return func(opt *option) {
		opt.timeLayout = timeLayout
	}
}

// WithConsole 关闭终端日志, 即日志不再向 os.Stdout 和 os.Stderr 输出
func WithConsole(disable bool) Option {
	return func(opt *option) {
		opt.disableConsole = disable
	}
}

var defaultLogger *zap.Logger

func Default() *zap.Logger {
	return defaultLogger
}

func SetDefault(logger *zap.Logger) {
	defaultLogger = logger
}

func New(logInfo *LoggersInfo) (*zap.Logger, error) {
	// 日志级别
	logLevel := WithLevel(logInfo.Level)
	console := WithConsole(logInfo.DisableConsole)
	// 日志文件
	logFile := WithLogFile(logInfo.File, logInfo.MaxSize, logInfo.MaxBackup, logInfo.MaxAge)
	// error日志文件, 如果没有, 则用logfile 加 error后缀.
	if logInfo.ErrFile == "" && logInfo.File != "" {
		logInfo.ErrFile = strings.Trim(logInfo.File, ".log") + "error.log"
	}
	errFile := WithErrorFile(logInfo.ErrFile, logInfo.MaxSize, logInfo.MaxBackup, logInfo.MaxAge)
	timeLayout := WithTimeLayout(logInfo.TimeLayout)
	domain := WithField("domain", logInfo.GetDomain())
	logger, err := NewJSONLogger(logLevel, logFile, errFile, console, timeLayout, domain)
	// 第一次初始化设为默认值, 如果多次初始化, 需要自己设置默认值
	if defaultLogger == nil {
		SetDefault(logger)
	}
	return logger, err
}

// NewJSONLogger 使用json日志
func NewJSONLogger(opts ...Option) (*zap.Logger, error) {
	opt := &option{level: DefaultLevel, fields: make(map[string]string)}
	for _, f := range opts {
		f(opt)
	}

	timeLayout := DefaultTimeLayout
	if opt.timeLayout != "" {
		timeLayout = opt.timeLayout
	}

	// similar to zap.NewProductionEncoderConfig()
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "loggerx", // used by loggerx.Named(key); optional; useless
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace", // use by zap.AddStacktrace; optional; useless
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder, // 大写日志级别
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(timeLayout))
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 采用短文件路径输出
	}

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// lowPriority usd by info\debug\warn
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl < zapcore.ErrorLevel
	})

	// highPriority usd by error\panic\fatal
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl >= zapcore.ErrorLevel
	})

	stdout := zapcore.Lock(os.Stdout) // lock for concurrent safe
	stderr := zapcore.Lock(os.Stderr) // lock for concurrent safe

	core := zapcore.NewTee()

	// 开启终端日志
	if !opt.disableConsole {
		core = zapcore.NewTee(
			zapcore.NewCore(jsonEncoder,
				zapcore.NewMultiWriteSyncer(stdout),
				lowPriority,
			),
			zapcore.NewCore(jsonEncoder,
				zapcore.NewMultiWriteSyncer(stderr),
				highPriority,
			),
		)
	}

	if opt.logFile != nil {
		core = zapcore.NewTee(core,
			zapcore.NewCore(jsonEncoder,
				zapcore.AddSync(opt.logFile),
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl >= opt.level
				}),
			),
		)
	}
	if opt.errorFile != nil {
		core = zapcore.NewTee(core,
			zapcore.NewCore(jsonEncoder,
				zapcore.AddSync(opt.errorFile),
				highPriority,
			),
		)
	}

	logger := zap.New(core,
		zap.AddCaller(),
		zap.ErrorOutput(stderr),
	)

	for key, value := range opt.fields {
		logger = logger.WithOptions(zap.Fields(zapcore.Field{Key: key, Type: zapcore.StringType, String: value}))
	}
	return logger, nil
}

var _ Meta = (*meta)(nil)

// Meta 一个包装zip.Field的键值对
type Meta interface {
	Key() string
	Value() interface{}
	meta()
}

type meta struct {
	key   string
	value interface{}
}

func (m *meta) Key() string {
	return m.key
}

func (m *meta) Value() interface{} {
	return m.value
}

func (m *meta) meta() {}

// NewMeta create meat
func NewMeta(key string, value interface{}) Meta {
	return &meta{key: key, value: value}
}

// WrapMeta 用来包装zap.Field
func WrapMeta(err error, metas ...Meta) (fields []zap.Field) {
	capacity := len(metas) + 1 // namespace meta
	if err != nil {
		capacity++
	}

	fields = make([]zap.Field, 0, capacity)
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	// 创建一个命名空间, 避免一些字段重名造成冲突
	fields = append(fields, zap.Namespace("meta"))
	for _, m := range metas {
		fields = append(fields, zap.Any(m.Key(), m.Value()))
	}

	return
}
