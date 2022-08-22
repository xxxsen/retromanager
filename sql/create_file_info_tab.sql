create file_info_tab (
    id bigint unsigned not null auto_increment,
    file_name varchar(256) not null,
    hash char(32) not null,
    file_size bigint unsigned not null, 
    create_time bigint unsigned not null,
    down_key char(16) unsigned not null,
    primary key(id),
    unique idx_downkey(down_key)
) ENGINE=InnoDB CHARACTER SET utf8mb4;