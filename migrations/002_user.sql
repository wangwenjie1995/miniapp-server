create table if not exists users (
    id bigint unsigned not null auto_increment primary key,
    username varchar(24) not null unique,
    password_hash varchar(72) not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp
)