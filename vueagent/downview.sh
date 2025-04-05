#!/bin/bash

# 设置工作目录
WORK_DIR="/home"
cd $WORK_DIR

# 1. 删除本地的 dist 文件夹（如果存在）
echo "正在删除旧的 dist 文件夹..."
if [ -d "dist" ]; then
  rm -rf dist
  echo "已删除旧的 dist 文件夹"
else
  echo "dist 文件夹不存在，跳过删除步骤"
fi

# 2. 下载远程的 agent.zip
echo "正在下载 agent.zip..."
wget -O agent.zip https://down.tapcode.work/agent.zip
if [ $? -ne 0 ]; then
  echo "下载失败，请检查网络连接或下载地址"
  exit 1
else
  echo "agent.zip 下载成功"
fi

# 3. 解压 agent.zip
echo "正在解压 agent.zip..."
unzip -o agent.zip
if [ $? -ne 0 ]; then
  echo "解压失败，请检查 agent.zip 文件是否损坏"
  exit 1
else
  echo "agent.zip 解压成功"
fi

# 清理下载的 zip 文件
rm agent.zip
echo "清理完成，安装/升级过程已完成"

exit 0