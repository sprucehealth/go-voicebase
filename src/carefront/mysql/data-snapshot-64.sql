-- MySQL dump 10.13  Distrib 5.6.13, for osx10.8 (x86_64)
--
-- Host: dev-db-3.ccvrwjdx3gvp.us-east-1.rds.amazonaws.com    Database: database_21545
-- ------------------------------------------------------
-- Server version	5.6.13-log

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `app_text`
--

DROP TABLE IF EXISTS `app_text`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `app_text` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `comment` varchar(600) DEFAULT NULL,
  `app_text_tag` varchar(250) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `app_text_tag` (`app_text_tag`)
) ENGINE=InnoDB AUTO_INCREMENT=294 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `app_text`
--

LOCK TABLES `app_text` WRITE;
/*!40000 ALTER TABLE `app_text` DISABLE KEYS */;
INSERT INTO `app_text` VALUES (1,'reason for visit with doctor','txt_visit_reason'),(2,'acne is the reason for visit','txt_acne_visit_reason'),(3,'something else is reason for visit','txt_something_else_visit_reason'),(4,'hint for typing a symptom or condition','txt_hint_type_symptom'),(5,'duration of acne','txt_acne_length'),(6,'0-6 months for acne length','txt_less_six_months'),(7,'6-12 months for acne length','txt_six_months_one_year_acne_length'),(8,'1-2 years for acne length','txt_one_two_year_acne_length'),(9,'2+ years for acne length','txt_two_plus_year_acne_length'),(10,'is your acne getting worse','txt_acne_worse'),(11,'acne is getting worse response','txt_yes'),(12,'acne is not getting worse response','txt_no'),(13,'helper text to describe what is making acne worse','txt_describe_changes_acne_worse'),(14,'hint text giving examples for what makes acne worse','txt_examples_changes_acne_worse'),(15,'select type of treatments tried for acne','txt_acne_treatments'),(16,'over the counter treatment for acne','txt_otc_acne_treatment'),(17,'prescription treatment for acne','txt_prescription_treatment'),(18,'no treatment tried for acne','txt_no_treatment_acne'),(19,'list medications tried for acne','txt_list_medications_acne'),(20,'type to add treatment','txt_type_add_treatment'),(21,'share anything else w.r.t acne','txt_anything_else_acne'),(22,'hint for anything else you\'d like to tell the doctor','txt_hint_anything_else_acne_treatment'),(23,'question for females to learn about family planning','txt_pregnancy_planning'),(26,'Are you allergic to any medications?','txt_allergic_to_medications'),(29,'hint to add a medication','txt_type_add_medication'),(30,'Your Skin History','txt_skin_history'),(31,'Your Medical History','txt_medical_history'),(32,'question to list medications','txt_list_medications'),(33,'hint to list medications','txt_hint_list_medications'),(34,'question to get social history','txt_get_social_history'),(35,'smoke tobacco','txt_smoke_social_history'),(36,'drink alocohol','txt_alcohol_social_history'),(37,'use tanning beds','txt_tanning_social_history'),(38,'question to learn whether patient has been diagnosed in the past','txt_diagnosed_skin_past'),(39,'listing past skin diagnosis for paitent to chose from','txt_alopecia_diagnosis'),(40,'listing past sking diagnoses for patient to chose from','txt_acne_diagnosis'),(41,'listing past sking diagnoses for patient to chose from','txt_eczema'),(42,'listing past sking diagnoses for patient to chose from','txt_psoriasis_diagnosis'),(43,'listing past sking diagnoses for patient to chose from','txt_rosacea_diagnosis'),(44,'listing past sking diagnoses for patient to chose from','txt_skin_cancer_diagnosis'),(45,'listing past sking diagnoses for patient to chose from','txt_other_diagnosis'),(46,'question to list any medical conditions that patient has been treated for','txt_list_medical_condition'),(47,'hint to prompt user to add a condition','txt_hint_add_condition'),(48,'medical condition list to chose from','txt_arthritis_condition'),(49,'medical condition list to chose from','txt_artificial_heart_valve_condition'),(50,'medical condition list to chose from','txt_artificial_joint_condition'),(51,'medical condition list to chose from','txt_asthma_condition'),(52,'medical condition list to chose from','txt_blood_clots_condition'),(53,'medical condition list to chose from','txt_diabetes_condition'),(54,'medical condition list to chose from','txt_epilepsy_condition'),(55,'medical condition list to chose from','txt_high_bp_condition'),(56,'medical condition list to chose from','txt_high_cholestrol_condition'),(57,'medical condition list to chose from','txt_hiv_condition'),(58,'medical condition list to chose from','txt_heart_attack_condition'),(59,'medical condition list to chose from','txt_heart_murmur_condition'),(60,'medical condition list to chose from','txt_irregular_heartbeat_condition'),(61,'medical condition list to chose from','txt_kidney_disease_condition'),(62,'medical condition list to chose from','txt_liver_disease_condition'),(63,'medical condition list to chose from','txt_lung_disease_condition'),(64,'medical condition list to chose from','txt_lupus_disease_condition'),(65,'medical condition list to chose from','txt_organ_transplant_disease_condition'),(66,'medical condition list to chose from','txt_pacemaker_disease_condition'),(67,'medical condition list to chose from','txt_thyroid_problems_condition'),(68,'medical condition list to chose from','txt_other_condition_condition'),(69,'medical condition list to chose from','txt_no_condition'),(70,'question to determine where the patient is experiencing acne','txt_acne_location'),(71,'face location for acne','txt_face_acne_location'),(72,'chest location for acne','txt_chest_acne_location'),(73,'back location for acne','txt_back_acne_location'),(74,'other locations for acne','txt_other_acne_location'),(75,'title for face section of photo tips','txt_face_photo_tips_title'),(76,'description for face section of photo taking','txt_photo_tips_description'),(77,'tip to remove glasses','txt_remove_glasses_tip'),(78,'tip to pull hair back','txt_pull_hair_back_tip'),(79,'tip to have no makeup','txt_no_makeup_tip'),(80,'title for chest section photo tips','txt_chest_photo_tips_title'),(81,'tip to remove jewellery','txt_remove_jewellery_photo_tip'),(82,'face front label','txt_face_front'),(83,'profile left label','txt_profile_left'),(84,'profile right label','txt_profile_right'),(85,'chest label','txt_chest'),(86,'back lebel','txt_back'),(87,'title for photo section','txt_photo_section_title'),(88,'short description of reason for visit','txt_short_reason_visit'),(89,'short description for length of time patient has been experiencing acne','txt_short_acne_length'),(90,'short description of other symptoms that the patient is attempting to use the app for ','txt_short_other_symptoms'),(91,'short description of whether or not acne is getting worse','txt_short_acne_worse'),(92,'short description of changes that would be making acne worse','txt_short_changes_acne_worse'),(93,'short description of previous types of treatments tried','txt_short_prev_type_treatment'),(94,'short description of previous list of treatments that have been tried','txt_short_prev_list_treatment'),(95,'short description of anything else patient would like to tell doctor about cane','txt_short_anything_else_acne'),(96,'short description of all the places that the patient marked acne is being present on','txt_short_photo_locations'),(97,'short description of whether patient is planning pregnancy','txt_short_pregnant'),(98,'short description of whether patient is alergic to medications','txt_allergic_medications'),(99,'short description to list any medications patient is currently taking','txt_short_list_medications'),(100,'short description to describe social history of patient','txt_short_social_history'),(101,'short description for previous skin diagnosis','txt_short_prev_skin_diagnosis'),(102,'short description for patient to describe medical conditions that they have been treated for','txt_short_medical_condition'),(103,'prompt to take photo of treatment','txt_take_photo_treatment'),(104,'short description for a list of medications that patient is allergic to','txt_short_allergic_medications_list'),(105,'short description for front face photo of patient','txt_short_face_photo'),(106,'short description for chest photos of patient','txt_short_chest_photo'),(107,'short description for back photo of patient','txt_short_back_photo'),(108,'short description for other photo of patient','txt_short_other_photo'),(109,'other lable for photo taking','txt_other'),(110,'how effective was this treatment','txt_effective_treatment'),(111,'answer option','txt_not_very'),(112,'answer option','txt_somewhat'),(113,'answer option','txt_very'),(114,'are you currently using this treatment','txt_current_treatment'),(115,'less than 1 month','txt_one_or_less'),(116,'2-5 months','txt_two_five_months'),(117,'6-11 months','txt_six_eleven_months'),(118,'12+ months','txt_twelve_plus_months'),(119,'not very effective','txt_not_very_effective'),(120,'somewhat effective','txt_somewhat_effective'),(121,'very effective','txt_very_effective'),(122,'currently using it','txt_current_using'),(123,'not currently using it','txt_not_currently_using'),(124,'Used for less than 1 month','txt_used_less_1_month'),(125,'Used for 2-5 months','txt_used_two_five_months'),(126,'Used for 6-11 months','txt_used_six_eleven_months'),(127,'Used for over a year','txt_used_twelve_plus_months'),(128,'question for length of treatment','txt_treatment_length'),(150,'txt for when you first started experiencing acne','txt_first_acne_experience'),(151,'txt response of during puberty','txt_during_puberty'),(152,'txt response of within last six months','txt_within_last_six_months'),(153,'txt response of 1-2 years ago','txt_one_two_years_ago'),(154,'txt response of more than 2 years ago','txt_more_than_two_years'),(155,'txt summary for onset of symptoms','txt_onset_symptoms'),(156,'txt for asking the user if they are experiencing acne symptoms','txt_acne_symtpoms'),(157,'txt for response of acne being painful to touch','txt_painful_touch'),(158,'txt for response of acne being scarring','txt_scarring'),(159,'txt for response of acne causing discoloration','txt_discoloration'),(160,'txt for summarizing additional symptoms','txt_additional_symptoms'),(161,'txt for asking female patients if their acne gets worse with periods','txt_acne_worse_period'),(162,'txt for asking female patients if their periods are regular','txt_periods_regular'),(163,'txt for summarizing information about txt_menstrual_cycle','txt_menstrual_cycle'),(164,'txt for question to descibe skin','txt_skin_description'),(165,'txt for response to skin description as normal','txt_normal_skin'),(166,'txt for response to skin description as oily','txt_oily_skin'),(167,'txt for response to skin description as dry','txt_dry_skin'),(168,'txt for response to skin description as combination','txt_combination_skin'),(169,'txt for summarizing skin type','txt_skin_type'),(170,'txt for determining whether patient has been allergic to topical medication','txt_allergy_topical_medication'),(171,'txt summary for determining whether patient has been allergic to topical medication','txt_summary_allergy_topical_medication'),(172,'txt for determining any other conditions patient may have been diagnosed for in the past','txt_other_condition_acne'),(173,'txt for determining any other conditions patient may have been diagnosed for in the past','txt_summary_other_condition_acne'),(174,'txt response for determining any other conditions patient may have been diagnosed for in the past','txt_gasitris'),(175,'txt response for determining any other conditions patient may have been diagnosed for in the past','txt_colitis'),(176,'txt response for determining any other conditions patient may have been diagnosed for in the past','txt_kidney_disease'),(177,'txt response for determining any other conditions patient may have been diagnosed for in the past','txt_lupus'),(178,'txt summary for treatment not effective','txt_answer_summary_not_effective'),(179,'txt summary for treatment somewhat effective','txt_answer_summary_somewhat_effective'),(180,'txt summary for treatment very effective','txt_answer_summary_very_effective'),(181,'txt summary for not currently using treatment','txt_answer_summary_not_using'),(182,'txt summary for using current treatment','txt_answer_summary_using'),(183,'txt summary for using treatment less than a month','txt_answer_summary_less_month'),(184,'txt summary for using treatment 2-5 months','txt_answer_summary_two_five_months'),(185,'txt summary for using treamtent 6-11 months','txt_answer_summary_six_eleven_months'),(186,'txt summary for using treatment 12+ months','txt_answer_summary_twelve_plus_months'),(187,'txt for prompting user to add treatment','txt_add_treatment'),(188,'txt for prompting user to add medication','txt_add_medication'),(189,'txt for prompting user to take a photo of the medication','txt_take_photo_medication'),(190,'txt for button when adding medication','txt_add_button_medication'),(191,'txt for button when adding treatment','txt_add_button_treatment'),(192,'txt for saving changes when adding medication or treatment','txt_save_changes'),(193,'txt for button to remove treatment','txt_remove_treatment'),(194,'txt for button to remove medication','txt_remove_medication'),(195,'what is your diagnosisa','txt_what_diagnosis'),(196,'acne vulgaris','txt_acne_vulgaris'),(197,'acne rosacea','txt_acne_rosacea'),(198,'how severe is the patients acne','txt_acne_severity'),(199,'acne severity mild','txt_acne_severity_mild'),(200,'acne severity moderate','txt_acne_severity_moderate'),(201,'acne severity severe','txt_acne_severity_severe'),(202,'type of acne','txt_acne_type'),(203,'acne whiteheads','txt_acne_whiteheads'),(204,'acne pustules','txt_acne_pustules'),(205,'acne nodules','txt_acne_nodules'),(206,'acne inflammatory','txt_acne_inflammatory'),(207,'acne blackheads','txt_acne_blackheads'),(208,'acne papules','txt_acne_papules'),(209,'acne cysts','txt_acne_cysts'),(210,'acne hormonal','txt_acne_hormonal'),(211,'select all apply','txt_select_all_apply'),(212,'dispense unit','txt_dispense_unit_Bag'),(213,'dispense unit','txt_dispense_unit_Bottle'),(214,'dispense unit','txt_dispense_unit_Box'),(215,'dispense unit','txt_dispense_unit_Capsule'),(216,'dispense unit','txt_dispense_unit_Cartridge'),(217,'dispense unit','txt_dispense_unit_Container'),(218,'dispense unit','txt_dispense_unit_Drop'),(219,'dispense unit','txt_dispense_unit_Gram'),(220,'dispense unit','txt_dispense_unit_Inhaler'),(221,'dispense unit','txt_dispense_unit_International'),(222,'dispense unit','txt_dispense_unit_Kit'),(223,'dispense unit','txt_dispense_unit_Liter'),(224,'dispense unit','txt_dispense_unit_Lozenge'),(225,'dispense unit','txt_dispense_unit_Milligram'),(226,'dispense unit','txt_dispense_unit_Milliliter'),(227,'dispense unit','txt_dispense_unit_Million_Units'),(228,'dispense unit','txt_dispense_unit_Mutually_Defined'),(229,'dispense unit','txt_dispense_unit_Fluid_Ounce'),(230,'dispense unit','txt_dispense_unit_Not_Specified'),(231,'dispense unit','txt_dispense_unit_Pack'),(232,'dispense unit','txt_dispense_unit_Packet'),(233,'dispense unit','txt_dispense_unit_Pint'),(234,'dispense unit','txt_dispense_unit_Suppository'),(235,'dispense unit','txt_dispense_unit_Syringe'),(236,'dispense unit','txt_dispense_unit_Tablespoon'),(237,'dispense unit','txt_dispense_unit_Tablet'),(238,'dispense unit','txt_dispense_unit_Teaspoon'),(239,'dispense unit','txt_dispense_unit_Transdermal_Patch'),(240,'dispense unit','txt_dispense_unit_Tube'),(241,'dispense unit','txt_dispense_unit_Unit'),(242,'dispense unit','txt_dispense_unit_Vial'),(243,'dispense unit','txt_dispense_unit_Each'),(244,'dispense unit','txt_dispense_unit_Gum'),(245,'dispense unit','txt_dispense_unit_Ampule'),(246,'dispense unit','txt_dispense_unit_Applicator'),(247,'dispense unit','txt_dispense_unit_Applicatorful'),(248,'dispense unit','txt_dispense_unit_Bar'),(249,'dispense unit','txt_dispense_unit_Bead'),(250,'dispense unit','txt_dispense_unit_Blister'),(251,'dispense unit','txt_dispense_unit_Block'),(252,'dispense unit','txt_dispense_unit_Bolus'),(253,'dispense unit','txt_dispense_unit_Can'),(254,'dispense unit','txt_dispense_unit_Canister'),(255,'dispense unit','txt_dispense_unit_Capler'),(256,'dispense unit','txt_dispense_unit_Carton'),(257,'dispense unit','txt_dispense_unit_Case'),(258,'dispense unit','txt_dispense_unit_Cassette'),(259,'dispense unit','txt_dispense_unit_Cylinder'),(260,'dispense unit','txt_dispense_unit_Disk'),(261,'dispense unit','txt_dispense_unit_Dose_Pack'),(262,'dispense unit','txt_dispense_unit_Dual_Packs'),(263,'dispense unit','txt_dispense_unit_Film'),(264,'dispense unit','txt_dispense_unit_Gallon'),(265,'dispense unit','txt_dispense_unit_Implant'),(266,'dispense unit','txt_dispense_unit_Inhalation'),(267,'dispense unit','txt_dispense_unit_Inhaler_Refill'),(268,'dispense unit','txt_dispense_unit_Insert'),(269,'dispense unit','txt_dispense_unit_Intravenous_Bag'),(270,'dispense unit','txt_dispense_unit_Milimeter'),(271,'dispense unit','txt_dispense_unit_Nebule'),(272,'dispense unit','txt_dispense_unit_Needle_Free_Injection'),(273,'dispense unit','txt_dispense_unit_Oscular_System'),(274,'dispense unit','txt_dispense_unit_Ounce'),(275,'dispense unit','txt_dispense_unit_Pad'),(276,'dispense unit','txt_dispense_unit_Paper'),(277,'dispense unit','txt_dispense_unit_Pouch'),(278,'dispense unit','txt_dispense_unit_Pound'),(279,'dispense unit','txt_dispense_unit_Puff'),(280,'dispense unit','txt_dispense_unit_Quart'),(281,'dispense unit','txt_dispense_unit_Ring'),(282,'dispense unit','txt_dispense_unit_Sachet'),(283,'dispense unit','txt_dispense_unit_Scoopful'),(284,'dispense unit','txt_dispense_unit_Sponge'),(285,'dispense unit','txt_dispense_unit_Spray'),(286,'dispense unit','txt_dispense_unit_Stick'),(287,'dispense unit','txt_dispense_unit_Strip'),(288,'dispense unit','txt_dispense_unit_Swab'),(289,'dispense unit','txt_dispense_unit_Tabminder'),(290,'dispense unit','txt_dispense_unit_Tampon'),(291,'dispense unit','txt_dispense_unit_Tray'),(292,'dispense unit','txt_dispense_unit_Troche'),(293,'dispense unit','txt_dispense_unit_Wafer');
/*!40000 ALTER TABLE `app_text` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `localized_text`
--

DROP TABLE IF EXISTS `localized_text`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `localized_text` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `language_id` int(10) unsigned NOT NULL,
  `ltext` varchar(600) NOT NULL,
  `app_text_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `language_id` (`language_id`,`app_text_id`),
  KEY `app_text_id` (`app_text_id`),
  CONSTRAINT `localized_text_ibfk_1` FOREIGN KEY (`app_text_id`) REFERENCES `app_text` (`id`) ON DELETE CASCADE,
  CONSTRAINT `localized_text_ibfk_2` FOREIGN KEY (`language_id`) REFERENCES `languages_supported` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=299 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `localized_text`
--

LOCK TABLES `localized_text` WRITE;
/*!40000 ALTER TABLE `localized_text` DISABLE KEYS */;
INSERT INTO `localized_text` VALUES (4,1,'What\'s the reason for your visit with Dr. Who today?',1),(5,1,'Something else',3),(6,1,'Acne',2),(7,1,'Type a symptom or condition',4),(8,1,'How long have you been experiencing acne symptoms?',5),(9,1,'0-6 months',6),(10,1,'6-123 months',7),(11,1,'1-2 years',8),(12,1,'2+ years',9),(13,1,'Is your acne getting worse?',10),(14,1,'Yes',11),(15,1,'No',12),(16,1,'Describe any recent changes that could be making your acne worse:',13),(18,1,'Examples: recreational activities, new cosmetics or toiletries, warmer weather, stress',14),(19,1,'Select what type of treatments you\'ve previously tried for your acne:',15),(20,1,'Over the counter',16),(21,1,'Prescription',17),(22,1,'No treatments tried',18),(23,1,'List any prescription or over the counter treatments that you\'ve tried for acne',19),(24,1,'Type to add a treatment',20),(25,1,'Is there anything else you\'d like to share about your acne with Dr. wHO?',21),(26,1,'This question is optional but feel free to share anything else about your acne that you think the doctor should know..',22),(27,1,'Are you pregnant, planning a pregnancy or nursing?',23),(28,1,'Are you allergic to any medications?',26),(29,1,'Type to add a medication',29),(30,1,'Your Skin History',30),(31,1,'Your Medical History',31),(32,1,'List any medications you are currently taking:',32),(33,1,'Include birth control, over the counter medications, vitamins or herbal supplements that you may be currently taking.',33),(34,1,'Select which if any of the following activities you do regularly:',34),(35,1,'Smoke tobacco',35),(36,1,'Drink alcohol',36),(37,1,'Use tanning beds or sunbath',37),(38,1,'Have you been diagnosed for a skin condition in the past?',38),(39,1,'Alopecia',39),(40,1,'Acne',40),(42,1,'Eczema',41),(43,1,'Psoriasis',42),(44,1,'Rosacea',43),(45,1,'Skin Cancer',44),(47,1,'Other',45),(48,1,'List any medical condition that you currently have or have been treated for:',46),(50,1,'Type to add a condition',47),(51,1,'Arthritis',48),(53,1,'Artifical Heart Valve',49),(55,1,'Artifical Joint',50),(56,1,'Asthma',51),(57,1,'Blood Clots',52),(58,1,'Diabetes',53),(59,1,'Epilepsy or Seizures',54),(60,1,'High Blood Pressure',55),(61,1,'High Cholestrol',56),(62,1,'HIV/AIDs',57),(63,1,'Heart Attack',58),(64,1,'Heart Murmur',59),(66,1,'Irregular Heartbeat',60),(67,1,'Kidney Disease',61),(68,1,'Liver Disease',62),(69,1,'Lung Disease',63),(70,1,'Lupus',64),(71,1,'Organ Transplant',65),(72,1,'Pacemaker',66),(73,1,'Thyroid Problems',67),(74,1,'Other Condition Not Listed',68),(75,1,'No past or present conditions',69),(76,1,'Photos for Diagnosis',87),(77,1,'We need to know where you\'re experiencing acne so we can take the right photos.',70),(78,1,'Face',71),(79,1,'Chest',72),(80,1,'Back',73),(81,1,'Other',74),(82,1,'Up First: Face Photos',75),(83,1,'Remember these photos are for diagnosis purposes. The clearer your photo the easier it is for the doctor to make a diagnosis.',76),(84,1,'Remove glasses or hats',77),(85,1,'Pull back any hair covering your face',78),(86,1,'No make up',79),(87,1,'Remve any jewellery or clothing that may be covering your chest (except under garments)',81),(88,1,'Next: Chest Photos',80),(89,1,'Reason for visit',88),(90,1,'Length of time with acne symptoms',89),(91,1,'Other symptoms or conditions patient wants diagnosed',90),(92,1,'Worsening symptoms',91),(93,1,'Recent changes making acne worse',92),(94,1,'Type of treatments',93),(95,1,'OTC and Prescriptions tried',94),(96,1,'Additional info patient shared',95),(97,1,'Location of symptoms',96),(98,1,'Pregnant/Nursing',97),(99,1,'Medication Allergies',98),(100,1,'Current medications',99),(101,1,'Social History',100),(102,1,'Skin Conditions',101),(103,1,'Other Conditions',102),(104,1,'Or take a photo of the treatment',103),(105,1,'Face photos of patient',105),(106,1,'Chest photos of patient',106),(107,1,'Back photos of patient',107),(108,1,'Other photos of patient',108),(109,1,'Other',109),(110,1,'Front',82),(111,1,'Profile Left',83),(112,1,'Profile Right',84),(113,1,'Chest',85),(114,1,'How effective was this treatment?',110),(115,1,'Not Very',111),(116,1,'Somewhat',112),(117,1,'Very',113),(118,1,'Are you currently using this treatment?',114),(119,1,'1 or less',115),(120,1,'2-5',116),(121,1,'6-11',117),(122,1,'12+',118),(123,1,'Not very effective',119),(124,1,'Somewhat effective',120),(125,1,'Very effective',121),(126,1,'Currently using it',122),(127,1,'Not currently using it',123),(128,1,'Used for less than 1 month',124),(129,1,'Used for 2-5 months',125),(131,1,'Used for 6-11 months',126),(132,1,'Used for over a year',127),(133,1,'Approximately how many months did you use this treatment for?',128),(154,1,'When did you first begin experiencing acne?',150),(155,1,'During puberty',151),(156,1,'Within the last six months',152),(157,1,'1-2 years ago',153),(158,1,'More than 2 years ago',154),(159,1,'Onset of symptoms',155),(160,1,'Are you experiencing any of the following symptoms with your acne?',156),(161,1,'Painful to the touch',157),(162,1,'Scarring',158),(163,1,'Discoloration',159),(164,1,'Additional Symptoms',160),(165,1,'Does your acne get worse with your period?',161),(166,1,'Are your periods regular?',162),(167,1,'Menstrual cycle',163),(168,1,'How would you describe your skin?',164),(169,1,'Normal',165),(170,1,'Oily',166),(171,1,'Dry',167),(172,1,'Combination',168),(173,1,'Skin type',169),(174,1,'Have you ever had an allergic reaction to a topical medication?',170),(175,1,'Topical Medication Allergies',171),(176,1,'Do you currently have or have been treated for any of the following conditions?',172),(177,1,'Other conditions',173),(178,1,'Gasitris',174),(179,1,'Colitis',175),(180,1,'Kidney Disease',176),(181,1,'Lupus',177),(182,1,'Medication Allergies',104),(183,1,'Not very effective',178),(184,1,'Somewhat effective',179),(185,1,'Very effective',180),(186,1,'Not currently using it',181),(187,1,'Currently using it',182),(188,1,'Used for less than one month',183),(189,1,'Used for 2-5 months',184),(190,1,'Used for 6-11 months',185),(191,1,'Used for 12+ months',186),(192,1,'Add a Treatment',187),(193,1,'Add a Medication',188),(194,1,'Or take a photo of the medication',189),(195,1,'Add Medication',190),(196,1,'Add Treatment',191),(197,1,'Save Changes',192),(198,1,'Remove Treatment',193),(199,1,'Remove Medication',194),(200,1,'What\'s your diagnosis?',195),(201,1,'Acne Vulgaris',196),(202,1,'Acne Rosacea',197),(203,1,'How severe is the patient\'s acne?',198),(204,1,'Mild',199),(205,1,'Moderate',200),(206,1,'Severe',201),(207,1,'What type of acne do they have?',202),(208,1,'Whiteheads',203),(209,1,'Pustules',204),(210,1,'Nodules',205),(211,1,'Inflammatory',206),(212,1,'Blackheads',207),(213,1,'Papules',208),(214,1,'Cysts',209),(215,1,'Hormonal',210),(216,1,'(select all that apply)',211),(217,1,'Bag',212),(218,1,'Bottle',213),(219,1,'Box',214),(220,1,'Capsule',215),(221,1,'Cartridge',216),(222,1,'Container',217),(223,1,'Drop',218),(224,1,'Gram',219),(225,1,'Inhaler',220),(226,1,'International',221),(227,1,'Kit',222),(228,1,'Liter',223),(229,1,'Lozenge',224),(230,1,'Milligram',225),(231,1,'Milliliter',226),(232,1,'Million Units',227),(233,1,'Mutually Defined',228),(234,1,'Fluid Ounce',229),(235,1,'Not Specified',230),(236,1,'Pack',231),(237,1,'Packet',232),(238,1,'Pint',233),(239,1,'Suppository',234),(240,1,'Syringe',235),(241,1,'Tablespoon',236),(242,1,'Tablet',237),(243,1,'Teaspoon',238),(244,1,'Transdermal Patch',239),(245,1,'Tube',240),(246,1,'Unit',241),(247,1,'Vial',242),(248,1,'Each',243),(249,1,'Gum',244),(250,1,'Ampule',245),(251,1,'Applicator',246),(252,1,'Applicatorful',247),(253,1,'Bar',248),(254,1,'Bead',249),(255,1,'Blister',250),(256,1,'Block',251),(257,1,'Bolus',252),(258,1,'Can',253),(259,1,'Canister',254),(260,1,'Capler',255),(261,1,'Carton',256),(262,1,'Case',257),(263,1,'Cassette',258),(264,1,'Cylinder',259),(265,1,'Disk',260),(266,1,'Dose Pack',261),(267,1,'Dual Packs',262),(268,1,'Film',263),(269,1,'Gallon',264),(270,1,'Implant',265),(271,1,'Inhalation',266),(272,1,'Inhaler Refill',267),(273,1,'Insert',268),(274,1,'Intravenous Bag',269),(275,1,'Milimeter',270),(276,1,'Nebule',271),(277,1,'Needle Free Injection',272),(278,1,'Oscular System',273),(279,1,'Ounce',274),(280,1,'Pad',275),(281,1,'Paper',276),(282,1,'Pouch',277),(283,1,'Pound',278),(284,1,'Puff',279),(285,1,'Quart',280),(286,1,'Ring',281),(287,1,'Sachet',282),(288,1,'Scoopful',283),(289,1,'Sponge',284),(290,1,'Spray',285),(291,1,'Stick',286),(292,1,'Strip',287),(293,1,'Swab',288),(294,1,'Tabminder',289),(295,1,'Tampon',290),(296,1,'Tray',291),(297,1,'Troche',292),(298,1,'Wafer',293);
/*!40000 ALTER TABLE `localized_text` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `answer_type`
--

DROP TABLE IF EXISTS `answer_type`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `answer_type` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `atype` varchar(250) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `otype` (`atype`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `answer_type`
--

LOCK TABLES `answer_type` WRITE;
/*!40000 ALTER TABLE `answer_type` DISABLE KEYS */;
INSERT INTO `answer_type` VALUES (5,'a_type_autocomplete_entry'),(4,'a_type_dropdown_entry'),(2,'a_type_free_text'),(1,'a_type_multiple_choice'),(11,'a_type_photo_entry_back'),(12,'a_type_photo_entry_chest'),(8,'a_type_photo_entry_face_left'),(7,'a_type_photo_entry_face_middle'),(10,'a_type_photo_entry_face_right'),(13,'a_type_photo_entry_other'),(6,'a_type_photo_to_text_entry'),(14,'a_type_segmented_control'),(3,'a_type_single_entry');
/*!40000 ALTER TABLE `answer_type` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `region`
--

DROP TABLE IF EXISTS `region`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `region` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `region_tag` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `region_tag` (`region_tag`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `region`
--

LOCK TABLES `region` WRITE;
/*!40000 ALTER TABLE `region` DISABLE KEYS */;
INSERT INTO `region` VALUES (1,'us-east-1'),(2,'us-west-1');
/*!40000 ALTER TABLE `region` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `health_condition`
--

DROP TABLE IF EXISTS `health_condition`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `health_condition` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `comment` varchar(600) NOT NULL,
  `health_condition_tag` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `treatment_tag` (`health_condition_tag`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `health_condition`
--

LOCK TABLES `health_condition` WRITE;
/*!40000 ALTER TABLE `health_condition` DISABLE KEYS */;
INSERT INTO `health_condition` VALUES (1,'health_condition_acne','health_condition_acne');
/*!40000 ALTER TABLE `health_condition` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `languages_supported`
--

DROP TABLE IF EXISTS `languages_supported`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `languages_supported` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `language` varchar(10) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `languages_supported`
--

LOCK TABLES `languages_supported` WRITE;
/*!40000 ALTER TABLE `languages_supported` DISABLE KEYS */;
INSERT INTO `languages_supported` VALUES (1,'en');
/*!40000 ALTER TABLE `languages_supported` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tips`
--

DROP TABLE IF EXISTS `tips`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tips` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `tips_text_id` int(10) unsigned NOT NULL,
  `tips_tag` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `tips_tag` (`tips_tag`),
  KEY `tips_text_id` (`tips_text_id`),
  CONSTRAINT `tips_ibfk_1` FOREIGN KEY (`tips_text_id`) REFERENCES `app_text` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tips`
--

LOCK TABLES `tips` WRITE;
/*!40000 ALTER TABLE `tips` DISABLE KEYS */;
INSERT INTO `tips` VALUES (1,77,'tip_remove_glasses'),(2,78,'tip_pull_hair_back'),(3,79,'tip_no_make_up'),(4,81,'tip_remove_jewellery');
/*!40000 ALTER TABLE `tips` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tips_section`
--

DROP TABLE IF EXISTS `tips_section`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tips_section` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `tips_section_tag` varchar(100) NOT NULL,
  `comment` varchar(500) DEFAULT NULL,
  `tips_title_text_id` int(10) unsigned NOT NULL,
  `tips_subtext_text_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `tips_section_tag` (`tips_section_tag`),
  KEY `tips_title_text_id` (`tips_title_text_id`),
  KEY `tips_subtext_text_id` (`tips_subtext_text_id`),
  CONSTRAINT `tips_section_ibfk_1` FOREIGN KEY (`tips_subtext_text_id`) REFERENCES `app_text` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tips_section`
--

LOCK TABLES `tips_section` WRITE;
/*!40000 ALTER TABLE `tips_section` DISABLE KEYS */;
INSERT INTO `tips_section` VALUES (1,'tips_section_face','tips for taking pictures of face',75,76),(2,'tips_section_chest','tips for taking pictures of chest',80,76);
/*!40000 ALTER TABLE `tips_section` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `section`
--

DROP TABLE IF EXISTS `section`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `section` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `section_title_app_text_id` int(10) unsigned NOT NULL,
  `comment` varchar(600) NOT NULL,
  `health_condition_id` int(10) unsigned DEFAULT NULL,
  `section_tag` varchar(250) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `section_tag` (`section_tag`),
  KEY `section_title_app_text_id` (`section_title_app_text_id`),
  KEY `health_condition_id` (`health_condition_id`),
  CONSTRAINT `section_ibfk_2` FOREIGN KEY (`health_condition_id`) REFERENCES `health_condition` (`id`),
  CONSTRAINT `section_ibfk_1` FOREIGN KEY (`section_title_app_text_id`) REFERENCES `app_text` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `section`
--

LOCK TABLES `section` WRITE;
/*!40000 ALTER TABLE `section` DISABLE KEYS */;
INSERT INTO `section` VALUES (1,30,'skin history section',1,'section_skin_history'),(2,31,'medical history section',NULL,'section_medical_history'),(3,87,'photos for diagnosis',1,'section_photo_diagnosis');
/*!40000 ALTER TABLE `section` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `screen_type`
--

DROP TABLE IF EXISTS `screen_type`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `screen_type` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `screen_type_tag` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `screen_type_tag` (`screen_type_tag`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `screen_type`
--

LOCK TABLES `screen_type` WRITE;
/*!40000 ALTER TABLE `screen_type` DISABLE KEYS */;
INSERT INTO `screen_type` VALUES (1,'screen_type_general'),(2,'screen_type_photo');
/*!40000 ALTER TABLE `screen_type` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `question_type`
--

DROP TABLE IF EXISTS `question_type`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `question_type` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `qtype` varchar(250) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `qtype` (`qtype`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `question_type`
--

LOCK TABLES `question_type` WRITE;
/*!40000 ALTER TABLE `question_type` DISABLE KEYS */;
INSERT INTO `question_type` VALUES (9,'q_type_autocomplete'),(3,'q_type_compound'),(2,'q_type_free_text'),(1,'q_type_multiple_choice'),(7,'q_type_multiple_photo'),(8,'q_type_segmented_control'),(5,'q_type_single_entry'),(4,'q_type_single_photo'),(6,'q_type_single_select');
/*!40000 ALTER TABLE `question_type` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `question`
--

DROP TABLE IF EXISTS `question`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `question` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `qtype_id` int(10) unsigned NOT NULL,
  `qtext_app_text_id` int(10) unsigned DEFAULT NULL,
  `qtext_short_text_id` int(10) unsigned DEFAULT NULL,
  `subtext_app_text_id` int(10) unsigned DEFAULT NULL,
  `question_tag` varchar(250) NOT NULL,
  `parent_question_id` int(10) unsigned DEFAULT NULL,
  `required` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `question_tag` (`question_tag`),
  KEY `qtype_id` (`qtype_id`),
  KEY `subtext_app_text_id` (`subtext_app_text_id`),
  KEY `qtext_app_text_id` (`qtext_app_text_id`),
  KEY `qtext_short_text_id` (`qtext_short_text_id`),
  KEY `parent_question_id` (`parent_question_id`),
  CONSTRAINT `question_ibfk_1` FOREIGN KEY (`qtype_id`) REFERENCES `question_type` (`id`),
  CONSTRAINT `question_ibfk_2` FOREIGN KEY (`subtext_app_text_id`) REFERENCES `app_text` (`id`),
  CONSTRAINT `question_ibfk_3` FOREIGN KEY (`qtext_app_text_id`) REFERENCES `app_text` (`id`),
  CONSTRAINT `question_ibfk_4` FOREIGN KEY (`qtext_short_text_id`) REFERENCES `app_text` (`id`),
  CONSTRAINT `question_ibfk_5` FOREIGN KEY (`parent_question_id`) REFERENCES `question` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `question`
--

LOCK TABLES `question` WRITE;
/*!40000 ALTER TABLE `question` DISABLE KEYS */;
INSERT INTO `question` VALUES (1,6,1,88,NULL,'q_reason_visit',NULL,NULL),(2,5,NULL,90,NULL,'q_condition_for_diagnosis',NULL,NULL),(3,6,5,89,NULL,'q_acne_length',NULL,NULL),(4,6,10,91,NULL,'q_acne_worse',NULL,NULL),(6,2,13,92,NULL,'q_changes_acne_worse',NULL,NULL),(7,1,15,93,NULL,'q_acne_prev_treatment_types',NULL,NULL),(8,9,19,94,NULL,'q_acne_prev_treatment_list',NULL,NULL),(9,2,21,95,NULL,'q_anything_else_acne',NULL,NULL),(10,6,23,97,NULL,'q_pregnancy_planning',NULL,NULL),(11,1,26,98,NULL,'q_allergic_medications',NULL,NULL),(12,9,29,104,NULL,'q_allergic_medication_entry',NULL,NULL),(13,9,32,99,33,'q_current_medications_entry',NULL,NULL),(14,1,34,100,NULL,'q_social_history',NULL,NULL),(15,6,38,101,NULL,'q_prev_skin_condition_diagnosis',NULL,NULL),(16,3,46,102,NULL,'q_prev_med_condition_diagnosis',NULL,NULL),(17,1,NULL,101,NULL,'q_list_prev_skin_condition_diagnosis',NULL,NULL),(18,1,70,96,NULL,'q_acne_location',NULL,NULL),(19,4,NULL,105,NULL,'q_face_photo_intake',NULL,NULL),(20,4,NULL,106,NULL,'q_chest_photo_intake',NULL,NULL),(21,4,NULL,107,NULL,'q_back_photo_intake',NULL,NULL),(22,7,NULL,108,NULL,'q_other_photo_intake',NULL,NULL),(24,8,110,NULL,NULL,'q_effective_treatment',8,NULL),(25,8,114,NULL,NULL,'q_using_treatment',8,NULL),(26,8,128,NULL,NULL,'q_length_treatment',8,NULL),(28,6,150,155,NULL,'q_onset_acne',NULL,1),(29,1,156,160,NULL,'q_acne_symptoms',NULL,1),(30,6,161,163,NULL,'q_acne_worse_period',NULL,0),(31,6,162,NULL,NULL,'q_periods_regular',30,0),(32,6,164,169,NULL,'q_skin_description',NULL,1),(33,6,170,171,NULL,'q_topical_allergic_medications',NULL,1),(34,6,172,173,NULL,'q_other_conditions_acne',NULL,1),(36,9,29,171,NULL,'q_topical_allergies_medication_entry',NULL,0),(37,6,195,NULL,NULL,'q_acne_diagnosis',NULL,1),(38,6,198,NULL,NULL,'q_acne_severity',NULL,1),(39,1,202,NULL,211,'q_acne_type',NULL,1);
/*!40000 ALTER TABLE `question` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `question_fields`
--

DROP TABLE IF EXISTS `question_fields`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `question_fields` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `question_field` varchar(250) NOT NULL,
  `question_id` int(10) unsigned NOT NULL,
  `app_text_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `question_id` (`question_id`),
  KEY `app_text_id` (`app_text_id`),
  KEY `question_field` (`question_field`,`question_id`),
  CONSTRAINT `question_fields_ibfk_1` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `question_fields_ibfk_2` FOREIGN KEY (`app_text_id`) REFERENCES `app_text` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=36 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `question_fields`
--

LOCK TABLES `question_fields` WRITE;
/*!40000 ALTER TABLE `question_fields` DISABLE KEYS */;
INSERT INTO `question_fields` VALUES (27,'add_text',12,188),(28,'placeholder_text',12,29),(29,'add_photo_text',12,189),(30,'add_text',13,188),(31,'placeholder_text',13,29),(32,'add_photo_text',13,189),(33,'add_text',36,188),(34,'placeholder_text',36,29),(35,'add_photo_text',36,189);
/*!40000 ALTER TABLE `question_fields` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `potential_answer`
--

DROP TABLE IF EXISTS `potential_answer`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `potential_answer` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `question_id` int(10) unsigned NOT NULL,
  `answer_localized_text_id` int(10) unsigned DEFAULT NULL,
  `atype_id` int(10) unsigned NOT NULL,
  `potential_answer_tag` varchar(250) NOT NULL,
  `ordering` int(10) unsigned NOT NULL,
  `answer_summary_text_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `potential_outcome_tag` (`potential_answer_tag`),
  UNIQUE KEY `question_id_2` (`question_id`,`ordering`),
  KEY `otype_id` (`atype_id`),
  KEY `outcome_localized_text` (`answer_localized_text_id`),
  KEY `answer_summary_text_id` (`answer_summary_text_id`),
  CONSTRAINT `potential_answer_ibfk_3` FOREIGN KEY (`answer_summary_text_id`) REFERENCES `app_text` (`id`),
  CONSTRAINT `potential_answer_ibfk_1` FOREIGN KEY (`atype_id`) REFERENCES `answer_type` (`id`),
  CONSTRAINT `potential_answer_ibfk_2` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=116 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `potential_answer`
--

LOCK TABLES `potential_answer` WRITE;
/*!40000 ALTER TABLE `potential_answer` DISABLE KEYS */;
INSERT INTO `potential_answer` VALUES (1,1,2,1,'a_acne',0,NULL),(2,1,3,1,'a_something_else',1,NULL),(3,2,NULL,3,'a_condition_entry',0,NULL),(4,3,6,1,'a_less_six_months',0,NULL),(5,3,7,1,'a_six_twelve_months',1,NULL),(6,3,8,1,'a_one_twa_years',2,NULL),(7,3,9,1,'a_twa_plus_years',3,NULL),(8,4,11,1,'a_yes_acne_worse',0,NULL),(9,4,12,1,'a_na_acne_worse',1,NULL),(12,7,17,1,'a_prescription_prev_treatment_type',0,NULL),(13,7,16,1,'a_otc_prev_treatment_type',1,NULL),(14,7,18,1,'a_na_prev_treatment_type',2,NULL),(18,10,11,1,'a_yes_pregnancy_planning',0,NULL),(19,10,12,1,'a_na_pregnancy_planning',1,NULL),(20,11,11,1,'a_yes_allergic_medications',0,NULL),(21,11,12,1,'a_na_allergic_medications',1,NULL),(24,14,35,1,'a_smoke_social_history',0,NULL),(25,14,36,1,'a_alcohol_social_history',1,NULL),(26,14,37,1,'a_tanning_social_history',2,NULL),(27,15,11,1,'a_yes_prev_skin_diagnosis',0,NULL),(28,15,12,1,'a_na_prev_skin_diagnosis',1,NULL),(29,17,39,1,'a_alopecia_skin_diagnosis',0,NULL),(30,17,40,1,'a_acne_skin_diagnosis',1,NULL),(31,17,41,1,'a_eczema_skin_diagnosis',2,NULL),(32,17,42,1,'a_psoriasis_skin_diagnosis',3,NULL),(33,17,43,1,'a_rosacea_skin_diagnosis',4,NULL),(34,17,44,1,'a_skin_cancer_diagnosis',5,NULL),(35,17,45,1,'a_other_skin_iagnosis',6,NULL),(36,16,48,1,'a_arthritis_diagnosis',0,NULL),(37,16,49,1,'a_heart_valve_diagnosis',1,NULL),(38,16,50,1,'a_artificial_join__diagnosis',2,NULL),(39,16,51,1,'a_asthma_diagnosis',3,NULL),(40,16,52,1,'a_blood_clots_diagnosis',4,NULL),(41,16,53,1,'a_diabetes_diagnosis',5,NULL),(42,16,54,1,'a_epilepsey_diagnosis',6,NULL),(43,16,55,1,'a_high_blood_pressure_diagnosis',7,NULL),(44,16,56,1,'a_high_cholestrol_diagnosis',8,NULL),(45,16,57,1,'a_hiv_diagnosis',9,NULL),(46,16,58,1,'a_heart_attack_diagnosis',10,NULL),(47,16,59,1,'a_heart_murmur_diagnosis',11,NULL),(48,16,60,1,'a_irregular_heart_beat_skin_diagnosis',12,NULL),(49,16,61,1,'a_kidney_disease_diagnosis',13,NULL),(50,16,62,1,'a_liver_disease_diagnosis',14,NULL),(51,16,63,1,'a_lung_disease_diagnosis',15,NULL),(52,16,64,1,'a_lupus_disease_diagnosis',16,NULL),(53,16,65,1,'a_organ_transplant_diagnosis',17,NULL),(55,16,66,1,'a_pacemaker_diagnosis',18,NULL),(56,16,67,1,'a_thyroid_diagnosis',19,NULL),(57,16,68,1,'a_other_skin_diagnosis',20,NULL),(58,16,69,1,'a_none_skin_diagnosis',21,NULL),(59,18,71,1,'a_face_acne_location',0,NULL),(60,18,72,1,'a_chest_acne_location',1,NULL),(61,18,73,1,'a_back_acne_location',2,NULL),(62,18,74,1,'a_other_acne_location',3,NULL),(63,19,82,7,'a_face_front_phota_intake',0,NULL),(64,19,84,10,'a_face_right_phota_intake',1,NULL),(65,19,83,8,'a_face_left_phota_intake',2,NULL),(66,20,85,12,'a_chest_phota_intake',0,NULL),(68,21,86,11,'a_back_phota_intake',0,NULL),(69,22,109,13,'a_other_phota_intake',0,NULL),(70,24,111,14,'a_effective_treatment_not_very',0,178),(71,24,112,14,'a_effective_treatment_somewhat',1,179),(72,24,113,14,'a_effective_treatment_very',2,180),(73,25,11,14,'a_using_treatment_yes',0,182),(75,25,12,14,'a_using_treatment_no',1,181),(76,26,115,14,'a_length_treatment_less_one',0,183),(77,26,116,14,'a_length_treatment_two_five_months',1,184),(78,26,117,14,'a_length_treatment_six_eleven_months',2,185),(79,26,118,14,'a_length_treatment_twelve_plus_months',3,186),(80,28,151,1,'a_puberty',0,NULL),(81,28,152,1,'a_onset_six_months',1,NULL),(82,28,153,1,'a_onset_one_two_years',2,NULL),(83,28,154,1,'a_onset_more_two_years',3,NULL),(84,29,157,1,'a_painful_touch',0,NULL),(85,29,158,1,'a_scarring',1,NULL),(86,29,159,1,'a_discoloration',2,NULL),(87,30,11,1,'a_acne_worse_yes',0,NULL),(88,30,12,1,'a_acne_worse_no',1,NULL),(89,31,11,1,'a_periods_regular_yes',0,NULL),(90,31,12,1,'a_periods_regular_no',1,NULL),(91,32,165,1,'a_normal_skin',0,NULL),(92,32,166,1,'a_oil_skin',1,NULL),(93,32,167,1,'a_dry_skin',2,NULL),(94,32,168,1,'a_combination_skin',3,NULL),(95,33,11,1,'a_topical_allergic_medication_yes',0,NULL),(96,33,12,1,'a_topical_allergic_medication_no',1,NULL),(97,34,174,1,'a_other_condition_acne_gastiris',0,NULL),(98,34,175,1,'a_other_condition_acne_colitis',1,NULL),(99,34,176,1,'a_other_condition_acne_kidney_condition',2,NULL),(100,34,177,1,'a_other_condition_acne_lupus',3,NULL),(102,37,196,1,'a_doctor_acne_vulgaris',0,NULL),(103,37,197,1,'a_doctor_acne_rosacea',1,NULL),(104,37,3,1,'a_doctor_acne_something_else',2,NULL),(105,38,199,1,'a_doctor_acne_severity_mild',0,NULL),(106,38,200,1,'a_doctor_acne_severity_moderate',1,NULL),(107,38,201,1,'a_doctor_acne_severity_severity',2,NULL),(108,39,203,1,'a_acne_whiteheads',0,NULL),(109,39,204,1,'a_acne_pustules',1,NULL),(110,39,205,1,'a_acne_nodules',2,NULL),(111,39,206,1,'a_acne_inflammatory',3,NULL),(112,39,207,1,'a_acne_blackheads',4,NULL),(113,39,208,1,'a_acne_papules',5,NULL),(114,39,209,1,'a_acne_cysts',6,NULL),(115,39,210,1,'a_acne_hormonal',7,NULL);
/*!40000 ALTER TABLE `potential_answer` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `photo_tips`
--

DROP TABLE IF EXISTS `photo_tips`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photo_tips` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `photo_tips_tag` varchar(100) NOT NULL,
  `photo_url_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `photo_tips_tag` (`photo_tips_tag`),
  KEY `photo_url_id` (`photo_url_id`),
  CONSTRAINT `photo_tips_ibfk_1` FOREIGN KEY (`photo_url_id`) REFERENCES `object_storage` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `photo_tips`
--

LOCK TABLES `photo_tips` WRITE;
/*!40000 ALTER TABLE `photo_tips` DISABLE KEYS */;
/*!40000 ALTER TABLE `photo_tips` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `patient_layout_version`
--

DROP TABLE IF EXISTS `patient_layout_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_layout_version` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `object_storage_id` int(10) unsigned NOT NULL,
  `language_id` int(10) unsigned NOT NULL,
  `layout_version_id` int(10) unsigned NOT NULL,
  `status` varchar(250) NOT NULL,
  `health_condition_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified_date` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `language_id` (`language_id`),
  KEY `object_storage_id` (`object_storage_id`),
  KEY `treatment_id` (`health_condition_id`),
  CONSTRAINT `patient_layout_version_ibfk_1` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `patient_layout_version_ibfk_2` FOREIGN KEY (`language_id`) REFERENCES `languages_supported` (`id`),
  CONSTRAINT `patient_layout_version_ibfk_3` FOREIGN KEY (`object_storage_id`) REFERENCES `object_storage` (`id`),
  CONSTRAINT `patient_layout_version_ibfk_4` FOREIGN KEY (`health_condition_id`) REFERENCES `health_condition` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=72 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `patient_layout_version`
--

LOCK TABLES `patient_layout_version` WRITE;
/*!40000 ALTER TABLE `patient_layout_version` DISABLE KEYS */;
INSERT INTO `patient_layout_version` VALUES (8,34,1,15,'DEPCRECATED',1,'2013-11-08 19:13:34','0000-00-00 00:00:00'),(9,36,1,16,'DEPCRECATED',1,'2013-11-08 19:13:34','0000-00-00 00:00:00'),(10,38,1,17,'DEPCRECATED',1,'2013-11-08 19:13:34','0000-00-00 00:00:00'),(11,40,1,18,'DEPCRECATED',1,'2013-11-08 19:13:34','0000-00-00 00:00:00'),(12,42,1,19,'DEPCRECATED',1,'2013-11-08 19:13:34','2013-11-08 19:22:26'),(13,44,1,20,'DEPCRECATED',1,'2013-11-08 19:22:21','2013-11-11 05:47:05'),(14,46,1,21,'DEPCRECATED',1,'2013-11-11 05:47:04','2013-11-11 05:57:41'),(15,48,1,22,'DEPCRECATED',1,'2013-11-11 05:57:40','2013-11-11 05:58:51'),(16,51,1,24,'DEPCRECATED',1,'2013-11-11 05:58:50','2013-11-11 06:02:30'),(17,53,1,25,'DEPCRECATED',1,'2013-11-11 06:02:29','2013-11-11 06:07:04'),(18,56,1,26,'DEPCRECATED',1,'2013-11-11 06:07:03','2013-11-12 15:02:06'),(19,58,1,27,'DEPCRECATED',1,'2013-11-12 15:02:05','2013-11-12 15:34:18'),(20,60,1,28,'DEPCRECATED',1,'2013-11-12 15:34:17','2013-11-12 15:34:49'),(21,63,1,29,'DEPCRECATED',1,'2013-11-12 15:34:48','2013-11-12 15:34:50'),(22,64,1,30,'DEPCRECATED',1,'2013-11-12 15:34:50','2013-11-12 15:38:15'),(23,66,1,31,'DEPCRECATED',1,'2013-11-12 15:38:15','2013-11-12 15:39:13'),(24,68,1,32,'DEPCRECATED',1,'2013-11-12 15:39:12','2013-11-12 17:02:20'),(25,70,1,33,'DEPCRECATED',1,'2013-11-12 17:02:19','2013-11-12 17:04:08'),(26,72,1,34,'DEPCRECATED',1,'2013-11-12 17:04:08','2013-11-12 17:15:20'),(27,74,1,35,'DEPCRECATED',1,'2013-11-12 17:15:19','2013-11-12 19:36:52'),(28,76,1,36,'DEPCRECATED',1,'2013-11-12 19:36:52','2013-11-17 00:30:53'),(29,106,1,37,'DEPCRECATED',1,'2013-11-17 00:30:52','2013-11-17 00:31:22'),(30,108,1,38,'DEPCRECATED',1,'2013-11-17 00:31:21','2013-11-17 00:48:23'),(31,110,1,39,'DEPCRECATED',1,'2013-11-17 00:48:22','2013-11-17 19:25:25'),(32,112,1,40,'DEPCRECATED',1,'2013-11-17 19:25:24','2013-11-17 19:28:23'),(33,114,1,41,'DEPCRECATED',1,'2013-11-17 19:28:22','2013-11-17 19:36:07'),(34,116,1,42,'DEPCRECATED',1,'2013-11-17 19:36:06','2013-11-20 01:30:08'),(35,121,1,44,'DEPCRECATED',1,'2013-11-20 01:30:07','2013-11-20 01:38:20'),(36,123,1,45,'DEPCRECATED',1,'2013-11-20 01:38:10','2013-11-20 21:04:04'),(37,125,1,46,'DEPCRECATED',1,'2013-11-20 21:04:03','2013-11-24 02:02:41'),(38,138,1,52,'DEPCRECATED',1,'2013-11-24 02:02:41','2013-11-24 02:05:20'),(39,140,1,53,'DEPCRECATED',1,'2013-11-24 02:05:19','2013-11-24 02:09:31'),(40,142,1,54,'DEPCRECATED',1,'2013-11-24 02:09:30','2013-11-24 02:11:37'),(41,144,1,55,'DEPCRECATED',1,'2013-11-24 02:11:36','2013-11-24 02:21:03'),(42,146,1,56,'DEPCRECATED',1,'2013-11-24 02:21:01','2013-12-03 21:23:08'),(43,175,1,63,'DEPCRECATED',1,'2013-12-03 21:23:07','2013-12-03 21:26:34'),(44,178,1,65,'DEPCRECATED',1,'2013-12-03 21:26:33','2013-12-03 21:28:43'),(45,180,1,66,'DEPCRECATED',1,'2013-12-03 21:28:42','2013-12-03 21:32:04'),(46,182,1,67,'DEPCRECATED',1,'2013-12-03 21:32:03','2013-12-03 21:50:59'),(47,184,1,68,'DEPCRECATED',1,'2013-12-03 21:50:58','2013-12-04 22:59:10'),(48,192,1,70,'DEPCRECATED',1,'2013-12-04 22:59:09','2013-12-04 23:02:03'),(49,194,1,71,'DEPCRECATED',1,'2013-12-04 23:02:02','2013-12-04 23:02:45'),(50,196,1,72,'DEPCRECATED',1,'2013-12-04 23:02:45','2013-12-04 23:41:39'),(51,198,1,73,'DEPCRECATED',1,'2013-12-04 23:41:38','2013-12-05 22:28:56'),(52,205,1,77,'DEPCRECATED',1,'2013-12-05 22:28:55','2013-12-05 22:32:06'),(53,207,1,78,'DEPCRECATED',1,'2013-12-05 22:32:06','2013-12-05 22:33:20'),(54,209,1,79,'DEPCRECATED',1,'2013-12-05 22:33:20','2013-12-05 22:34:30'),(55,212,1,80,'DEPCRECATED',1,'2013-12-05 22:34:29','2013-12-05 22:34:31'),(56,213,1,81,'DEPCRECATED',1,'2013-12-05 22:34:31','2013-12-05 22:35:59'),(57,216,1,82,'DEPCRECATED',1,'2013-12-05 22:35:59','2013-12-05 22:37:00'),(58,218,1,84,'DEPCRECATED',1,'2013-12-05 22:37:00','2013-12-05 22:51:08'),(59,220,1,85,'DEPCRECATED',1,'2013-12-05 22:51:08','2013-12-06 00:07:53'),(60,223,1,86,'DEPCRECATED',1,'2013-12-06 00:07:53','2013-12-06 00:25:03'),(61,225,1,87,'DEPCRECATED',1,'2013-12-06 00:25:03','2013-12-06 00:29:35'),(62,227,1,88,'DEPCRECATED',1,'2013-12-06 00:29:34','2013-12-06 00:34:07'),(63,229,1,89,'DEPCRECATED',1,'2013-12-06 00:34:06','2013-12-06 03:43:30'),(64,231,1,90,'DEPCRECATED',1,'2013-12-06 03:43:30','2013-12-16 07:12:05'),(65,260,1,92,'DEPCRECATED',1,'2013-12-16 07:12:05','2013-12-16 17:37:01'),(66,266,1,93,'DEPCRECATED',1,'2013-12-16 17:37:01','2013-12-17 01:18:28'),(68,275,1,95,'DEPCRECATED',1,'2013-12-17 01:18:28','2013-12-17 01:33:23'),(69,277,1,96,'DEPCRECATED',1,'2013-12-17 01:33:22','2013-12-17 02:25:14'),(70,285,1,100,'DEPCRECATED',1,'2013-12-17 02:25:13','2013-12-17 04:43:49'),(71,287,1,101,'ACTIVE',1,'2013-12-17 04:43:49','2013-12-17 04:43:50');
/*!40000 ALTER TABLE `patient_layout_version` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `object_storage`
--

DROP TABLE IF EXISTS `object_storage`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `object_storage` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `bucket` varchar(100) NOT NULL,
  `storage_key` varchar(100) NOT NULL,
  `status` varchar(100) NOT NULL,
  `region_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified_date` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `region_id` (`region_id`,`storage_key`,`bucket`,`status`),
  CONSTRAINT `object_storage_ibfk_1` FOREIGN KEY (`region_id`) REFERENCES `region` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=309 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `object_storage`
--

LOCK TABLES `object_storage` WRITE;
/*!40000 ALTER TABLE `object_storage` DISABLE KEYS */;
INSERT INTO `object_storage` VALUES (33,'carefront-layout','1383846854','ACTIVE',1,'2013-11-08 18:59:58','2013-11-08 19:09:51'),(34,'carefront-client-layout','1383846875','ACTIVE',1,'2013-11-08 18:59:58','0000-00-00 00:00:00'),(35,'carefront-layout','1383852476','ACTIVE',1,'2013-11-08 18:59:58','0000-00-00 00:00:00'),(36,'carefront-client-layout','1383852497','ACTIVE',1,'2013-11-08 18:59:58','0000-00-00 00:00:00'),(37,'carefront-layout','1383934286','ACTIVE',1,'2013-11-08 18:59:58','0000-00-00 00:00:00'),(38,'carefront-client-layout','1383934320','ACTIVE',1,'2013-11-08 18:59:58','0000-00-00 00:00:00'),(39,'carefront-layout','1383937634','0',1,'2013-11-08 19:07:15','0000-00-00 00:00:00'),(40,'carefront-client-layout','1383937647','0',1,'2013-11-08 19:07:28','0000-00-00 00:00:00'),(41,'carefront-layout','1383937817','ACTIVE',1,'2013-11-08 19:10:18','2013-11-08 19:10:19'),(42,'carefront-client-layout','1383937831','ACTIVE',1,'2013-11-08 19:10:33','2013-11-08 19:10:34'),(43,'carefront-layout','1383938460','ACTIVE',1,'2013-11-08 19:21:02','2013-11-08 19:21:05'),(44,'carefront-client-layout','1383938531','ACTIVE',1,'2013-11-08 19:22:13','2013-11-08 19:22:20'),(45,'carefront-layout','1384148805','ACTIVE',1,'2013-11-11 05:46:45','2013-11-11 05:46:47'),(46,'carefront-client-layout','1384148822','ACTIVE',1,'2013-11-11 05:47:02','2013-11-11 05:47:03'),(47,'carefront-layout','1384149443','ACTIVE',1,'2013-11-11 05:57:23','2013-11-11 05:57:25'),(48,'carefront-client-layout','1384149458','ACTIVE',1,'2013-11-11 05:57:38','2013-11-11 05:57:40'),(49,'carefront-layout','1384149498','ACTIVE',1,'2013-11-11 05:58:18','2013-11-11 05:58:19'),(50,'carefront-layout','1384149512','ACTIVE',1,'2013-11-11 05:58:32','2013-11-11 05:58:34'),(51,'carefront-client-layout','1384149529','ACTIVE',1,'2013-11-11 05:58:49','2013-11-11 05:58:50'),(52,'carefront-layout','1384149729','ACTIVE',1,'2013-11-11 06:02:09','2013-11-11 06:02:11'),(53,'carefront-client-layout','1384149748','ACTIVE',1,'2013-11-11 06:02:28','2013-11-11 06:02:29'),(54,'carefront-layout','1384149808','ACTIVE',1,'2013-11-11 06:03:28','2013-11-11 06:03:30'),(55,'carefront-layout','1384150007','ACTIVE',1,'2013-11-11 06:06:48','2013-11-11 06:06:48'),(56,'carefront-client-layout','1384150021','ACTIVE',1,'2013-11-11 06:07:01','2013-11-11 06:07:03'),(57,'carefront-layout','1384268515','ACTIVE',1,'2013-11-12 15:01:55','2013-11-12 15:01:55'),(58,'carefront-client-layout','1384268525','ACTIVE',1,'2013-11-12 15:02:05','2013-11-12 15:02:05'),(59,'carefront-layout','1384270446','ACTIVE',1,'2013-11-12 15:34:06','2013-11-12 15:34:06'),(60,'carefront-client-layout','1384270457','ACTIVE',1,'2013-11-12 15:34:17','2013-11-12 15:34:17'),(61,'carefront-layout','1384270477','ACTIVE',1,'2013-11-12 15:34:37','2013-11-12 15:34:38'),(62,'carefront-layout','1384270479','ACTIVE',1,'2013-11-12 15:34:39','2013-11-12 15:34:39'),(63,'carefront-client-layout','1384270487','ACTIVE',1,'2013-11-12 15:34:47','2013-11-12 15:34:48'),(64,'carefront-client-layout','1384270489','ACTIVE',1,'2013-11-12 15:34:49','2013-11-12 15:34:49'),(65,'carefront-layout','1384270684','ACTIVE',1,'2013-11-12 15:38:04','2013-11-12 15:38:04'),(66,'carefront-client-layout','1384270694','ACTIVE',1,'2013-11-12 15:38:14','2013-11-12 15:38:15'),(67,'carefront-layout','1384270742','ACTIVE',1,'2013-11-12 15:39:02','2013-11-12 15:39:03'),(68,'carefront-client-layout','1384270751','ACTIVE',1,'2013-11-12 15:39:12','2013-11-12 15:39:12'),(69,'carefront-layout','1384275729','ACTIVE',1,'2013-11-12 17:02:09','2013-11-12 17:02:09'),(70,'carefront-client-layout','1384275739','ACTIVE',1,'2013-11-12 17:02:19','2013-11-12 17:02:19'),(71,'carefront-layout','1384275838','ACTIVE',1,'2013-11-12 17:03:58','2013-11-12 17:03:58'),(72,'carefront-client-layout','1384275847','ACTIVE',1,'2013-11-12 17:04:07','2013-11-12 17:04:07'),(73,'carefront-layout','1384276508','ACTIVE',1,'2013-11-12 17:15:08','2013-11-12 17:15:09'),(74,'carefront-client-layout','1384276518','ACTIVE',1,'2013-11-12 17:15:19','2013-11-12 17:15:19'),(75,'carefront-layout','1384285001','ACTIVE',1,'2013-11-12 19:36:41','2013-11-12 19:36:42'),(76,'carefront-client-layout','1384285011','ACTIVE',1,'2013-11-12 19:36:51','2013-11-12 19:36:51'),(77,'carefront-cases','19/51','CREATING',1,'2013-11-12 23:11:19','0000-00-00 00:00:00'),(78,'carefront-cases','19/52','CREATING',1,'2013-11-12 23:14:52','0000-00-00 00:00:00'),(79,'carefront-cases','19/54','CREATING',2,'2013-11-12 23:17:17','0000-00-00 00:00:00'),(80,'carefront-cases','19/55','ACTIVE',2,'2013-11-12 23:19:52','2013-11-12 23:19:53'),(81,'carefront-cases','19/56','ACTIVE',2,'2013-11-12 23:20:53','2013-11-12 23:20:53'),(82,'carefront-cases','19/57.1','CREATING',1,'2013-11-13 00:28:26','0000-00-00 00:00:00'),(83,'carefront-cases','19/0.1','CREATING',1,'2013-11-13 00:28:54','0000-00-00 00:00:00'),(84,'carefront-cases','19/59.1','CREATING',1,'2013-11-13 02:32:21','0000-00-00 00:00:00'),(85,'carefront-cases','19/60.1','ACTIVE',2,'2013-11-13 02:33:23','2013-11-13 02:33:24'),(86,'carefront-cases','19/0.1','ACTIVE',2,'2013-11-13 02:36:29','2013-11-13 02:36:31'),(87,'carefront-cases','19/0.1','CREATING',2,'2013-11-15 01:28:17','0000-00-00 00:00:00'),(90,'carefront-cases','19/65.1','ACTIVE',2,'2013-11-15 18:57:49','2013-11-15 18:57:55'),(91,'carefront-cases','6/66.out','ACTIVE',2,'2013-11-15 23:28:54','2013-11-15 23:28:55'),(92,'carefront-cases','6/67.out','ACTIVE',2,'2013-11-15 23:29:21','2013-11-15 23:29:25'),(93,'carefront-cases','6/68.out','ACTIVE',2,'2013-11-15 23:30:45','2013-11-15 23:30:49'),(94,'carefront-cases','6/69.out','ACTIVE',2,'2013-11-15 23:31:03','2013-11-15 23:31:05'),(95,'carefront-cases','6/70.out','ACTIVE',2,'2013-11-16 01:32:12','2013-11-16 01:32:13'),(96,'carefront-cases','6/71.out','ACTIVE',2,'2013-11-16 01:32:42','2013-11-16 01:32:43'),(97,'carefront-cases','6/72.out','ACTIVE',2,'2013-11-16 01:34:39','2013-11-16 01:34:39'),(98,'carefront-cases','6/73.out','ACTIVE',2,'2013-11-16 01:34:50','2013-11-16 01:34:51'),(99,'carefront-cases','6/74.out','ACTIVE',2,'2013-11-16 01:34:57','2013-11-16 01:34:57'),(100,'carefront-cases','48/75.out','ACTIVE',2,'2013-11-16 01:47:49','2013-11-16 01:47:51'),(101,'carefront-cases','48/76.out','ACTIVE',2,'2013-11-16 01:49:33','2013-11-16 01:49:33'),(102,'carefront-cases','50/77.out','ACTIVE',2,'2013-11-16 22:16:18','2013-11-16 22:16:19'),(103,'carefront-cases','50/0.out','ACTIVE',2,'2013-11-16 23:55:02','2013-11-16 23:55:03'),(104,'carefront-cases','50/83.out','ACTIVE',2,'2013-11-16 23:56:03','2013-11-16 23:56:04'),(105,'carefront-layout','1384648241','ACTIVE',1,'2013-11-17 00:30:41','2013-11-17 00:30:41'),(106,'carefront-client-layout','1384648251','ACTIVE',1,'2013-11-17 00:30:51','2013-11-17 00:30:52'),(107,'carefront-layout','1384648263','ACTIVE',1,'2013-11-17 00:31:07','2013-11-17 00:31:08'),(108,'carefront-client-layout','1384648276','ACTIVE',1,'2013-11-17 00:31:20','2013-11-17 00:31:20'),(109,'carefront-layout','1384649283','ACTIVE',1,'2013-11-17 00:48:07','2013-11-17 00:48:07'),(110,'carefront-client-layout','1384649297','ACTIVE',1,'2013-11-17 00:48:21','2013-11-17 00:48:22'),(111,'carefront-layout','1384716307','ACTIVE',1,'2013-11-17 19:25:07','2013-11-17 19:25:08'),(112,'carefront-client-layout','1384716322','ACTIVE',1,'2013-11-17 19:25:23','2013-11-17 19:25:23'),(113,'carefront-layout','1384716486','ACTIVE',1,'2013-11-17 19:28:06','2013-11-17 19:28:07'),(114,'carefront-client-layout','1384716500','ACTIVE',1,'2013-11-17 19:28:21','2013-11-17 19:28:22'),(115,'carefront-layout','1384716950','ACTIVE',1,'2013-11-17 19:35:51','2013-11-17 19:35:52'),(116,'carefront-client-layout','1384716965','ACTIVE',1,'2013-11-17 19:36:05','2013-11-17 19:36:06'),(117,'carefront-cases','53/96.out','ACTIVE',2,'2013-11-18 06:17:55','2013-11-18 06:17:57'),(118,'carefront-cases','53/97.out','ACTIVE',2,'2013-11-18 19:34:45','2013-11-18 19:34:47'),(119,'carefront-layout','1384910982','ACTIVE',1,'2013-11-20 01:29:41','2013-11-20 01:29:42'),(120,'carefront-layout','1384911003','ACTIVE',1,'2013-11-20 01:30:03','2013-11-20 01:30:04'),(121,'carefront-client-layout','1384911006','ACTIVE',1,'2013-11-20 01:30:06','2013-11-20 01:30:07'),(122,'carefront-layout','1384911441','ACTIVE',1,'2013-11-20 01:37:22','2013-11-20 01:37:24'),(123,'carefront-client-layout','1384911488','ACTIVE',1,'2013-11-20 01:38:08','2013-11-20 01:38:09'),(124,'carefront-layout','1384981429','ACTIVE',1,'2013-11-20 21:03:49','2013-11-20 21:03:50'),(125,'carefront-client-layout','1384981442','ACTIVE',1,'2013-11-20 21:04:02','2013-11-20 21:04:02'),(126,'carefront-cases','84/412.out','CREATING',1,'2013-11-23 17:43:25','0000-00-00 00:00:00'),(127,'carefront-doctor-layout-useast','1385249163','CREATING',1,'2013-11-23 23:26:03','0000-00-00 00:00:00'),(128,'carefront-doctor-layout-useast','1385249388','CREATING',1,'2013-11-23 23:29:48','0000-00-00 00:00:00'),(129,'carefront-doctor-layout-useast','1385249461','CREATING',1,'2013-11-23 23:31:02','0000-00-00 00:00:00'),(130,'carefront-doctor-layout-useast','1385249539','CREATING',1,'2013-11-23 23:32:19','0000-00-00 00:00:00'),(131,'carefront-doctor-layout-useast','1385249735','ACTIVE',1,'2013-11-23 23:35:35','2013-11-23 23:35:36'),(132,'carefront-doctor-layout-useast','1385250120','ACTIVE',1,'2013-11-23 23:42:00','2013-11-23 23:42:02'),(133,'carefront-doctor-layout-useast','1385250840','ACTIVE',1,'2013-11-23 23:54:00','2013-11-23 23:54:01'),(134,'carefront-doctor-layout-useast','1385250968','ACTIVE',1,'2013-11-23 23:56:08','2013-11-23 23:56:09'),(135,'carefront-cases-useast','84/413.out','CREATING',1,'2013-11-24 01:28:29','0000-00-00 00:00:00'),(136,'carefront-cases-useast','84/414.out','ACTIVE',1,'2013-11-24 01:30:25','2013-11-24 01:30:28'),(137,'carefront-layout','1385258558','ACTIVE',1,'2013-11-24 02:02:38','2013-11-24 02:02:39'),(138,'carefront-client-layout','1385258560','ACTIVE',1,'2013-11-24 02:02:40','2013-11-24 02:02:41'),(139,'carefront-layout','1385258716','ACTIVE',1,'2013-11-24 02:05:16','2013-11-24 02:05:17'),(140,'carefront-client-layout','1385258718','ACTIVE',1,'2013-11-24 02:05:18','2013-11-24 02:05:19'),(141,'carefront-layout','1385258945','ACTIVE',1,'2013-11-24 02:09:05','2013-11-24 02:09:06'),(142,'carefront-client-layout','1385258969','ACTIVE',1,'2013-11-24 02:09:29','2013-11-24 02:09:30'),(143,'carefront-layout','1385259067','ACTIVE',1,'2013-11-24 02:11:07','2013-11-24 02:11:09'),(144,'carefront-client-layout','1385259093','ACTIVE',1,'2013-11-24 02:11:34','2013-11-24 02:11:35'),(145,'carefront-layout','1385259622','ACTIVE',1,'2013-11-24 02:20:23','2013-11-24 02:20:23'),(146,'carefront-client-layout','1385259659','ACTIVE',1,'2013-11-24 02:21:00','2013-11-24 02:21:01'),(147,'carefront-doctor-layout-useast','1385260577','ACTIVE',1,'2013-11-24 02:36:17','2013-11-24 02:36:18'),(148,'carefront-doctor-layout-useast','1385260584','ACTIVE',1,'2013-11-24 02:36:24','2013-11-24 02:36:25'),(149,'carefront-doctor-visual-layout-useast','1385261177','ACTIVE',1,'2013-11-24 02:46:17','2013-11-24 02:46:18'),(150,'carefront-doctor-layout-useast','1385261184','ACTIVE',1,'2013-11-24 02:46:24','2013-11-24 02:46:25'),(151,'carefront-doctor-visual-layout-useast','1385332622','ACTIVE',1,'2013-11-24 22:37:03','2013-11-24 22:37:03'),(152,'carefront-doctor-layout-useast','1385332629','ACTIVE',1,'2013-11-24 22:37:10','2013-11-24 22:37:10'),(153,'carefront-cases-useast','84/415.out','ACTIVE',1,'2013-11-24 22:38:54','2013-11-24 22:38:57'),(154,'carefront-doctor-visual-layout-useast','1385335170','ACTIVE',1,'2013-11-24 23:19:30','2013-11-24 23:19:31'),(155,'carefront-doctor-layout-useast','1385335176','ACTIVE',1,'2013-11-24 23:19:36','2013-11-24 23:19:37'),(156,'carefront-doctor-visual-layout-useast','1385335246','ACTIVE',1,'2013-11-24 23:20:47','2013-11-24 23:20:47'),(157,'carefront-doctor-layout-useast','1385335253','ACTIVE',1,'2013-11-24 23:20:53','2013-11-24 23:20:54'),(158,'carefront-doctor-visual-layout-useast','1385335611','ACTIVE',1,'2013-11-24 23:26:51','2013-11-24 23:26:52'),(159,'carefront-doctor-layout-useast','1385335617','ACTIVE',1,'2013-11-24 23:26:58','2013-11-24 23:26:59'),(160,'carefront-cases-useast','86/416.out','ACTIVE',1,'2013-11-25 05:55:06','2013-11-25 05:55:08'),(161,'carefront-cases-useast','88/417.out','ACTIVE',1,'2013-11-25 22:42:41','2013-11-25 22:42:42'),(162,'cases-bucket-integ','271/869.jpg','ACTIVE',1,'2013-12-02 05:17:11','2013-12-02 05:17:11'),(163,'cases-bucket-integ','272/870.jpg','ACTIVE',1,'2013-12-02 05:24:15','2013-12-02 05:24:15'),(164,'cases-bucket-integ','273/871.jpg','ACTIVE',1,'2013-12-02 05:26:20','2013-12-02 05:26:21'),(165,'cases-bucket-integ','274/872.jpg','ACTIVE',1,'2013-12-02 05:28:58','2013-12-02 05:28:59'),(166,'cases-bucket-integ','275/873.jpg','ACTIVE',1,'2013-12-02 05:30:00','2013-12-02 05:30:01'),(167,'cases-bucket-integ','276/874.jpg','ACTIVE',1,'2013-12-02 05:32:53','2013-12-02 05:32:54'),(168,'cases-bucket-integ','277/875.jpg','ACTIVE',1,'2013-12-02 05:34:03','2013-12-02 05:34:04'),(169,'cases-bucket-integ','278/876.jpg','ACTIVE',1,'2013-12-02 05:36:03','2013-12-02 05:36:04'),(170,'cases-bucket-integ','279/877.jpg','ACTIVE',1,'2013-12-02 05:38:05','2013-12-02 05:38:05'),(171,'cases-bucket-integ','280/878.jpg','ACTIVE',1,'2013-12-02 05:39:06','2013-12-02 05:39:07'),(172,'cases-bucket-integ','281/879.jpg','ACTIVE',1,'2013-12-02 05:41:21','2013-12-02 05:41:22'),(173,'cases-bucket-integ','287/900.jpg','ACTIVE',1,'2013-12-02 05:43:48','2013-12-02 05:43:48'),(174,'carefront-layout','1386105770','ACTIVE',1,'2013-12-03 21:22:51','2013-12-03 21:22:52'),(175,'carefront-client-layout','1386105786','ACTIVE',1,'2013-12-03 21:23:07','2013-12-03 21:23:07'),(176,'carefront-layout','1386105911','ACTIVE',1,'2013-12-03 21:25:11','2013-12-03 21:25:11'),(177,'carefront-layout','1386105977','ACTIVE',1,'2013-12-03 21:26:17','2013-12-03 21:26:18'),(178,'carefront-client-layout','1386105992','ACTIVE',1,'2013-12-03 21:26:32','2013-12-03 21:26:33'),(179,'carefront-layout','1386106104','ACTIVE',1,'2013-12-03 21:28:24','2013-12-03 21:28:25'),(180,'carefront-client-layout','1386106121','ACTIVE',1,'2013-12-03 21:28:41','2013-12-03 21:28:42'),(181,'carefront-layout','1386106307','ACTIVE',1,'2013-12-03 21:31:47','2013-12-03 21:31:47'),(182,'carefront-client-layout','1386106322','ACTIVE',1,'2013-12-03 21:32:03','2013-12-03 21:32:03'),(183,'carefront-layout','1386107436','ACTIVE',1,'2013-12-03 21:50:36','2013-12-03 21:50:37'),(184,'carefront-client-layout','1386107457','ACTIVE',1,'2013-12-03 21:50:57','2013-12-03 21:50:58'),(185,'cases-bucket-integ','295/904.jpg','ACTIVE',1,'2013-12-03 22:42:15','2013-12-03 22:42:15'),(186,'cases-bucket-integ','302/908.jpg','ACTIVE',1,'2013-12-03 22:46:06','2013-12-03 22:46:07'),(187,'cases-bucket-integ','311/912.jpg','ACTIVE',1,'2013-12-04 00:00:34','2013-12-04 00:00:34'),(188,'cases-bucket-integ','323/938.jpg','ACTIVE',1,'2013-12-04 00:49:50','2013-12-04 00:49:51'),(189,'carefront-doctor-visual-layout-useast','1386192995','ACTIVE',1,'2013-12-04 21:36:36','2013-12-04 21:36:37'),(190,'carefront-doctor-layout-useast','1386193001','ACTIVE',1,'2013-12-04 21:36:42','2013-12-04 21:36:43'),(191,'carefront-layout','1386197945','ACTIVE',1,'2013-12-04 22:59:07','2013-12-04 22:59:07'),(192,'carefront-client-layout','1386197947','ACTIVE',1,'2013-12-04 22:59:09','2013-12-04 22:59:09'),(193,'carefront-layout','1386198118','ACTIVE',1,'2013-12-04 23:02:00','2013-12-04 23:02:01'),(194,'carefront-client-layout','1386198120','ACTIVE',1,'2013-12-04 23:02:02','2013-12-04 23:02:02'),(195,'carefront-layout','1386198161','ACTIVE',1,'2013-12-04 23:02:43','2013-12-04 23:02:43'),(196,'carefront-client-layout','1386198163','ACTIVE',1,'2013-12-04 23:02:44','2013-12-04 23:02:45'),(197,'carefront-layout','1386200480','ACTIVE',1,'2013-12-04 23:41:20','2013-12-04 23:41:21'),(198,'carefront-client-layout','1386200497','ACTIVE',1,'2013-12-04 23:41:37','2013-12-04 23:41:38'),(199,'carefront-doctor-visual-layout-useast','1386201694','ACTIVE',1,'2013-12-05 00:01:34','2013-12-05 00:01:35'),(200,'carefront-doctor-layout-useast','1386201701','ACTIVE',1,'2013-12-05 00:01:41','2013-12-05 00:01:41'),(201,'cases-bucket-integ','333/963.jpg','ACTIVE',1,'2013-12-05 00:35:31','2013-12-05 00:35:32'),(202,'carefront-layout','1386282428','ACTIVE',1,'2013-12-05 22:27:09','2013-12-05 22:27:09'),(203,'carefront-layout','1386282502','ACTIVE',1,'2013-12-05 22:28:22','2013-12-05 22:28:23'),(204,'carefront-layout','1386282533','ACTIVE',1,'2013-12-05 22:28:53','2013-12-05 22:28:53'),(205,'carefront-client-layout','1386282534','ACTIVE',1,'2013-12-05 22:28:55','2013-12-05 22:28:55'),(206,'carefront-layout','1386282723','ACTIVE',1,'2013-12-05 22:32:03','2013-12-05 22:32:04'),(207,'carefront-client-layout','1386282725','ACTIVE',1,'2013-12-05 22:32:05','2013-12-05 22:32:05'),(208,'carefront-layout','1386282797','ACTIVE',1,'2013-12-05 22:33:17','2013-12-05 22:33:18'),(209,'carefront-client-layout','1386282799','ACTIVE',1,'2013-12-05 22:33:19','2013-12-05 22:33:19'),(210,'carefront-layout','1386282867','ACTIVE',1,'2013-12-05 22:34:27','2013-12-05 22:34:28'),(211,'carefront-layout','1386282868','ACTIVE',1,'2013-12-05 22:34:29','2013-12-05 22:34:29'),(212,'carefront-client-layout','1386282869','ACTIVE',1,'2013-12-05 22:34:29','2013-12-05 22:34:29'),(213,'carefront-client-layout','1386282870','ACTIVE',1,'2013-12-05 22:34:30','2013-12-05 22:34:31'),(214,'carefront-layout','1386282907','ACTIVE',1,'2013-12-05 22:35:07','2013-12-05 22:35:08'),(215,'carefront-layout','1386282955','ACTIVE',1,'2013-12-05 22:35:55','2013-12-05 22:35:56'),(216,'carefront-client-layout','1386282957','ACTIVE',1,'2013-12-05 22:35:58','2013-12-05 22:35:58'),(217,'carefront-layout','1386282995','ACTIVE',1,'2013-12-05 22:36:36','2013-12-05 22:36:36'),(218,'carefront-client-layout','1386283019','ACTIVE',1,'2013-12-05 22:36:59','2013-12-05 22:37:00'),(219,'carefront-layout','1386283845','ACTIVE',1,'2013-12-05 22:50:46','2013-12-05 22:50:46'),(220,'carefront-client-layout','1386283866','ACTIVE',1,'2013-12-05 22:51:07','2013-12-05 22:51:07'),(221,'cases-bucket-integ','341/984.jpg','ACTIVE',1,'2013-12-05 23:03:53','2013-12-05 23:03:53'),(222,'carefront-layout','1386288449','ACTIVE',1,'2013-12-06 00:07:29','2013-12-06 00:07:30'),(223,'carefront-client-layout','1386288472','ACTIVE',1,'2013-12-06 00:07:52','2013-12-06 00:07:52'),(224,'carefront-layout','1386289480','ACTIVE',1,'2013-12-06 00:24:40','2013-12-06 00:24:41'),(225,'carefront-client-layout','1386289501','ACTIVE',1,'2013-12-06 00:25:01','2013-12-06 00:25:02'),(226,'carefront-layout','1386289752','ACTIVE',1,'2013-12-06 00:29:12','2013-12-06 00:29:13'),(227,'carefront-client-layout','1386289773','ACTIVE',1,'2013-12-06 00:29:33','2013-12-06 00:29:34'),(228,'carefront-layout','1386290024','ACTIVE',1,'2013-12-06 00:33:44','2013-12-06 00:33:45'),(229,'carefront-client-layout','1386290045','ACTIVE',1,'2013-12-06 00:34:05','2013-12-06 00:34:06'),(230,'carefront-layout','1386301390','ACTIVE',1,'2013-12-06 03:43:10','2013-12-06 03:43:10'),(231,'carefront-client-layout','1386301409','ACTIVE',1,'2013-12-06 03:43:30','2013-12-06 03:43:30'),(232,'cases-bucket-integ','349/998.jpg','ACTIVE',1,'2013-12-06 03:50:55','2013-12-06 03:50:55'),(233,'cases-bucket-integ','356/1005.jpg','ACTIVE',1,'2013-12-06 03:58:39','2013-12-06 03:58:41'),(234,'cases-bucket-integ','363/1027.jpg','ACTIVE',1,'2013-12-06 04:00:49','2013-12-06 04:00:51'),(235,'cases-bucket-integ','370/1049.jpg','ACTIVE',1,'2013-12-06 04:02:59','2013-12-06 04:03:00'),(236,'cases-bucket-integ','377/1071.jpg','ACTIVE',1,'2013-12-06 04:05:08','2013-12-06 04:05:08'),(237,'cases-bucket-integ','384/1093.jpg','ACTIVE',1,'2013-12-06 04:07:01','2013-12-06 04:07:02'),(238,'cases-bucket-integ','392/1115.jpg','ACTIVE',1,'2013-12-06 19:26:37','2013-12-06 19:26:37'),(239,'cases-bucket-integ','400/1137.jpg','ACTIVE',1,'2013-12-06 19:28:20','2013-12-06 19:28:21'),(240,'cases-bucket-integ','408/1159.jpg','ACTIVE',1,'2013-12-06 19:34:01','2013-12-06 19:34:01'),(241,'cases-bucket-integ','416/1178.jpg','ACTIVE',1,'2013-12-06 19:34:40','2013-12-06 19:34:41'),(242,'cases-bucket-integ','424/1188.jpg','ACTIVE',1,'2013-12-06 19:35:12','2013-12-06 19:35:12'),(243,'cases-bucket-integ','433/1209.jpg','ACTIVE',1,'2013-12-06 19:35:42','2013-12-06 19:35:43'),(244,'carefront-cases-useast','439/1226.md','ACTIVE',1,'2013-12-07 02:39:52','2013-12-07 02:39:52'),(245,'cases-bucket-integ','440/1227.jpg','ACTIVE',1,'2013-12-07 02:48:09','2013-12-07 02:48:09'),(246,'cases-bucket-integ','441/1228.jpg','ACTIVE',1,'2013-12-07 02:48:34','2013-12-07 02:48:35'),(247,'cases-bucket-integ','457/1271.jpg','ACTIVE',1,'2013-12-09 23:05:39','2013-12-09 23:05:40'),(248,'cases-bucket-integ','467/1293.jpg','ACTIVE',1,'2013-12-09 23:11:32','2013-12-09 23:11:32'),(249,'cases-bucket-integ','478/1316.jpg','ACTIVE',1,'2013-12-12 03:18:15','2013-12-12 03:18:16'),(250,'carefront-doctor-visual-layout-useast','1386926142','ACTIVE',1,'2013-12-13 09:15:41','2013-12-13 09:15:42'),(251,'carefront-doctor-layout-useast','1386926158','ACTIVE',1,'2013-12-13 09:15:57','2013-12-13 09:15:59'),(252,'cases-bucket-integ','489/1338.jpg','ACTIVE',1,'2013-12-13 18:26:31','2013-12-13 18:26:31'),(253,'cases-bucket-integ','499/1360.jpg','ACTIVE',1,'2013-12-14 20:57:22','2013-12-14 20:57:23'),(254,'cases-bucket-integ','511/1383.jpg','ACTIVE',1,'2013-12-16 00:35:21','2013-12-16 00:35:22'),(255,'cases-bucket-integ','521/1405.jpg','ACTIVE',1,'2013-12-16 06:41:46','2013-12-16 06:41:46'),(256,'cases-bucket-integ','531/1427.jpg','ACTIVE',1,'2013-12-16 06:43:59','2013-12-16 06:44:00'),(257,'cases-bucket-integ','541/1449.jpg','ACTIVE',1,'2013-12-16 07:02:33','2013-12-16 07:02:33'),(258,'cases-bucket-integ','551/1471.jpg','ACTIVE',1,'2013-12-16 07:06:36','2013-12-16 07:06:36'),(259,'carefront-layout','1387177900','ACTIVE',1,'2013-12-16 07:11:40','2013-12-16 07:11:41'),(260,'carefront-client-layout','1387177924','ACTIVE',1,'2013-12-16 07:12:04','2013-12-16 07:12:05'),(261,'cases-bucket-integ','561/1493.jpg','ACTIVE',1,'2013-12-16 07:14:42','2013-12-16 07:14:43'),(262,'cases-bucket-integ','571/1515.jpg','ACTIVE',1,'2013-12-16 07:19:56','2013-12-16 07:19:56'),(263,'cases-bucket-integ','581/1537.jpg','ACTIVE',1,'2013-12-16 07:30:55','2013-12-16 07:30:55'),(264,'cases-bucket-integ','591/1559.jpg','ACTIVE',1,'2013-12-16 17:13:41','2013-12-16 17:13:42'),(265,'carefront-layout','1387215398','ACTIVE',1,'2013-12-16 17:36:39','2013-12-16 17:36:39'),(266,'carefront-client-layout','1387215420','ACTIVE',1,'2013-12-16 17:37:00','2013-12-16 17:37:01'),(267,'cases-bucket-integ','601/1581.jpg','ACTIVE',1,'2013-12-16 21:51:52','2013-12-16 21:51:53'),(268,'cases-bucket-integ','615/1608.jpg','ACTIVE',1,'2013-12-16 21:58:39','2013-12-16 21:58:40'),(269,'cases-bucket-integ','625/1630.jpg','ACTIVE',1,'2013-12-16 23:19:36','2013-12-16 23:19:37'),(270,'carefront-doctor-visual-layout-useast','1387240525','ACTIVE',1,'2013-12-17 00:35:25','2013-12-17 00:35:26'),(271,'carefront-layout','1387242799','ACTIVE',1,'2013-12-17 01:13:19','2013-12-17 01:13:20'),(272,'carefront-client-layout','1387242819','ACTIVE',1,'2013-12-17 01:13:39','2013-12-17 01:13:40'),(273,'carefront-layout','1387242996','ACTIVE',1,'2013-12-17 01:16:36','2013-12-17 01:16:37'),(274,'carefront-layout','1387243085','ACTIVE',1,'2013-12-17 01:18:05','2013-12-17 01:18:06'),(275,'carefront-client-layout','1387243106','ACTIVE',1,'2013-12-17 01:18:26','2013-12-17 01:18:27'),(276,'carefront-layout','1387243980','ACTIVE',1,'2013-12-17 01:33:00','2013-12-17 01:33:01'),(277,'carefront-client-layout','1387244001','ACTIVE',1,'2013-12-17 01:33:21','2013-12-17 01:33:22'),(278,'carefront-doctor-visual-layout-useast','1387245485','ACTIVE',1,'2013-12-17 01:58:05','2013-12-17 01:58:06'),(279,'carefront-doctor-layout-useast','1387245493','ACTIVE',1,'2013-12-17 01:58:13','2013-12-17 01:58:14'),(280,'carefront-doctor-visual-layout-useast','1387245757','ACTIVE',1,'2013-12-17 02:02:37','2013-12-17 02:02:38'),(281,'carefront-doctor-layout-useast','1387245764','ACTIVE',1,'2013-12-17 02:02:45','2013-12-17 02:02:45'),(282,'carefront-doctor-visual-layout-useast','1387245882','ACTIVE',1,'2013-12-17 02:04:42','2013-12-17 02:04:42'),(283,'carefront-doctor-layout-useast','1387245889','ACTIVE',1,'2013-12-17 02:04:49','2013-12-17 02:04:50'),(284,'carefront-layout','1387247095','ACTIVE',1,'2013-12-17 02:24:56','2013-12-17 02:24:56'),(285,'carefront-client-layout','1387247112','ACTIVE',1,'2013-12-17 02:25:12','2013-12-17 02:25:13'),(286,'carefront-layout','1387255404','ACTIVE',1,'2013-12-17 04:43:24','2013-12-17 04:43:25'),(287,'carefront-client-layout','1387255428','ACTIVE',1,'2013-12-17 04:43:48','2013-12-17 04:43:49'),(288,'carefront-doctor-visual-layout-useast','1387255517','ACTIVE',1,'2013-12-17 04:45:17','2013-12-17 04:45:17'),(289,'carefront-doctor-layout-useast','1387255528','ACTIVE',1,'2013-12-17 04:45:29','2013-12-17 04:45:29'),(290,'carefront-doctor-visual-layout-useast','1387256149','ACTIVE',1,'2013-12-17 04:55:49','2013-12-17 04:55:50'),(291,'carefront-doctor-layout-useast','1387256152','ACTIVE',1,'2013-12-17 04:55:53','2013-12-17 04:55:53'),(292,'carefront-doctor-visual-layout-useast','1387256326','ACTIVE',1,'2013-12-17 04:58:46','2013-12-17 04:58:46'),(293,'carefront-doctor-layout-useast','1387256329','ACTIVE',1,'2013-12-17 04:58:49','2013-12-17 04:58:50'),(294,'carefront-doctor-visual-layout-useast','1387256699','ACTIVE',1,'2013-12-17 05:04:59','2013-12-17 05:04:59'),(295,'carefront-doctor-layout-useast','1387256702','ACTIVE',1,'2013-12-17 05:05:02','2013-12-17 05:05:03'),(296,'carefront-doctor-visual-layout-useast','1387257840','ACTIVE',1,'2013-12-17 05:24:00','2013-12-17 05:24:01'),(297,'carefront-doctor-layout-useast','1387257845','ACTIVE',1,'2013-12-17 05:24:05','2013-12-17 05:24:05'),(298,'carefront-doctor-visual-layout-useast','1387287697','ACTIVE',1,'2013-12-17 13:41:36','2013-12-17 13:41:36'),(299,'carefront-doctor-layout-useast','1387287701','ACTIVE',1,'2013-12-17 13:41:40','2013-12-17 13:41:40'),(300,'cases-bucket-integ','635/1652.jpg','ACTIVE',1,'2013-12-17 13:49:40','2013-12-17 13:49:41'),(301,'cases-bucket-integ','645/1674.jpg','ACTIVE',1,'2013-12-17 13:52:01','2013-12-17 13:52:02'),(302,'carefront-doctor-visual-layout-useast','1387324196','ACTIVE',1,'2013-12-17 23:49:56','2013-12-17 23:49:57'),(303,'carefront-doctor-layout-useast','1387324200','ACTIVE',1,'2013-12-17 23:50:00','2013-12-17 23:50:01'),(304,'carefront-doctor-visual-layout-useast','1387324640','ACTIVE',1,'2013-12-17 23:57:20','2013-12-17 23:57:21'),(305,'carefront-doctor-layout-useast','1387324644','ACTIVE',1,'2013-12-17 23:57:24','2013-12-17 23:57:25'),(306,'cases-bucket-integ','660/1696.jpg','ACTIVE',1,'2013-12-18 17:31:53','2013-12-18 17:31:53'),(307,'cases-bucket-integ','688/1718.jpg','ACTIVE',1,'2013-12-18 22:02:31','2013-12-18 22:02:32'),(308,'cases-bucket-integ','700/1740.jpg','ACTIVE',1,'2013-12-18 22:45:14','2013-12-18 22:45:14');
/*!40000 ALTER TABLE `object_storage` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `layout_version`
--

DROP TABLE IF EXISTS `layout_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `layout_version` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `object_storage_id` int(10) unsigned NOT NULL,
  `syntax_version` int(10) unsigned NOT NULL,
  `health_condition_id` int(10) unsigned NOT NULL,
  `comment` varchar(600) DEFAULT NULL,
  `status` varchar(250) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified_date` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  `role` varchar(250) DEFAULT NULL,
  `layout_purpose` varchar(250) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `object_storage_id` (`object_storage_id`,`syntax_version`,`health_condition_id`,`status`),
  KEY `treatment_id` (`health_condition_id`),
  CONSTRAINT `layout_version_ibfk_1` FOREIGN KEY (`health_condition_id`) REFERENCES `health_condition` (`id`),
  CONSTRAINT `layout_version_ibfk_2` FOREIGN KEY (`object_storage_id`) REFERENCES `object_storage` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=110 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `layout_version`
--

LOCK TABLES `layout_version` WRITE;
/*!40000 ALTER TABLE `layout_version` DISABLE KEYS */;
INSERT INTO `layout_version` VALUES (15,33,1,1,'automatically generated','DEPCRECATED','2013-11-08 19:13:06','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(16,35,1,1,'automatically generated','DEPCRECATED','2013-11-08 19:13:06','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(17,37,1,1,'automatically generated','DEPCRECATED','2013-11-08 19:13:06','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(18,39,1,1,'automatically generated','DEPCRECATED','2013-11-08 19:13:06','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(19,41,1,1,'automatically generated','DEPCRECATED','2013-11-08 19:13:06','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(20,43,1,1,'automatically generated','DEPCRECATED','2013-11-08 19:21:07','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(21,45,1,1,'automatically generated','DEPCRECATED','2013-11-11 05:46:47','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(22,47,1,1,'automatically generated','DEPCRECATED','2013-11-11 05:57:25','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(23,49,1,1,'automatically generated','CREATING','2013-11-11 05:58:19','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(24,50,1,1,'automatically generated','DEPCRECATED','2013-11-11 05:58:34','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(25,52,1,1,'automatically generated','DEPCRECATED','2013-11-11 06:02:11','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(26,55,1,1,'automatically generated','DEPCRECATED','2013-11-11 06:06:49','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(27,57,1,1,'automatically generated','DEPCRECATED','2013-11-12 15:01:56','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(28,59,1,1,'automatically generated','DEPCRECATED','2013-11-12 15:34:07','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(29,61,1,1,'automatically generated','DEPCRECATED','2013-11-12 15:34:38','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(30,62,1,1,'automatically generated','DEPCRECATED','2013-11-12 15:34:40','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(31,65,1,1,'automatically generated','DEPCRECATED','2013-11-12 15:38:05','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(32,67,1,1,'automatically generated','DEPCRECATED','2013-11-12 15:39:03','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(33,69,1,1,'automatically generated','DEPCRECATED','2013-11-12 17:02:09','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(34,71,1,1,'automatically generated','DEPCRECATED','2013-11-12 17:03:58','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(35,73,1,1,'automatically generated','DEPCRECATED','2013-11-12 17:15:09','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(36,75,1,1,'automatically generated','DEPCRECATED','2013-11-12 19:36:42','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(37,105,1,1,'automatically generated','DEPCRECATED','2013-11-17 00:30:41','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(38,107,1,1,'automatically generated','DEPCRECATED','2013-11-17 00:31:08','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(39,109,1,1,'automatically generated','DEPCRECATED','2013-11-17 00:48:08','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(40,111,1,1,'automatically generated','DEPCRECATED','2013-11-17 19:25:08','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(41,113,1,1,'automatically generated','DEPCRECATED','2013-11-17 19:28:07','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(42,115,1,1,'automatically generated','DEPCRECATED','2013-11-17 19:35:52','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(43,119,1,1,'automatically generated','CREATING','2013-11-20 01:29:43','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(44,120,1,1,'automatically generated','DEPCRECATED','2013-11-20 01:30:04','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(45,122,1,1,'automatically generated','DEPCRECATED','2013-11-20 01:37:27','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(46,124,1,1,'automatically generated','DEPCRECATED','2013-11-20 21:03:50','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(48,131,1,1,'automatically generated','CREATING','2013-11-23 23:35:36','2013-12-17 01:02:18','DOCTOR','REVIEW'),(49,132,1,1,'automatically generated','CREATING','2013-11-23 23:42:03','2013-12-17 01:02:18','DOCTOR','REVIEW'),(50,133,1,1,'automatically generated','CREATING','2013-11-23 23:54:01','2013-12-17 01:02:18','DOCTOR','REVIEW'),(51,134,1,1,'automatically generated','DEPCRECATED','2013-11-23 23:56:09','2013-12-17 01:02:18','DOCTOR','REVIEW'),(52,137,1,1,'automatically generated','DEPCRECATED','2013-11-24 02:02:39','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(53,139,1,1,'automatically generated','DEPCRECATED','2013-11-24 02:05:17','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(54,141,1,1,'automatically generated','DEPCRECATED','2013-11-24 02:09:07','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(55,143,1,1,'automatically generated','DEPCRECATED','2013-11-24 02:11:10','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(56,145,1,1,'automatically generated','DEPCRECATED','2013-11-24 02:20:24','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(57,147,1,1,'automatically generated','DEPCRECATED','2013-11-24 02:36:19','2013-12-17 01:02:18','DOCTOR','REVIEW'),(58,149,1,1,'automatically generated','DEPCRECATED','2013-11-24 02:46:18','2013-12-17 01:02:18','DOCTOR','REVIEW'),(59,151,1,1,'automatically generated','DEPCRECATED','2013-11-24 22:37:04','2013-12-17 01:02:18','DOCTOR','REVIEW'),(60,154,1,1,'automatically generated','DEPCRECATED','2013-11-24 23:19:31','2013-12-17 01:02:18','DOCTOR','REVIEW'),(61,156,1,1,'automatically generated','DEPCRECATED','2013-11-24 23:20:48','2013-12-17 01:02:18','DOCTOR','REVIEW'),(62,158,1,1,'automatically generated','DEPCRECATED','2013-11-24 23:26:52','2013-12-17 01:02:18','DOCTOR','REVIEW'),(63,174,1,1,'automatically generated','DEPCRECATED','2013-12-03 21:22:52','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(64,176,1,1,'automatically generated','CREATING','2013-12-03 21:25:12','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(65,177,1,1,'automatically generated','DEPCRECATED','2013-12-03 21:26:18','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(66,179,1,1,'automatically generated','DEPCRECATED','2013-12-03 21:28:26','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(67,181,1,1,'automatically generated','DEPCRECATED','2013-12-03 21:31:48','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(68,183,1,1,'automatically generated','DEPCRECATED','2013-12-03 21:50:37','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(69,189,1,1,'automatically generated','DEPCRECATED','2013-12-04 21:36:38','2013-12-17 01:02:18','DOCTOR','REVIEW'),(70,191,1,1,'automatically generated','DEPCRECATED','2013-12-04 22:59:08','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(71,193,1,1,'automatically generated','DEPCRECATED','2013-12-04 23:02:01','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(72,195,1,1,'automatically generated','DEPCRECATED','2013-12-04 23:02:44','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(73,197,1,1,'automatically generated','DEPCRECATED','2013-12-04 23:41:21','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(74,199,1,1,'automatically generated','DEPCRECATED','2013-12-05 00:01:35','2013-12-17 01:02:18','DOCTOR','REVIEW'),(75,202,1,1,'automatically generated','CREATING','2013-12-05 22:27:09','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(76,203,1,1,'automatically generated','CREATING','2013-12-05 22:28:23','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(77,204,1,1,'automatically generated','DEPCRECATED','2013-12-05 22:28:54','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(78,206,1,1,'automatically generated','DEPCRECATED','2013-12-05 22:32:04','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(79,208,1,1,'automatically generated','DEPCRECATED','2013-12-05 22:33:18','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(80,210,1,1,'automatically generated','DEPCRECATED','2013-12-05 22:34:28','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(81,211,1,1,'automatically generated','DEPCRECATED','2013-12-05 22:34:29','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(82,214,1,1,'automatically generated','DEPCRECATED','2013-12-05 22:35:09','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(83,215,1,1,'automatically generated','CREATING','2013-12-05 22:35:56','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(84,217,1,1,'automatically generated','DEPCRECATED','2013-12-05 22:36:37','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(85,219,1,1,'automatically generated','DEPCRECATED','2013-12-05 22:50:46','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(86,222,1,1,'automatically generated','DEPCRECATED','2013-12-06 00:07:30','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(87,224,1,1,'automatically generated','DEPCRECATED','2013-12-06 00:24:41','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(88,226,1,1,'automatically generated','DEPCRECATED','2013-12-06 00:29:13','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(89,228,1,1,'automatically generated','DEPCRECATED','2013-12-06 00:33:45','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(90,230,1,1,'automatically generated','DEPCRECATED','2013-12-06 03:43:11','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(91,250,1,1,'automatically generated','DEPCRECATED','2013-12-13 09:15:42','2013-12-17 02:04:50','DOCTOR','REVIEW'),(92,259,1,1,'automatically generated','DEPCRECATED','2013-12-16 07:11:41','2013-12-17 01:02:18','PATIENT','CONDITION_INTAKE'),(93,265,1,1,'automatically generated','DEPCRECATED','2013-12-16 17:36:39','2013-12-17 01:18:28','PATIENT','CONDITION_INTAKE'),(95,274,1,1,'automatically generated','DEPCRECATED','2013-12-17 01:18:06','2013-12-17 01:33:22','PATIENT','CONDITION_INTAKE'),(96,276,1,1,'automatically generated','DEPCRECATED','2013-12-17 01:33:01','2013-12-17 02:25:13','PATIENT','CONDITION_INTAKE'),(97,278,1,1,'automatically generated','CREATING','2013-12-17 01:58:06','0000-00-00 00:00:00','DOCTOR','REVIEW'),(98,280,1,1,'automatically generated','CREATING','2013-12-17 02:02:38','0000-00-00 00:00:00','DOCTOR','REVIEW'),(99,282,1,1,'automatically generated','DEPCRECATED','2013-12-17 02:04:43','2013-12-17 04:45:30','DOCTOR','REVIEW'),(100,284,1,1,'automatically generated','DEPCRECATED','2013-12-17 02:24:57','2013-12-17 04:43:49','PATIENT','CONDITION_INTAKE'),(101,286,1,1,'automatically generated','ACTIVE','2013-12-17 04:43:25','2013-12-17 04:43:50','PATIENT','CONDITION_INTAKE'),(102,288,1,1,'automatically generated','ACTIVE','2013-12-17 04:45:18','2013-12-17 04:45:31','DOCTOR','REVIEW'),(103,290,1,1,'automatically generated','DEPCRECATED','2013-12-17 04:55:50','2013-12-17 04:58:51','DOCTOR','DIAGNOSE'),(104,292,1,1,'automatically generated','DEPCRECATED','2013-12-17 04:58:47','2013-12-17 05:05:03','DOCTOR','DIAGNOSE'),(105,294,1,1,'automatically generated','DEPCRECATED','2013-12-17 05:05:00','2013-12-17 05:24:06','DOCTOR','DIAGNOSE'),(106,296,1,1,'automatically generated','DEPCRECATED','2013-12-17 05:24:01','2013-12-17 13:41:41','DOCTOR','DIAGNOSE'),(107,298,1,1,'automatically generated','DEPCRECATED','2013-12-17 13:41:36','2013-12-17 23:50:01','DOCTOR','DIAGNOSE'),(108,302,1,1,'automatically generated','DEPCRECATED','2013-12-17 23:49:57','2013-12-17 23:57:25','DOCTOR','DIAGNOSE'),(109,304,1,1,'automatically generated','ACTIVE','2013-12-17 23:57:21','2013-12-17 23:57:25','DOCTOR','DIAGNOSE');
/*!40000 ALTER TABLE `layout_version` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `dr_layout_version`
--

DROP TABLE IF EXISTS `dr_layout_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_layout_version` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `object_storage_id` int(10) unsigned NOT NULL,
  `layout_version_id` int(10) unsigned NOT NULL,
  `status` varchar(250) NOT NULL,
  `modified_date` timestamp NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `health_condition_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `object_storage_id` (`object_storage_id`),
  KEY `health_condition_id` (`health_condition_id`),
  CONSTRAINT `dr_layout_version_ibfk_3` FOREIGN KEY (`health_condition_id`) REFERENCES `health_condition` (`id`),
  CONSTRAINT `dr_layout_version_ibfk_1` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `dr_layout_version_ibfk_2` FOREIGN KEY (`object_storage_id`) REFERENCES `object_storage` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dr_layout_version`
--

LOCK TABLES `dr_layout_version` WRITE;
/*!40000 ALTER TABLE `dr_layout_version` DISABLE KEYS */;
INSERT INTO `dr_layout_version` VALUES (1,132,49,'CREATING','0000-00-00 00:00:00','2013-11-23 23:42:03',1),(2,133,50,'CREATING','0000-00-00 00:00:00','2013-11-23 23:54:02',1),(3,134,51,'DEPCRECATED','2013-11-24 02:36:26','2013-11-23 23:56:10',1),(4,148,57,'DEPCRECATED','2013-11-24 02:46:25','2013-11-24 02:36:25',1),(5,150,58,'DEPCRECATED','2013-11-24 22:37:11','2013-11-24 02:46:25',1),(6,152,59,'DEPCRECATED','2013-11-24 23:19:38','2013-11-24 22:37:11',1),(7,155,60,'DEPCRECATED','2013-11-24 23:20:55','2013-11-24 23:19:37',1),(8,157,61,'DEPCRECATED','2013-11-24 23:26:59','2013-11-24 23:20:54',1),(9,159,62,'DEPCRECATED','2013-12-04 21:36:44','2013-11-24 23:26:59',1),(10,190,69,'DEPCRECATED','2013-12-05 00:01:42','2013-12-04 21:36:43',1),(11,200,74,'DEPCRECATED','2013-12-13 09:16:00','2013-12-05 00:01:42',1),(12,251,91,'DEPCRECATED','2013-12-17 02:04:50','2013-12-13 09:15:59',1),(13,279,97,'CREATING','0000-00-00 00:00:00','2013-12-17 01:58:14',1),(14,281,98,'CREATING','0000-00-00 00:00:00','2013-12-17 02:02:45',1),(15,283,99,'DEPCRECATED','2013-12-17 04:45:30','2013-12-17 02:04:50',1),(16,289,102,'ACTIVE','2013-12-17 04:45:31','2013-12-17 04:45:30',1),(17,291,103,'DEPCRECATED','2013-12-17 04:58:51','2013-12-17 04:55:53',1),(18,293,104,'DEPCRECATED','2013-12-17 05:05:03','2013-12-17 04:58:50',1),(19,295,105,'DEPCRECATED','2013-12-17 05:24:06','2013-12-17 05:05:03',1),(20,297,106,'DEPCRECATED','2013-12-17 13:41:41','2013-12-17 05:24:05',1),(21,299,107,'DEPCRECATED','2013-12-17 23:50:01','2013-12-17 13:41:40',1),(22,303,108,'DEPCRECATED','2013-12-17 23:57:25','2013-12-17 23:50:01',1),(23,305,109,'ACTIVE','2013-12-17 23:57:26','2013-12-17 23:57:25',1);
/*!40000 ALTER TABLE `dr_layout_version` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `care_providing_state`
--

DROP TABLE IF EXISTS `care_providing_state`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `care_providing_state` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `state` varchar(100) NOT NULL,
  `health_condition_id` int(10) unsigned NOT NULL,
  `long_state` varchar(250) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `health_condition_id` (`health_condition_id`),
  CONSTRAINT `care_providing_state_ibfk_1` FOREIGN KEY (`health_condition_id`) REFERENCES `health_condition` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `care_providing_state`
--

LOCK TABLES `care_providing_state` WRITE;
/*!40000 ALTER TABLE `care_providing_state` DISABLE KEYS */;
INSERT INTO `care_providing_state` VALUES (1,'CA',1,'California');
/*!40000 ALTER TABLE `care_providing_state` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `dispense_unit`
--

DROP TABLE IF EXISTS `dispense_unit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dispense_unit` (
  `id` int(10) unsigned NOT NULL,
  `dispense_unit_text_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dispense_unit_text_id` (`dispense_unit_text_id`),
  CONSTRAINT `dispense_unit_ibfk_1` FOREIGN KEY (`dispense_unit_text_id`) REFERENCES `app_text` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dispense_unit`
--

LOCK TABLES `dispense_unit` WRITE;
/*!40000 ALTER TABLE `dispense_unit` DISABLE KEYS */;
INSERT INTO `dispense_unit` VALUES (1,212),(2,213),(3,214),(4,215),(5,216),(6,217),(7,218),(8,219),(9,220),(10,221),(11,222),(12,223),(13,224),(14,225),(15,226),(16,227),(17,228),(18,229),(19,230),(20,231),(21,232),(22,233),(23,234),(24,235),(25,236),(26,237),(27,238),(28,239),(29,240),(30,241),(31,242),(32,243),(33,244),(34,245),(35,246),(36,247),(37,248),(38,249),(39,250),(40,251),(41,252),(42,253),(43,254),(44,255),(45,256),(46,257),(47,258),(48,259),(49,260),(50,261),(51,262),(52,263),(53,264),(54,265),(55,266),(56,267),(57,268),(58,269),(59,270),(60,271),(61,272),(62,273),(63,274),(64,275),(65,276),(66,277),(67,278),(68,279),(69,280),(70,281),(71,282),(72,283),(73,284),(74,285),(75,286),(76,287),(77,288),(78,289),(79,290),(80,291),(81,292),(82,293);
/*!40000 ALTER TABLE `dispense_unit` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `drug_name`
--

DROP TABLE IF EXISTS `drug_name`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_name` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(150) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `drug_name`
--

LOCK TABLES `drug_name` WRITE;
/*!40000 ALTER TABLE `drug_name` DISABLE KEYS */;
INSERT INTO `drug_name` VALUES (1,'Benzoyl Peroxide Topical');
/*!40000 ALTER TABLE `drug_name` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `drug_route`
--

DROP TABLE IF EXISTS `drug_route`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_route` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(150) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `drug_route`
--

LOCK TABLES `drug_route` WRITE;
/*!40000 ALTER TABLE `drug_route` DISABLE KEYS */;
INSERT INTO `drug_route` VALUES (1,'topical'),(2,'compounding');
/*!40000 ALTER TABLE `drug_route` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `drug_form`
--

DROP TABLE IF EXISTS `drug_form`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_form` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(150) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `drug_form`
--

LOCK TABLES `drug_form` WRITE;
/*!40000 ALTER TABLE `drug_form` DISABLE KEYS */;
INSERT INTO `drug_form` VALUES (1,'powder'),(2,'bar'),(3,'cream'),(4,'foam'),(5,'gel'),(6,'kit'),(7,'liquid'),(8,'lotion'),(9,'pad');
/*!40000 ALTER TABLE `drug_form` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `drug_supplemental_instruction`
--

DROP TABLE IF EXISTS `drug_supplemental_instruction`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_supplemental_instruction` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `text` varchar(150) NOT NULL,
  `drug_name_id` int(10) unsigned NOT NULL,
  `drug_form_id` int(10) unsigned DEFAULT NULL,
  `drug_route_id` int(10) unsigned DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `drug_route_id` (`drug_route_id`),
  CONSTRAINT `drug_supplemental_instruction_ibfk_1` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `drug_supplemental_instruction_ibfk_2` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `drug_supplemental_instruction_ibfk_3` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `drug_supplemental_instruction`
--

LOCK TABLES `drug_supplemental_instruction` WRITE;
/*!40000 ALTER TABLE `drug_supplemental_instruction` DISABLE KEYS */;
INSERT INTO `drug_supplemental_instruction` VALUES (1,'Benzoyl peroxide level instruction 1',1,NULL,NULL,'ACTIVE','2013-12-28 19:11:16'),(2,'Benzoyl peroxide level instruction 2',1,NULL,NULL,'ACTIVE','2013-12-28 19:11:16'),(3,'Benzoyl peroxide and route topical level instruction 1',1,NULL,1,'ACTIVE','2013-12-28 19:11:16'),(4,'Benzoyl peroxide and route compounding level instruction 1',1,NULL,2,'ACTIVE','2013-12-28 19:11:16'),(5,'Benzoyl peroxide, route topical and form cream level instruction 1',1,3,1,'ACTIVE','2013-12-28 19:11:16'),(6,'Benzoyl peroxide, route topical and form gel level instruction 1',1,5,1,'ACTIVE','2013-12-28 19:11:16'),(7,'Benzoyl peroxide, route topical and form liquid level instruction 1',1,7,1,'ACTIVE','2013-12-28 19:11:16'),(8,'Benzoyl peroxide level instruction 1',1,NULL,NULL,'ACTIVE','2013-12-30 13:26:48'),(9,'Benzoyl peroxide level instruction 2',1,NULL,NULL,'ACTIVE','2013-12-30 13:26:48'),(10,'Benzoyl peroxide and route topical level instruction 1',1,NULL,1,'ACTIVE','2013-12-30 13:26:48'),(11,'Benzoyl peroxide and route compounding level instruction 1',1,NULL,2,'ACTIVE','2013-12-30 13:26:48'),(12,'Benzoyl peroxide, route topical and form cream level instruction 1',1,3,1,'ACTIVE','2013-12-30 13:26:49'),(13,'Benzoyl peroxide, route topical and form gel level instruction 1',1,5,1,'ACTIVE','2013-12-30 13:26:49'),(14,'Benzoyl peroxide, route topical and form liquid level instruction 1',1,7,1,'ACTIVE','2013-12-30 13:26:49');
/*!40000 ALTER TABLE `drug_supplemental_instruction` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2014-01-03  6:01:52
