package server

import (
	"commentservice/conf"
	"commentservice/rpc/comment/pb"
	"context"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type CommentService struct {
}

var (
	mcCli    *memcache.Client
	redisCli redis.Cmdable
	dbCli    *gorm.DB
	slaveCli *gorm.DB
)

func InitService(config *conf.Conf) error {
	var err error
	mcCli = conf.GetMC(config.MC.Addr)
	redisCli = conf.GetRedisCluster(config.RedisCluster.Addr)
	dbCli, err = conf.GetGorm(fmt.Sprintf(conf.MysqlAddr, config.Mysql.User, config.Mysql.Password, config.Mysql.Host, config.Mysql.Port, config.Mysql.DB))
	if err != nil {
		return err
	}
	slaveCli, err = conf.GetGorm(fmt.Sprintf(conf.MysqlAddr, config.Slave.User, config.Slave.Password, config.Slave.Host, config.Slave.Port, config.Slave.DB))
	return err
}

func (cs *CommentService) GetCommentList(ctx context.Context, req *comment_service.GetCommentListRequest, res *comment_service.GetCommentListResponse) error {
	return nil
}

func (cs *CommentService) PublishComment(ctx context.Context, req *comment_service.PublishCommentRequest, res *comment_service.EmptyResponse) error {
	err := publishComment(ctx, req.CommentInfo.CommentId, req.CommentInfo.PCommentId, req.CommentInfo.ArticleId, req.CommentInfo.Uid, req.CommentInfo.ReplyUid, req.CommentInfo.Content)
	return err
}

func (cs *CommentService) DeleteComment(ctx context.Context, req *comment_service.PublishCommentRequest, res *comment_service.EmptyResponse) error {
	err := deleteComment(ctx, req.CommentInfo.CommentId, req.CommentInfo.PCommentId, req.CommentInfo.ArticleId)
	return err
}

func (cs *CommentService) GetCommentCount(ctx context.Context, req *comment_service.GetCountRequest, res *comment_service.GetCountResponse) error {
	cntMap, err := getCommentCount(ctx, req.ArticleIds)
	if err != nil {
		return err
	}
	res.LikeCount = cntMap
	return nil
}

func (cs *CommentService) GetLikeCount(ctx context.Context, req *comment_service.GetCountRequest, res *comment_service.GetCountResponse) error {
	cntMap, err := getLikeCount(ctx, req.ArticleIds)
	if err != nil {
		return err
	}
	res.LikeCount = cntMap
	return nil
}

func (cs *CommentService) GetLikeState(ctx context.Context, req *comment_service.GetLikeStateRequest, res *comment_service.GetLikeStateResponse) error {
	oks, err := getLikeState(ctx, req.ArticleIds, req.Uid)
	if err != nil {
		return err
	}
	res.Ok = oks
	return nil
}
func (cs *CommentService) LikePoint(ctx context.Context, req *comment_service.LikePointRequest, res *comment_service.EmptyResponse) error {
	err := likePoint(ctx, req.LikeInfo.ArticleId, req.LikeInfo.Uid)
	return err
}

func (cs *CommentService) CancelLike(ctx context.Context, req *comment_service.LikePointRequest, res *comment_service.EmptyResponse) error {
	err := cancelLike(ctx, req.LikeInfo.ArticleId, req.LikeInfo.Uid)
	return err
}
