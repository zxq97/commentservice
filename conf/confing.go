package conf

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"
)

const (
	ApiConfPath     = "./conf/yaml/api.yaml"
	ArticleConfPath = "./conf/yaml/article.yaml"
	ASyncConfPath   = "./conf/yaml/async.yaml"
	CommentConfPath = "./conf/yaml/comment.yaml"
	RemindConfPath  = "./conf/yaml/remind.yaml"
	SocialConfPath  = "./conf/yaml/social.yaml"
	UserConfPath    = "./conf/yaml/user.yaml"

	MysqlAddr = "%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True"
)

type MysqlConf struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DB       string `yaml:"db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type RedisConf struct {
	Addr string `yaml:"addr"`
	DB   int    `yaml:"db"`
}

type MCConf struct {
	Addr string `yaml:"addr"`
}

type GrpcConf struct {
	Addr string `yaml:"addr"`
	Name string `yaml:"name"`
}

type EtcdConf struct {
	Addr string `yaml:"addr"`
}

type KafkaConf struct {
	Addr string `yaml:"addr"`
}

type Conf struct {
	Mysql MysqlConf `yaml:"mysql"`
	Redis RedisConf `yaml:"redis"`
	MC    MCConf    `yaml:"MC"`
	Grpc  GrpcConf  `yaml:"grpc"`
	Etcd  EtcdConf  `yaml:"etcd"`
	Kafka KafkaConf `yaml:"kafka"`
}

func LoadYaml(path string) (*Conf, error) {
	conf := new(Conf)
	y, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(y, conf)
	return conf, err
}

func GetMC(addr string) *memcache.Client {
	return memcache.New(addr)
}

func GetRedis(addr string, db int) *redis.Client {
	return redis.NewClient(
		&redis.Options{
			Addr: addr,
			DB:   db,
		})
}

func GetMysql(addr string) (sqlbuilder.Database, error) {
	dsn, err := mysql.ParseURL(addr)
	if err != nil {
		return nil, err
	}
	r, err := mysql.Open(dsn)
	if err != nil {
		return nil, err
	}
	return r, nil
}
