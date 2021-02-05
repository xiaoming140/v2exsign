####改
# 白嫖 Github Action 自动 v2ex 签到

## 使用说明
Fork 本项目，点击你 Fork 后的仓库右上角的 settings，点击其中的 secrets。

点击 New secret，Name 填 v2exCookie，Value 填 v2ex 的 cookie。

cookie 可通过如下方法获取。

打开 v2ex 的页面，按下 f12，点击 network，刷新一次页面，复制其中 cookie 中的所有内容。

![](https://i.loli.net/2020/10/20/zxf34BjosKPeXCM.png)

之后点击仓库上方的 Actions，点击`I understand my workflows, go ahead and enable them`。

再点击左侧的 `Auto v2ex sign`，再点击 `Enable workflow`

最后打开 README.md 任意编辑一次，提交。

然后就会在每天的 8 点和 20 点尝试签到，签到失败或者 cookie 失效大概会发送邮件给你吧。

还可以额外添加一个 sckey，添加后签到成功就会通过微信推送。

sckey 来这获取 https://sc.ftqq.com/3.version
