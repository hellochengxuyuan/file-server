package mq

import (
	"fileStore_server/config"
	"github.com/streadway/amqp"
	"log"
)

var (
	conn    *amqp.Connection
	channel *amqp.Channel
)

// 如果异常关闭，会接收通知
var notifyClose chan *amqp.Error

func init() {
	// 是否开启异步转移功能，开启时才初始化rabbitMQ连接
	if !config.AsyncTransferEnable {
		return
	}

	if initChannel() {
		channel.NotifyClose(notifyClose)
	}
	// 断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

func initChannel() bool {
	var (
		err error
	)

	// 1.判断channel是否已经创建过
	if channel != nil {
		return true
	}

	// 2.获得rabbitmq的一个连接
	if conn, err = amqp.Dial(config.RabbitURL); err != nil {
		log.Println(err.Error())
		return false
	}

	// 3.打开一个channel(，用于消息的发布与接收等
	if channel, err = conn.Channel(); err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

// 发布消息
func Publish(exchange string, routingKey string, msg []byte) bool {
	var (
		err error
	)

	// 1. 判断channel是否正常
	if !initChannel() {
		return false
	}

	// 2. 执行消息发布动作
	// 第三个参数是如果没有合适的队列，是否把这消息返回个消息生产者，如果false会被丢弃
	err = channel.Publish(
		exchange,
		routingKey,
		false, // 如果没有对应的queue，就丢弃这条
		false,
		amqp.Publishing{
			ContentType: "text/plain", // 普通文本格式
			Body:        msg,
		})
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}
