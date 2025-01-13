alter table report modify column state enum ('SKIPPED', 'CHANGED', 'FAILED', 'UNCHANGED') not null;
