package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"qzone-history/internal/domain/entity"
	"qzone-history/internal/domain/usecase"
	"qzone-history/pkg/loghub"
	"qzone-history/pkg/offset"
	"qzone-history/pkg/paths"
	"time"
)

type App struct {
	authUseCase           usecase.AuthUseCase
	momentUseCase         usecase.MomentUseCase
	boardMessageUseCase   usecase.BoardMessageUseCase
	friendUseCase         usecase.FriendUseCase
	exportUseCase         usecase.ExportUseCase
	activityUseCase       usecase.ActivityUseCase
	reconstructionUseCase usecase.ReconstructionUseCase
}

func NewApp(
	authUseCase usecase.AuthUseCase,
	momentUseCase usecase.MomentUseCase,
	boardMessageUseCase usecase.BoardMessageUseCase,
	friendUseCase usecase.FriendUseCase,
	exportUseCase usecase.ExportUseCase,
	activityUseCase usecase.ActivityUseCase,
	reconstructionUseCase usecase.ReconstructionUseCase,
) *App {
	return &App{
		authUseCase:           authUseCase,
		momentUseCase:         momentUseCase,
		boardMessageUseCase:   boardMessageUseCase,
		friendUseCase:         friendUseCase,
		exportUseCase:         exportUseCase,
		activityUseCase:       activityUseCase,
		reconstructionUseCase: reconstructionUseCase,
	}
}

func (a *App) Auth() usecase.AuthUseCase {
	return a.authUseCase
}

func (a *App) RunPipeline(ctx context.Context, user *entity.User, opts RunOptions) error {
	hub := loghub.Default()
	hub.SetStatus(func(s *loghub.Status) {
		s.Running = true
		s.Done = false
		s.Error = ""
		s.UserQQ = user.QQ
		s.TargetYear = opts.TargetYear
		s.MaxOffset = opts.MaxOffset
		s.Phase = "准备中"
		s.ProgressPercent = 0
	})
	_, estMax := offset.EstimateScan(opts.TargetYear, opts.MaxOffset)
	hub.BeginTimedProgress(estMax)
	progressCtx, stopProgress := context.WithCancel(ctx)
	defer stopProgress()
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-progressCtx.Done():
				return
			case <-ticker.C:
				hub.TouchProgress()
			}
		}
	}()

	if _, err := paths.EnsureUserDir(user.QQ); err != nil {
		return logError("创建用户目录失败", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	// 恢复流程按“当前可见数据 -> 活动痕迹 -> 重建结果 -> 本地导出”的顺序执行。
	// 当前说说或留言接口失败时不立即终止，是因为活动流中仍可能保留已删除内容的点赞、
	// 评论和浏览痕迹；让后续阶段继续运行，通常比把某一个接口当作唯一真相更完整。
	hub.Log("登录成功，获取未删除说说...")
	hub.SetStatus(func(s *loghub.Status) { s.Phase = "获取未删除说说" })
	count, err := a.momentUseCase.ImportVisibleMoments(ctx, *user)
	if err != nil && !errors.Is(err, context.Canceled) {
		hub.Logf("获取未删除说说失败（将继续尝试活动记录）: %v", err)
	} else if err == nil {
		hub.Logf("已导入 %d 条未删除说说", count)
	}
	if err := ctx.Err(); err != nil {
		hub.Log("任务已停止")
		return err
	}

	// 活动记录是重建已删除说说的核心证据源。它不是一份完整的说说快照，而是同一内容
	// 在不同时间产生的事件碎片，因此这里先尽可能深扫并持久化，下一阶段再统一聚合。
	hub.Log("获取活动记录（用于恢复已删说说）...")
	hub.SetStatus(func(s *loghub.Status) { s.Phase = "扫描活动记录" })
	_, err = a.activityUseCase.FetchActivities(ctx, *user, opts.MaxOffset, opts.TargetYear)
	if errors.Is(err, context.Canceled) {
		hub.Log("任务已停止")
		return err
	}
	if err != nil {
		hub.Logf("获取活动记录失败（将继续导出已有数据）: %v", err)
	}
	if err := ctx.Err(); err != nil {
		hub.Log("任务已停止")
		return err
	}

	hub.Log("获取留言板消息...")
	hub.SetStatus(func(s *loghub.Status) { s.Phase = "获取留言板" })
	boardCount, err := a.boardMessageUseCase.ImportBoardMessages(ctx, *user)
	if err != nil && !errors.Is(err, context.Canceled) {
		hub.Logf("获取留言板失败（将继续导出已有数据）: %v", err)
	} else if err == nil {
		hub.Logf("已导入 %d 条留言板消息", boardCount)
	}
	if err := ctx.Err(); err != nil {
		hub.Log("任务已停止")
		return err
	}

	// 重建阶段只读取已经落盘的活动记录。这样网络抓取与数据推断相互解耦：
	// 即使后续导出失败或用户重新运行，已经抓到的历史证据仍可复用。
	hub.Log("开始数据重建过程...")
	hub.SetStatus(func(s *loghub.Status) { s.Phase = "重建数据" })
	err = a.reconstructionUseCase.ReconstructMomentsFromActivities(ctx, user.QQ)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			hub.Log("任务已停止")
			return err
		}
		return logError("重建 Moments 失败", err)
	}
	if err := ctx.Err(); err != nil {
		hub.Log("任务已停止")
		return err
	}

	// 留言板接口返回的记录优先保留，活动记录只负责补缺。两路数据最终通过稳定的
	// 去重键合并，避免同一条留言既来自官方列表又来自“与我相关”时重复出现。
	if boardCount == 0 {
		hub.Log("留言板 API 无数据，将从活动记录补充留言")
	} else {
		hub.Logf("留言板 API 已导入 %d 条，继续从活动记录合并补充", boardCount)
	}
	err = a.reconstructionUseCase.ReconstructBoardMessagesFromActivities(ctx, user.QQ)
	if err != nil {
		return logError("合并留言板失败", err)
	}
	hub.Log("留言板已与活动记录合并完成")

	hub.Log("导出用户数据到 JSON 格式...")
	hub.SetStatus(func(s *loghub.Status) { s.Phase = "导出 JSON" })
	err = a.exportUseCase.ExportUserDataToJSON(ctx, user.QQ)
	if err != nil {
		return logError("导出用户数据到 JSON 失败", err)
	}

	hub.Log("导出 HTML 浏览页...")
	hub.SetStatus(func(s *loghub.Status) { s.Phase = "导出 HTML" })
	err = a.exportUseCase.ExportUserDataToHTML(ctx, user.QQ)
	if err != nil {
		return logError("导出 HTML 浏览页失败", err)
	}

	hub.Log("导出原始活动记录到 JSON 格式...")
	hub.SetStatus(func(s *loghub.Status) { s.Phase = "导出活动记录" })
	err = a.exportActivitiesToJSON(ctx, user.QQ)
	if err != nil {
		return logError("导出活动记录到 JSON 失败", err)
	}

	viewer := paths.ViewerHTMLPath(user.QQ)
	hub.SetStatus(func(s *loghub.Status) {
		s.Running = false
		s.Done = true
		s.Phase = "完成"
		s.ProgressPercent = 100
		s.ViewerPath = viewer
	})
	hub.Logf("全部完成！浏览页: %s", viewer)
	return nil
}

