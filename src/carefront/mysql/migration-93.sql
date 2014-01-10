start transaction;
update localized_text set ltext = 'Add Medication' where app_text_id = (select app_text_id from question_fields where question_field = 'add_text' and question_id = (select id from question where question_tag='q_current_medications_entry') );

insert into app_text (app_text_tag, comment) values ('txt_months_current_medication', 'question text for number of months medication has been taken for');
insert into localized_text (language_id, ltext, app_text_id) values (1, 'How many months have you been taking this medication for?', (select id from app_text where app_text_tag='txt_months_current_medication'));
insert into question (qtype_id, qtext_app_text_id, question_tag) values ((select id from question_type where qtype='q_type_segmented_control'), (select id from app_text where app_text_tag='txt_months_current_medication'), 'q_length_current_medication');
update question as q1 inner join question as q2 on q1.id = q2.id set q1.parent_question_id = q2.id where q2.question_tag='q_current_medications_entry';
insert into potential_answer (question_id, answer_localized_text_id, atype_id, potential_answer_tag, ordering, answer_summary_text_id, status) values ((select id from question where question_tag='q_length_current_medication'), (select id from app_text where app_text_tag='txt_used_less_1_month'), (select id from answer_type where atype='a_type_segmented_control'), 'a_length_current_medication_less_than_month',0, (select id from app_text where app_text_tag='txt_answer_summary_less_month'),'ACTIVE');
insert into potential_answer (question_id, answer_localized_text_id, atype_id, potential_answer_tag, ordering, answer_summary_text_id, status) values ((select id from question where question_tag='q_length_current_medication'), (select id from app_text where app_text_tag='txt_used_two_five_months'), (select id from answer_type where atype='a_type_segmented_control'), 'a_length_current_medication_two_five_months',1, (select id from app_text where app_text_tag='txt_answer_summary_two_five_months'),'ACTIVE');
insert into potential_answer (question_id, answer_localized_text_id, atype_id, potential_answer_tag, ordering, answer_summary_text_id, status) values ((select id from question where question_tag='q_length_current_medication'), (select id from app_text where app_text_tag='txt_used_six_eleven_months'), (select id from answer_type where atype='a_type_segmented_control'), 'a_length_current_medication_six_eleven_months',2, (select id from app_text where app_text_tag='txt_answer_summary_six_eleven_months'),'ACTIVE');
insert into potential_answer (question_id, answer_localized_text_id, atype_id, potential_answer_tag, ordering, answer_summary_text_id, status) values ((select id from question where question_tag='q_length_current_medication'), (select id from app_text where app_text_tag='txt_used_twelve_plus_months'), (select id from answer_type where atype='a_type_segmented_control'), 'a_length_current_medication_twelve_plus_months',3, (select id from app_text where app_text_tag='txt_answer_summary_twelve_plus_months'),'ACTIVE');
commit;