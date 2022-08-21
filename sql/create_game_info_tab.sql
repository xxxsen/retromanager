create game_info_tab (
    id bigint unsigned not null auto_increment,
    platform int unsigned not null,
    display_name varchar(256) not null,
    file_name varchar(256) not null,
    file_size bigint unsigned not null,
    desc text not null,
    create_time bigint unsigned not null,
    update_time bigint unsigned not null,
    hash char(32) not null,
    extinfo blob not null,
    primary key(id),
    key idx_hash(hash),
    key idx_platform_createtime(platform, create_time)
) ENGINE=InnoDB CHARACTER SET utf8mb4;