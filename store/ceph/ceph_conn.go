package ceph

import (
	//"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

var (
	cephConn *s3.S3
)

func GetCephConnection() *s3.S3 {
	// 判断之前是否已经初始化过了
	if cephConn != nil {
		return cephConn
	}

	// 1. 初始化ceph的一些信息
	//aws.Auth{}
	// 2. 创建S3类型的连接
	return cephConn
}
