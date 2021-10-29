create database if not exists testname default character set utf8mb4;
create user if not exists `testuser`@'%' identified with mysql_native_password by 'testpass';
grant all on `testname`.* to 'testuser'@`%`;

use testname;

create table if not exists `test_result`
(
  `challid`     int               not null,
  `name`        varchar(255)      not null,
  `result`      int               not null,
  `timestamp`   datetime           not null
);
