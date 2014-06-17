create table treatment_plan_parent (
	treatment_plan_id int unsigned not null,
	parent_id int unsigned not null,
	parent_type varchar(100) not null,
	primary key (treatment_plan_id),
	foreign key (treatment_plan_id) references treatment_plan(id) on delete cascade
) character set utf8;

insert into treatment_plan_parent (treatment_plan_id, parent_id, parent_type) 
	select treatment_plan.id, treatment_plan.patient_visit_id, 'PATIENT_VISIT' from treatment_plan;

create table treatment_plan_patient_visit_mapping (
	treatment_plan_id int unsigned not null,
	patient_visit_id int unsigned not null,
	foreign key (treatment_plan_id) references treatment_plan(id) on delete cascade,
	foreign key (patient_visit_id) references patient_visit(id),
	primary key (treatment_plan_id,patient_visit_id)
) character set utf8;


insert into treatment_plan_patient_visit_mapping (treatment_plan_id, patient_visit_id)
	select treatment_plan.id, treatment_plan.patient_visit_id from treatment_plan;

alter table treatment_plan drop foreign key treatment_plan_ibfk_1;
alter table treatment_plan drop column patient_visit_id;

create table treatment_plan_content_source (
	treatment_plan_id int unsigned not null,
	content_source_id int unsigned not null,
	content_source_type varchar(100) not null,
	doctor_id int unsigned not null,
	has_deviated tinyint(1) not null default 0,
	deviated_date timestamp(6),
	primary key (treatment_plan_id),
	foreign key (treatment_plan_id) references treatment_plan(id) on delete cascade,
	foreign key (doctor_id) references doctor(id)
) character set utf8;

insert into treatment_plan_content_source (treatment_plan_id, doctor_id, content_source_id, content_source_type) 
	select treatment_plan_favorite_mapping.treatment_plan_id, dr_favorite_treatment_plan.doctor_id, treatment_plan_favorite_mapping.dr_favorite_treatment_plan_id, 'FAVORITE_TREATMENT_PLAN' 
	from treatment_plan_favorite_mapping 
	inner join dr_favorite_treatment_plan on dr_favorite_treatment_plan.id = dr_favorite_treatment_plan_id;

drop table treatment_plan_favorite_mapping;

