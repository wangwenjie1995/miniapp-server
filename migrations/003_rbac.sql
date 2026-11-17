create table if not exists roles (
    id bigint unsigned not null auto_increment primary key,
    code varchar(64) not null unique,
    name varchar(64) not null unique,
    description varchar(255) not null default '',
    status tinyint unsigned not null default 1 comment '0 禁用。 1 启用', 
    is_builtin tinyint unsigned not null default 0 comment '0 自定义角色，1 系统内置角色',
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp
);

create table if not exists permissions (
    id bigint unsigned not null auto_increment primary key,
    code varchar(64) not null unique,
    name varchar(64) not null,
    module varchar(255) not null,
    description varchar(255) not null default '',
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp
);

create table if not exists user_roles (
    user_id bigint unsigned not null,
    role_id bigint unsigned not null,
    created_at timestamp default current_timestamp,

    primary key (user_id, role_id), 
    key idx_user_roles_role_id(role_id),
    
    constraint fk_user_roles_user
        foreign key (user_id) references users(id)
        on delete restrict,

    constraint fk_user_roles_role
        foreign key (role_id) references roles(id)
        on delete restrict
);

create table if not exists role_permissions (
    role_id bigint unsigned not null,
    permission_id bigint unsigned not null,
    created_at timestamp default current_timestamp,

    primary key (role_id, permission_id),
    key idx_role_permissions_permission_id(permission_id),

    constraint fk_role_permissions_role
        foreign key (role_id) references roles(id)
        on delete restrict,
    
    constraint fk_role_permissions_permission
        foreign key (permission_id) references permissions(id)
        on delete restrict
);

insert into roles (code, name, description, status, is_builtin)
    values ('user', '普通用户', '小程序普通用户', 1, 1);
insert into roles (code, name, description, status, is_builtin)
    values ('admin', '管理员', '小程序管理员:管理角色和权限', 1, 1);

insert into permissions (code ,name, module, description)
    values('user:list', '查看用户列表', 'user', '查看系统用户列表');

insert into role_permissions (role_id, permission_id)
    select roles.id, permissions.id
    from roles, permissions
    where roles.code = 'admin'
        and permissions.code = 'user:list'
        and not exists (
        select 1
        from role_permissions
        where role_permissions.role_id = roles.id
            and role_permissions.permission_id = permissions.id
        );


insert into user_roles (user_id, role_id)
    select users.id, roles.id
    from users, roles
    where users.username = 'root'
        and roles.code = 'admin'
        and not exists (
            select 1
            from user_roles
            where user_roles.user_id = users.id
                and user_roles.role_id = roles.id
        );

insert into user_roles (user_id, role_id)
    select users.id, roles.id
    from users, roles
    where roles.code = 'user'
    and not exists (
        select 1
        from user_roles
        where user_roles.user_id = users.id
            and user_roles.role_id = roles.id
    );