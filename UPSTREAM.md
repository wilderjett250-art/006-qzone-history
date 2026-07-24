# 上游来源与修改记录

## 上游来源

- 项目：`ZHChen2000/qzone-history`
- 地址：https://github.com/ZHChen2000/qzone-history
- 基线提交：`666f8dd4e7fb3ad88248f7818e2f95c16f48adb6`
- 上游作者：ZHChen
- 许可证：Apache License 2.0

`LICENSE` 与上游文件 SHA256 一致。上游 `README.md` 内容已原样保存在 `README.upstream.md`。

## Project 006 修改

1. 排除上游仓库中的预编译 `qzone-history-gui.exe`。
2. 排除 `session.db`、个人 QQ 数据目录、恢复数据库和导出文件。
3. 在以下环节增加原理型中文注释：
   - 多策略活动深扫；
   - 跨接口事件去重；
   - 相对时间推断；
   - 已删除说说与留言重建；
   - Offset 推荐与耗时估算；
   - 本机 Web 控制台和任务取消机制。
4. 增加 Project 006 中英文 README。

这些修改用于提高源码可读性、可审计性和交付清晰度，没有改变恢复业务逻辑。
