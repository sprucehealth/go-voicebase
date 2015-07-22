-- MySQL dump 10.13  Distrib 5.6.25, for osx10.10 (x86_64)
--
-- Host: localhost    Database: database_30826
-- ------------------------------------------------------
-- Server version	5.6.25

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
  `role_type_id` int(10) unsigned NOT NULL,
  `registration_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `last_opened_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `two_factor_enabled` tinyint(1) NOT NULL DEFAULT '0',
  `account_code` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  KEY `role_type_id` (`role_type_id`),
  KEY `account_code` (`account_code`),
  CONSTRAINT `account_ibfk_1` FOREIGN KEY (`role_type_id`) REFERENCES `role_type` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=94 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_app_version`
--

DROP TABLE IF EXISTS `account_app_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_app_version` (
  `account_id` int(10) unsigned NOT NULL,
  `major` int(10) unsigned NOT NULL,
  `minor` int(10) unsigned NOT NULL,
  `patch` int(10) unsigned NOT NULL,
  `platform` varchar(32) NOT NULL,
  `platform_version` varchar(32) NOT NULL,
  `device` varchar(128) NOT NULL,
  `device_model` varchar(128) NOT NULL,
  `last_modified_date` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `build` varchar(32) NOT NULL,
  PRIMARY KEY (`account_id`,`platform`,`device`,`device_model`),
  CONSTRAINT `account_app_version_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_available_permission`
--

DROP TABLE IF EXISTS `account_available_permission`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_available_permission` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(60) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=31 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_credit`
--

DROP TABLE IF EXISTS `account_credit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_credit` (
  `account_id` int(10) unsigned NOT NULL,
  `credit` int(10) unsigned NOT NULL,
  `last_checked_account_credit_history_id` int(10) unsigned NOT NULL,
  `last_modified_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`account_id`),
  KEY `last_checked_patient_credit_history_id` (`last_checked_account_credit_history_id`),
  CONSTRAINT `account_credit_ibfk_2` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`),
  CONSTRAINT `account_credit_ibfk_3` FOREIGN KEY (`last_checked_account_credit_history_id`) REFERENCES `account_credit_history` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_credit_history`
--

DROP TABLE IF EXISTS `account_credit_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_credit_history` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account_id` int(10) unsigned NOT NULL,
  `credit` int(11) NOT NULL,
  `description` varchar(256) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`account_id`),
  CONSTRAINT `account_credit_history_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_device`
--

DROP TABLE IF EXISTS `account_device`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_device` (
  `account_id` int(10) unsigned NOT NULL,
  `device_id` varchar(128) NOT NULL,
  `verified` tinyint(1) NOT NULL,
  `verified_tstamp` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`account_id`,`device_id`),
  CONSTRAINT `account_device_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_email_optout`
--

DROP TABLE IF EXISTS `account_email_optout`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_email_optout` (
  `account_id` int(10) unsigned NOT NULL,
  `type` varchar(255) NOT NULL,
  `time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`account_id`,`type`),
  CONSTRAINT `account_email_optout_account` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_email_sent`
--

DROP TABLE IF EXISTS `account_email_sent`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_email_sent` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `account_id` int(10) unsigned NOT NULL,
  `type` varchar(255) NOT NULL,
  `time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `account_type` (`account_id`,`type`),
  CONSTRAINT `account_email_sent_account` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_group`
--

DROP TABLE IF EXISTS `account_group`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_group` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(60) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_group_member`
--

