package mq

import (
	"github.com/streadway/amqp"
	"log"
)

var done chan bool

// 开始监听队列，获取消息
// 第3个参数为外部调用者指定的一个处理消息的callback函数
func StartConsume(qName string, cName string, callback func(msg []byte) bool) {
	var (
		//msgs       <-chan amqp.Delivery
		msg        amqp.Delivery
		err        error
		processSuc bool
	)

	// 1. 通过channel.Consume获得消息信道
	msgs, err := channel.Consume(
		qName,
		cName,
		true,  // 自动回复ack信息，为true则自动回复一个确认信号给发布者那边
		false, // 指定是否为唯一的消息消费者
		false, // rabbitMQ只能设置为fals
		false, // false表示会有阻塞直到有消息过来
		nil)

	if err != nil {
		log.Println(err.Error())
		return
	}

	done = make(chan bool)

	go func() {
		// 2. 循环获取队列的消息
		for msg = range msgs {
			// 3. 调用callback方法来处理新的消息
			processSuc = callback(msg.Body)
			if !processSuc {
				// TODO: 将任务写到另一个队列，用于异常情况的重试
			}
		}
	}()

	// done没有新的消息过来，则会一直阻塞
	<-done

	// 关闭channel通道
	channel.Close()
}
