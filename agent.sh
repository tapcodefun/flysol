#!/bin/bash

# 定义日志文件路径
LOG_FILE="install.log"

# 清空或创建日志文件
> "$LOG_FILE"

# 打印日志函数
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# 更新包列表
log "开始更新包列表..."
sudo apt-get update | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "包列表更新成功！"
else
    log "包列表更新失败！"
    exit 1
fi

# 安装lsof
log "开始安装 lsof..."
sudo DEBIAN_FRONTEND=noninteractive apt-get -o Dpkg::Options::="--force-confold" install -y lsof | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "lsof 安装成功！"
else
    log "lsof 安装失败！"
    exit 1
fi

# 安装unzip
log "开始安装 unzip..."
sudo DEBIAN_FRONTEND=noninteractive apt-get -o Dpkg::Options::="--force-confold" install -y unzip | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "unzip 安装成功！"
else
    log "unzip 安装失败！"
    exit 1
fi

# 安装jq
log "开始安装 jq..."
sudo DEBIAN_FRONTEND=noninteractive apt-get -o Dpkg::Options::="--force-confold" install -y jq | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "jq 安装成功！"
else
    log "jq 安装失败！"
    exit 1
fi

# 下载 rust-mev-bot-1.0.5.zip
log "开始下载 rust-mev-bot-1.0.5.zip..."
cd /home && wget https://sourceforge.net/projects/rust-mev-bot/files/rust-mev-bot-1.0.5.zip -O rust-mev-bot-1.0.5.zip | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "rust-mev-bot-1.0.5.zip 下载成功！"
else
    log "rust-mev-bot-1.0.5.zip 下载失败！"
    exit 1
fi

# 解压 rust-mev-bot-1.0.5.zip
log "开始解压 rust-mev-bot-1.0.5.zip..."
cd /home && unzip -o rust-mev-bot-1.0.5.zip -d bot | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "rust-mev-bot-1.0.5.zip 解压成功！"
else
    log "rust-mev-bot-1.0.5.zip 解压失败！"
    exit 1
fi

# 进入 /home/bot 目录
log "进入 /home/bot 目录..."
cd /home/bot || { log "无法进入 /home/bot 目录！"; exit 1; }

# 重命名配置文件
log "重命名 config.yaml.example 为 config.yaml..."
mv config.yaml.example config.yaml | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "配置文件重命名成功！"
else
    log "配置文件重命名失败！"
    exit 1
fi

# 赋予 run.sh 执行权限
log "赋予 run.sh 执行权限..."
chmod +x run.sh | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "run.sh 执行权限设置成功！"
else
    log "run.sh 执行权限设置失败！"
    exit 1
fi

# 赋予 upgrade.sh 执行权限
log "赋予 upgrade.sh 执行权限..."
chmod +x upgrade.sh | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "upgrade.sh 执行权限设置成功！"
else
    log "upgrade.sh 执行权限设置失败！"
    exit 1
fi

# 下载并安装 yq
log "开始下载并安装 yq..."
wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "yq 下载成功！"
else
    log "yq 下载失败！"
    exit 1
fi

log "赋予 yq 执行权限..."
chmod +x /usr/local/bin/yq | tee -a "$LOG_FILE"
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    log "yq 执行权限设置成功！"
else
    log "yq 执行权限设置失败！"
    exit 1
fi

log "所有操作已完成！日志已保存到 $LOG_FILE。"