DROP TABLE IF EXISTS `account_group_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_group_member` (
  `group_id` int(10) unsigned NOT NULL,
  `account_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`account_id`,`group_id`),
  KEY `group_id` (`group_id`),
  CONSTRAINT `account_group_member_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `account_group` (`id`) ON DELETE CASCADE,
  CONSTRAINT `account_group_member_ibfk_2` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_group_permission`
--

DROP TABLE IF EXISTS `account_group_permission`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_group_permission` (
  `group_id` int(10) unsigned NOT NULL,
  `permission_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`group_id`,`permission_id`),
  KEY `permission_id` (`permission_id`),
  CONSTRAINT `account_group_permission_ibfk_1` FOREIGN KEY (`group_id`) REFERENCES `account_group` (`id`) ON DELETE CASCADE,
  CONSTRAINT `account_group_permission_ibfk_2` FOREIGN KEY (`permission_id`) REFERENCES `account_available_permission` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_phone`
--

DROP TABLE IF EXISTS `account_phone`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_phone` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account_id` int(10) unsigned NOT NULL,
  `phone` varchar(64) NOT NULL,
  `phone_type` varchar(32) NOT NULL,
  `status` varchar(32) NOT NULL,
  `verified` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `account_phone_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_promotion`
--

DROP TABLE IF EXISTS `account_promotion`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_promotion` (
  `account_id` int(10) unsigned NOT NULL,
  `promotion_code_id` int(10) unsigned NOT NULL,
  `promotion_group_id` int(10) unsigned NOT NULL,
  `promo_type` varchar(32) NOT NULL,
  `promo_data` blob NOT NULL,
  `expires` timestamp NULL DEFAULT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` varchar(32) NOT NULL,
  `id` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`),
  KEY `promotion_code_id` (`promotion_code_id`),
  KEY `promotion_group_id` (`promotion_group_id`),
  KEY `fk_account_promotion_account_id` (`account_id`),
  CONSTRAINT `account_promotion_ibfk_3` FOREIGN KEY (`promotion_group_id`) REFERENCES `promotion_group` (`id`),
  CONSTRAINT `fk_account_promotion_account_id` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`),
  CONSTRAINT `fk_account_promotion_promotion_code_id` FOREIGN KEY (`promotion_code_id`) REFERENCES `promotion_code` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_referral_tracking`
--

DROP TABLE IF EXISTS `account_referral_tracking`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_referral_tracking` (
  `promotion_code_id` int(10) unsigned NOT NULL,
  `claiming_account_id` int(10) unsigned NOT NULL,
  `referring_account_id` int(10) unsigned NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` varchar(32) NOT NULL,
  PRIMARY KEY (`claiming_account_id`),
  KEY `referring_account_id` (`referring_account_id`),
  KEY `promotion_code_id` (`promotion_code_id`),
  CONSTRAINT `account_referral_tracking_ibfk_1` FOREIGN KEY (`referring_account_id`) REFERENCES `account` (`id`),
  CONSTRAINT `account_referral_tracking_ibfk_3` FOREIGN KEY (`promotion_code_id`) REFERENCES `promotion_code` (`id`),
  CONSTRAINT `account_referral_tracking_ibfk_4` FOREIGN KEY (`claiming_account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `account_timezone`
--

DROP TABLE IF EXISTS `account_timezone`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `account_timezone` (
  `account_id` int(10) unsigned NOT NULL,
  `tz_name` varchar(256) NOT NULL,
  PRIMARY KEY (`account_id`),
  CONSTRAINT `account_timezone_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `additional_question_fields`
--

DROP TABLE IF EXISTS `additional_question_fields`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `additional_question_fields` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `question_id` int(10) unsigned NOT NULL,
  `json` blob NOT NULL,
  `language_id` int(10) unsigned DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `fk_additional_answer_fields_question_id` (`question_id`),
  CONSTRAINT `fk_additional_answer_fields_question_id` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=129 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `address`
--

DROP TABLE IF EXISTS `address`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `address` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `address_line_1` varchar(500) NOT NULL,
  `address_line_2` varchar(500) NOT NULL,
  `city` varchar(500) NOT NULL,
  `state` varchar(500) NOT NULL,
  `country` varchar(500) NOT NULL,
  `zip_code` varchar(500) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `advice`
--

DROP TABLE IF EXISTS `advice`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `advice` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `dr_advice_point_id` int(10) unsigned DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `treatment_plan_id` int(10) unsigned NOT NULL,
  `text` varchar(2048) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_advice_point_id` (`dr_advice_point_id`),
  KEY `treatment_plan_id` (`treatment_plan_id`),
  CONSTRAINT `advice_ibfk_2` FOREIGN KEY (`dr_advice_point_id`) REFERENCES `dr_advice_point` (`id`),
  CONSTRAINT `advice_ibfk_3` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `advice_point`
--

DROP TABLE IF EXISTS `advice_point`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `advice_point` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `text` varchar(2048) NOT NULL,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `analytics_report`
--

DROP TABLE IF EXISTS `analytics_report`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `analytics_report` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `owner_account_id` int(10) unsigned NOT NULL,
  `name` varchar(200) NOT NULL,
  `query` text NOT NULL,
  `presentation` text NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `owner_account_id` (`owner_account_id`),
  CONSTRAINT `analytics_report_ibfk_1` FOREIGN KEY (`owner_account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
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
  `deprecated` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `otype` (`atype`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8;
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
) ENGINE=InnoDB AUTO_INCREMENT=512 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `app_version_layout_mapping`
--

DROP TABLE IF EXISTS `app_version_layout_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `app_version_layout_mapping` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `app_major` int(10) unsigned NOT NULL,
  `app_minor` int(10) unsigned NOT NULL,
  `app_patch` int(10) unsigned NOT NULL,
  `layout_major` int(10) unsigned NOT NULL,
  `platform` varchar(64) NOT NULL,
  `role` varchar(64) NOT NULL,
  `purpose` varchar(64) NOT NULL,
  `sku_id` int(10) unsigned NOT NULL,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `major_platform_role_sku_pathway_key` (`layout_major`,`platform`,`role`,`purpose`,`sku_id`,`clinical_pathway_id`),
  KEY `sku_id` (`sku_id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  CONSTRAINT `app_version_layout_mapping_ibfk_2` FOREIGN KEY (`sku_id`) REFERENCES `sku` (`id`),
  CONSTRAINT `app_version_layout_mapping_ibfk_3` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
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
  `created` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `expires` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `platform` varchar(128) NOT NULL,
  `extended` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`token`),
  KEY `account_platform` (`account_id`,`platform`),
  CONSTRAINT `auth_token_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `bank_account`
--

DROP TABLE IF EXISTS `bank_account`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `bank_account` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account_id` int(10) unsigned NOT NULL,
  `stripe_recipient_id` varchar(128) NOT NULL,
  `default_account` tinyint(1) NOT NULL,
  `verified` tinyint(1) NOT NULL DEFAULT '0',
  `verify_amount_1` int(11) DEFAULT NULL,
  `verify_amount_2` int(11) DEFAULT NULL,
  `verify_transfer1_id` varchar(128) DEFAULT NULL,
  `verify_transfer2_id` varchar(128) DEFAULT NULL,
  `verify_expires` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `bank_account_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `care_provider_profile`
--

DROP TABLE IF EXISTS `care_provider_profile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `care_provider_profile` (
  `account_id` int(10) unsigned NOT NULL,
  `full_name` varchar(250) NOT NULL,
  `why_spruce` text NOT NULL,
  `qualifications` text NOT NULL,
  `medical_school` text NOT NULL,
  `residency` text NOT NULL,
  `fellowship` text NOT NULL,
  `experience` text NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `undergraduate_school` text NOT NULL,
  `graduate_school` text NOT NULL,
  PRIMARY KEY (`account_id`),
  CONSTRAINT `care_provider_profile_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `care_provider_state_elligibility`
--

DROP TABLE IF EXISTS `care_provider_state_elligibility`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `care_provider_state_elligibility` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `role_type_id` int(10) unsigned NOT NULL,
  `provider_id` int(10) unsigned NOT NULL,
  `care_providing_state_id` int(10) unsigned NOT NULL,
  `notify` tinyint(1) NOT NULL DEFAULT '0',
  `unavailable` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `provider_id` (`provider_id`,`role_type_id`,`care_providing_state_id`),
  KEY `care_providing_state_id` (`care_providing_state_id`),
  KEY `eligible_doctor_lookup` (`role_type_id`,`care_providing_state_id`,`unavailable`),
  CONSTRAINT `care_provider_state_elligibility_ibfk_1` FOREIGN KEY (`role_type_id`) REFERENCES `role_type` (`id`),
  CONSTRAINT `care_provider_state_elligibility_ibfk_2` FOREIGN KEY (`care_providing_state_id`) REFERENCES `care_providing_state` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `care_providing_state`
--

DROP TABLE IF EXISTS `care_providing_state`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `care_providing_state` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `state` varchar(100) NOT NULL,
  `long_state` varchar(250) NOT NULL,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  KEY `state` (`state`),
  CONSTRAINT `care_providing_state_ibfk_2` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `care_providing_state_notification`
--

DROP TABLE IF EXISTS `care_providing_state_notification`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `care_providing_state_notification` (
  `care_providing_state_id` int(10) unsigned NOT NULL,
  `last_notified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`care_providing_state_id`),
  KEY `last_notified` (`last_notified`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `case_notification`
--

DROP TABLE IF EXISTS `case_notification`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `case_notification` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_case_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `notification_type` varchar(100) NOT NULL,
  `uid` varchar(100) NOT NULL,
  `data` blob,
  PRIMARY KEY (`id`),
  UNIQUE KEY `patient_case_id_2` (`patient_case_id`,`uid`),
  KEY `patient_case_id` (`patient_case_id`),
  CONSTRAINT `case_notification_ibfk_1` FOREIGN KEY (`patient_case_id`) REFERENCES `patient_case` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `clinical_pathway`
--

DROP TABLE IF EXISTS `clinical_pathway`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `clinical_pathway` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `tag` varchar(64) NOT NULL,
  `name` varchar(250) NOT NULL,
  `medicine_branch` varchar(250) NOT NULL,
  `status` varchar(32) NOT NULL,
  `details_json` blob,
  `stp_json` blob,
  PRIMARY KEY (`id`),
  UNIQUE KEY `tag` (`tag`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `clinical_pathway_menu`
--

DROP TABLE IF EXISTS `clinical_pathway_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `clinical_pathway_menu` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `json` blob NOT NULL,
  `status` varchar(32) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `status` (`status`,`created`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `common_diagnosis_set`
--

DROP TABLE IF EXISTS `common_diagnosis_set`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `common_diagnosis_set` (
  `pathway_id` int(10) unsigned NOT NULL,
  `title` varchar(600) NOT NULL,
  PRIMARY KEY (`pathway_id`),
  KEY `pathway_id` (`pathway_id`),
  CONSTRAINT `common_diagnosis_set_ibfk_1` FOREIGN KEY (`pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `common_diagnosis_set_item`
--

DROP TABLE IF EXISTS `common_diagnosis_set_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `common_diagnosis_set_item` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `diagnosis_code_id` varchar(32) NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `diagnosis_code_id` (`diagnosis_code_id`,`pathway_id`),
  KEY `pathway_id` (`pathway_id`,`active`),
  CONSTRAINT `common_diagnosis_set_item_ibfk_1` FOREIGN KEY (`pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `communication_preference`
--

DROP TABLE IF EXISTS `communication_preference`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `communication_preference` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account_id` int(10) unsigned NOT NULL,
  `communication_type` varchar(50) NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `account_id` (`account_id`,`communication_type`),
  CONSTRAINT `communication_preference_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `communication_snooze`
--

DROP TABLE IF EXISTS `communication_snooze`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `communication_snooze` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account_id` int(10) unsigned NOT NULL,
  `start_hour` int(10) unsigned NOT NULL,
  `num_hours` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `communication_snooze_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `credit_card`
--

DROP TABLE IF EXISTS `credit_card`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `credit_card` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `third_party_card_id` varchar(100) DEFAULT NULL,
  `type` varchar(100) NOT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `address_id` int(10) unsigned DEFAULT NULL,
  `is_default` tinyint(1) NOT NULL,
  `label` varchar(200) DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  `fingerprint` varchar(200) DEFAULT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `payment_service_customer_id` varchar(500) DEFAULT NULL,
  `apple_pay` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `address_id` (`address_id`),
  KEY `patient_id` (`patient_id`),
  CONSTRAINT `credit_card_ibfk_2` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `credit_card_ibfk_3` FOREIGN KEY (`address_id`) REFERENCES `address` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `deny_refill_reason`
--

DROP TABLE IF EXISTS `deny_refill_reason`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `deny_refill_reason` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `reason_code` varchar(100) NOT NULL,
  `reason` varchar(150) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `deprecated_object_storage`
--

DROP TABLE IF EXISTS `deprecated_object_storage`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `deprecated_object_storage` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `bucket` varchar(100) NOT NULL,
  `storage_key` varchar(100) NOT NULL,
  `status` varchar(100) NOT NULL,
  `region_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `modified_date` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `region_id` (`region_id`,`storage_key`,`bucket`,`status`),
  CONSTRAINT `deprecated_object_storage_ibfk_1` FOREIGN KEY (`region_id`) REFERENCES `region` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=826 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `diagnosis_details_intake`
--

DROP TABLE IF EXISTS `diagnosis_details_intake`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `diagnosis_details_intake` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `doctor_id` int(10) unsigned NOT NULL,
  `visit_diagnosis_item_id` int(10) unsigned NOT NULL,
  `layout_version_id` int(10) unsigned NOT NULL,
  `answered_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `question_id` int(10) unsigned NOT NULL,
  `potential_answer_id` int(10) unsigned DEFAULT NULL,
  `answer_text` text,
  `parent_question_id` int(10) unsigned DEFAULT NULL,
  `parent_info_intake_id` int(10) unsigned DEFAULT NULL,
  `client_clock` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `visit_diagnosis_item_id` (`visit_diagnosis_item_id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `question_id` (`question_id`),
  KEY `potential_answer_id` (`potential_answer_id`),
  KEY `parent_question_id` (`parent_question_id`),
  KEY `parent_info_intake_id` (`parent_info_intake_id`),
  CONSTRAINT `diagnosis_details_intake_ibfk_1` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `diagnosis_details_intake_ibfk_2` FOREIGN KEY (`visit_diagnosis_item_id`) REFERENCES `visit_diagnosis_item` (`id`),
  CONSTRAINT `diagnosis_details_intake_ibfk_4` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `diagnosis_details_intake_ibfk_5` FOREIGN KEY (`potential_answer_id`) REFERENCES `potential_answer` (`id`),
  CONSTRAINT `diagnosis_details_intake_ibfk_6` FOREIGN KEY (`parent_question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `diagnosis_details_intake_ibfk_7` FOREIGN KEY (`parent_info_intake_id`) REFERENCES `diagnosis_details_intake` (`id`),
  CONSTRAINT `diagnosis_details_intake_ibfk_8` FOREIGN KEY (`layout_version_id`) REFERENCES `diagnosis_details_layout` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `diagnosis_details_layout`
--

DROP TABLE IF EXISTS `diagnosis_details_layout`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `diagnosis_details_layout` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(64) NOT NULL,
  `layout` blob NOT NULL,
  `diagnosis_code_id` varchar(32) NOT NULL,
  `major` int(10) unsigned NOT NULL,
  `minor` int(10) unsigned NOT NULL,
  `patch` int(10) unsigned NOT NULL,
  `active` tinyint(1) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `template_layout_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `diagnosis_code_id` (`diagnosis_code_id`,`major`,`minor`,`patch`),
  KEY `template_layout_id` (`template_layout_id`),
  CONSTRAINT `diagnosis_details_layout_ibfk_2` FOREIGN KEY (`template_layout_id`) REFERENCES `diagnosis_details_layout_template` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `diagnosis_details_layout_template`
--

DROP TABLE IF EXISTS `diagnosis_details_layout_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `diagnosis_details_layout_template` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(64) NOT NULL,
  `layout` blob NOT NULL,
  `diagnosis_code_id` varchar(32) NOT NULL,
  `major` int(10) unsigned NOT NULL,
  `minor` int(10) unsigned NOT NULL,
  `patch` int(10) unsigned NOT NULL,
  `active` tinyint(1) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `diagnosis_code_id` (`diagnosis_code_id`,`major`,`minor`,`patch`),
  KEY `diagnosis_code_id_2` (`diagnosis_code_id`,`active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `diagnosis_intake`
--

DROP TABLE IF EXISTS `diagnosis_intake`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `diagnosis_intake` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_visit_id` int(10) unsigned NOT NULL,
  `question_id` int(10) unsigned NOT NULL,
  `potential_answer_id` int(10) unsigned DEFAULT NULL,
  `answer_text` mediumtext,
  `layout_version_id` int(10) unsigned NOT NULL,
  `answered_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `doctor_id` int(10) unsigned NOT NULL,
  `parent_info_intake_id` int(10) unsigned DEFAULT NULL,
  `parent_question_id` int(10) unsigned DEFAULT NULL,
  `client_clock` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_visit_id` (`patient_visit_id`),
  KEY `question_id` (`question_id`),
  KEY `potential_answer_id` (`potential_answer_id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `parent_info_intake_id` (`parent_info_intake_id`),
  KEY `parent_question_id` (`parent_question_id`),
  CONSTRAINT `diagnosis_intake_ibfk_1` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`),
  CONSTRAINT `diagnosis_intake_ibfk_2` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `diagnosis_intake_ibfk_3` FOREIGN KEY (`potential_answer_id`) REFERENCES `potential_answer` (`id`),
  CONSTRAINT `diagnosis_intake_ibfk_4` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `diagnosis_intake_ibfk_5` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `diagnosis_intake_ibfk_7` FOREIGN KEY (`parent_question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `diagnosis_intake_ibfk_8` FOREIGN KEY (`parent_info_intake_id`) REFERENCES `info_intake` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `diagnosis_layout_version`
--

DROP TABLE IF EXISTS `diagnosis_layout_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `diagnosis_layout_version` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `layout_version_id` int(10) unsigned NOT NULL,
  `layout_blob_storage_id` int(10) unsigned NOT NULL,
  `major` int(10) unsigned NOT NULL,
  `minor` int(10) unsigned NOT NULL,
  `patch` int(10) unsigned NOT NULL,
  `language_id` int(10) unsigned NOT NULL,
  `status` varchar(64) NOT NULL,
  `modified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `layout_blob_storage_id` (`layout_blob_storage_id`),
  KEY `language_id` (`language_id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  CONSTRAINT `diagnosis_layout_version_ibfk_1` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `diagnosis_layout_version_ibfk_2` FOREIGN KEY (`layout_blob_storage_id`) REFERENCES `layout_blob_storage` (`id`),
  CONSTRAINT `diagnosis_layout_version_ibfk_4` FOREIGN KEY (`language_id`) REFERENCES `languages_supported` (`id`),
  CONSTRAINT `diagnosis_layout_version_ibfk_5` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
-- Table structure for table `dntf_mapping`
--

DROP TABLE IF EXISTS `dntf_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dntf_mapping` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `treatment_id` int(10) unsigned DEFAULT NULL,
  `unlinked_dntf_treatment_id` int(10) unsigned DEFAULT NULL,
  `rx_refill_request_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `treatment_id` (`treatment_id`),
  KEY `rx_refill_request_id` (`rx_refill_request_id`),
  KEY `unlinked_dntf_treatment_id` (`unlinked_dntf_treatment_id`),
  CONSTRAINT `dntf_mapping_ibfk_1` FOREIGN KEY (`treatment_id`) REFERENCES `treatment` (`id`),
  CONSTRAINT `dntf_mapping_ibfk_2` FOREIGN KEY (`rx_refill_request_id`) REFERENCES `rx_refill_request` (`id`),
  CONSTRAINT `dntf_mapping_ibfk_3` FOREIGN KEY (`unlinked_dntf_treatment_id`) REFERENCES `unlinked_dntf_treatment` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `doctor`
--

DROP TABLE IF EXISTS `doctor`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `doctor` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `first_name` varchar(250) NOT NULL,
  `last_name` varchar(250) NOT NULL,
  `gender` varchar(250) NOT NULL,
  `account_id` int(10) unsigned NOT NULL,
  `dea_number` varchar(250) DEFAULT NULL,
  `npi_number` varchar(250) DEFAULT NULL,
  `status` varchar(250) NOT NULL,
  `clinician_id` int(10) unsigned DEFAULT NULL,
  `dob_month` int(10) unsigned NOT NULL,
  `dob_year` int(10) unsigned NOT NULL,
  `dob_day` int(10) unsigned NOT NULL,
  `middle_name` varchar(100) DEFAULT NULL,
  `prefix` varchar(100) DEFAULT NULL,
  `suffix` varchar(100) DEFAULT NULL,
  `short_title` varchar(300) DEFAULT NULL,
  `long_title` varchar(300) DEFAULT NULL,
  `short_display_name` varchar(300) NOT NULL,
  `long_display_name` varchar(600) NOT NULL,
  `small_thumbnail_id` varchar(250) DEFAULT NULL,
  `large_thumbnail_id` varchar(250) DEFAULT NULL,
  `hero_image_id` varchar(250) DEFAULT NULL,
  `primary_cc` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `doctor_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `doctor_address_selection`
--

DROP TABLE IF EXISTS `doctor_address_selection`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `doctor_address_selection` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `address_id` int(10) unsigned NOT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `address_id` (`address_id`),
  CONSTRAINT `doctor_address_selection_ibfk_1` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `doctor_address_selection_ibfk_2` FOREIGN KEY (`address_id`) REFERENCES `address` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `doctor_attribute`
--

DROP TABLE IF EXISTS `doctor_attribute`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `doctor_attribute` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `doctor_id` int(10) unsigned NOT NULL,
  `name` varchar(64) NOT NULL,
  `value` varchar(1024) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `doctor_id` (`doctor_id`,`name`),
  CONSTRAINT `doctor_attribute_ibfk_1` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `doctor_case_notification`
--

DROP TABLE IF EXISTS `doctor_case_notification`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `doctor_case_notification` (
  `doctor_id` int(10) unsigned NOT NULL,
  `last_notified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`doctor_id`),
  KEY `last_notified` (`last_notified`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `doctor_medical_license`
--

DROP TABLE IF EXISTS `doctor_medical_license`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `doctor_medical_license` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `doctor_id` int(10) unsigned NOT NULL,
  `state` char(2) NOT NULL,
  `license_number` varchar(64) NOT NULL,
  `status` varchar(32) NOT NULL,
  `expiration_date` date DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `doctor_id` (`doctor_id`,`state`),
  CONSTRAINT `doctor_medical_license_ibfk_1` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `doctor_queue`
--

DROP TABLE IF EXISTS `doctor_queue`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `doctor_queue` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `doctor_id` int(10) unsigned NOT NULL,
  `status` varchar(100) NOT NULL,
  `event_type` varchar(100) NOT NULL,
  `enqueue_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `completed_date` timestamp NULL DEFAULT NULL,
  `item_id` int(10) unsigned NOT NULL,
  `description` varchar(2000) NOT NULL,
  `action_url` varchar(2000) NOT NULL,
  `short_description` text NOT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `tags` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `status_enqueue_date` (`status`,`enqueue_date`),
  KEY `doctor_status_enqueue_date` (`doctor_id`,`status`,`enqueue_date`),
  CONSTRAINT `doctor_queue_ibfk_1` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `doctor_transaction`
--

DROP TABLE IF EXISTS `doctor_transaction`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `doctor_transaction` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `doctor_id` int(10) unsigned NOT NULL,
  `item_cost_id` int(10) unsigned DEFAULT NULL,
  `item_id` int(10) unsigned NOT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `sku_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `doctor_id_2` (`doctor_id`,`item_id`),
  KEY `patient_id` (`patient_id`),
  KEY `item_cost_id` (`item_cost_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `sku_id` (`sku_id`),
  KEY `created` (`created`),
  CONSTRAINT `doctor_transaction_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `doctor_transaction_ibfk_2` FOREIGN KEY (`item_cost_id`) REFERENCES `item_cost` (`id`),
  CONSTRAINT `doctor_transaction_ibfk_3` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `doctor_transaction_ibfk_4` FOREIGN KEY (`sku_id`) REFERENCES `sku` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_advice_point`
--

DROP TABLE IF EXISTS `dr_advice_point`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_advice_point` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `text` varchar(2048) NOT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `source_id` int(10) unsigned DEFAULT NULL,
  `modified_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `source_id` (`source_id`),
  KEY `doctor_id_2` (`doctor_id`),
  CONSTRAINT `dr_advice_point_ibfk_1` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `dr_advice_point_ibfk_2` FOREIGN KEY (`source_id`) REFERENCES `dr_advice_point` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_drug_supplemental_instruction`
--

DROP TABLE IF EXISTS `dr_drug_supplemental_instruction`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_drug_supplemental_instruction` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `doctor_id` int(10) unsigned NOT NULL,
  `text` varchar(150) NOT NULL,
  `drug_name_id` int(10) unsigned NOT NULL,
  `drug_form_id` int(10) unsigned DEFAULT NULL,
  `drug_route_id` int(10) unsigned DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `drug_supplemental_instruction_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `drug_route_id` (`drug_route_id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `drug_supplemental_instruction_id` (`drug_supplemental_instruction_id`),
  CONSTRAINT `dr_drug_supplemental_instruction_ibfk_1` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `dr_drug_supplemental_instruction_ibfk_2` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`),
  CONSTRAINT `dr_drug_supplemental_instruction_ibfk_3` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `dr_drug_supplemental_instruction_ibfk_4` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `dr_drug_supplemental_instruction_ibfk_5` FOREIGN KEY (`drug_supplemental_instruction_id`) REFERENCES `drug_supplemental_instruction` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_drug_supplemental_instruction_selected_state`
--

DROP TABLE IF EXISTS `dr_drug_supplemental_instruction_selected_state`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_drug_supplemental_instruction_selected_state` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_name_id` int(10) unsigned NOT NULL,
  `drug_form_id` int(10) unsigned NOT NULL,
  `drug_route_id` int(10) unsigned NOT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  `dr_drug_supplemental_instruction_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `drug_route_id` (`drug_route_id`),
  KEY `dr_drug_supplemental_instruction_id` (`dr_drug_supplemental_instruction_id`),
  KEY `drug_name_id` (`drug_name_id`,`drug_form_id`,`drug_route_id`,`doctor_id`,`dr_drug_supplemental_instruction_id`),
  CONSTRAINT `dr_drug_supplemental_instruction_selected_state_ibfk_1` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `dr_drug_supplemental_instruction_selected_state_ibfk_2` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `dr_drug_supplemental_instruction_selected_state_ibfk_3` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `dr_drug_supplemental_instruction_selected_state_ibfk_4` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`),
  CONSTRAINT `dr_drug_supplemental_instruction_selected_state_ibfk_5` FOREIGN KEY (`dr_drug_supplemental_instruction_id`) REFERENCES `dr_drug_supplemental_instruction` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_advice`
--

DROP TABLE IF EXISTS `dr_favorite_advice`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_advice` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `text` varchar(2048) NOT NULL,
  `dr_advice_point_id` int(10) unsigned DEFAULT NULL,
  `dr_favorite_treatment_plan_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_advice_point_id` (`dr_advice_point_id`),
  KEY `dr_favorite_treatment_plan_id` (`dr_favorite_treatment_plan_id`),
  CONSTRAINT `dr_favorite_advice_ibfk_1` FOREIGN KEY (`dr_advice_point_id`) REFERENCES `dr_advice_point` (`id`),
  CONSTRAINT `dr_favorite_advice_ibfk_2` FOREIGN KEY (`dr_favorite_treatment_plan_id`) REFERENCES `dr_favorite_treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_patient_visit_follow_up`
--

DROP TABLE IF EXISTS `dr_favorite_patient_visit_follow_up`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_patient_visit_follow_up` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `follow_up_date` date NOT NULL,
  `follow_up_value` int(10) unsigned NOT NULL,
  `follow_up_unit` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  `dr_favorite_treatment_plan_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_favorite_treatment_plan_id` (`dr_favorite_treatment_plan_id`),
  CONSTRAINT `dr_favorite_patient_visit_follow_up_ibfk_1` FOREIGN KEY (`dr_favorite_treatment_plan_id`) REFERENCES `dr_favorite_treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_regimen`
--

DROP TABLE IF EXISTS `dr_favorite_regimen`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_regimen` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `text` varchar(2048) NOT NULL,
  `dr_regimen_step_id` int(10) unsigned DEFAULT NULL,
  `dr_favorite_treatment_plan_id` int(10) unsigned NOT NULL,
  `dr_favorite_regimen_section_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_favorite_treatment_plan_id` (`dr_favorite_treatment_plan_id`),
  KEY `dr_regimen_step_id` (`dr_regimen_step_id`),
  CONSTRAINT `dr_favorite_regimen_ibfk_2` FOREIGN KEY (`dr_regimen_step_id`) REFERENCES `dr_regimen_step` (`id`),
  CONSTRAINT `dr_favorite_regimen_ibfk_3` FOREIGN KEY (`dr_favorite_treatment_plan_id`) REFERENCES `dr_favorite_treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_regimen_section`
--

DROP TABLE IF EXISTS `dr_favorite_regimen_section`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_regimen_section` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(500) NOT NULL,
  `creation_date` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `dr_favorite_treatment_plan_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_favorite_treatment_plan_id` (`dr_favorite_treatment_plan_id`),
  CONSTRAINT `dr_favorite_regimen_section_ibfk_1` FOREIGN KEY (`dr_favorite_treatment_plan_id`) REFERENCES `dr_favorite_treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_treatment`
--

DROP TABLE IF EXISTS `dr_favorite_treatment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_treatment` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `dr_favorite_treatment_plan_id` int(10) unsigned NOT NULL,
  `drug_internal_name` varchar(250) NOT NULL,
  `dispense_value` decimal(21,10) NOT NULL,
  `dispense_unit_id` int(10) unsigned NOT NULL,
  `refills` int(10) unsigned NOT NULL,
  `substitutions_allowed` tinyint(4) DEFAULT NULL,
  `days_supply` int(10) unsigned DEFAULT NULL,
  `pharmacy_notes` varchar(250) DEFAULT NULL,
  `patient_instructions` varchar(150) NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  `dosage_strength` varchar(250) NOT NULL,
  `type` varchar(150) NOT NULL,
  `drug_name_id` int(10) unsigned DEFAULT NULL,
  `drug_form_id` int(10) unsigned DEFAULT NULL,
  `drug_route_id` int(10) unsigned DEFAULT NULL,
  `is_controlled_substance` tinyint(4) DEFAULT NULL,
  `generic_drug_name_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_favorite_treatment_plan_id` (`dr_favorite_treatment_plan_id`),
  KEY `dispense_unit_id` (`dispense_unit_id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `drug_route_id` (`drug_route_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `generic_drug_name_id` (`generic_drug_name_id`),
  CONSTRAINT `dr_favorite_treatment_ibfk_2` FOREIGN KEY (`dispense_unit_id`) REFERENCES `dispense_unit` (`id`),
  CONSTRAINT `dr_favorite_treatment_ibfk_3` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `dr_favorite_treatment_ibfk_4` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`),
  CONSTRAINT `dr_favorite_treatment_ibfk_5` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `dr_favorite_treatment_ibfk_6` FOREIGN KEY (`dr_favorite_treatment_plan_id`) REFERENCES `dr_favorite_treatment_plan` (`id`) ON DELETE CASCADE,
  CONSTRAINT `dr_favorite_treatment_ibfk_7` FOREIGN KEY (`generic_drug_name_id`) REFERENCES `drug_name` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_treatment_drug_db_id`
--

DROP TABLE IF EXISTS `dr_favorite_treatment_drug_db_id`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_treatment_drug_db_id` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_db_id` varchar(100) NOT NULL,
  `drug_db_id_tag` varchar(100) NOT NULL,
  `dr_favorite_treatment_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_favorite_treatment_id` (`dr_favorite_treatment_id`),
  CONSTRAINT `dr_favorite_treatment_drug_db_id_ibfk_1` FOREIGN KEY (`dr_favorite_treatment_id`) REFERENCES `dr_favorite_treatment` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_treatment_plan`
--

DROP TABLE IF EXISTS `dr_favorite_treatment_plan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_treatment_plan` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(600) NOT NULL,
  `modified_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `note` text CHARACTER SET utf8mb4,
  `parent_id` int(10) unsigned DEFAULT NULL,
  `creator_id` int(10) unsigned DEFAULT NULL,
  `lifecycle` varchar(20) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_parent_treatment_plan_id` (`parent_id`),
  KEY `fk_creator_id_doctor` (`creator_id`),
  CONSTRAINT `fk_creator_id_doctor` FOREIGN KEY (`creator_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `fk_parent_treatment_plan_id` FOREIGN KEY (`parent_id`) REFERENCES `dr_favorite_treatment_plan` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_treatment_plan_membership`
--

DROP TABLE IF EXISTS `dr_favorite_treatment_plan_membership`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_treatment_plan_membership` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `dr_favorite_treatment_plan_id` int(10) unsigned NOT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `plan_doctor_clinical_pathway` (`dr_favorite_treatment_plan_id`,`doctor_id`,`clinical_pathway_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `dr_favorite_treatment_plan_membership_clinical_pathway_id` (`clinical_pathway_id`),
  CONSTRAINT `dr_favorite_treatment_plan_membership_clinical_pathway_id` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`),
  CONSTRAINT `dr_favorite_treatment_plan_membership_doctor_id` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `dr_favorite_treatment_plan_membership_plan_id` FOREIGN KEY (`dr_favorite_treatment_plan_id`) REFERENCES `dr_favorite_treatment_plan` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_treatment_plan_resource_guide`
--

DROP TABLE IF EXISTS `dr_favorite_treatment_plan_resource_guide`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_treatment_plan_resource_guide` (
  `dr_favorite_treatment_plan_id` int(10) unsigned NOT NULL,
  `resource_guide_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`dr_favorite_treatment_plan_id`,`resource_guide_id`),
  KEY `resource_guide_id` (`resource_guide_id`),
  CONSTRAINT `dr_favorite_treatment_plan_resource_guide_ibfk_1` FOREIGN KEY (`dr_favorite_treatment_plan_id`) REFERENCES `dr_favorite_treatment_plan` (`id`) ON DELETE CASCADE,
  CONSTRAINT `dr_favorite_treatment_plan_resource_guide_ibfk_2` FOREIGN KEY (`resource_guide_id`) REFERENCES `resource_guide` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_treatment_plan_scheduled_message`
--

DROP TABLE IF EXISTS `dr_favorite_treatment_plan_scheduled_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_treatment_plan_scheduled_message` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `dr_favorite_treatment_plan_id` int(10) unsigned NOT NULL,
  `scheduled_days` int(10) unsigned NOT NULL,
  `message` text NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_favorite_treatment_plan_id` (`dr_favorite_treatment_plan_id`),
  CONSTRAINT `dr_favorite_treatment_plan_scheduled_message_ibfk_1` FOREIGN KEY (`dr_favorite_treatment_plan_id`) REFERENCES `dr_favorite_treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_favorite_treatment_plan_scheduled_message_attachment`
--

DROP TABLE IF EXISTS `dr_favorite_treatment_plan_scheduled_message_attachment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_favorite_treatment_plan_scheduled_message_attachment` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `dr_favorite_treatment_plan_scheduled_message_id` bigint(20) unsigned NOT NULL,
  `item_type` varchar(64) NOT NULL,
  `item_id` bigint(20) unsigned NOT NULL,
  `title` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_favorite_treatment_plan_scheduled_message_id` (`dr_favorite_treatment_plan_scheduled_message_id`),
  CONSTRAINT `dr_favorite_treatment_plan_scheduled_message_attachment_ibfk_1` FOREIGN KEY (`dr_favorite_treatment_plan_scheduled_message_id`) REFERENCES `dr_favorite_treatment_plan_scheduled_message` (`id`) ON DELETE CASCADE
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
  `layout_version_id` int(10) unsigned NOT NULL,
  `status` varchar(250) NOT NULL,
  `modified_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `layout_blob_storage_id` int(10) unsigned DEFAULT NULL,
  `major` int(10) unsigned NOT NULL,
  `minor` int(10) unsigned NOT NULL,
  `patch` int(10) unsigned NOT NULL,
  `language_id` int(10) unsigned NOT NULL,
  `sku_id` int(10) unsigned NOT NULL,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `layout_blob_storage_id` (`layout_blob_storage_id`),
  KEY `language_id` (`language_id`),
  KEY `sku_id` (`sku_id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  CONSTRAINT `dr_layout_version_ibfk_1` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `dr_layout_version_ibfk_5` FOREIGN KEY (`layout_blob_storage_id`) REFERENCES `layout_blob_storage` (`id`),
  CONSTRAINT `dr_layout_version_ibfk_6` FOREIGN KEY (`language_id`) REFERENCES `languages_supported` (`id`),
  CONSTRAINT `dr_layout_version_ibfk_7` FOREIGN KEY (`sku_id`) REFERENCES `sku` (`id`),
  CONSTRAINT `dr_layout_version_ibfk_8` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=55 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_regimen_step`
--

DROP TABLE IF EXISTS `dr_regimen_step`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_regimen_step` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `text` varchar(2048) NOT NULL,
  `drug_name_id` int(10) unsigned DEFAULT NULL,
  `drug_form_id` int(10) unsigned DEFAULT NULL,
  `drug_route_id` int(10) unsigned DEFAULT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `source_id` int(10) unsigned DEFAULT NULL,
  `modified_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `drug_route_id` (`drug_route_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `source_id` (`source_id`),
  KEY `doctor_id_2` (`doctor_id`),
  CONSTRAINT `dr_regimen_step_ibfk_1` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `dr_regimen_step_ibfk_2` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `dr_regimen_step_ibfk_3` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`),
  CONSTRAINT `dr_regimen_step_ibfk_4` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `dr_regimen_step_ibfk_5` FOREIGN KEY (`source_id`) REFERENCES `dr_regimen_step` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_treatment_template`
--

DROP TABLE IF EXISTS `dr_treatment_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_treatment_template` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(600) NOT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  `status` varchar(100) NOT NULL,
  `drug_internal_name` varchar(250) NOT NULL,
  `dispense_value` decimal(21,10) NOT NULL,
  `dispense_unit_id` int(10) unsigned NOT NULL,
  `refills` int(10) unsigned NOT NULL,
  `substitutions_allowed` tinyint(4) NOT NULL,
  `days_supply` int(10) unsigned DEFAULT NULL,
  `pharmacy_notes` varchar(250) DEFAULT NULL,
  `patient_instructions` varchar(150) NOT NULL,
  `dosage_strength` varchar(250) NOT NULL,
  `type` varchar(150) NOT NULL,
  `drug_name_id` int(10) unsigned NOT NULL,
  `drug_form_id` int(10) unsigned DEFAULT NULL,
  `drug_route_id` int(10) unsigned DEFAULT NULL,
  `erx_sent_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `erx_id` int(10) unsigned DEFAULT NULL,
  `pharmacy_id` int(10) unsigned DEFAULT NULL,
  `erx_last_filled_date` timestamp(6) NULL DEFAULT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `is_controlled_substance` tinyint(4) DEFAULT NULL,
  `generic_drug_name_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `dispense_unit_id` (`dispense_unit_id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `drug_route_id` (`drug_route_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `pharmacy_id` (`pharmacy_id`),
  KEY `generic_drug_name_id` (`generic_drug_name_id`),
  CONSTRAINT `dr_treatment_template_ibfk_1` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `dr_treatment_template_ibfk_2` FOREIGN KEY (`dispense_unit_id`) REFERENCES `dispense_unit` (`id`),
  CONSTRAINT `dr_treatment_template_ibfk_3` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `dr_treatment_template_ibfk_4` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`),
  CONSTRAINT `dr_treatment_template_ibfk_5` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `dr_treatment_template_ibfk_6` FOREIGN KEY (`pharmacy_id`) REFERENCES `pharmacy_selection` (`id`),
  CONSTRAINT `dr_treatment_template_ibfk_7` FOREIGN KEY (`generic_drug_name_id`) REFERENCES `drug_name` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dr_treatment_template_drug_db_id`
--

DROP TABLE IF EXISTS `dr_treatment_template_drug_db_id`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dr_treatment_template_drug_db_id` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_db_id_tag` varchar(100) NOT NULL,
  `drug_db_id` varchar(100) NOT NULL,
  `dr_treatment_template_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_treatment_template_id` (`dr_treatment_template_id`),
  CONSTRAINT `dr_treatment_template_drug_db_id_ibfk_1` FOREIGN KEY (`dr_treatment_template_id`) REFERENCES `dr_treatment_template` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_description`
--

DROP TABLE IF EXISTS `drug_description`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_description` (
  `drug_name_strength` varchar(250) NOT NULL,
  `json` blob NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`drug_name_strength`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_details`
--

DROP TABLE IF EXISTS `drug_details`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_details` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `ndc` varchar(12) DEFAULT NULL,
  `json` blob NOT NULL,
  `modified_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `generic_drug_name` varchar(250) DEFAULT NULL,
  `drug_route` varchar(250) DEFAULT NULL,
  `drug_form` varchar(250) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `drug_details_name_route` (`generic_drug_name`,`drug_route`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_form`
--

DROP TABLE IF EXISTS `drug_form`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_form` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(150) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `drug_form` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_name`
--

DROP TABLE IF EXISTS `drug_name`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_name` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(150) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `drug_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=85 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `drug_route`
--

DROP TABLE IF EXISTS `drug_route`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drug_route` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(150) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `drug_route` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

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
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
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
-- Table structure for table `email_campaign_state`
--

DROP TABLE IF EXISTS `email_campaign_state`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `email_campaign_state` (
  `key` varchar(64) NOT NULL,
  `data` blob NOT NULL,
  PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `email_sender`
--

DROP TABLE IF EXISTS `email_sender`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `email_sender` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `email` varchar(64) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `email_template`
--

DROP TABLE IF EXISTS `email_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `email_template` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(128) NOT NULL,
  `name` varchar(200) NOT NULL,
  `sender_id` int(10) unsigned NOT NULL,
  `subject_template` varchar(1024) NOT NULL,
  `body_text_template` text NOT NULL,
  `body_html_template` text NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '0',
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `sender_id` (`sender_id`),
  KEY `type` (`type`),
  CONSTRAINT `email_template_ibfk_1` FOREIGN KEY (`sender_id`) REFERENCES `email_sender` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `erx_status_events`
--

DROP TABLE IF EXISTS `erx_status_events`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `erx_status_events` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `treatment_id` int(10) unsigned NOT NULL,
  `erx_status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  `event_details` varchar(500) DEFAULT NULL,
  `reported_timestamp` timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `treatment_id` (`treatment_id`),
  CONSTRAINT `erx_status_events_ibfk_1` FOREIGN KEY (`treatment_id`) REFERENCES `treatment` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `form_doctor_interest`
--

DROP TABLE IF EXISTS `form_doctor_interest`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `form_doctor_interest` (
  `name` varchar(250) NOT NULL,
  `email` varchar(250) NOT NULL,
  `states` varchar(250) NOT NULL,
  `comment` varchar(4000) NOT NULL,
  `request_id` bigint(20) unsigned NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `source` varchar(64) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `form_notify_me`
--

DROP TABLE IF EXISTS `form_notify_me`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `form_notify_me` (
  `email` varchar(250) NOT NULL,
  `state` char(2) NOT NULL,
  `platform` varchar(128) NOT NULL,
  `request_id` bigint(20) unsigned NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `source` varchar(64) NOT NULL,
  `unique_key` varchar(128) DEFAULT NULL,
  UNIQUE KEY `unique_key` (`unique_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `health_log`
--

DROP TABLE IF EXISTS `health_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `health_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `uid` varchar(128) NOT NULL,
  `tstamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `type` varchar(64) NOT NULL,
  `data` blob NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `patient_id` (`patient_id`,`uid`),
  CONSTRAINT `health_log_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `info_intake`
--

DROP TABLE IF EXISTS `info_intake`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `info_intake` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_visit_id` int(10) unsigned NOT NULL,
  `question_id` int(10) unsigned NOT NULL,
  `potential_answer_id` int(10) unsigned DEFAULT NULL,
  `answer_text` mediumtext,
  `layout_version_id` int(10) unsigned NOT NULL,
  `answered_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `patient_id` int(10) unsigned NOT NULL,
  `object_storage_id` int(10) unsigned DEFAULT NULL,
  `parent_info_intake_id` int(10) unsigned DEFAULT NULL,
  `summary_localized_text_id` int(10) unsigned DEFAULT NULL,
  `parent_question_id` int(10) unsigned DEFAULT NULL,
  `client_clock` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_visit_id` (`patient_visit_id`),
  KEY `question_id` (`question_id`),
  KEY `potential_answer_id` (`potential_answer_id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `patient_id` (`patient_id`),
  KEY `parent_info_intake_id` (`parent_info_intake_id`),
  KEY `summary_localized_text_id` (`summary_localized_text_id`),
  KEY `parent_question_id` (`parent_question_id`),
  CONSTRAINT `info_intake_ibfk_10` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`),
  CONSTRAINT `info_intake_ibfk_11` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `info_intake_ibfk_12` FOREIGN KEY (`parent_info_intake_id`) REFERENCES `info_intake` (`id`) ON DELETE CASCADE,
  CONSTRAINT `info_intake_ibfk_2` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `info_intake_ibfk_3` FOREIGN KEY (`potential_answer_id`) REFERENCES `potential_answer` (`id`),
  CONSTRAINT `info_intake_ibfk_4` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `info_intake_ibfk_8` FOREIGN KEY (`summary_localized_text_id`) REFERENCES `app_text` (`id`),
  CONSTRAINT `info_intake_ibfk_9` FOREIGN KEY (`parent_question_id`) REFERENCES `question` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=502 DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `item_cost`
--

DROP TABLE IF EXISTS `item_cost`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `item_cost` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `status` varchar(32) NOT NULL,
  `creation_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `sku_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `sku_id` (`sku_id`),
  CONSTRAINT `item_cost_ibfk_1` FOREIGN KEY (`sku_id`) REFERENCES `sku` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
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
-- Table structure for table `layout_blob_storage`
--

DROP TABLE IF EXISTS `layout_blob_storage`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `layout_blob_storage` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `layout` mediumblob,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=51 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `layout_version`
--

DROP TABLE IF EXISTS `layout_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `layout_version` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `status` varchar(250) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `modified_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `role` varchar(250) DEFAULT NULL,
  `layout_purpose` varchar(250) DEFAULT NULL,
  `layout_blob_storage_id` int(10) unsigned DEFAULT NULL,
  `major` int(10) unsigned NOT NULL,
  `minor` int(10) unsigned NOT NULL,
  `patch` int(10) unsigned NOT NULL,
  `sku_id` int(10) unsigned DEFAULT NULL,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `layout_blob_storage_id` (`layout_blob_storage_id`),
  KEY `sku_id` (`sku_id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  CONSTRAINT `layout_version_ibfk_3` FOREIGN KEY (`layout_blob_storage_id`) REFERENCES `layout_blob_storage` (`id`),
  CONSTRAINT `layout_version_ibfk_4` FOREIGN KEY (`sku_id`) REFERENCES `sku` (`id`),
  CONSTRAINT `layout_version_ibfk_5` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=221 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `line_item`
--

DROP TABLE IF EXISTS `line_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `line_item` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `currency` varchar(10) NOT NULL,
  `description` varchar(300) NOT NULL,
  `amount` int(11) NOT NULL,
  `item_cost_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `item_cost_id` (`item_cost_id`),
  CONSTRAINT `line_item_ibfk_1` FOREIGN KEY (`item_cost_id`) REFERENCES `item_cost` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
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
) ENGINE=InnoDB AUTO_INCREMENT=516 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `media`
--

DROP TABLE IF EXISTS `media`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `media` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `uploaded_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `uploader_id` bigint(20) unsigned NOT NULL,
  `mimetype` varchar(128) NOT NULL,
  `url` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `uploader_id` (`uploader_id`),
  CONSTRAINT `media_ibfk_1` FOREIGN KEY (`uploader_id`) REFERENCES `person` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `media_claim`
--

DROP TABLE IF EXISTS `media_claim`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `media_claim` (
  `media_id` int(10) unsigned NOT NULL,
  `claimer_type` varchar(64) DEFAULT NULL,
  `claimer_id` bigint(20) unsigned DEFAULT NULL,
  UNIQUE KEY `media_id` (`media_id`,`claimer_type`,`claimer_id`),
  KEY `claimer_type` (`claimer_type`,`claimer_id`),
  CONSTRAINT `media_claim_ibfk_1` FOREIGN KEY (`media_id`) REFERENCES `media` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `migrations`
--

DROP TABLE IF EXISTS `migrations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `migrations` (
  `migration_id` int(10) unsigned NOT NULL,
  `migration_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `migration_user` varchar(100) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `notification_prompt_status`
--

DROP TABLE IF EXISTS `notification_prompt_status`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `notification_prompt_status` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account_id` int(10) unsigned NOT NULL,
  `prompt_status` varchar(100) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `notification_prompt_status_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `parked_account`
--

DROP TABLE IF EXISTS `parked_account`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `parked_account` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(250) NOT NULL,
  `state` varchar(250) NOT NULL,
  `promotion_code_id` int(10) unsigned NOT NULL,
  `account_created` tinyint(1) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_modified_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  KEY `promotion_code_id` (`promotion_code_id`),
  CONSTRAINT `parked_account_ibfk_1` FOREIGN KEY (`promotion_code_id`) REFERENCES `promotion_code` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
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
  `gender` varchar(500) NOT NULL,
  `status` varchar(500) NOT NULL,
  `account_id` int(10) unsigned NOT NULL,
  `erx_patient_id` int(10) unsigned DEFAULT NULL,
  `prefix` varchar(100) DEFAULT NULL,
  `middle_name` varchar(100) DEFAULT NULL,
  `suffix` varchar(100) DEFAULT NULL,
  `payment_service_customer_id` varchar(200) DEFAULT NULL,
  `dob_month` int(10) unsigned NOT NULL,
  `dob_year` int(10) unsigned NOT NULL,
  `dob_day` int(10) unsigned NOT NULL,
  `training` tinyint(1) NOT NULL,
  `has_parental_consent` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `patient_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=91 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_address_selection`
--

DROP TABLE IF EXISTS `patient_address_selection`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_address_selection` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `address_id` int(10) unsigned NOT NULL,
  `label` varchar(100) DEFAULT NULL,
  `is_default` tinyint(1) NOT NULL,
  `is_updated_by_doctor` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`),
  KEY `address_id` (`address_id`),
  CONSTRAINT `patient_address_selection_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `patient_address_selection_ibfk_2` FOREIGN KEY (`address_id`) REFERENCES `address` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_agreement`
--

DROP TABLE IF EXISTS `patient_agreement`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_agreement` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `agreement_type` varchar(100) NOT NULL,
  `status` varchar(100) NOT NULL,
  `agreement_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `agreed` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`),
  CONSTRAINT `patient_agreement_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_alerts`
--

DROP TABLE IF EXISTS `patient_alerts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_alerts` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `alert` varchar(1024) NOT NULL,
  `question_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `patient_visit_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_visit_id` (`patient_visit_id`),
  KEY `question_id` (`question_id`),
  CONSTRAINT `patient_alerts_ibfk_2` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`),
  CONSTRAINT `patient_alerts_ibfk_3` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_care_provider_assignment`
--

DROP TABLE IF EXISTS `patient_care_provider_assignment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_care_provider_assignment` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `role_type_id` int(10) unsigned NOT NULL,
  `provider_id` int(10) unsigned NOT NULL,
  `status` varchar(250) NOT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `expires` timestamp(6) NULL DEFAULT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `role_type_id` (`role_type_id`,`provider_id`,`patient_id`),
  KEY `patient_id` (`patient_id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  CONSTRAINT `patient_care_provider_assignment_ibfk_3` FOREIGN KEY (`role_type_id`) REFERENCES `role_type` (`id`),
  CONSTRAINT `patient_care_provider_assignment_ibfk_4` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `patient_care_provider_assignment_ibfk_6` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_case`
--

DROP TABLE IF EXISTS `patient_case`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_case` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  `name` varchar(250) NOT NULL,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  `closed_date` timestamp NULL DEFAULT NULL,
  `claimed` tinyint(1) NOT NULL DEFAULT '0',
  `timeout_date` timestamp NULL DEFAULT NULL,
  `requested_doctor_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  KEY `patient_id_status_key` (`patient_id`,`status`),
  KEY `timeout_date_index` (`timeout_date`),
  KEY `fk_requested_doctor_doctor_id` (`requested_doctor_id`),
  CONSTRAINT `fk_requested_doctor_doctor_id` FOREIGN KEY (`requested_doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `patient_case_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `patient_case_ibfk_3` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_case_care_provider_assignment`
--

DROP TABLE IF EXISTS `patient_case_care_provider_assignment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_case_care_provider_assignment` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_case_id` int(10) unsigned NOT NULL,
  `provider_id` int(10) unsigned NOT NULL,
  `role_type_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  `expires` timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `case_role_provider` (`patient_case_id`,`role_type_id`,`provider_id`),
  KEY `role_provider_status` (`role_type_id`,`provider_id`,`status`),
  CONSTRAINT `patient_case_care_provider_assignment_ibfk_1` FOREIGN KEY (`patient_case_id`) REFERENCES `patient_case` (`id`),
  CONSTRAINT `patient_case_care_provider_assignment_ibfk_2` FOREIGN KEY (`role_type_id`) REFERENCES `role_type` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_case_message`
--

DROP TABLE IF EXISTS `patient_case_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_case_message` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `tstamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `person_id` bigint(20) unsigned NOT NULL,
  `body` mediumtext NOT NULL,
  `patient_case_id` int(10) unsigned NOT NULL,
  `private` tinyint(1) NOT NULL DEFAULT '0',
  `event_text` mediumtext NOT NULL,
  PRIMARY KEY (`id`),
  KEY `person_id` (`person_id`),
  KEY `case_tstamp` (`patient_case_id`,`tstamp`),
  CONSTRAINT `patient_case_message_ibfk_2` FOREIGN KEY (`person_id`) REFERENCES `person` (`id`),
  CONSTRAINT `patient_case_message_ibfk_3` FOREIGN KEY (`patient_case_id`) REFERENCES `patient_case` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_case_message_attachment`
--

DROP TABLE IF EXISTS `patient_case_message_attachment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_case_message_attachment` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `message_id` bigint(20) unsigned DEFAULT NULL,
  `item_type` varchar(64) NOT NULL,
  `item_id` bigint(20) unsigned NOT NULL,
  `title` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `message_id` (`message_id`),
  CONSTRAINT `patient_case_message_attachment_ibfk_1` FOREIGN KEY (`message_id`) REFERENCES `patient_case_message` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_case_message_participant`
--

DROP TABLE IF EXISTS `patient_case_message_participant`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_case_message_participant` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `patient_case_id` int(10) unsigned NOT NULL,
  `person_id` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `patient_case_id` (`patient_case_id`,`person_id`),
  KEY `person_id` (`person_id`),
  CONSTRAINT `patient_case_message_participant_ibfk_1` FOREIGN KEY (`patient_case_id`) REFERENCES `patient_case` (`id`),
  CONSTRAINT `patient_case_message_participant_ibfk_2` FOREIGN KEY (`person_id`) REFERENCES `person` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_case_message_read`
--

DROP TABLE IF EXISTS `patient_case_message_read`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_case_message_read` (
  `message_id` bigint(20) unsigned NOT NULL,
  `person_id` bigint(20) unsigned NOT NULL,
  `timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`message_id`,`person_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_case_note`
--

DROP TABLE IF EXISTS `patient_case_note`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_case_note` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `case_id` int(10) unsigned NOT NULL,
  `author_doctor_id` int(10) unsigned NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `note_text` text CHARACTER SET utf8mb4 NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_case_note_patient_case` (`case_id`),
  KEY `patient_case_note_doctor` (`author_doctor_id`),
  CONSTRAINT `patient_case_note_doctor` FOREIGN KEY (`author_doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `patient_case_note_patient_case` FOREIGN KEY (`case_id`) REFERENCES `patient_case` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_charge_item`
--

DROP TABLE IF EXISTS `patient_charge_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_charge_item` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `currency` varchar(10) NOT NULL,
  `description` varchar(300) NOT NULL,
  `amount` int(11) NOT NULL,
  `patient_receipt_id` int(10) unsigned NOT NULL,
  `creation_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `patient_receipt_id` (`patient_receipt_id`),
  CONSTRAINT `patient_charge_item_ibfk_1` FOREIGN KEY (`patient_receipt_id`) REFERENCES `patient_receipt` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_doctor_layout_mapping`
--

DROP TABLE IF EXISTS `patient_doctor_layout_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_doctor_layout_mapping` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `dr_major` int(10) unsigned NOT NULL,
  `dr_minor` int(10) unsigned NOT NULL,
  `patient_major` int(10) unsigned NOT NULL,
  `patient_minor` int(10) unsigned NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `sku_id` int(10) unsigned NOT NULL,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `dr_patient_sku_pathway` (`dr_major`,`dr_minor`,`patient_major`,`patient_minor`,`sku_id`,`clinical_pathway_id`),
  KEY `sku_id` (`sku_id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  CONSTRAINT `patient_doctor_layout_mapping_ibfk_2` FOREIGN KEY (`sku_id`) REFERENCES `sku` (`id`),
  CONSTRAINT `patient_doctor_layout_mapping_ibfk_3` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_emergency_contact`
--

DROP TABLE IF EXISTS `patient_emergency_contact`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_emergency_contact` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `full_name` varchar(1024) NOT NULL,
  `phone_number` varchar(30) NOT NULL,
  `relationship` varchar(100) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`),
  CONSTRAINT `patient_emergency_contact_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_exported_medical_record`
--

DROP TABLE IF EXISTS `patient_exported_medical_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_exported_medical_record` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `status` varchar(32) NOT NULL,
  `error` varchar(256) DEFAULT NULL,
  `storage_url` varchar(512) DEFAULT NULL,
  `requested_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `completed_timestamp` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`),
  CONSTRAINT `patient_exported_medical_record_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_feedback`
--

DROP TABLE IF EXISTS `patient_feedback`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_feedback` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `feedback_for` varchar(32) NOT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `rating` int(11) NOT NULL,
  `comment` text CHARACTER SET utf8mb4,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`,`feedback_for`),
  CONSTRAINT `patient_feedback_patient` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_layout_version`
--

DROP TABLE IF EXISTS `patient_layout_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_layout_version` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `language_id` int(10) unsigned NOT NULL,
  `layout_version_id` int(10) unsigned NOT NULL,
  `status` varchar(250) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `modified_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `layout_blob_storage_id` int(10) unsigned DEFAULT NULL,
  `major` int(10) unsigned NOT NULL,
  `minor` int(10) unsigned NOT NULL,
  `patch` int(10) unsigned NOT NULL,
  `sku_id` int(10) unsigned NOT NULL,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `language_id` (`language_id`),
  KEY `layout_blob_storage_id` (`layout_blob_storage_id`),
  KEY `sku_id` (`sku_id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  CONSTRAINT `patient_layout_version_ibfk_1` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `patient_layout_version_ibfk_2` FOREIGN KEY (`language_id`) REFERENCES `languages_supported` (`id`),
  CONSTRAINT `patient_layout_version_ibfk_6` FOREIGN KEY (`layout_blob_storage_id`) REFERENCES `layout_blob_storage` (`id`),
  CONSTRAINT `patient_layout_version_ibfk_7` FOREIGN KEY (`sku_id`) REFERENCES `sku` (`id`),
  CONSTRAINT `patient_layout_version_ibfk_8` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=150 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_location`
--

DROP TABLE IF EXISTS `patient_location`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_location` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `zip_code` varchar(100) NOT NULL,
  `city` varchar(150) DEFAULT NULL,
  `state` varchar(150) DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`),
  CONSTRAINT `patient_location_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_parent`
--

DROP TABLE IF EXISTS `patient_parent`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_parent` (
  `patient_id` int(10) unsigned NOT NULL,
  `parent_patient_id` int(10) unsigned NOT NULL,
  `consented` tinyint(1) NOT NULL DEFAULT '0',
  `relationship` varchar(128) NOT NULL,
  PRIMARY KEY (`patient_id`,`parent_patient_id`),
  KEY `patient_parent_parent` (`parent_patient_id`),
  CONSTRAINT `patient_parent_parent` FOREIGN KEY (`parent_patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `patient_parent_patient` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_pcp`
--

DROP TABLE IF EXISTS `patient_pcp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_pcp` (
  `patient_id` int(10) unsigned NOT NULL,
  `physician_name` varchar(500) NOT NULL,
  `phone_number` varchar(30) NOT NULL,
  `practice_name` varchar(300) NOT NULL,
  `email` varchar(300) NOT NULL,
  `fax_number` varchar(300) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`patient_id`),
  CONSTRAINT `patient_pcp_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_pharmacy_selection`
--

DROP TABLE IF EXISTS `patient_pharmacy_selection`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_pharmacy_selection` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `pharmacy_id` varchar(300) DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  `erx_pharmacy_id` int(10) unsigned DEFAULT NULL,
  `pharmacy_selection_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`),
  KEY `pharmacy_selection_id` (`pharmacy_selection_id`),
  CONSTRAINT `patient_pharmacy_selection_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `patient_pharmacy_selection_ibfk_2` FOREIGN KEY (`pharmacy_selection_id`) REFERENCES `pharmacy_selection` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_receipt`
--

DROP TABLE IF EXISTS `patient_receipt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_receipt` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `item_id` int(10) unsigned NOT NULL,
  `receipt_reference_id` varchar(32) NOT NULL,
  `stripe_charge_id` varchar(32) DEFAULT NULL,
  `creation_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_modified_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `status` varchar(32) NOT NULL,
  `item_cost_id` int(10) unsigned NOT NULL,
  `sku_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `patient_id` (`patient_id`,`item_id`),
  UNIQUE KEY `patient_id_2` (`patient_id`,`item_id`,`sku_id`),
  KEY `item_cost_id` (`item_cost_id`),
  KEY `sku_id` (`sku_id`),
  KEY `creation_timestamp` (`creation_timestamp`),
  CONSTRAINT `patient_receipt_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `patient_receipt_ibfk_3` FOREIGN KEY (`item_cost_id`) REFERENCES `item_cost` (`id`),
  CONSTRAINT `patient_receipt_ibfk_4` FOREIGN KEY (`sku_id`) REFERENCES `sku` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
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
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `closed_date` timestamp NULL DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  `layout_version_id` int(10) unsigned NOT NULL,
  `submitted_date` timestamp NULL DEFAULT NULL,
  `patient_case_id` int(10) unsigned DEFAULT NULL,
  `last_modified_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `sku_id` int(10) unsigned NOT NULL,
  `followup` tinyint(1) NOT NULL DEFAULT '0',
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`),
  KEY `layout_version_id` (`layout_version_id`),
  KEY `sku_id` (`sku_id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  KEY `case_status` (`patient_case_id`,`status`,`submitted_date`),
  CONSTRAINT `fk_patient_visit_patient_case_id` FOREIGN KEY (`patient_case_id`) REFERENCES `patient_case` (`id`),
  CONSTRAINT `patient_visit_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `patient_visit_ibfk_3` FOREIGN KEY (`layout_version_id`) REFERENCES `layout_version` (`id`),
  CONSTRAINT `patient_visit_ibfk_4` FOREIGN KEY (`sku_id`) REFERENCES `sku` (`id`),
  CONSTRAINT `patient_visit_ibfk_5` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=89 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_visit_event`
--

DROP TABLE IF EXISTS `patient_visit_event`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_visit_event` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_visit_id` int(10) unsigned DEFAULT NULL,
  `event` varchar(100) NOT NULL,
  `status` varchar(100) NOT NULL,
  `message` varchar(600) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_visit_id` (`patient_visit_id`),
  CONSTRAINT `patient_visit_event_ibfk_1` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_visit_follow_up`
--

DROP TABLE IF EXISTS `patient_visit_follow_up`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_visit_follow_up` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `doctor_id` int(10) unsigned NOT NULL,
  `follow_up_date` date NOT NULL,
  `follow_up_value` int(10) unsigned NOT NULL,
  `follow_up_unit` varchar(100) NOT NULL,
  `status` varchar(100) NOT NULL,
  `treatment_plan_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `treatment_plan_id` (`treatment_plan_id`),
  CONSTRAINT `patient_visit_follow_up_ibfk_2` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `patient_visit_follow_up_ibfk_3` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `patient_visit_message`
--

DROP TABLE IF EXISTS `patient_visit_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `patient_visit_message` (
  `patient_visit_id` int(10) unsigned NOT NULL,
  `message` mediumtext NOT NULL,
  `creation_timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`patient_visit_id`),
  CONSTRAINT `patient_visit_message_ibfk_1` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pending_task`
--

DROP TABLE IF EXISTS `pending_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pending_task` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(100) NOT NULL,
  `item_id` int(10) unsigned NOT NULL,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `person`
--

DROP TABLE IF EXISTS `person`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `person` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `role_id` int(10) unsigned NOT NULL,
  `role_type_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `role_type_id_2` (`role_type_id`,`role_id`),
  KEY `role_type_id` (`role_type_id`),
  CONSTRAINT `person_ibfk_1` FOREIGN KEY (`role_type_id`) REFERENCES `role_type` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pharmacy_dispensed_treatment`
--

DROP TABLE IF EXISTS `pharmacy_dispensed_treatment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pharmacy_dispensed_treatment` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_internal_name` varchar(250) NOT NULL,
  `dispense_value` decimal(21,10) NOT NULL,
  `refills` int(11) NOT NULL,
  `substitutions_allowed` tinyint(1) NOT NULL,
  `days_supply` int(10) unsigned DEFAULT NULL,
  `pharmacy_notes` varchar(250) DEFAULT NULL,
  `pharmacy_id` int(10) unsigned NOT NULL,
  `patient_instructions` varchar(150) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  `dosage_strength` varchar(250) NOT NULL,
  `type` varchar(150) NOT NULL,
  `drug_name_id` int(10) unsigned DEFAULT NULL,
  `drug_form_id` int(10) unsigned DEFAULT NULL,
  `drug_route_id` int(10) unsigned DEFAULT NULL,
  `erx_id` int(10) unsigned NOT NULL,
  `erx_last_filled_date` timestamp NULL DEFAULT NULL,
  `erx_sent_date` timestamp NULL DEFAULT NULL,
  `dispense_unit` varchar(100) NOT NULL,
  `requested_treatment_id` int(10) unsigned DEFAULT NULL,
  `doctor_id` int(10) unsigned DEFAULT NULL,
  `is_controlled_substance` tinyint(4) DEFAULT NULL,
  `generic_drug_name_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `drug_route_id` (`drug_route_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `pharmacy_id` (`pharmacy_id`),
  KEY `unlinked_requested_treatment_id` (`requested_treatment_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `generic_drug_name_id` (`generic_drug_name_id`),
  CONSTRAINT `pharmacy_dispensed_treatment_ibfk_2` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `pharmacy_dispensed_treatment_ibfk_3` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`),
  CONSTRAINT `pharmacy_dispensed_treatment_ibfk_4` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `pharmacy_dispensed_treatment_ibfk_6` FOREIGN KEY (`pharmacy_id`) REFERENCES `pharmacy_selection` (`id`),
  CONSTRAINT `pharmacy_dispensed_treatment_ibfk_7` FOREIGN KEY (`requested_treatment_id`) REFERENCES `requested_treatment` (`id`),
  CONSTRAINT `pharmacy_dispensed_treatment_ibfk_8` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `pharmacy_dispensed_treatment_ibfk_9` FOREIGN KEY (`generic_drug_name_id`) REFERENCES `drug_name` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pharmacy_dispensed_treatment_drug_db_id`
--

DROP TABLE IF EXISTS `pharmacy_dispensed_treatment_drug_db_id`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pharmacy_dispensed_treatment_drug_db_id` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_db_id` varchar(100) NOT NULL,
  `drug_db_id_tag` varchar(100) NOT NULL,
  `pharmacy_dispensed_treatment_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `pharmacy_dispensed_treatment_id` (`pharmacy_dispensed_treatment_id`),
  CONSTRAINT `pharmacy_dispensed_treatment_drug_db_id_ibfk_1` FOREIGN KEY (`pharmacy_dispensed_treatment_id`) REFERENCES `pharmacy_dispensed_treatment` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pharmacy_selection`
--

DROP TABLE IF EXISTS `pharmacy_selection`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pharmacy_selection` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `pharmacy_id` int(10) unsigned NOT NULL,
  `address_line_1` varchar(500) DEFAULT NULL,
  `address_line_2` varchar(500) DEFAULT NULL,
  `source` varchar(100) NOT NULL,
  `city` varchar(100) DEFAULT NULL,
  `state` varchar(100) DEFAULT NULL,
  `country` varchar(100) DEFAULT NULL,
  `phone` varchar(100) DEFAULT NULL,
  `zip_code` varchar(100) DEFAULT NULL,
  `lat` varchar(100) DEFAULT NULL,
  `lng` varchar(100) DEFAULT NULL,
  `name` varchar(500) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `photo`
--

DROP TABLE IF EXISTS `photo`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photo` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `uploaded` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `uploader_id` bigint(20) unsigned NOT NULL,
  `mimetype` varchar(128) NOT NULL,
  `url` varchar(255) NOT NULL,
  `claimer_type` varchar(64) DEFAULT NULL,
  `claimer_id` bigint(20) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `uploader_id` (`uploader_id`),
  CONSTRAINT `photo_ibfk_1` FOREIGN KEY (`uploader_id`) REFERENCES `person` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `photo_intake_section`
--

DROP TABLE IF EXISTS `photo_intake_section`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photo_intake_section` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `section_name` text NOT NULL,
  `question_id` int(10) unsigned NOT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `patient_visit_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `client_clock` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `question_id` (`question_id`),
  KEY `patient_id` (`patient_id`),
  KEY `patient_visit_id` (`patient_visit_id`),
  CONSTRAINT `photo_intake_section_ibfk_1` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `photo_intake_section_ibfk_2` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `photo_intake_section_ibfk_3` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `photo_intake_slot`
--

DROP TABLE IF EXISTS `photo_intake_slot`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photo_intake_slot` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `photo_slot_id` int(10) unsigned NOT NULL,
  `photo_id` int(10) unsigned NOT NULL,
  `photo_slot_name` varchar(150) NOT NULL,
  `photo_intake_section_id` int(10) unsigned NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  KEY `photo_slot_id` (`photo_slot_id`),
  KEY `photo_id` (`photo_id`),
  KEY `photo_intake_section_id` (`photo_intake_section_id`),
  CONSTRAINT `photo_intake_slot_ibfk_1` FOREIGN KEY (`photo_slot_id`) REFERENCES `photo_slot` (`id`),
  CONSTRAINT `photo_intake_slot_ibfk_4` FOREIGN KEY (`photo_id`) REFERENCES `media` (`id`),
  CONSTRAINT `photo_intake_slot_ibfk_5` FOREIGN KEY (`photo_intake_section_id`) REFERENCES `photo_intake_section` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `photo_slot`
--

DROP TABLE IF EXISTS `photo_slot`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photo_slot` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `question_id` int(10) unsigned NOT NULL,
  `required` tinyint(1) NOT NULL,
  `status` varchar(100) NOT NULL,
  `placeholder_image_tag` varchar(100) DEFAULT NULL,
  `ordering` int(10) unsigned NOT NULL,
  `language_id` int(10) unsigned DEFAULT '1',
  `name_text` varchar(600) DEFAULT NULL,
  `photo_slot_type` varchar(60) NOT NULL,
  `client_data` blob,
  PRIMARY KEY (`id`),
  KEY `question_id` (`question_id`),
  KEY `fk_photo_slot_languages_supported_id` (`language_id`),
  CONSTRAINT `fk_photo_slot_languages_supported_id` FOREIGN KEY (`language_id`) REFERENCES `languages_supported` (`id`),
  CONSTRAINT `photo_slot_ibfk_1` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `photo_slot_type`
--

DROP TABLE IF EXISTS `photo_slot_type`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `photo_slot_type` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `slot_type` varchar(100) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8;
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
  `potential_answer_tag` varchar(250) NOT NULL,
  `ordering` int(10) unsigned NOT NULL,
  `answer_summary_text_id` int(10) unsigned DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  `to_alert` tinyint(1) DEFAULT NULL,
  `language_id` int(10) unsigned NOT NULL DEFAULT '1',
  `answer_text` varchar(600) DEFAULT NULL,
  `answer_summary_text` varchar(600) DEFAULT NULL,
  `answer_type` varchar(60) NOT NULL,
  `client_data` blob,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_potential_answer_tag_quid_order` (`potential_answer_tag`,`question_id`,`ordering`,`language_id`),
  KEY `outcome_localized_text` (`answer_localized_text_id`),
  KEY `answer_summary_text_id` (`answer_summary_text_id`),
  KEY `fk_potential_answer_languages_supported_id` (`language_id`),
  KEY `fk_question_question_id` (`question_id`),
  CONSTRAINT `fk_potential_answer_languages_supported_id` FOREIGN KEY (`language_id`) REFERENCES `languages_supported` (`id`),
  CONSTRAINT `fk_question_question_id` FOREIGN KEY (`question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `potential_answer_ibfk_3` FOREIGN KEY (`answer_summary_text_id`) REFERENCES `app_text` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=258 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `promo_code_prefix`
--

DROP TABLE IF EXISTS `promo_code_prefix`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `promo_code_prefix` (
  `prefix` varchar(32) NOT NULL,
  `status` varchar(32) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`prefix`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `promotion`
--

DROP TABLE IF EXISTS `promotion`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `promotion` (
  `promotion_code_id` int(10) unsigned NOT NULL,
  `promo_type` varchar(32) NOT NULL,
  `promo_data` blob NOT NULL,
  `promotion_group_id` int(10) unsigned NOT NULL,
  `expires` timestamp NULL DEFAULT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`promotion_code_id`),
  KEY `promotion_group_id` (`promotion_group_id`),
  CONSTRAINT `promotion_ibfk_1` FOREIGN KEY (`promotion_group_id`) REFERENCES `promotion_group` (`id`),
  CONSTRAINT `promotion_ibfk_2` FOREIGN KEY (`promotion_code_id`) REFERENCES `promotion_code` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `promotion_code`
--

DROP TABLE IF EXISTS `promotion_code`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `promotion_code` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(32) NOT NULL,
  `is_referral` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `promotion_group`
--

DROP TABLE IF EXISTS `promotion_group`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `promotion_group` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `max_allowed_promos` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `promotion_referral_route`
--

DROP TABLE IF EXISTS `promotion_referral_route`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `promotion_referral_route` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `promotion_code_id` int(10) unsigned NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `priority` int(10) unsigned NOT NULL,
  `lifecycle` varchar(25) NOT NULL DEFAULT 'ACTIVE',
  `gender` varchar(1) DEFAULT NULL,
  `age_lower` int(10) unsigned DEFAULT NULL,
  `age_upper` int(10) unsigned DEFAULT NULL,
  `state` varchar(2) DEFAULT NULL,
  `pharmacy` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `promotion_referral_route_priority_idx` (`priority`),
  KEY `promotion_referral_route_gender_idx` (`gender`),
  KEY `promotion_referral_route_age_lower_idx` (`age_lower`),
  KEY `promotion_referral_route_age_upper_idx` (`age_upper`),
  KEY `promotion_referral_route_state_idx` (`state`),
  KEY `promotion_referral_route_pharmacy_idx` (`pharmacy`),
  KEY `promotion_referral_route_promotion_code` (`promotion_code_id`),
  CONSTRAINT `promotion_referral_route_promotion_code` FOREIGN KEY (`promotion_code_id`) REFERENCES `promotion_code` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `push_config`
--

DROP TABLE IF EXISTS `push_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `push_config` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `account_id` int(10) unsigned NOT NULL,
  `device_token` varbinary(500) NOT NULL,
  `push_endpoint` varchar(300) NOT NULL,
  `platform` varchar(100) NOT NULL,
  `platform_version` varchar(100) NOT NULL,
  `app_type` varchar(100) NOT NULL,
  `app_env` varchar(100) NOT NULL,
  `app_version` varchar(100) NOT NULL,
  `device` varchar(100) NOT NULL,
  `device_model` varchar(100) NOT NULL,
  `device_id` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  UNIQUE KEY `device_token` (`device_token`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `push_config_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `question`
--

DROP TABLE IF EXISTS `question`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `question` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `qtext_app_text_id` int(10) unsigned DEFAULT NULL,
  `qtext_short_text_id` int(10) unsigned DEFAULT NULL,
  `subtext_app_text_id` int(10) unsigned DEFAULT NULL,
  `question_tag` varchar(250) NOT NULL,
  `parent_question_id` int(10) unsigned DEFAULT NULL,
  `required` tinyint(1) DEFAULT NULL,
  `formatted_field_tags` varchar(150) DEFAULT NULL,
  `to_alert` tinyint(1) DEFAULT NULL,
  `alert_app_text_id` int(10) unsigned DEFAULT NULL,
  `qtext_has_tokens` tinyint(1) DEFAULT NULL,
  `language_id` int(10) unsigned DEFAULT '1',
  `version` int(10) unsigned NOT NULL DEFAULT '1',
  `summary_text` varchar(600) DEFAULT NULL,
  `subtext_text` varchar(600) DEFAULT NULL,
  `question_text` varchar(600) DEFAULT NULL,
  `alert_text` varchar(600) DEFAULT NULL,
  `question_type` varchar(60) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_question_question_tag_version` (`question_tag`,`version`,`language_id`),
  KEY `subtext_app_text_id` (`subtext_app_text_id`),
  KEY `qtext_app_text_id` (`qtext_app_text_id`),
  KEY `qtext_short_text_id` (`qtext_short_text_id`),
  KEY `parent_question_id` (`parent_question_id`),
  KEY `alert_app_text_id` (`alert_app_text_id`),
  KEY `fk_question_languages_supported_id` (`language_id`),
  CONSTRAINT `fk_question_languages_supported_id` FOREIGN KEY (`language_id`) REFERENCES `languages_supported` (`id`),
  CONSTRAINT `question_ibfk_2` FOREIGN KEY (`subtext_app_text_id`) REFERENCES `app_text` (`id`),
  CONSTRAINT `question_ibfk_3` FOREIGN KEY (`qtext_app_text_id`) REFERENCES `app_text` (`id`),
  CONSTRAINT `question_ibfk_4` FOREIGN KEY (`qtext_short_text_id`) REFERENCES `app_text` (`id`),
  CONSTRAINT `question_ibfk_5` FOREIGN KEY (`parent_question_id`) REFERENCES `question` (`id`),
  CONSTRAINT `question_ibfk_6` FOREIGN KEY (`alert_app_text_id`) REFERENCES `app_text` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=95 DEFAULT CHARSET=utf8;
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
  `deprecated` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `qtype` (`qtype`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `referral_program`
--

DROP TABLE IF EXISTS `referral_program`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `referral_program` (
  `referral_program_template_id` int(10) unsigned DEFAULT NULL,
  `account_id` int(10) unsigned NOT NULL,
  `promotion_code_id` int(10) unsigned NOT NULL,
  `referral_type` varchar(32) NOT NULL,
  `referral_data` blob NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` varchar(32) NOT NULL,
  `promotion_referral_route_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`account_id`,`promotion_code_id`),
  KEY `promotion_code_id` (`promotion_code_id`),
  KEY `referral_program_template_id` (`referral_program_template_id`),
  KEY `referral_program_promotion_referral_route` (`promotion_referral_route_id`),
  CONSTRAINT `referral_program_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`),
  CONSTRAINT `referral_program_ibfk_2` FOREIGN KEY (`promotion_code_id`) REFERENCES `promotion_code` (`id`),
  CONSTRAINT `referral_program_ibfk_3` FOREIGN KEY (`referral_program_template_id`) REFERENCES `referral_program_template` (`id`),
  CONSTRAINT `referral_program_promotion_referral_route` FOREIGN KEY (`promotion_referral_route_id`) REFERENCES `promotion_referral_route` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `referral_program_template`
--

DROP TABLE IF EXISTS `referral_program_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `referral_program_template` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `referral_type` varchar(32) NOT NULL,
  `referral_data` blob NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` varchar(32) NOT NULL,
  `role_type_id` int(10) unsigned NOT NULL,
  `promotion_code_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `role_type_id` (`role_type_id`,`status`),
  KEY `referral_program_template_promotion_code` (`promotion_code_id`),
  CONSTRAINT `referral_program_template_ibfk_1` FOREIGN KEY (`role_type_id`) REFERENCES `role_type` (`id`),
  CONSTRAINT `referral_program_template_promotion_code` FOREIGN KEY (`promotion_code_id`) REFERENCES `promotion_code` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `regimen`
--

DROP TABLE IF EXISTS `regimen`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `regimen` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `dr_regimen_step_id` int(10) unsigned DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `treatment_plan_id` int(10) unsigned NOT NULL,
  `text` varchar(2048) NOT NULL,
  `regimen_section_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_regimen_step_id` (`dr_regimen_step_id`),
  KEY `treatment_plan_id` (`treatment_plan_id`),
  CONSTRAINT `regimen_ibfk_2` FOREIGN KEY (`dr_regimen_step_id`) REFERENCES `dr_regimen_step` (`id`),
  CONSTRAINT `regimen_ibfk_3` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `regimen_section`
--

DROP TABLE IF EXISTS `regimen_section`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `regimen_section` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(500) NOT NULL,
  `creation_date` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `treatment_plan_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `treatment_plan_id` (`treatment_plan_id`),
  CONSTRAINT `regimen_section_ibfk_1` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `regimen_step`
--

DROP TABLE IF EXISTS `regimen_step`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `regimen_step` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `text` varchar(2048) NOT NULL,
  `drug_name_id` int(10) unsigned DEFAULT NULL,
  `drug_form_id` int(10) unsigned DEFAULT NULL,
  `drug_route_id` int(10) unsigned DEFAULT NULL,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `drug_route_id` (`drug_route_id`),
  CONSTRAINT `regimen_step_ibfk_1` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `regimen_step_ibfk_2` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `regimen_step_ibfk_3` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8;
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
-- Table structure for table `requested_treatment`
--

DROP TABLE IF EXISTS `requested_treatment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `requested_treatment` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_internal_name` varchar(250) NOT NULL,
  `dispense_value` decimal(21,10) NOT NULL,
  `refills` int(11) NOT NULL,
  `substitutions_allowed` tinyint(1) NOT NULL,
  `days_supply` int(10) unsigned DEFAULT NULL,
  `pharmacy_id` int(10) unsigned NOT NULL,
  `pharmacy_notes` varchar(250) DEFAULT NULL,
  `patient_instructions` varchar(150) NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` varchar(100) NOT NULL,
  `dosage_strength` varchar(250) NOT NULL,
  `type` varchar(150) NOT NULL,
  `drug_name_id` int(10) unsigned DEFAULT NULL,
  `drug_form_id` int(10) unsigned DEFAULT NULL,
  `drug_route_id` int(10) unsigned DEFAULT NULL,
  `erx_id` int(10) unsigned DEFAULT NULL,
  `erx_last_filled_date` timestamp NULL DEFAULT NULL,
  `erx_sent_date` timestamp NULL DEFAULT NULL,
  `dispense_unit` varchar(100) NOT NULL,
  `originating_treatment_id` int(10) unsigned DEFAULT NULL,
  `doctor_id` int(10) unsigned DEFAULT NULL,
  `is_controlled_substance` tinyint(4) DEFAULT NULL,
  `generic_drug_name_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `drug_route_id` (`drug_route_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `pharmacy_id` (`pharmacy_id`),
  KEY `originating_treatment_id` (`originating_treatment_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `generic_drug_name_id` (`generic_drug_name_id`),
  CONSTRAINT `requested_treatment_ibfk_1` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `requested_treatment_ibfk_2` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`),
  CONSTRAINT `requested_treatment_ibfk_3` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `requested_treatment_ibfk_5` FOREIGN KEY (`pharmacy_id`) REFERENCES `pharmacy_selection` (`id`),
  CONSTRAINT `requested_treatment_ibfk_6` FOREIGN KEY (`originating_treatment_id`) REFERENCES `treatment` (`id`),
  CONSTRAINT `requested_treatment_ibfk_7` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `requested_treatment_ibfk_8` FOREIGN KEY (`generic_drug_name_id`) REFERENCES `drug_name` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `requested_treatment_drug_db_id`
--

DROP TABLE IF EXISTS `requested_treatment_drug_db_id`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `requested_treatment_drug_db_id` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_db_id` varchar(100) NOT NULL,
  `drug_db_id_tag` varchar(100) NOT NULL,
  `requested_treatment_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `unlinked_requested_treatment_id` (`requested_treatment_id`),
  CONSTRAINT `requested_treatment_drug_db_id_ibfk_1` FOREIGN KEY (`requested_treatment_id`) REFERENCES `requested_treatment` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `resource_guide`
--

DROP TABLE IF EXISTS `resource_guide`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `resource_guide` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `section_id` int(10) unsigned NOT NULL,
  `ordinal` int(11) NOT NULL,
  `title` varchar(256) NOT NULL,
  `photo_url` varchar(256) NOT NULL,
  `layout` blob NOT NULL,
  `creation_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `modified_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `active` tinyint(1) NOT NULL DEFAULT '0',
  `tag` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `resource_guide_tag` (`tag`),
  KEY `section_id` (`section_id`),
  KEY `resource_guide_active_ordinal` (`active`,`ordinal`),
  CONSTRAINT `resource_guide_ibfk_1` FOREIGN KEY (`section_id`) REFERENCES `resource_guide_section` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `resource_guide_section`
--

DROP TABLE IF EXISTS `resource_guide_section`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `resource_guide_section` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `ordinal` int(11) NOT NULL,
  `title` varchar(256) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `role_type`
--

DROP TABLE IF EXISTS `role_type`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `role_type` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `role_type_tag` varchar(250) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `rx_refill_request`
--

DROP TABLE IF EXISTS `rx_refill_request`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rx_refill_request` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `erx_request_queue_item_id` int(10) unsigned DEFAULT NULL,
  `reference_number` varchar(100) DEFAULT NULL,
  `pharmacy_rx_reference_number` varchar(100) DEFAULT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `request_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `doctor_id` int(10) unsigned NOT NULL,
  `dispensed_treatment_id` int(10) unsigned NOT NULL,
  `requested_treatment_id` int(10) unsigned DEFAULT NULL,
  `erx_id` int(10) unsigned DEFAULT NULL,
  `approved_refill_amount` int(10) unsigned DEFAULT NULL,
  `comments` varchar(500) DEFAULT NULL,
  `denial_reason_id` int(10) unsigned DEFAULT NULL,
  `creation_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`),
  KEY `dispensed_treatment_id` (`dispensed_treatment_id`),
  KEY `unlinked_requested_treatment_id` (`requested_treatment_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `patient_id` (`patient_id`),
  KEY `denial_reason_id` (`denial_reason_id`),
  KEY `erx_request_queue_item_id` (`erx_request_queue_item_id`),
  CONSTRAINT `rx_refill_request_ibfk_2` FOREIGN KEY (`dispensed_treatment_id`) REFERENCES `pharmacy_dispensed_treatment` (`id`),
  CONSTRAINT `rx_refill_request_ibfk_3` FOREIGN KEY (`requested_treatment_id`) REFERENCES `requested_treatment` (`id`),
  CONSTRAINT `rx_refill_request_ibfk_4` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `rx_refill_request_ibfk_5` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `rx_refill_request_ibfk_6` FOREIGN KEY (`denial_reason_id`) REFERENCES `deny_refill_reason` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `rx_refill_status_events`
--

DROP TABLE IF EXISTS `rx_refill_status_events`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `rx_refill_status_events` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `rx_refill_request_id` int(10) unsigned NOT NULL,
  `rx_refill_status` varchar(100) NOT NULL,
  `status` varchar(100) NOT NULL,
  `rx_refill_status_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `reported_timestamp` timestamp(6) NULL DEFAULT NULL,
  `event_details` varchar(500) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `rx_refill_request_id` (`rx_refill_request_id`),
  KEY `status` (`status`),
  CONSTRAINT `rx_refill_status_events_ibfk_1` FOREIGN KEY (`rx_refill_request_id`) REFERENCES `rx_refill_request` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `scheduled_message`
--

DROP TABLE IF EXISTS `scheduled_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `scheduled_message` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_id` int(10) unsigned NOT NULL,
  `message_type` varchar(64) NOT NULL,
  `message_json` blob NOT NULL,
  `event` varchar(64) NOT NULL,
  `status` varchar(64) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `scheduled` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `completed` timestamp NULL DEFAULT NULL,
  `error` varchar(512) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_id` (`patient_id`),
  KEY `scheduled_message_status_scheduled` (`status`,`scheduled`),
  CONSTRAINT `scheduled_message_ibfk_1` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `scheduled_message_template`
--

DROP TABLE IF EXISTS `scheduled_message_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `scheduled_message_template` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` text NOT NULL,
  `event` varchar(64) NOT NULL,
  `schedule_period` int(10) unsigned NOT NULL,
  `message` text NOT NULL,
  `creator_account_id` int(10) unsigned DEFAULT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `creator_account_id` (`creator_account_id`),
  CONSTRAINT `scheduled_message_template_ibfk_1` FOREIGN KEY (`creator_account_id`) REFERENCES `account` (`id`)
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
  `section_tag` varchar(250) NOT NULL,
  `clinical_pathway_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `section_tag` (`section_tag`),
  KEY `section_title_app_text_id` (`section_title_app_text_id`),
  KEY `clinical_pathway_id` (`clinical_pathway_id`),
  CONSTRAINT `section_ibfk_1` FOREIGN KEY (`section_title_app_text_id`) REFERENCES `app_text` (`id`),
  CONSTRAINT `section_ibfk_3` FOREIGN KEY (`clinical_pathway_id`) REFERENCES `clinical_pathway` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sku`
--

DROP TABLE IF EXISTS `sku`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sku` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `sku_category_id` int(10) unsigned NOT NULL,
  `type` varchar(128) NOT NULL,
  `status` varchar(32) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `type` (`type`),
  KEY `sku_category_id` (`sku_category_id`),
  CONSTRAINT `sku_ibfk_1` FOREIGN KEY (`sku_category_id`) REFERENCES `sku_category` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sku_category`
--

DROP TABLE IF EXISTS `sku_category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sku_category` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(32) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `type` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `state`
--

DROP TABLE IF EXISTS `state`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `state` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `full_name` varchar(300) NOT NULL,
  `abbreviation` varchar(10) NOT NULL,
  `country` varchar(300) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=101 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tag`
--

DROP TABLE IF EXISTS `tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tag` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `tag_text` varchar(255) DEFAULT NULL,
  `common` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `tag_text` (`tag_text`),
  KEY `tag_common_idx` (`common`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tag_membership`
--

DROP TABLE IF EXISTS `tag_membership`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tag_membership` (
  `tag_id` int(10) unsigned NOT NULL,
  `case_id` int(10) unsigned NOT NULL DEFAULT '0',
  `trigger_time` timestamp NULL DEFAULT NULL,
  `hidden` tinyint(1) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`tag_id`,`case_id`),
  KEY `case_id` (`case_id`),
  KEY `tag_membership_created_idx` (`created`),
  CONSTRAINT `tag_membership_ibfk_1` FOREIGN KEY (`tag_id`) REFERENCES `tag` (`id`) ON DELETE CASCADE,
  CONSTRAINT `tag_membership_ibfk_2` FOREIGN KEY (`case_id`) REFERENCES `patient_case` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tag_saved_search`
--

DROP TABLE IF EXISTS `tag_saved_search`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tag_saved_search` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(50) NOT NULL,
  `query` text NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `tag_saved_search_title` (`title`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `temp_auth_token`
--

DROP TABLE IF EXISTS `temp_auth_token`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `temp_auth_token` (
  `token` varchar(128) NOT NULL,
  `expires` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `purpose` varchar(32) NOT NULL,
  `account_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`purpose`,`token`),
  KEY `expires` (`expires`),
  KEY `account_id` (`account_id`),
  CONSTRAINT `temp_auth_token_ibfk_1` FOREIGN KEY (`account_id`) REFERENCES `account` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
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

--
-- Table structure for table `training_case`
--

DROP TABLE IF EXISTS `training_case`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `training_case` (
  `training_case_set_id` int(10) unsigned NOT NULL,
  `patient_visit_id` int(10) unsigned NOT NULL,
  `template_name` varchar(64) NOT NULL,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`patient_visit_id`),
  KEY `training_case_set_id` (`training_case_set_id`),
  CONSTRAINT `training_case_ibfk_1` FOREIGN KEY (`training_case_set_id`) REFERENCES `training_case_set` (`id`) ON DELETE CASCADE,
  CONSTRAINT `training_case_ibfk_2` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `training_case_set`
--

DROP TABLE IF EXISTS `training_case_set`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `training_case_set` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` varchar(32) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment`
--

DROP TABLE IF EXISTS `treatment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `treatment_plan_id` int(10) unsigned NOT NULL,
  `drug_internal_name` varchar(250) NOT NULL,
  `dispense_value` decimal(21,10) NOT NULL,
  `dispense_unit_id` int(10) unsigned NOT NULL,
  `refills` int(10) unsigned NOT NULL,
  `substitutions_allowed` tinyint(1) DEFAULT NULL,
  `days_supply` int(10) unsigned DEFAULT NULL,
  `pharmacy_notes` varchar(250) DEFAULT NULL,
  `patient_instructions` varchar(150) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  `dosage_strength` varchar(250) NOT NULL,
  `type` varchar(150) NOT NULL,
  `drug_name_id` int(10) unsigned NOT NULL,
  `drug_form_id` int(10) unsigned NOT NULL,
  `drug_route_id` int(10) unsigned NOT NULL,
  `erx_sent_date` timestamp NULL DEFAULT NULL,
  `erx_id` int(10) unsigned DEFAULT NULL,
  `pharmacy_id` int(10) unsigned DEFAULT NULL,
  `erx_last_filled_date` timestamp NULL DEFAULT NULL,
  `is_controlled_substance` tinyint(4) DEFAULT NULL,
  `generic_drug_name_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `erx_id` (`erx_id`),
  KEY `treatment_plan_id` (`treatment_plan_id`),
  KEY `dispense_unit_id` (`dispense_unit_id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `drug_route_id` (`drug_route_id`),
  KEY `pharmacy_id` (`pharmacy_id`),
  KEY `generic_drug_name_id` (`generic_drug_name_id`),
  CONSTRAINT `treatment_ibfk_10` FOREIGN KEY (`generic_drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `treatment_ibfk_3` FOREIGN KEY (`dispense_unit_id`) REFERENCES `dispense_unit` (`id`),
  CONSTRAINT `treatment_ibfk_5` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `treatment_ibfk_6` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `treatment_ibfk_7` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`),
  CONSTRAINT `treatment_ibfk_8` FOREIGN KEY (`pharmacy_id`) REFERENCES `pharmacy_selection` (`id`),
  CONSTRAINT `treatment_ibfk_9` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_dr_template_selection`
--

DROP TABLE IF EXISTS `treatment_dr_template_selection`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_dr_template_selection` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `treatment_id` int(10) unsigned NOT NULL,
  `dr_treatment_template_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `dr_favorite_treatment_id` (`dr_treatment_template_id`),
  KEY `treatment_id` (`treatment_id`),
  CONSTRAINT `treatment_dr_template_selection_ibfk_1` FOREIGN KEY (`dr_treatment_template_id`) REFERENCES `dr_treatment_template` (`id`),
  CONSTRAINT `treatment_dr_template_selection_ibfk_2` FOREIGN KEY (`treatment_id`) REFERENCES `treatment` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_drug_db_id`
--

DROP TABLE IF EXISTS `treatment_drug_db_id`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_drug_db_id` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_db_id_tag` varchar(100) NOT NULL,
  `drug_db_id` varchar(100) DEFAULT NULL,
  `treatment_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `treatment_id` (`treatment_id`),
  CONSTRAINT `treatment_drug_db_id_ibfk_1` FOREIGN KEY (`treatment_id`) REFERENCES `treatment` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_instructions`
--

DROP TABLE IF EXISTS `treatment_instructions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_instructions` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `treatment_id` int(10) unsigned NOT NULL,
  `dr_drug_instruction_id` int(10) unsigned NOT NULL,
  `status` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `treatment_id` (`treatment_id`),
  KEY `dr_drug_instruction_id` (`dr_drug_instruction_id`),
  CONSTRAINT `treatment_instructions_ibfk_2` FOREIGN KEY (`dr_drug_instruction_id`) REFERENCES `dr_drug_supplemental_instruction` (`id`),
  CONSTRAINT `treatment_instructions_ibfk_3` FOREIGN KEY (`treatment_id`) REFERENCES `treatment` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_plan`
--

DROP TABLE IF EXISTS `treatment_plan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_plan` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `doctor_id` int(10) unsigned DEFAULT NULL,
  `sent_date` timestamp NULL DEFAULT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `patient_case_id` int(10) unsigned DEFAULT NULL,
  `last_modified_date` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `note` text CHARACTER SET utf8mb4,
  `patient_viewed` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `patient_id` (`patient_id`),
  CONSTRAINT `treatment_plan_ibfk_2` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `treatment_plan_ibfk_3` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_plan_content_source`
--

DROP TABLE IF EXISTS `treatment_plan_content_source`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_plan_content_source` (
  `treatment_plan_id` int(10) unsigned NOT NULL,
  `content_source_id` int(10) unsigned NOT NULL,
  `content_source_type` varchar(100) NOT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  `has_deviated` tinyint(1) NOT NULL DEFAULT '0',
  `deviated_date` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`treatment_plan_id`),
  KEY `doctor_id` (`doctor_id`),
  CONSTRAINT `treatment_plan_content_source_ibfk_1` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE,
  CONSTRAINT `treatment_plan_content_source_ibfk_2` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_plan_parent`
--

DROP TABLE IF EXISTS `treatment_plan_parent`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_plan_parent` (
  `treatment_plan_id` int(10) unsigned NOT NULL,
  `parent_id` int(10) unsigned NOT NULL,
  `parent_type` varchar(100) NOT NULL,
  PRIMARY KEY (`treatment_plan_id`),
  CONSTRAINT `treatment_plan_parent_ibfk_1` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_plan_patient_visit_mapping`
--

DROP TABLE IF EXISTS `treatment_plan_patient_visit_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_plan_patient_visit_mapping` (
  `treatment_plan_id` int(10) unsigned NOT NULL,
  `patient_visit_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`treatment_plan_id`,`patient_visit_id`),
  KEY `patient_visit_id` (`patient_visit_id`),
  CONSTRAINT `treatment_plan_patient_visit_mapping_ibfk_1` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE,
  CONSTRAINT `treatment_plan_patient_visit_mapping_ibfk_2` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_plan_resource_guide`
--

DROP TABLE IF EXISTS `treatment_plan_resource_guide`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_plan_resource_guide` (
  `treatment_plan_id` int(10) unsigned NOT NULL,
  `resource_guide_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`treatment_plan_id`,`resource_guide_id`),
  KEY `resource_guide_id` (`resource_guide_id`),
  CONSTRAINT `treatment_plan_resource_guide_ibfk_1` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE,
  CONSTRAINT `treatment_plan_resource_guide_ibfk_2` FOREIGN KEY (`resource_guide_id`) REFERENCES `resource_guide` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_plan_scheduled_message`
--

DROP TABLE IF EXISTS `treatment_plan_scheduled_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_plan_scheduled_message` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `treatment_plan_id` int(10) unsigned NOT NULL,
  `scheduled_days` int(10) unsigned NOT NULL,
  `message` text NOT NULL,
  `scheduled_message_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `treatment_plan_id` (`treatment_plan_id`),
  CONSTRAINT `treatment_plan_scheduled_message_ibfk_1` FOREIGN KEY (`treatment_plan_id`) REFERENCES `treatment_plan` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `treatment_plan_scheduled_message_attachment`
--

DROP TABLE IF EXISTS `treatment_plan_scheduled_message_attachment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `treatment_plan_scheduled_message_attachment` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `treatment_plan_scheduled_message_id` bigint(20) unsigned NOT NULL,
  `item_type` varchar(64) NOT NULL,
  `item_id` bigint(20) unsigned NOT NULL,
  `title` varchar(256) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `treatment_plan_scheduled_message_id` (`treatment_plan_scheduled_message_id`),
  CONSTRAINT `treatment_plan_scheduled_message_attachment_ibfk_1` FOREIGN KEY (`treatment_plan_scheduled_message_id`) REFERENCES `treatment_plan_scheduled_message` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `unclaimed_case_queue`
--

DROP TABLE IF EXISTS `unclaimed_case_queue`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `unclaimed_case_queue` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `care_providing_state_id` int(10) unsigned NOT NULL,
  `event_type` varchar(100) NOT NULL,
  `item_id` int(10) unsigned NOT NULL,
  `patient_case_id` int(10) unsigned NOT NULL,
  `status` varchar(100) NOT NULL,
  `locked` tinyint(1) NOT NULL DEFAULT '0',
  `doctor_id` int(10) unsigned DEFAULT NULL,
  `enqueue_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `expires` timestamp(6) NULL DEFAULT NULL,
  `description` varchar(2000) NOT NULL,
  `action_url` varchar(2000) NOT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `short_description` text NOT NULL,
  `tags` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `patient_case_id` (`patient_case_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `care_providing_state_id` (`care_providing_state_id`),
  KEY `locked` (`locked`,`care_providing_state_id`),
  KEY `locked_2` (`locked`,`enqueue_date`),
  CONSTRAINT `unclaimed_case_queue_ibfk_1` FOREIGN KEY (`care_providing_state_id`) REFERENCES `care_providing_state` (`id`),
  CONSTRAINT `unclaimed_case_queue_ibfk_2` FOREIGN KEY (`patient_case_id`) REFERENCES `patient_case` (`id`),
  CONSTRAINT `unclaimed_case_queue_ibfk_3` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `unlinked_dntf_treatment`
--

DROP TABLE IF EXISTS `unlinked_dntf_treatment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `unlinked_dntf_treatment` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_internal_name` varchar(250) NOT NULL,
  `dispense_value` decimal(21,10) NOT NULL,
  `dispense_unit_id` int(10) unsigned NOT NULL,
  `refills` int(10) unsigned NOT NULL,
  `days_supply` int(10) unsigned DEFAULT NULL,
  `pharmacy_notes` varchar(250) DEFAULT NULL,
  `substitutions_allowed` tinyint(4) DEFAULT NULL,
  `patient_instructions` varchar(150) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  `dosage_strength` varchar(250) NOT NULL,
  `type` varchar(150) NOT NULL,
  `drug_name_id` int(10) unsigned DEFAULT NULL,
  `drug_form_id` int(10) unsigned DEFAULT NULL,
  `drug_route_id` int(10) unsigned DEFAULT NULL,
  `erx_sent_date` timestamp NULL DEFAULT NULL,
  `erx_id` int(10) unsigned DEFAULT NULL,
  `pharmacy_id` int(10) unsigned DEFAULT NULL,
  `erx_last_filled_date` timestamp(6) NULL DEFAULT NULL,
  `patient_id` int(10) unsigned NOT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  `is_controlled_substance` tinyint(4) DEFAULT NULL,
  `generic_drug_name_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `dispense_unit_id` (`dispense_unit_id`),
  KEY `drug_name_id` (`drug_name_id`),
  KEY `drug_form_id` (`drug_form_id`),
  KEY `drug_route_id` (`drug_route_id`),
  KEY `pharmacy_id` (`pharmacy_id`),
  KEY `patient_id` (`patient_id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `generic_drug_name_id` (`generic_drug_name_id`),
  CONSTRAINT `unlinked_dntf_treatment_ibfk_1` FOREIGN KEY (`dispense_unit_id`) REFERENCES `dispense_unit` (`id`),
  CONSTRAINT `unlinked_dntf_treatment_ibfk_2` FOREIGN KEY (`drug_name_id`) REFERENCES `drug_name` (`id`),
  CONSTRAINT `unlinked_dntf_treatment_ibfk_3` FOREIGN KEY (`drug_form_id`) REFERENCES `drug_form` (`id`),
  CONSTRAINT `unlinked_dntf_treatment_ibfk_4` FOREIGN KEY (`drug_route_id`) REFERENCES `drug_route` (`id`),
  CONSTRAINT `unlinked_dntf_treatment_ibfk_5` FOREIGN KEY (`pharmacy_id`) REFERENCES `pharmacy_selection` (`id`),
  CONSTRAINT `unlinked_dntf_treatment_ibfk_6` FOREIGN KEY (`patient_id`) REFERENCES `patient` (`id`),
  CONSTRAINT `unlinked_dntf_treatment_ibfk_7` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`),
  CONSTRAINT `unlinked_dntf_treatment_ibfk_8` FOREIGN KEY (`generic_drug_name_id`) REFERENCES `drug_name` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `unlinked_dntf_treatment_drug_db_id`
--

DROP TABLE IF EXISTS `unlinked_dntf_treatment_drug_db_id`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `unlinked_dntf_treatment_drug_db_id` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `drug_db_id` varchar(100) NOT NULL,
  `drug_db_id_tag` varchar(100) NOT NULL,
  `unlinked_dntf_treatment_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `unlinked_dntf_treatment_id` (`unlinked_dntf_treatment_id`),
  CONSTRAINT `unlinked_dntf_treatment_drug_db_id_ibfk_1` FOREIGN KEY (`unlinked_dntf_treatment_id`) REFERENCES `unlinked_dntf_treatment` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `unlinked_dntf_treatment_status_events`
--

DROP TABLE IF EXISTS `unlinked_dntf_treatment_status_events`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `unlinked_dntf_treatment_status_events` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `unlinked_dntf_treatment_id` int(10) unsigned NOT NULL,
  `erx_status` varchar(100) NOT NULL,
  `creation_date` timestamp(6) NULL DEFAULT CURRENT_TIMESTAMP(6),
  `status` varchar(100) NOT NULL,
  `event_details` varchar(500) DEFAULT NULL,
  `reported_timestamp` timestamp(6) NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `unlinked_dntf_treatment_id` (`unlinked_dntf_treatment_id`),
  CONSTRAINT `unlinked_dntf_treatment_status_events_ibfk_1` FOREIGN KEY (`unlinked_dntf_treatment_id`) REFERENCES `unlinked_dntf_treatment` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `visit_diagnosis`
--

DROP TABLE IF EXISTS `visit_diagnosis`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `visit_diagnosis` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_visit_id` int(10) unsigned NOT NULL,
  `status` varchar(32) NOT NULL,
  `diagnosis` text NOT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `patient_visit_id` (`patient_visit_id`),
  KEY `doctor_id` (`doctor_id`,`patient_visit_id`),
  CONSTRAINT `visit_diagnosis_ibfk_1` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`),
  CONSTRAINT `visit_diagnosis_ibfk_2` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `visit_diagnosis_item`
--

DROP TABLE IF EXISTS `visit_diagnosis_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `visit_diagnosis_item` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `visit_diagnosis_set_id` int(10) unsigned NOT NULL,
  `diagnosis_code_id` varchar(32) NOT NULL,
  `layout_version_id` int(10) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `visit_diagnosis_set_id` (`visit_diagnosis_set_id`),
  KEY `diagnosis_code_id` (`diagnosis_code_id`),
  KEY `layout_version_id` (`layout_version_id`),
  CONSTRAINT `visit_diagnosis_item_ibfk_1` FOREIGN KEY (`visit_diagnosis_set_id`) REFERENCES `visit_diagnosis_set` (`id`),
  CONSTRAINT `visit_diagnosis_item_ibfk_3` FOREIGN KEY (`layout_version_id`) REFERENCES `diagnosis_details_layout` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `visit_diagnosis_set`
--

DROP TABLE IF EXISTS `visit_diagnosis_set`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `visit_diagnosis_set` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `patient_visit_id` int(10) unsigned NOT NULL,
  `doctor_id` int(10) unsigned NOT NULL,
  `notes` text NOT NULL,
  `unsuitable` tinyint(1) NOT NULL,
  `unsuitable_reason` text NOT NULL,
  `active` tinyint(1) NOT NULL DEFAULT '0',
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `doctor_id` (`doctor_id`),
  KEY `patient_visit_id` (`patient_visit_id`,`active`),
  CONSTRAINT `visit_diagnosis_set_ibfk_1` FOREIGN KEY (`patient_visit_id`) REFERENCES `patient_visit` (`id`),
  CONSTRAINT `visit_diagnosis_set_ibfk_2` FOREIGN KEY (`doctor_id`) REFERENCES `doctor` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2015-07-21 15:44:44
