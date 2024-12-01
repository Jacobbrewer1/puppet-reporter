create table log_message
(
    id        int primary key auto_increment,
    report_id int  not null,
    message   text not null,
    constraint log_message_report_id_fk
        foreign key (report_id) references report (id)
);

