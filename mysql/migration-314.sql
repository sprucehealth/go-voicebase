alter table bank_account modify column verify_expires timestamp not null default current_timestamp;