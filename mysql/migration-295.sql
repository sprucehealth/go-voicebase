update question set qtype_id=(select id from question_type where qtype='q_type_free_text') where question_tag in ('q_anything_else_prev_acne_prescription', 'q_anything_else_prev_acne_otc', 'q_acne_otc_product_tried');





