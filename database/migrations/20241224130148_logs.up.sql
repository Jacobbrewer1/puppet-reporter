create table log_message
(
    id        int auto_increment,
    report_id int  not null,
    message   text not null,
    primary key (id),
    constraint log_message_report_id_fk
        foreign key (report_id) references report (id)
);

