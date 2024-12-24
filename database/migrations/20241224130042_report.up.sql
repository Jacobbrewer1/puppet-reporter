create table report
(
    id             int auto_increment,
    hash           text           not null unique,
    host           text           not null,
    puppet_version decimal(10, 2) not null,
    environment    text           not null,
    state          text           not null,
    executed_at    datetime       not null,
    runtime        int            not null,
    failed         int            not null,
    changed        int            not null,
    skipped        int            not null,
    total          int            not null,
    primary key (id)
);

