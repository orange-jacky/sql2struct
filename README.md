# sql2struct

>mysql 创建table sql语句, 直接生成go对应的struct, orm选用的是xorm
>支持sql文件中包含多个create table语句

>create table语句中字段类型, 只写了经常用到的, 可以在sql2struct.go文件的HFiledtype中添加自己需要的类型转换


#使用方式
>1.下载项目
>2.go run *.go  xxx.sql


#example: 

a.sql的内容是
```
create table tb_user_match_policy
(
   id                   int not null auto_increment,
   user_id              int not null comment '用户ID',
   policy_id            int not null comment '政策ID',
   publish_time         datetime comment '政策发布时间',
   primary key (id)
)
auto_increment = 10000;

alter table tb_user_match_policy comment '政策精准推送表';

/*==============================================================*/
/* Table: tb_user_type                                          */
/*==============================================================*/
create table tb_user_type
(
   id                   int not null comment '用户类型ID',
   name                 varchar(20) not null comment '用户类型名称',
   state                int default 1 comment '状态 0：无效, 1: 有效(默认)',
   primary key (id)
)
```

执行转换
go run *.go  a.sql
输出
```
type TbUserMatchPolicy struct { //tb_user_match_policy
	Id          int64     `json:"id" xorm:"id"`                     //
	UserId      int64     `json:"user_id" xorm:"user_id"`           //用户ID
	PolicyId    int64     `json:"policy_id" xorm:"policy_id"`       //政策ID
	PublishTime time.Time `json:"publish_time" xorm:"publish_time"` //政策发布时间
}


type TbUserType struct { //tb_user_type
	Id    int64  `json:"id" xorm:"id"`       //用户类型ID
	Name  string `json:"name" xorm:"name"`   //用户类型名称
	State int    `json:"state" xorm:"state"` //状态 0：无效, 1: 有效(默认)
}
```




