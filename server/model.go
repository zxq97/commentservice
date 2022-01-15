package server

type Comment struct {
	ArticleID  int64  `json:"article_id"`
	CommentID  int64  `json:"comment_id"`
	UID        int64  `json:"uid"`
	PCommentID int64  `json:"p_comment_id"`
	ReplyUID   int64  `json:"reply_uid"`
	Content    string `json:"content"`
}

type CommentCount struct {
	ArticleID    int64 `json:"article_id"`
	CommentCount int64 `json:"comment_count"`
}

type LikePoint struct {
	ArticleID int64 `json:"article_id"`
	UID       int64 `json:"uid"`
}

type LikeCount struct {
	ArticleID int64 `json:"article_id"`
	LikeCount int64 `json:"like_count"`
}

func (t *Comment) TableName() string {
	return "comment"
}

func (t *CommentCount) TableName() string {
	return "comment_count"
}

func (t *LikePoint) TableName() string {
	return "like_point"
}

func (t *LikeCount) TableName() string {
	return "like_count"
}

type CommentReply struct {
	Comment Comment `json:"comment"`
	Reply   Comment `json:"reply"`
}
