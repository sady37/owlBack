部署环境
JDK version：Jdk1.8
OS : linux (centos7 or ubuntu18)
DB : mysql5.7以上
MQTT broker: mosquitto v1.6
Redis：5.0.8
使用到的端口
port	Remarks
8090	Http port 
1883	Mosquitto(mqtt) port
29010	TCP地址服务端口。在给设备配网时使用
28070	TCP通信服务端口

安装redis
请参考https://redis.io/docs/getting-started/installation/install-redis-on-linux/
安装mqtt中间件mosquitto
请参考 https://mosquitto.org/

mosquitto 最新版本是2.x，但推荐使用1. 6版本
安装redis
请自行查询安装方式
导入数据库
将下列sql文件依次导入mysql：
tb_data_struct.sql 
pro_user_struct.sql
pro_version_struct.sql
z1_manage_struct.sql
tb_system.sql
sdk_config.sql
device_information.sql （注意：设备信息SQL，需支持人员另外发送。）
日志中如有报错“'sleepace tb data.sleep data original' doesn't exist”请执行以下SQL：

CREATE DATABASE sleepace_pro_data;
USE sleepace_pro_data;

CREATE TABLE `sleep_data_original` (
  `seqId` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) DEFAULT NULL,
  `filePath` varchar(255) DEFAULT NULL,
  `startTime` int(11) DEFAULT NULL,
  `createTime` datetime DEFAULT NULL,
  PRIMARY KEY (`seqId`)
) ENGINE=InnoDB AUTO_INCREMENT=387 DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;

日志中如有报错 `Unknown column 'channelId' in 'field list'`（AlarmTask:179 BadSqlGrammarException），表示 `sleepace_tb_system.alert_record` 缺 `channelId` 字段，AlarmTask 写库失败 → 整个 alarmNotify 推送链路停摆（alarmLeftBed/alarmSitup/alarmHeartRate* 等都不会推 MQTT）。补字段：

```sql
USE sleepace_tb_system;
ALTER TABLE `alert_record` ADD COLUMN `channelId` INT(10) DEFAULT 0;
```
安装Sleepace服务 
1、将sleepace-service.zip解压到你要安装的目录
2、修改 sleepace-service/classes/config.properties 
A.	将 ’pro.tcp.ip’ 属性修改为当前服务器IP
 
B.	配置appId和seretKey(对应开发阶段sleepace提供的服务器配置 )
 
C.	如果需要设备一直上报实时数据（实时的心率、呼吸率），需要将channelId配置到auto.realtime.businessId中

 

3、修改sleepace-service/classes/application.properties配置文件
A.	修改数据库配置
将 DB_IP 和 DB_Port修改为mysql的IP和端口
将BUserName and DBPassword 修改为MySql的账号和密码
 
B.	修改mqtt 连接配置
根据mqtt broker安装信息，修改url、账号和密码
 

C.	修改redis配置
	如果你的redis使用哨兵模式安装，修改
 

	如果你的redis为单机，请修改
     
	如果你的redis为集群，请修改
 
4、设置服务日志保存时间
log.delete 是否自动删除过期日志，默认false
log.expire  当log.delete为true时，日志保存多少天，默认30天

 

5、修改 ‘sleepace-service/bin’ 下的脚本权限
chmod 755 *.sh
 


启动sleepace服务
执行 ‘sleepace-service/bin/’下的startup.sh
 

如果你在‘sleepace-service/logs/server.log’看到一下日志，说明服务启动成功了


 
关闭sleepace服务
执行 ‘sleepace-service/bin/’下的 shutdown.sh

 
如何将设备导入数据库
设备信息需要先到如数据库才能使用。I导入方式有两种:
1.	在MySql直接导入我们提供的SQL文件
2.	通过后台系统导入
目前后台系统还没开发完成，所以只能在MySql中直接导入我们提供的SQL文件

连接设备
配网时将设备连接到以下端口（config.properties文件中）。
pro.address.tcp.port=29010


备注：
升级设备
由于您手上的设备固件版本可以不是最新的，建议下载最新固件包（ https://www.yuque.com/sleepace/elder/kpwsruaclvm1oc7c ），并通过固件上传接口将固件上传。固件上传后，当您的绑定设备或设备重新上线后，默认情况下，服务器会将您的设备升级到您上传的固件版本（您也可以通过调用升级接口升级设备）


兼容非24小时监测设备
本服务主要支持24小时检测报告设备（如：BM80701、m901等），如果你接入非24小时检测报告的设备（如：Z400twp-3）,需要执行一下操作来进行兼容：
1、将sleepace_pro_data.sql导入mysql中
2、再application.properties中reportAuto的数据库注释，并正确配置
 
