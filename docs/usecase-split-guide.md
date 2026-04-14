# 用例图拆分说明

为了便于在 PPT 中展示，已经将原来的总体用例图拆分成 5 张较小的图。

## 文件对应关系

1. [usecase-public.puml](D:\desk\软件工程\SEUOJ\docs\usecase-public.puml)
- 适合放在“公开访问 / 基础浏览”这一页
- 包含：题库、题目、题解、比赛、公告、论坛、排行榜

2. [usecase-judge.puml](D:\desk\软件工程\SEUOJ\docs\usecase-judge.puml)
- 适合放在“在线评测主链路”这一页
- 包含：注册登录、运行样例、提交代码、查看提交

3. [usecase-contest.puml](D:\desk\软件工程\SEUOJ\docs\usecase-contest.puml)
- 适合放在“比赛需求”这一页
- 包含：报名比赛、比赛题目、比赛提交、榜单、比赛公告、比赛管理

4. [usecase-teaching-split.puml](D:\desk\软件工程\SEUOJ\docs\usecase-teaching-split.puml)
- 适合放在“教学需求”这一页
- 包含：题单、班级、作业/考试、题解管理

5. [usecase-community-admin.puml](D:\desk\软件工程\SEUOJ\docs\usecase-community-admin.puml)
- 适合放在“社区与平台治理”这一页
- 包含：论坛发帖回帖、论坛管理、题目管理、用户管理、公告管理、重判、统计

## 如果只想放 2 张图

建议：
- 一张放 [usecase-judge.puml](D:\desk\软件工程\SEUOJ\docs\usecase-judge.puml)
- 一张放 [usecase-teaching-split.puml](D:\desk\软件工程\SEUOJ\docs\usecase-teaching-split.puml)

这样能覆盖“训练评测 + 教学管理”两个最有区分度的部分。

## 如果想保留总图

总图仍然保留在：
- [usecase-overall.puml](D:\desk\软件工程\SEUOJ\docs\usecase-overall.puml)

适合用于：
- 先总览系统
- 再用拆分图逐页展开
