# go-gpt-client
基于GPT3.5的对话生成器的GO语言客户端实现，界面使用[Fyne](https://fyne.io/)实现，支持实时显示输出。

## 预览
![预览图](https://github.com/CxZMoE/go-gpt3-fyne/raw/main/preview/home.png)

## 使用方式
在下方输入框中输入内容，按下左Alt+Enter发送消息。

## 注意
如果闪退了，可能是因为网络问题，或者魔法失效了。  
可以看下程序目录下的.log文件，然后提issue，或者自己修改代码。  

`程序根目录下需要有以下文件： `

+ auth.json
``` json
// 实际使用需要把注释去掉
{
    "apiKey": "你的APIKEY",
    "model": "gpt-3.5-turbo", 
    "username": "你的用户名",
    "capacity": 20, //整型，保存的历史消息条数，超过这个数量会清除之前的，为了防止出现token超量的问题。
    "proxy":"http://127.0.0.1:7890" // 代理的地址，根据你的情况如果直连不行，就必须设置。
}
```

+ FONT.TTF 用于显示中文字体

## 编译
``` powershell
# 开启CGO功能
go env -w CGO_ENABLED=1

git clone github.com/CxZMoE/go-gpt-client
cd go-gpt-client

# 不显示控制台
go build -o go-gpt-client.exe -ldflags -H=windowsgui
# 显示控制台
go build -o go-gpt-client.exe 
```

## 贡献
欢迎fork本仓库，提交pull reqeust