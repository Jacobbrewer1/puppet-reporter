alter table resource modify column status enum ('skipped', 'changed', 'failed', 'unchanged') not null;