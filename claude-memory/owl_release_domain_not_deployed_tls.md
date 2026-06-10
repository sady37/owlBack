---
name: owl_release_domain_not_deployed_tls
description: 2026-06-02 owlCare/web 登录失败真因=owl.wisefido.com(release)未部署+自签名证书兜底;test sandbox 正常
metadata: 
  node_type: memory
  type: project
  originSessionId: 47a4cb19-7f35-4bee-a27d-6dda608ef53e
---

2026-06-02 排查 "owlCare app/web 用 admin/Ts123@123 登录失败、没返回 tenant_name:demo、app 不报错但进不去"。

**真因(不是凭证/不是代码/不是后端):** `owl.wisefido.com`(release 域名)在 nginx 里**根本没有 server block**。它和 `test.wisefido.com` 同解析到 `47.77.194.143`(本机,NAT 后公网 IP),无显式 default_server → 按字母序最先加载的 `owl-web-care-ssl` 块成兜底 → 给 owl.wisefido.com 上**自签名证书**(`/etc/nginx/sites-available/owlcert/server.crt`,subject==issuer O=wisefido,SAN 只含 IP+test.wisefido.com,不含 owl)。自签名 + SAN 不匹配双重 TLS 失败 → 浏览器/iOS(Alamofire+ATS)握手失败 → 无响应 → "登录失败/进不去/无 tenant_name"。且该兜底块只 proxy `/api/v1/`→17217(老 v1),没接 v2 `/auth/api`,所以 release **整体未部署**不只是证书。

**鉴别点:** TLS 失败=无响应(curl HTTP 000 / err 60 unable to get local issuer cert),不是 code≠2000 的业务 Fail;凭证错才会 bump users.failed_login_count(admin=0=从没失败过)。

**对照健康环境:** `test.wisefido.com`(sandbox,用户称 **Alitest**=阿里云测试环境)= owlfront nginx 块 + Let's Encrypt 证书 + proxy owlback v2,实测 admin/Ts123@123 → code 2000 + tenant_name:demo。用户确认:**app 设计支持双域名(product/test),但当前只有 test.wisefido.com 真正在跑**,owl.wisefido.com(product)未部署 → 默认指 product 的客户端必然连不上。

**止血(已采纳):** 都切 test 用。Web 直接开 https://test.wisefido.com;owlCare app **登录页连点 Logo 6 次(2s 内)** 弹出服务器卡片 → 切 Test(slogan 圆点橙=test/绿=product),baseURL 存 UserDefaults serverAddressKey。app 默认 baseURL=https://owl.wisefido.com(product)是踩坑根源(AppConfig.swift:24 / LoginViewController.swift:35)。

**治本(待办,需用户确认是生产变更):** certbot 2.9.0 已装 → 签 owl.wisefido.com 证书 + 加 server_name owl.wisefido.com 的 nginx 块(镜像 owlfront,proxy 8080/v2)+ 部署 release 前端 + reload。release 用哪个前端构建/release 模式开关是用户决策。参 [[deploy_mode_by_domain]]。

附:本会话还给 owlFront [axios/index.ts](../../../owl/owlFront/src/utils/http/axios/index.ts) 补了 error 弹窗(原 errorMessageMode modal/message 是 TODO 桩只 console.error,登录失败界面无提示只抛 index.ts:179)——与本 TLS 根因无关,是顺手修的可见性缺口。
