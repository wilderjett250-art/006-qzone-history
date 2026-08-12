# 006 QQ 空间历史恢复 | Qzone History

> 把本人或已授权账号的 QQ 空间活动线索整理为可搜索的离线历史档案。
>
> **English:** A practical, runnable project with a documented workflow for the problem described above.

## 项目展示 / Demo

![工具界面](docs/images/gui-overview.png)

## 解决什么问题 / Problem

解决当前页面只显示部分内容、删除说说难以从活动记录中追溯以及历史数据难以离线保存的问题。

**English:** This project addresses the problem above with a reproducible local workflow.

## 有什么用 / Use

扫码登录后扫描活动流、留言板和可见内容，进行事件去重、内容重建，并导出 JSON/HTML。

**English:** Run the workflow locally, inspect the output, and extend the project from the provided source.

## 高光亮点 / Highlights

- 扫码登录和本地会话
- 活动流深度扫描与时间范围定位
- 点赞/评论/浏览/转发事件聚合
- Windows EXE、JSON 和离线 HTML

## 技术名词 / Tech

`Go · SQLite · Local Web UI · QQ Space API · HTML`

## 从 ZIP 开始复现 / Reproduce from ZIP

1. 从 Release 下载 ZIP 并解压。
2. 双击 qzone-history-gui.exe。
3. 点击获取二维码，用手机 QQ 扫码并确认。
4. 选择年份和扫描范围，开始恢复。
5. 在结果入口打开生成的 JSON 或 _view.html。
6. 只处理本人或明确授权的数据，输出文件不要上传仓库。

**Expected result:** 浏览器显示登录和扫描进度，完成后得到本地时间线和导出文件。

## 目录提示 / Notes

- 先阅读本 README，再按项目内更详细的中文/英文文档补充配置。
- 不要把真实密码、Token、数据库业务数据和本机运行结果提交回仓库。
- 下载 ZIP 后的第一次运行应使用测试数据或示例图片，确认链路正常后再接入自己的环境。

[English documentation](README.en.md)
