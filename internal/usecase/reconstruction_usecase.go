package usecase

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"qzone-history/internal/domain/entity"
	"qzone-history/internal/domain/repository"
	"qzone-history/internal/domain/usecase"
	"strings"
)

type reconstructionUseCase struct {
	activityRepo     repository.ActivityRepository
	momentRepo       repository.MomentRepository
	boardMessageRepo repository.BoardMessageRepository
}

func NewReconstructionUseCase(
	activityRepo repository.ActivityRepository,
	momentRepo repository.MomentRepository,
	boardMessageRepo repository.BoardMessageRepository,
) usecase.ReconstructionUseCase {
	return &reconstructionUseCase{
		activityRepo:     activityRepo,
		momentRepo:       momentRepo,
		boardMessageRepo: boardMessageRepo,
	}
}

func (u *reconstructionUseCase) ReconstructMomentsFromActivities(ctx context.Context, userQQ string) error {
	// 已删除说说本体可能不再出现在说说接口中，但点赞、评论、浏览或转发活动仍会
	// 携带内容、参与者和时间。重建就是把这些事件碎片按同一条说说聚合成可浏览记录。
	activities, err := u.activityRepo.FindByUserQQ(ctx, userQQ, -1, 0)
	if err != nil {
		return err
	}

	// 每个聚合对象先保存一份“最可信/最完整”的基础字段，随后把其它事件提供的
	// 点赞数、浏览数、评论和更完整的正文或图片合并进来。
	momentMap := make(map[string]*entity.Moment)

	for _, activity := range activities {
		if activity.Type == entity.TypeBoardMessage || activity.Type == entity.TypeBoardReply {
			continue
		}
		switch activity.Type {
		case entity.TypeMoment, entity.TypeLike, entity.TypeView, entity.TypeComment, entity.TypeForward:
		default:
			continue
		}

		momentKey := generateMomentKey(activity)
		moment, ok := momentMap[momentKey]
		if !ok {
			moment = &entity.Moment{
				ID:              momentKey,
				SenderQQ:        activity.SenderQQ,
				UserQQ:          activity.ReceiverQQ,
				Content:         activity.Content,
				Timestamp:       activity.Timestamp,
				TimeText:        activity.TimeText,
				ImageURLs:       activity.ImageURLs,
				IsReconstructed: true,
			}
			momentMap[momentKey] = moment
		}

		switch activity.Type {
		case entity.TypeMoment:
			updateMomentFromActivity(moment, activity)
		case entity.TypeLike:
			moment.Likes++
		case entity.TypeView:
			moment.Views++
		case entity.TypeComment:
			comment := reconstructCommentFromActivity(activity)
			moment.Comments = append(moment.Comments, comment)
		}
	}

	for _, moment := range momentMap {
		if err := u.momentRepo.UpsertMoment(ctx, *moment); err != nil {
			return err
		}
	}

	return nil
}

func generateMomentKey(activity entity.Activity) string {
	// QQ 空间活动记录没有稳定的原始说说 ID。内容与接收者组合在同一空间内通常稳定，
	// 再哈希为固定长度键可用于数据库主键和跨扫描去重。代价是同一用户重复发布完全
	// 相同正文时可能被合并，这是缺少原始 ID 时的保守折中。
	key := activity.Content + activity.ReceiverQQ
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

func updateMomentFromActivity(moment *entity.Moment, activity entity.Activity) {
	// 多个活动碎片的信息质量不同：保留更长正文、最早时间和更丰富图片集合，
	// 而不是让后到但信息更少的事件覆盖已经恢复出的内容。
	if len(activity.Content) > len(moment.Content) {
		moment.Content = activity.Content
	}
	if activity.Timestamp.Before(moment.Timestamp) || moment.Timestamp.IsZero() {
		moment.Timestamp = activity.Timestamp
		moment.TimeText = activity.TimeText
	}
	if len(activity.ImageURLs) > len(moment.ImageURLs) {
		moment.ImageURLs = activity.ImageURLs
	}
}

func reconstructCommentFromActivity(activity entity.Activity) entity.Comment {
	return entity.Comment{
		UserQQ:    activity.SenderQQ,
		Content:   activity.Content,
		Timestamp: activity.Timestamp,
		TimeText:  activity.TimeText,
	}
}

func (u *reconstructionUseCase) ReconstructBoardMessagesFromActivities(ctx context.Context, userQQ string) error {
	activities, err := u.activityRepo.FindByUserQQ(ctx, userQQ, -1, 0)
	if err != nil {
		return err
	}

	existing, err := u.boardMessageRepo.FindByUserQQ(ctx, userQQ, -1, 0)
	if err != nil {
		return err
	}

	// 先把留言板 API 的正式记录建立索引，再用活动流候选补缺。这样能以官方列表为
	// 主数据源，同时保留 API 已不可见但活动中仍有痕迹的旧留言。
	apiKeys := make(map[string]struct{}, len(existing))
	for _, msg := range existing {
		apiKeys[boardDedupKey(msg)] = struct{}{}
	}

	added := 0
	for _, activity := range activities {
		if !isBoardActivity(activity) {
			continue
		}
		candidate := reconstructBoardMessageFromActivity(activity)
		key := boardDedupKey(candidate)
		if _, ok := apiKeys[key]; ok {
			continue
		}
		_ = candidate.BeforeCreate(nil)
		if err := u.boardMessageRepo.Insert(ctx, candidate); err != nil {
			return err
		}
		apiKeys[key] = struct{}{}
		added++
	}

	_ = added
	return nil
}

func isBoardActivity(a entity.Activity) bool {
	return a.Type == entity.TypeBoardMessage || a.Type == entity.TypeBoardReply
}

func reconstructBoardMessageFromActivity(activity entity.Activity) entity.BoardMessage {
	content := strings.TrimSpace(activity.Content)
	if activity.Type == entity.TypeBoardReply && content != "" {
		content = "[回复] " + content
	}
	return entity.BoardMessage{
		UserQQ:     activity.ReceiverQQ,
		SenderQQ:   activity.SenderQQ,
		SenderName: activity.SenderName,
		Content:    content,
		Timestamp:  activity.Timestamp,
		TimeText:   activity.TimeText,
	}
}

func boardDedupKey(m entity.BoardMessage) string {
	// 留言活动同样缺少跨接口统一 ID，因此使用发送者、日期和正文前缀构造近似键。
	// 截断正文可以控制索引大小；加入日期可降低常见短句在不同时间被误合并的概率。
	day := m.TimeText
	if !m.Timestamp.IsZero() {
		day = m.Timestamp.Format("2006-01-02")
	}
	norm := strings.TrimSpace(m.Content)
	if len(norm) > 80 {
		norm = norm[:80]
	}
	return m.SenderQQ + "|" + day + "|" + norm
}
