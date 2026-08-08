QQ 空间历史恢复工具：普通用户使用说明

一、下载

请从 GitHub 仓库的 Releases 页面下载 Windows ZIP：
https://github.com/wilderjett250-art/006-qzone-history-recovery/releases/tag/v0.0.4

二、打开

1. 右键 ZIP，选择“解压全部”。
2. 打开解压后的文件夹。
3. 双击 qzone-history-gui.exe。
4. 浏览器会自动打开工具页面。

三、恢复记录

1. 点击“获取/刷新二维码”。
2. 用手机 QQ 扫描二维码并确认登录。
3. 选择要查看的年份。
4. 点击开始恢复，等待页面显示完成。
5. 点击结果入口查看恢复的时间线。

四、重要提醒

- 只用于本人或已经获得授权的 QQ 空间。
- 程序只在本机运行，登录会话和恢复结果保存在 EXE 所在文件夹。
- session.db、app.db、*_export.json、*_activities.json、*_view.html 可能包含个人信息，不要上传或发给别人。
- Windows 第一次运行如果提示“未知发布者”，确认下载来源后选择“更多信息 → 仍要运行”。

项目来源

本项目是基于 ZHChen2000/qzone-history 的下游整理版本。核心恢复逻辑属于上游项目；本仓库补充了源码清理、原理注释、中文/英文说明和 Windows 构建验证。
