# zabbix-saml-bypass-exp-
cve-2022-23131 exp

```
fofa: app="ZABBIX-监控系统" && body="saml" 
```

![image-20220218164224691](docs/image-20220218164224691.png)

1. replace [zbx_signed_session] to  [cookie] 
2. sign in with Single Sign-On (SAML)

![image-20220218164332289](docs/image-20220218164332289.png)

