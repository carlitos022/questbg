-- Configuracion propia. Los INSERT IGNORE conservan los cambios del operador.
CREATE TABLE IF NOT EXISTS `pvp_bg_quest_settings` (
 `id` TINYINT UNSIGNED NOT NULL PRIMARY KEY,
 `npc_entry` INT UNSIGNED NOT NULL DEFAULT 91049,
 `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
 `sync_revision` BIGINT UNSIGNED NOT NULL DEFAULT 0,
 CHECK (`id`=1), CHECK (`enabled` IN (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
INSERT IGNORE INTO `pvp_bg_quest_settings` (`id`) VALUES (1);

CREATE TABLE IF NOT EXISTS `pvp_bg_quest_config` (
 `quest_id` INT UNSIGNED NOT NULL PRIMARY KEY,
 `objective_type` TINYINT UNSIGNED NOT NULL COMMENT '1=victorias, 2=bajas PvP',
 `battleground_type` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=cualquiera; 1=AV,2=WSG,3=AB,7=EotS,9=SotA,30=IoC',
 `required_count` SMALLINT UNSIGNED NOT NULL DEFAULT 1,
 `credit_entry` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'Victoria: 0 utiliza quest_id como credito exclusivo',
 `min_level` TINYINT UNSIGNED NOT NULL DEFAULT 80,
 `repeat_mode` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0=una vez,1=diaria,2=semanal',
 `require_opposite_faction` TINYINT UNSIGNED NOT NULL DEFAULT 1,
 `block_same_account` TINYINT UNSIGNED NOT NULL DEFAULT 1,
 `victim_cooldown_seconds` INT UNSIGNED NOT NULL DEFAULT 60,
 `victim_credit_limit` SMALLINT UNSIGNED NOT NULL DEFAULT 3 COMMENT '0=sin limite; por victima, mision y partida',
 `min_win_participation_seconds` INT UNSIGNED NOT NULL DEFAULT 0,
 `enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1,
 `title` VARCHAR(255) NOT NULL,
 `log_description` TEXT NOT NULL,
 `description` TEXT NOT NULL,
 `objective_text` VARCHAR(255) NOT NULL,
 CHECK (`objective_type` IN (1,2)),
 CHECK (`battleground_type` IN (0,1,2,3,7,9,30)),
 CHECK (`required_count` BETWEEN 1 AND 65535),
 CHECK (`objective_type`=1 OR `required_count`<=255),
 CHECK (`min_level` BETWEEN 1 AND 80),
 CHECK (`repeat_mode` IN (0,1,2)),
 CHECK (`enabled` IN (0,1)),
 CHECK (`require_opposite_faction` IN (0,1)),
 CHECK (`block_same_account` IN (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `pvp_bg_quest_rewards` (
 `quest_id` INT UNSIGNED NOT NULL PRIMARY KEY,
 `honor_points` INT UNSIGNED NOT NULL DEFAULT 0,
 `arena_points` SMALLINT UNSIGNED NOT NULL DEFAULT 0,
 `item1` INT UNSIGNED NOT NULL DEFAULT 0,
 `item1_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0,
 `item2` INT UNSIGNED NOT NULL DEFAULT 0,
 `item2_count` SMALLINT UNSIGNED NOT NULL DEFAULT 0,
 `frost_emblems` SMALLINT UNSIGNED NOT NULL DEFAULT 0,
 CONSTRAINT `fk_pvp_bg_quest_rewards_config` FOREIGN KEY (`quest_id`)
 REFERENCES `pvp_bg_quest_config`(`quest_id`) ON DELETE RESTRICT,
 CHECK (`honor_points`<=2147483647),
 CHECK ((`item1`=0 AND `item1_count`=0) OR (`item1`>0 AND `item1_count`>0)),
 CHECK ((`item2`=0 AND `item2_count`=0) OR (`item2`>0 AND `item2_count`>0))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `pvp_bg_quest_managed` (
 `quest_id` INT UNSIGNED NOT NULL PRIMARY KEY,
 `credit_entry` INT UNSIGNED NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO `pvp_bg_quest_config`
(`quest_id`,`objective_type`,`battleground_type`,`required_count`,`title`,`log_description`,`description`,`objective_text`)
VALUES
(91050,2,0,50,'Dominio de los Campos de Batalla','Mata 50 jugadores enemigos dentro de cualquier Battleground.','Derrota a 50 jugadores de la faccion contraria dentro de Campos de Batalla. No cuentan arenas ni bajas en el mundo abierto.','Jugadores enemigos derrotados'),
(91051,2,7,50,'Carniceria en Ojo de la Tormenta','Mata 50 jugadores enemigos exclusivamente en Ojo de la Tormenta.','Derrota a 50 jugadores de la faccion contraria en Ojo de la Tormenta. Las bajas de otras BG no cuentan.','Enemigos derrotados en Ojo de la Tormenta'),
(91052,1,7,1,'Victoria: Ojo de la Tormenta','Gana una partida de Ojo de la Tormenta.','Participa y consigue una victoria para tu equipo.','Gana Ojo de la Tormenta'),
(91053,1,2,1,'Victoria: Garganta Grito de Guerra','Gana una partida de Garganta Grito de Guerra.','Participa y consigue una victoria para tu equipo.','Gana Garganta Grito de Guerra'),
(91054,1,3,1,'Victoria: Cuenca de Arathi','Gana una partida de Cuenca de Arathi.','Participa y consigue una victoria para tu equipo.','Gana Cuenca de Arathi'),
(91055,1,1,1,'Victoria: Valle de Alterac','Gana una partida de Valle de Alterac.','Participa y consigue una victoria para tu equipo.','Gana Valle de Alterac'),
(91056,1,9,1,'Victoria: Playa de los Ancestros','Gana una partida de Playa de los Ancestros.','Participa y consigue una victoria para tu equipo.','Gana Playa de los Ancestros'),
(91057,1,30,1,'Victoria: Isla de la Conquista','Gana una partida de Isla de la Conquista.','Participa y consigue una victoria para tu equipo.','Gana Isla de la Conquista'),
(91058,1,0,1,'Victoria: Campos de Batalla','Gana una partida en cualquier Battleground.','Consigue una victoria para tu equipo en cualquier Campo de Batalla.','Gana un Campo de Batalla');

INSERT IGNORE INTO `pvp_bg_quest_rewards` (`quest_id`,`honor_points`,`arena_points`,`item1`,`item1_count`,`item2`,`item2_count`,`frost_emblems`)
VALUES (91050,5000,50,40093,2,33448,3,3),(91051,6500,75,40093,2,33448,3,5),(91052,7500,100,40093,2,33448,3,5),
(91053,5000,50,40093,2,33448,3,3),(91054,5000,50,40093,2,33448,3,3),(91055,5000,50,40093,2,33448,3,3),
(91056,5000,50,40093,2,33448,3,3),(91057,5000,50,40093,2,33448,3,3),(91058,5000,50,40093,2,33448,3,3);

-- Todas las recompensas se reflejan en las quests nativas: el core controla el turn-in.
DROP PROCEDURE IF EXISTS `pvp_bg_quests_sync`;
DELIMITER //
CREATE PROCEDURE `pvp_bg_quests_sync`()
BEGIN
 DECLARE `v_npc` INT UNSIGNED;
 DECLARE `v_enabled` TINYINT UNSIGNED;
 DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
 SELECT `npc_entry`,`enabled` INTO `v_npc`,`v_enabled` FROM `pvp_bg_quest_settings` WHERE `id`=1;
 IF `v_npc` IS NULL OR NOT EXISTS (SELECT 1 FROM `creature_template` WHERE `entry`=`v_npc` AND (`npcflag` & 2)=2) THEN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='PvPBGQ: npc_entry no existe o no es questgiver';
 END IF;
 IF EXISTS (SELECT 1 FROM `pvp_bg_quest_config` `q` JOIN `quest_template` `t` ON `t`.`ID`=`q`.`quest_id`
  LEFT JOIN `pvp_bg_quest_managed` `m` ON `m`.`quest_id`=`q`.`quest_id` WHERE `m`.`quest_id` IS NULL) THEN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='PvPBGQ: ID de quest ocupado fuera del modulo';
 END IF;
 IF EXISTS (SELECT 1 FROM `pvp_bg_quest_rewards` `r`
  WHERE (`r`.`item1`<>0 AND NOT EXISTS (SELECT 1 FROM `item_template` WHERE `entry`=`r`.`item1`))
     OR (`r`.`item2`<>0 AND NOT EXISTS (SELECT 1 FROM `item_template` WHERE `entry`=`r`.`item2`))
     OR (`r`.`frost_emblems`>0 AND NOT EXISTS (SELECT 1 FROM `item_template` WHERE `entry`=49426))) THEN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='PvPBGQ: una recompensa referencia un item inexistente';
 END IF;
 IF EXISTS (SELECT 1 FROM `pvp_bg_quest_config` `q` JOIN `creature_template` `t`
  ON `t`.`entry`=IF(`q`.`credit_entry`=0,`q`.`quest_id`,`q`.`credit_entry`)
  WHERE `q`.`objective_type`=1 AND `t`.`subname`<>'PvPBGQ credit') THEN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='PvPBGQ: entry de credito ocupado fuera del modulo';
 END IF;
 IF EXISTS (SELECT IF(`credit_entry`=0,`quest_id`,`credit_entry`) FROM `pvp_bg_quest_config`
  WHERE `objective_type`=1 GROUP BY IF(`credit_entry`=0,`quest_id`,`credit_entry`) HAVING COUNT(*)>1) THEN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='PvPBGQ: dos victorias comparten credit_entry';
 END IF;
 START TRANSACTION;
 INSERT INTO `creature_template`
 (`entry`,`name`,`subname`,`minlevel`,`maxlevel`,`faction`,`npcflag`,`unit_flags`,`unit_class`,`type`,`AIName`,`ScriptName`)
 SELECT IF(`credit_entry`=0,`quest_id`,`credit_entry`),`objective_text`,'PvPBGQ credit',80,80,35,0,33554434,1,7,'',''
 FROM `pvp_bg_quest_config` WHERE `objective_type`=1
 ON DUPLICATE KEY UPDATE `name`=VALUES(`name`);
 INSERT IGNORE INTO `creature_template_model`
 (`CreatureID`,`Idx`,`CreatureDisplayID`,`DisplayScale`,`Probability`,`VerifiedBuild`)
 SELECT IF(`credit_entry`=0,`quest_id`,`credit_entry`),0,11686,1,1,NULL
 FROM `pvp_bg_quest_config` WHERE `objective_type`=1;
 INSERT INTO `quest_template`
 (`ID`,`QuestType`,`QuestLevel`,`MinLevel`,`QuestSortID`,`QuestInfoID`,`Flags`,`RequiredPlayerKills`,
  `LogTitle`,`LogDescription`,`QuestDescription`,`AreaDescription`,`QuestCompletionLog`,
  `RequiredNpcOrGo1`,`RequiredNpcOrGoCount1`,`ObjectiveText1`,
  `RewardHonor`,`RewardKillHonor`,`RewardArenaPoints`,
  `RewardItem1`,`RewardAmount1`,`RewardItem2`,`RewardAmount2`,`RewardItem3`,`RewardAmount3`,`RewardItem4`,`RewardAmount4`)
 SELECT `q`.`quest_id`,2,80,`q`.`min_level`,-25,41,
 IF(`q`.`repeat_mode`=1,4096,IF(`q`.`repeat_mode`=2,32768,0)),
 IF(`q`.`objective_type`=2,`q`.`required_count`,0),
 `q`.`title`,`q`.`log_description`,`q`.`description`,`q`.`objective_text`,'Objetivo cumplido. Regresa al Comandante de Battlegrounds.',
 IF(`q`.`objective_type`=1,IF(`q`.`credit_entry`=0,`q`.`quest_id`,`q`.`credit_entry`),0),
 IF(`q`.`objective_type`=1,`q`.`required_count`,0),`q`.`objective_text`,
 COALESCE(`r`.`honor_points`,0),0,COALESCE(`r`.`arena_points`,0),
 CASE WHEN COALESCE(`r`.`item1`,0)>0 THEN `r`.`item1` WHEN COALESCE(`r`.`item2`,0)>0 THEN `r`.`item2` WHEN COALESCE(`r`.`frost_emblems`,0)>0 THEN 49426 ELSE 0 END,
 CASE WHEN COALESCE(`r`.`item1`,0)>0 THEN `r`.`item1_count` WHEN COALESCE(`r`.`item2`,0)>0 THEN `r`.`item2_count` ELSE COALESCE(`r`.`frost_emblems`,0) END,
 CASE WHEN COALESCE(`r`.`item1`,0)>0 AND COALESCE(`r`.`item2`,0)>0 THEN `r`.`item2` WHEN (COALESCE(`r`.`item1`,0)>0 OR COALESCE(`r`.`item2`,0)>0) AND COALESCE(`r`.`frost_emblems`,0)>0 THEN 49426 ELSE 0 END,
 CASE WHEN COALESCE(`r`.`item1`,0)>0 AND COALESCE(`r`.`item2`,0)>0 THEN `r`.`item2_count` WHEN COALESCE(`r`.`item1`,0)>0 OR COALESCE(`r`.`item2`,0)>0 THEN COALESCE(`r`.`frost_emblems`,0) ELSE 0 END,
 IF(COALESCE(`r`.`item1`,0)>0 AND COALESCE(`r`.`item2`,0)>0 AND COALESCE(`r`.`frost_emblems`,0)>0,49426,0),
 IF(COALESCE(`r`.`item1`,0)>0 AND COALESCE(`r`.`item2`,0)>0,COALESCE(`r`.`frost_emblems`,0),0),0,0
 FROM `pvp_bg_quest_config` `q` LEFT JOIN `pvp_bg_quest_rewards` `r` ON `r`.`quest_id`=`q`.`quest_id`
 ON DUPLICATE KEY UPDATE
 `QuestType`=VALUES(`QuestType`),`QuestLevel`=VALUES(`QuestLevel`),`MinLevel`=VALUES(`MinLevel`),
 `QuestSortID`=VALUES(`QuestSortID`),`QuestInfoID`=VALUES(`QuestInfoID`),`Flags`=VALUES(`Flags`),
 `RequiredPlayerKills`=VALUES(`RequiredPlayerKills`),`LogTitle`=VALUES(`LogTitle`),`LogDescription`=VALUES(`LogDescription`),
 `QuestDescription`=VALUES(`QuestDescription`),`AreaDescription`=VALUES(`AreaDescription`),
 `QuestCompletionLog`=VALUES(`QuestCompletionLog`),`RequiredNpcOrGo1`=VALUES(`RequiredNpcOrGo1`),
 `RequiredNpcOrGoCount1`=VALUES(`RequiredNpcOrGoCount1`),`ObjectiveText1`=VALUES(`ObjectiveText1`),
 `RewardHonor`=VALUES(`RewardHonor`),`RewardKillHonor`=0,`RewardArenaPoints`=VALUES(`RewardArenaPoints`),
 `RewardItem1`=VALUES(`RewardItem1`),`RewardAmount1`=VALUES(`RewardAmount1`),
 `RewardItem2`=VALUES(`RewardItem2`),`RewardAmount2`=VALUES(`RewardAmount2`),
 `RewardItem3`=VALUES(`RewardItem3`),`RewardAmount3`=VALUES(`RewardAmount3`),
 `RewardItem4`=0,`RewardAmount4`=0;
 INSERT INTO `quest_template_addon` (`ID`,`SpecialFlags`)
 SELECT `quest_id`,IF(`repeat_mode`=0,0,1) FROM `pvp_bg_quest_config`
 ON DUPLICATE KEY UPDATE `SpecialFlags`=VALUES(`SpecialFlags`);
 INSERT INTO `quest_offer_reward` (`ID`,`RewardText`)
 SELECT `quest_id`,'Excelente. Aqui tienes tu recompensa.' FROM `pvp_bg_quest_config`
 ON DUPLICATE KEY UPDATE `RewardText`=VALUES(`RewardText`);
 INSERT INTO `quest_request_items` (`ID`,`CompletionText`)
 SELECT `quest_id`,`log_description` FROM `pvp_bg_quest_config`
 ON DUPLICATE KEY UPDATE `CompletionText`=VALUES(`CompletionText`);
 INSERT INTO `pvp_bg_quest_managed` (`quest_id`,`credit_entry`)
 SELECT `quest_id`,IF(`objective_type`=1,IF(`credit_entry`=0,`quest_id`,`credit_entry`),0) FROM `pvp_bg_quest_config`
 ON DUPLICATE KEY UPDATE `credit_entry`=VALUES(`credit_entry`);
 DELETE FROM `creature_queststarter` WHERE `quest` IN (SELECT `quest_id` FROM `pvp_bg_quest_managed`);
 DELETE FROM `creature_questender` WHERE `quest` IN (SELECT `quest_id` FROM `pvp_bg_quest_managed`);
 INSERT INTO `creature_queststarter` (`id`,`quest`)
 SELECT `v_npc`,`quest_id` FROM `pvp_bg_quest_config` WHERE `enabled`=1 AND `v_enabled`=1;
 INSERT INTO `creature_questender` (`id`,`quest`)
 SELECT `v_npc`,`quest_id` FROM `pvp_bg_quest_config` WHERE `enabled`=1 AND `v_enabled`=1;
 UPDATE `pvp_bg_quest_settings` SET `sync_revision`=`sync_revision`+1 WHERE `id`=1;
 COMMIT;
END//
DELIMITER ;
