use carefront;
update question set qtype_id = (select id from question_type where qtype='q_type_single_select') where question_tag in ('q_acne_rosacea_type','q_acne_type');