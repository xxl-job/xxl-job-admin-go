#!/bin/bash
# GLUE-Shell 任务示例：在「任务管理」新建任务，GLUE 类型选 Shell，粘贴本脚本。
# 调度中心把脚本下发执行器，执行器用 bash 运行；$1 为任务参数。
echo "[$(date '+%F %T')] GLUE Shell 开始，参数: $1"
echo "当前主机: $(hostname)"
echo "磁盘使用:"; df -h | head -3
echo "[$(date '+%F %T')] GLUE Shell 结束"
exit 0
