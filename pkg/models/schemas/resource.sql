create table resource
(
    id        int primary key auto_increment,
    report_id int  not null,
    status    enum ('skipped', 'changed', 'failed', 'unchanged') not null,
    name      text not null,
    type      text not null,
    file      text not null,
    line      int  not null,
    constraint resource_report_id_fk
        foreign key (report_id) references report (id)
);

