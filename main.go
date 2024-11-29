package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"freessl-go/pkg"
	"freessl-go/pkg/parse_config"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"
)

var (
	wg              sync.WaitGroup
	cfg             parse_config.Config
	err             error
	instanceLocks   sync.Map
	RunPath, _      = os.Getwd()
	ProviderSupport = map[string]func(cfg parse_config.Provider) (challenge.Provider, error){
		"aliDNS": func(cfg parse_config.Provider) (challenge.Provider, error) {
			_ = os.Setenv("ALICLOUD_ACCESS_KEY", cfg.AccessKey)
			_ = os.Setenv("ALICLOUD_SECRET_KEY", cfg.SecretKey)
			provider, err := alidns.NewDNSProvider()
			if err != nil {
				return nil, err
			}
			return provider, nil
		},
		//"dnsPod": func(cfg parse_config.Provider) (challenge.Provider, error) {
		//	_ = os.Setenv("DNSPOD_API_KEY", cfg.AccessKey)
		//	provider, err := dnspod.NewDNSProvider()
		//	if err != nil {
		//		return nil, err
		//	}
		//	return provider, nil
		//},
		"tencentCloud": func(cfg parse_config.Provider) (challenge.Provider, error) {
			_ = os.Setenv("TENCENTCLOUD_SECRET_ID", cfg.AccessKey)
			_ = os.Setenv("TENCENTCLOUD_SECRET_KEY", cfg.SecretKey)
			provider, err := tencentcloud.NewDNSProvider()
			if err != nil {
				return nil, err
			}
			return provider, nil
		},
	}
)

func InstanceTask(instance parse_config.Provider, lock *sync.Mutex) {
	lock.Lock()
	defer wg.Done()
	defer lock.Unlock()
	var CertName string
	var KeyName string
	if instance.SaveSSLName == "" {
		sslBaseName := instance.Domains[0]
		CertName = sslBaseName + ".pem"
		KeyName = sslBaseName + ".key"
	} else {
		CertName = instance.SaveSSLName + ".pem"
		KeyName = instance.SaveSSLName + ".key"
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	CertBasePath := path.Join(RunPath, cfg.CertStoragePath)
	err = pkg.DirExists(CertBasePath)
	if err != nil {
		log.Fatal(err)
	}
	myCert := &pkg.CertParams{
		CertPath:       path.Join(CertBasePath, CertName),
		KeyPath:        path.Join(CertBasePath, KeyName),
		RenewBeforeDay: instance.RenewBeforeDay,
		Domains:        instance.Domains,
		Key:            privateKey,
		Email:          instance.Email,
	}
	myCert.Log("证书存储路径: ", CertBasePath)
	dnsProviderFunc, ok := ProviderSupport[instance.ProviderName]
	if !ok {
		myCert.Log("不支持的provider：", instance.ProviderName)
		return
	}
	if myCert.Valid() {
		return
	}
	dnsProvider, status := dnsProviderFunc(instance)
	if status != nil {
		myCert.Log("创建dns provider失败: ", status)
		return
	}

	dnsProviderParams := pkg.DnsProviderParams{
		ProviderApiKey: pkg.ProviderApiKey{
			AccessKey: instance.AccessKey,
			SecretKey: instance.SecretKey,
		},
		Provider: dnsProvider,
	}
	if myCert.NewObtain(dnsProviderParams) {
		if instance.Hook != "" {
			myCert.Log("执行hook：%s", instance.Hook)
			cmd := exec.Command(instance.Hook)
			output, err := cmd.Output()
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			myCert.Log("hook输出：", output)
		}
		return
	} else {
		myCert.Log("颁发证书失败, 退出")
		return
	}
}

func main() {
	log.Println("start time: ", time.Now().UTC())
	cfg, err = parse_config.ParseConfig()
	if err != nil {
		log.Fatal("加载配置文件错误.")
	}
	for _, instance := range cfg.Providers {
		if !instance.Enable {
			log.Printf("跳过域名托管: %+v", instance.Domains)
			continue
		}
		lock, _ := instanceLocks.LoadOrStore(instance.Domains[0], &sync.Mutex{})
		instanceLock := lock.(*sync.Mutex)
		wg.Add(1)
		go InstanceTask(instance, instanceLock)
	}
	wg.Wait()
	os.Exit(0)
}
