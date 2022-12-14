create table game_info_tab (
    id bigint unsigned not null auto_increment,
    platform tinyint unsigned not null,
    display_name varchar(256) not null,
    file_name varchar(256) not null,
    file_size bigint unsigned not null,
    detail text not null,
    create_time bigint unsigned not null,
    update_time bigint unsigned not null,
    hash char(32) not null,
    extinfo blob not null,
    down_key varchar(64) not null,
    state tinyint unsigned not null,
    primary key(id),
    unique key idx_hash(hash),
    key idx_state_platform_createtime(state, platform, create_time),
    key idx_updatetime(update_time)
) ENGINE=InnoDB CHARACTER SET utf8mb4;