func (a *App) exportActivitiesToJSON(ctx context.Context, userQQ string) error {
	activities, err := a.activityUseCase.GetAllActivities(ctx, userQQ)
	if err != nil {
		return err
	}
	filename := paths.ActivitiesJSONPath(userQQ)
	data, err := json.MarshalIndent(activities, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func logError(message string, err error) error {
	hub := loghub.Default()
	if err != nil {
		hub.Logf("%s: %v", message, err)
		hub.SetStatus(func(s *loghub.Status) {
			s.Running = false
			s.Error = message + ": " + err.Error()
			s.Phase = "失败"
		})
	} else {
		hub.Log(message)
		hub.SetStatus(func(s *loghub.Status) {
			s.Running = false
			s.Error = message
			s.Phase = "失败"
		})
	}
	return err
}

func WaitQRLogin(ctx context.Context, auth usecase.AuthUseCase, qrsig string) (*entity.User, error) {
	hub := loghub.Default()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		status, res, err := auth.CheckQRCodeLoginStatus(ctx, qrsig)
		if err != nil {
			return nil, err
		}
		switch status {
		case entity.LoginStatusSuccess:
			hub.Log("二维码扫描成功，完成登录...")
			user, err := auth.CompleteLogin(ctx, res)
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Second)
			return user, nil
		case entity.LoginStatusScanning:
			hub.Log("二维码认证中...")
		case entity.LoginStatusExpired:
			return nil, errQRExpired
		}
		time.Sleep(2 * time.Second)
	}
}

var errQRExpired = &qrExpiredError{}

type qrExpiredError struct{}

func (e *qrExpiredError) Error() string { return "二维码已失效，请刷新" }
