-- MySQL dump 10.13  Distrib 5.6.13, for osx10.8 (x86_64)
--
-- Host: dev-db-3.ccvrwjdx3gvp.us-east-1.rds.amazonaws.com    Database: database_6864
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
-- Table structure for table `account`
--

DROP TABLE IF EXISTS `account`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(250) DEFAULT NULL,
  `password` varbinary(250) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=94 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
) ENGINE=InnoDB AUTO_INCREMENT=169 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `auth_token`
--

DROP TABLE IF EXISTS `auth_token`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_token` (
  `token` varbinary(250) NOT NULL DEFAULT '',
  `account_id` int(10) unsigned NOT NULL,
  `created` timestamp NOT NULL,
  `expires` timestamp NOT NULL,
  PRIMARY KEY (`token`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `auth_token_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
  PRIMARY KEY (`id`),
  UNIQUE KEY `object_storage_id` (`object_storage_id`,`syntax_version`,`health_condition_id`,`status`),
  KEY `treatment_id` (`health_condition_id`),
  CONSTRAINT `layout_version_ibfk_1` FOREIGN KEY (`health_condition_id`) REFERENCES `health_condition` (`id`),
  CONSTRAINT `layout_version_ibfk_2` FOREIGN KEY (`object_storage_id`) REFERENCES `object_storage` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=63 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
) ENGINE=InnoDB AUTO_INCREMENT=175 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
) ENGINE=InnoDB AUTO_INCREMENT=162 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient`
--

DROP TABLE IF EXISTS `patient`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `first_name` varchar(500) NOT NULL,
  `last_name` varchar(500) NOT NULL,
  `dob` date NOT NULL,
  `gender` varchar(500) NOT NULL,
  `zip_code` varchar(500) NOT NULL,
  `status` varchar(500) NOT NULL,
  `account_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `patient_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=91 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_info_intake`
--

DROP TABLE IF EXISTS `patient_info_intake`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_info_intake` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_visit_id` int(10) unsigned DEFAULT NULL,
  `question_id` int(10) unsigned NOT NULL,
  `potential_answer_id` int(10) unsigned NOT NULL,
  `answer_text` varchar(600) DEFAULT NULL,
  `layout_version_id` int(10) unsigned NOT NULL,
  `answered_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` varchar(100) NOT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `object_storage_id` int(10) unsigned DEFAULT NULL,
  `parent_info_intake_id` int(10) unsigned DEFAULT NULL,
  `summary_localized_text_id` int(10) unsigned DEFAULT NULL,
  `parent_question_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_visit_id` (`patient_visit_id`),
  KEY `question_id` (`question_id`),
  KEY `potential_answer_id` (`potential_answer_id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `patient_id` (`patient_id`),
  KEY `object_storage_id` (`object_storage_id`),
  KEY `parent_info_intake_id` (`parent_info_intake_id`),
  KEY `summary_localized_text_id` (`summary_localized_text_id`),
  KEY `parent_question_id` (`parent_question_id`),
  CONSTRAINT `patient_info_intake_ibfk_9` FOREIGN KEY (`parent_question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `patient_info_intake_ibfk_1` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`),
  CONSTRAINT `patient_info_intake_ibfk_2` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `patient_info_intake_ibfk_3` FOREIGN KEY (`potential_answer_id`) REFERENCES `potential_answer` (`id`),
  CONSTRAINT `patient_info_intake_ibfk_4` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `patient_info_intake_ibfk_5` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `patient_info_intake_ibfk_6` FOREIGN KEY (`object_storage_id`) REFERENCES `object_storage` (`id`),
  CONSTRAINT `patient_info_intake_ibfk_7` FOREIGN KEY (`parent_info_intake_id`) REFERENCES `patient_info_intake` (`id`),
  CONSTRAINT `patient_info_intake_ibfk_8` FOREIGN KEY (`summary_localized_text_id`) REFERENCES `app_text` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=502 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
) ENGINE=InnoDB AUTO_INCREMENT=43 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_visit`
--

DROP TABLE IF EXISTS `patient_visit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_visit` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `opened_date` timestamp NULL DEFAULT NULL,
  `closed_date` timestamp NULL DEFAULT NULL,
  `health_condition_id` int(10) unsigned NOT NULL,
  `status` varchar(100) NOT NULL,
  `layout_version_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`),
  KEY `treatment_id` (`health_condition_id`),
  KEY `layout_version_id` (`layout_version_id`),
  CONSTRAINT `patient_visit_ibfk_3` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `patient_visit_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `patient_visit_ibfk_2` FOREIGN KEY (`health_condition_id`) REFERENCES `health_condition` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=89 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
) ENGINE=InnoDB AUTO_INCREMENT=102 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
) ENGINE=InnoDB AUTO_INCREMENT=35 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
  CONSTRAINT `question_fields_ibfk_1` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `question_fields_ibfk_2` FOREIGN KEY (`app_text_id`) REFERENCES `app_text` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2013-12-05 14:48:50
