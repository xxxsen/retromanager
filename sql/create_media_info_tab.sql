create media_info_tab (
    id bigint unsigned not null auto_increment,
    file_name varchar(256) not null,
    hash varchar(32) not null,
    file_size bigint unsigned not null, 
    create_time bigint unsigned not null,
    file_type tinyint unsigned not null,
    primary key(id),
    unique idx_filetype_hash(file_type, hash)
) ENGINE=InnoDB CHARACTER SET utf8mb4;