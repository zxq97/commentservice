package server

import (
	"commentservice/global"
	"commentservice/util/cast"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis"
	"time"
)

const (
	CommentReplyCount = 1

	McCommentTTl = 60 * 5
	McKeyComment = "comment_service_comment_%v" // comment_id

	RedisKeySArticleLikeList     = "comment_service_article_like_list_%v"     // article_id uid ctime
	RedisKeyArticleCommentCount  = "comment_service_article_comment_count_%v" // article_id count
	RedisKeyArticleLikeCount     = "comment_service_article_like_count_%v"    // article_id count
	RedisKeyZArticleComment      = "comment_service_comment_%v"               // article_id comment_id ctime
	RedisKeyZArticleCommentReply = "comment_service_comment_reply_%v"         // comment_id comment_id ctime
)

func cacheLikePoint(ctx context.Context, articleID, uid int64) error {
	key := fmt.Sprintf(RedisKeySArticleLikeList, articleID)
	res, err := redisCli.SAdd(key, uid).Result()
	if err != nil {
		global.ExcLog.Printf("ctx %v cacheLikePoint articleid %v uid %v err %v", ctx, articleID, err)
		return err
	}
	if res == 1 {
		key = fmt.Sprintf(RedisKeyArticleLikeCount, articleID)
		err = redisCli.Incr(key).Err()
		if err != nil {
			global.ExcLog.Printf("ctx %v cacheaddlikecount articleid %v uid %v err %v", ctx, articleID, uid, err)
			return err
		}
	}
	return nil
}

func cacheCancelPoint(ctx context.Context, articleID, uid int64) error {
	key := fmt.Sprintf(RedisKeySArticleLikeList, articleID)
	err := redisCli.SRem(key, uid).Err()
	if err != nil {
		global.ExcLog.Printf("ctx %v cacheCancelPoint articleid %v uid %v err %v", ctx, articleID, uid, err)
		return err
	}

	return err
}

func cacheGetLikeState(ctx context.Context, articleIDs []int64, uid int64) (map[int64]bool, []int64, error) {
	cmdMap := make(map[int64]*redis.BoolCmd)
	pipe := redisCli.Pipeline()
	for _, v := range articleIDs {
		key := fmt.Sprintf(RedisKeySArticleLikeList, v)
		cmdMap[v] = pipe.SIsMember(key, uid)
	}
	_, err := pipe.Exec()
	if err != nil && err != redis.Nil {
		global.ExcLog.Printf("ctx %v cacheGetLikeState articleids %v uid %v err %v", ctx, articleIDs, uid, err)
		return nil, articleIDs, err
	}
	okMap := make(map[int64]bool, len(articleIDs))
	missed := make([]int64, 0)
	for _, v := range articleIDs {
		okMap[v], err = cmdMap[v].Result()
		if err != nil {
			missed = append(missed, v)
		}
	}
	return okMap, missed, nil
}

func cacheGetCount(ctx context.Context, keyTmp string, articleIDs []int64) (map[int64]int64, []int64, error) {
	cmdMap := make(map[int64]*redis.StringCmd, len(articleIDs))
	pipe := redisCli.Pipeline()
	for _, v := range articleIDs {
		key := fmt.Sprintf(keyTmp, v)
		cmdMap[v] = pipe.Get(key)
	}
	_, err := pipe.Exec()
	if err != nil && err != redis.Nil {
		global.ExcLog.Printf("ctx %v keytmp %v cacheGetCount articleids %v err %v", ctx, keyTmp, articleIDs, err)
		return nil, articleIDs, err
	}
	cntMap := make(map[int64]int64, len(articleIDs))
	missed := make([]int64, 0)
	for _, v := range articleIDs {
		val, err := cmdMap[v].Result()
		if err != nil {
			missed = append(missed, v)
			continue
		}
		cntMap[v] = cast.ParseInt(val, 0)
	}
	return cntMap, missed, nil
}

func cachePublishComment(ctx context.Context, commentID, pCommentID, articleID int64) error {
	var (
		key string
		err error
	)
	z := redis.Z{
		Member: commentID,
		Score:  float64(time.Now().Unix()),
	}
	global.InfoLog.Printf("ctx %v cachepublishcomment commentid %v pcommentid %v articleid %v", ctx, commentID, pCommentID, articleID)
	if pCommentID == 0 {
		key = fmt.Sprintf(RedisKeyZArticleComment, articleID)
		err = redisCli.ZAdd(key, z).Err()
	} else {
		key = fmt.Sprintf(RedisKeyZArticleCommentReply, pCommentID)
		err = redisCli.ZAdd(key, z).Err()
	}
	global.InfoLog.Printf("ctx cachepublisherr %v", err)
	if err != nil {
		global.ExcLog.Printf("ctx %v cachePublishComment commentid %v pcommentid %v articleid %v err %v", ctx, commentID, pCommentID, articleID, err)
	}
	return err
}

