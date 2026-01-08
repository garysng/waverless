这是一些帮助数据方便在接下来开发过程中用来定位问题

当前后端服务部署在集群 wavespeed-test 命名空间下：waverless
前端服务在本地启动测试，通过 kubectl port-forward -n wavespeed-test svc/waverless-svc 8080:80 连接

1. 服务部署方式，本地修改完，我自行去其他环境打包部署环境.
2. 本地连接测试数据库查询数据的方式： /opt/homebrew/opt/mysql-client@8.0/bin/mysql -h sg-cdb-qmk985xh.sql.tencentcdb.com -P 63938 -u wavespeed -p'wavespeed&123!'  数据库是waverless
   