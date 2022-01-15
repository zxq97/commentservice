package server

import (
	"commentservice/global"
	"context"
	"github.com/jinzhu/gorm"
)

func dbLikePoint(ctx context.Context, articleID, uid int64) error {
	like := LikePoint{
		ArticleID: articleID,
		UID:       uid,
	}
	likeCount := LikeCount{
		ArticleID: articleID,
		LikeCount: 1,
	}
	tx := dbCli.Begin()
	defer tx.Rollback()
	err := dbCli.Create(like).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbLikePoint articleid %v uid %v err %v", ctx, articleID, uid, err)
		return err
	}
	err = dbCli.Model(&likeCount).Where("article_id = ", articleID).Update("like_count", gorm.Expr("like_count + 1")).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbaddlikecount articleid %v uid %v err %v", ctx, articleID, uid, err)
		return err
	}
	tx.Commit()
	return nil
}

func dbCancelPoint(ctx context.Context, articleID, uid int64) error {
	likeCount := LikeCount{
		ArticleID: articleID,
		LikeCount: 1,
	}
	tx := dbCli.Begin()
	defer tx.Rollback()
	err := dbCli.Where("article_id = ? and uid = ?", articleID, uid).Delete(&LikePoint{}).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbCancelPoint articleid %v uid %v err %v", ctx, articleID, uid, err)
		return err
	}
	err = dbCli.Model(&likeCount).Where("article_id = ", articleID).Update("like_count", gorm.Expr("like_count - 1")).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbreducelikecount articleid %v uid %v err %v", ctx, articleID, uid, err)
		return err
	}
	tx.Commit()
	return nil
}

func dbGetLikeState(ctx context.Context, articleIDs []int64, uid int64) (map[int64]bool, error) {
	likes := []*LikePoint{}
	err := slaveCli.Model(&LikePoint{}).Where("article_id in (?) amd uid = ?", articleIDs, uid).Find(&likes).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbGetLikeState articleids %v uid %v err %v", ctx, articleIDs, uid, err)
		return nil, err
	}
	okMap := make(map[int64]bool, len(articleIDs))
	for _, v := range likes {
		okMap[v.ArticleID] = true
	}
	return okMap, nil
}

func dbGetLikeCount(ctx context.Context, articleIDs []int64) (map[int64]int64, error) {
	counts := []*LikeCount{}
	err := slaveCli.Model(&LikeCount{}).Where("article_id in (?)", articleIDs).Find(&counts).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbGetLikeCount articleids %v err %v", ctx, articleIDs, err)
		return nil, err
	}
	cntMap := make(map[int64]int64, len(articleIDs))
	for _, v := range counts {
		cntMap[v.ArticleID] = v.LikeCount
	}
	return cntMap, nil
}

func dbGetCommentCount(ctx context.Context, articleIDs []int64) (map[int64]int64, error) {
	counts := []*CommentCount{}
	err := slaveCli.Model(&CommentCount{}).Where("article_id in (?)", articleIDs).Find(&counts).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbGetLikeCount articleids %v err %v", ctx, articleIDs, err)
		return nil, err
	}
	cntMap := make(map[int64]int64, len(articleIDs))
	for _, v := range counts {
		cntMap[v.ArticleID] = v.CommentCount
	}
	return cntMap, nil
}

func dbPublishComment(ctx context.Context, comment *Comment) error {
	commentCount := CommentCount{
		ArticleID:    comment.ArticleID,
		CommentCount: 1,
	}
	tx := dbCli.Begin()
	defer tx.Rollback()
	err := dbCli.Create(comment).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbPublishComment comment %#v err %v", ctx, comment, err)
		return err
	}
	err = dbCli.Model(&commentCount).Where("article_id = ", comment.ArticleID).Update("comment_count", gorm.Expr("comment_count + 1")).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbaddlikecomment articleid %v uid %v err %v", ctx, comment.ArticleID, comment.UID, err)
		return err
	}
	tx.Commit()
	return nil
}

func dbDeleteComment(ctx context.Context, commentID int64) error {
	err := dbCli.Where("comment_id = ? or p_comment_id = ?", commentID, commentID).Delete(&Comment{}).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbDeleteComment commentid %v err %v", ctx, commentID, err)
	}
	return err
}

func dbGetComment(ctx context.Context, commentIDs []int64) (map[int64]*Comment, error) {
	comments := []*Comment{}
	err := slaveCli.Model(&Comment{}).Where("comment_id in (?)", commentIDs).Find(&comments).Error
	if err != nil {
		global.ExcLog.Printf("ctx %v dbGetComment commentids %v err %v", ctx, comments, err)
		return nil, err
	}
	comMap := make(map[int64]*Comment, len(comments))
	for _, v := range comments {
		comMap[v.CommentID] = v
	}
	return comMap, nil
}