func cacheSetComment(ctx context.Context, comment *Comment) error {
	buf, err := json.Marshal(comment)
	if err != nil {
		global.ExcLog.Printf("ctx %v marshal comment %#v err %v", ctx, comment, err)
		return err
	}
	global.InfoLog.Printf("ctx %v setcommentcache comment %#v", ctx, comment)
	key := fmt.Sprintf(McKeyComment, comment.CommentID)
	err = mcCli.Set(&memcache.Item{Key: key, Value: buf, Expiration: McCommentTTl})
	global.InfoLog.Printf("ctx %v setcacheerr %v", err)
	if err != nil {
		global.ExcLog.Printf("ctx %v cacheSeyComment comment %#v err %v", ctx, comment, err)
	}
	return err
}

func cacheDeleteComment(ctx context.Context, commentID, pCommentID, articleID int64) error {
	key := fmt.Sprintf(RedisKeyZArticleComment, articleID)
	rkey := fmt.Sprintf(RedisKeyZArticleCommentReply, pCommentID)
	var err error
	if pCommentID == 0 {
		pipe := redisCli.Pipeline()
		pipe.ZRem(key, commentID)
		pipe.Del(rkey)
		_, err = pipe.Exec()
	} else {
		err = redisCli.ZRem(rkey, commentID).Err()
	}
	if err != nil {
		global.ExcLog.Printf("ctx %v cachedeletecommentlist commentid %v pcommentid %v articleid %v err %v", ctx, commentID, pCommentID, articleID, err)
	}
	key = fmt.Sprintf(McKeyComment, commentID)
	err = mcCli.Delete(key)
	if err != nil {
		global.ExcLog.Printf("ctx %v cachedeletecomment commentid %v err %v", ctx, commentID, err)
	}
	return err
}

func cacheGetCommentList(ctx context.Context, articleID, cursor, offset int64) (map[int64][]int64, bool, error) {
	key := fmt.Sprintf(RedisKeyZArticleComment, articleID)
	val, err := redisCli.ZRevRange(key, cursor, cursor+offset).Result()
	if err != nil {
		global.ExcLog.Printf("ctx %v cacheGetCommentList articleid %v cursor %v offset %v err %v", ctx, articleID, cursor, offset, err)
		return nil, false, err
	}
	var hasMore bool
	if len(val) > int(offset) {
		hasMore = true
	}
	replyMap := make(map[int64][]int64, len(val))
	cmdMap := make(map[int64]*redis.StringSliceCmd, len(val))
	pipe := redisCli.Pipeline()
	for k, v := range val {
		if k == len(val)-1 {
			break
		}
		commendID := cast.ParseInt(v, 0)
		replyMap[commendID] = make([]int64, 0, 2)
		key = fmt.Sprintf(RedisKeyZArticleCommentReply, commendID)
		cmdMap[commendID] = pipe.ZRevRange(key, 0, CommentReplyCount)
	}
	_, err = pipe.Exec()
	if err != nil && err != redis.Nil {
		global.ExcLog.Printf("ctx %v cacheGetCommentList articleid %v cursor %v offset %v err %v", ctx, articleID, cursor, offset)
		return nil, false, err
	}
	for k, v := range cmdMap {
		val, err = v.Result()
		if err != nil && err != redis.Nil {
			global.ExcLog.Printf("ctx %v cachegetcommentreplylist commentid %v err %v", ctx, k, err)
			continue
		}
		for _, sid := range val {
			replyMap[k] = append(replyMap[k], cast.ParseInt(sid, 0))
		}
	}
	return replyMap, hasMore, nil
}

func cacheGetComment(ctx context.Context, commentIDs []int64) (map[int64]*Comment, []int64, error) {
	keys := make([]string, 0, len(commentIDs))
	for _, v := range commentIDs {
		keys = append(keys, fmt.Sprintf(McKeyComment, v))
	}
	res, err := mcCli.GetMulti(keys)
	if err != nil {
		return nil, commentIDs, err
	}
	commentMap := make(map[int64]*Comment, len(commentIDs))
	for _, v := range res {
		comment := Comment{}
		err = json.Unmarshal(v.Value, &comment)
		if err != nil {
			global.ExcLog.Printf("ctx %v cachegetcomment commentid %v err %v", ctx, v.Value, err)
			continue
		}
		commentMap[comment.CommentID] = &comment
	}
	missed := make([]int64, 0, len(commentIDs))
	for _, v := range commentIDs {
		if _, ok := commentMap[v]; !ok {
			missed = append(missed, v)
		}
	}
	return commentMap, missed, nil
}
