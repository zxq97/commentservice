package server

import (
	comment_service "commentservice/rpc/comment/pb"
	"commentservice/util/concurrent"
	"context"
)

func likePoint(ctx context.Context, articleID, uid int64) error {
	err := dbLikePoint(ctx, articleID, uid)
	if err != nil {
		return err
	}
	concurrent.Go(func() {
		_ = cacheLikePoint(ctx, articleID, uid)
	})
	return nil
}

func cancelLike(ctx context.Context, articleID, uid int64) error {
	err := dbCancelPoint(ctx, articleID, uid)
	if err != nil {
		return err
	}
	err = cacheCancelPoint(ctx, articleID, uid)
	return err
}

func getLikeState(ctx context.Context, articleIDs []int64, uid int64) (map[int64]bool, error) {
	okMap, missed, err := cacheGetLikeState(ctx, articleIDs, uid)
	if err != nil || len(missed) != 0 {
		dbOkMap, err := dbGetLikeState(ctx, missed, uid)
		if err != nil {
			return nil, err
		}
		for k, v := range dbOkMap {
			okMap[k] = v
		}
	}
	return okMap, nil
}

func getLikeCount(ctx context.Context, articleIDs []int64) (map[int64]int64, error) {
	cntMap, missed, err := cacheGetCount(ctx, RedisKeyArticleLikeCount, articleIDs)
	if err != nil || len(missed) != 0 {
		dbMap, err := dbGetLikeCount(ctx, missed)
		if err != nil {
			return nil, err
		}
		for k, v := range dbMap {
			cntMap[k] = v
		}
	}
	return cntMap, nil
}

func getCommentCount(ctx context.Context, articleIDs []int64) (map[int64]int64, error) {
	cntMap, missed, err := cacheGetCount(ctx, RedisKeyArticleCommentCount, articleIDs)
	if err != nil || len(missed) != 0 {
		dbMap, err := dbGetCommentCount(ctx, missed)
		if err != nil {
			return nil, err
		}
		for k, v := range dbMap {
			cntMap[k] = v
		}
	}
	return cntMap, nil
}

func publishComment(ctx context.Context, commentID, pCommentID, articleID, uid, replyUID int64, content string) error {
	comment := Comment{
		CommentID:  commentID,
		PCommentID: pCommentID,
		ArticleID:  articleID,
		UID:        uid,
		ReplyUID:   replyUID,
		Content:    content,
	}
	err := dbPublishComment(ctx, &comment)
	if err != nil {
		return err
	}
	concurrent.Go(func() {
		_ = cachePublishComment(ctx, commentID, pCommentID, articleID)
	})
	concurrent.Go(func() {
		_ = cacheSetComment(ctx, &comment)
	})
	return nil
}

func deleteComment(ctx context.Context, commentID, pCommentID, articleID int64) error {
	err := dbDeleteComment(ctx, commentID)
	if err != nil {
		return err
	}
	err = cacheDeleteComment(ctx, commentID, pCommentID, articleID)
	return err
}

func getCommentList(ctx context.Context, articleID, cursor, offset int64) (map[int64]*comment_service.ReplyInfo, bool, error) {
	//replyMap, hasMore, err := cacheGetCommentList(ctx, articleID, cursor, offset)
	//if err != nil || replyMap == nil {
	//	// todo db查询
	//}
	//commentIDs := make([]int64, 0, len(replyMap)*3)
	//for k, v := range replyMap {
	//	commentIDs = append(commentIDs, k)
	//	for _, id := range v {
	//		commentIDs = append(commentIDs, id)
	//	}
	//}
	//commentMap, missend, err := cacheGetComment(ctx, commentIDs)
	//if err != nil || len(missend) != 0 {
	//	dbCommentMap, err := dbGetComment(ctx, missend)
	//	if err != nil {
	//		return nil, false, err
	//	}
	//	for k, v := range dbCommentMap {
	//		commentMap[k] = v
	//	}
	//}
	//res := make(map[int64]*comment_service.ReplyInfo, offset)
	//for k, v := range replyMap {
	//	res[k] =
	//}
	return nil, false, nil
}
