package main

import (
	"commentservice/conf"
	"commentservice/rpc/comment/pb"
	"commentservice/server"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/registry/etcd"
)

var (
	CommentConf *conf.Conf
	err         error
)

func main() {
	CommentConf, err = conf.LoadYaml(conf.CommentConfPath)
	if err != nil {
		panic(err)
	}

	err = server.InitService(CommentConf)
	if err != nil {
		panic(err)
	}

	etcdRegistry := etcd.NewRegistry(func(options *registry.Options) {
		options.Addrs = CommentConf.Etcd.Addr
	})

	service := micro.NewService(
		micro.Name(CommentConf.Grpc.Name),
		micro.Address(CommentConf.Grpc.Addr),
		micro.Registry(etcdRegistry),
	)
	service.Init()
	err = comment_service.RegisterCommentServerHandler(
		service.Server(),
		new(server.CommentService),
	)
	if err != nil {
		panic(err)
	}
	err = service.Run()
	if err != nil {
		panic(err)
	}
}
