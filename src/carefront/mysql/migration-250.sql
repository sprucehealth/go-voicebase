update localized_text set ltext='Back' where app_text_id = (select qtext_app_text_id from question where question_tag='q_back_photo_section');