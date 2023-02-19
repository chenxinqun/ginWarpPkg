package environment

import (
	"fmt"
	"strings"
)

var (
	active Environment
	dev    Environment = &environment{value: "dev"}
	test   Environment = &environment{value: "test"}
	pre    Environment = &environment{value: "pre"}
	pro    Environment = &environment{value: "pro"}
)

var _ Environment = (*environment)(nil)

// Environment 环境配置.
type Environment interface {
	Value() string
	IsDev() bool
	IsTest() bool
	IsPre() bool
	IsPro() bool
}

type environment struct {
	value            string
	businessPlatform string
}

func (e *environment) getBusinessPlatform() string {
	return e.businessPlatform
}
func (e *environment) setBusinessPlatform(value string) {
	e.businessPlatform = value
}

func (e *environment) Value() string {
	return e.value
}

// IsDev 开发环境.
func (e *environment) IsDev() bool {
	return e.value == "dev"
}

// IsTest 测试环境.
func (e *environment) IsTest() bool {
	return e.value == "test"
}

// IsPre 预发布环境.
func (e *environment) IsPre() bool {
	return e.value == "pre"
}

// IsPro 生产环境.
func (e *environment) IsPro() bool {
	return e.value == "pro"
}

func InitEnv(env string) {
	abc := strings.TrimSpace(env)
	switch strings.ToLower(abc) {
	case "dev":
		active = dev
	case "test":
		active = test
	case "pre":
		active = pre
	case "pro":
		active = pro
	default:
		if env != "" {
			active = &environment{value: env}
		} else {
			active = dev
			fmt.Println("警告: 命令行参数 '-env' 没找到, 已经设置默认值 'dev'.")
		}
	}
}

//BusinessPlatform 这个业务平台是后加的, 为了提升兼容性, 就没有把获取方法加在 Environment interface中. 而是加在 environment 包中.
func BusinessPlatform() string {
	return active.(*environment).getBusinessPlatform()
}

func InitBusinessPlatform(businessPlatform string) {
	if BusinessPlatform() == "" {
		active.(*environment).setBusinessPlatform(businessPlatform)
	}
}

// Active 当前配置的env
func Active() Environment {
	return active
}
