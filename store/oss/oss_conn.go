package oss

import (
	cfg "fileStore_server/config"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	ossCli *oss.Client
)

// 创建ossCli对象
func Client() *oss.Client {
	var (
		err error
	)
	if ossCli != nil {
		return ossCli
	}

	if ossCli, err = oss.New(cfg.OSSEndpoint, cfg.OSSAccesskeyID,
		cfg.OSSAccesskeySercet); err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return ossCli
}

// 获取bucket存储空间
func Bucket() *oss.Bucket {
	var (
		cli    *oss.Client
		bucket *oss.Bucket
		err    error
	)

	cli = Client()
	if cli != nil {
		if bucket, err = cli.Bucket(cfg.OSSBucket); err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return bucket
	}
	return nil
}

// 临时授权下载url
func DownloadURL(objName string) (signUrl string) {
	var (
		err error
	)

	// 有效时间3600秒
	if signUrl, err = Bucket().SignURL(objName, oss.HTTPGet, 3600); err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return signUrl
}

// BuildLifecycleRule : 针对指定bucket设置生命周期规则
func BuildLifecycleRule(bucketName string) {
	// 表示前缀为test的对象(文件)距最后修改时间30天后过期。
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest1}

	Client().SetBucketLifecycle(bucketName, rules)
}
