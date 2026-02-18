@echo off
title Claude via CC-Switch
echo Connecting to CC-Switch (Local Proxy)...

:: 1. 强制清除旧的登录凭证 (解决 Auth conflict 报错)
set "ANTHROPIC_AUTH_TOKEN="

:: 2. 指向 CC-Switch 的本地端口
:: 只要 CC Switch 开着，这个地址对所有模型都是通用的，不用改
set "ANTHROPIC_BASE_URL=http://127.0.0.1:15721/v1"

:: 3. 设置假 Key
:: 无论你用什么模型，CC Switch 都会在后台把它替换成真的 Key
set "ANTHROPIC_API_KEY=sk-generic-key"

:: 4. 启动 Claude
:: %* 允许你在命令行后面加参数，比如 "UniversalStart.bat /init"
call claude %*

:: 退出时暂停，方便看报错（如果不需要可以删掉这一行）
pause