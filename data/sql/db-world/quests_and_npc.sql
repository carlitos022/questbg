-- NPC neutral. Este archivo puede volver a importarse sin duplicar spawns.
INSERT INTO `creature_template`
(`entry`,`name`,`subname`,`minlevel`,`maxlevel`,`faction`,`npcflag`,`unit_class`,`type`,`AIName`,`ScriptName`)
VALUES (91049,'Comandante de Battlegrounds','Misiones PvP',80,80,35,2,1,7,'','')
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`),`subname`=VALUES(`subname`),`npcflag`=2,`ScriptName`='';
INSERT IGNORE INTO `creature_template_model`
(`CreatureID`,`Idx`,`CreatureDisplayID`,`DisplayScale`,`Probability`,`VerifiedBuild`)
VALUES (91049,0,1736,1,1,NULL);
CALL `pvp_bg_quests_sync`();
INSERT INTO `creature`
(`id`,`map`,`zoneId`,`areaId`,`spawnMask`,`phaseMask`,`equipment_id`,`position_x`,`position_y`,`position_z`,`orientation`,
 `spawntimesecs`,`wander_distance`,`currentwaypoint`,`curhealth`,`curmana`,`MovementType`,`npcflag`,`unit_flags`,`dynamicflags`,
 `ScriptName`,`VerifiedBuild`,`CreateObject`,`Comment`)
SELECT 91049,0,0,0,1,1,0,-8833.38,628.62,94.00,0.80,300,0,0,10000,0,0,0,0,0,'',NULL,0,'PvP BG Quest Master - Stormwind'
WHERE NOT EXISTS (SELECT 1 FROM `creature` WHERE `id`=91049 AND `Comment`='PvP BG Quest Master - Stormwind');
INSERT INTO `creature`
(`id`,`map`,`zoneId`,`areaId`,`spawnMask`,`phaseMask`,`equipment_id`,`position_x`,`position_y`,`position_z`,`orientation`,
 `spawntimesecs`,`wander_distance`,`currentwaypoint`,`curhealth`,`curmana`,`MovementType`,`npcflag`,`unit_flags`,`dynamicflags`,
 `ScriptName`,`VerifiedBuild`,`CreateObject`,`Comment`)
SELECT 91049,1,0,0,1,1,0,1601.20,-4378.70,9.98,2.80,300,0,0,10000,0,0,0,0,0,'',NULL,0,'PvP BG Quest Master - Orgrimmar'
WHERE NOT EXISTS (SELECT 1 FROM `creature` WHERE `id`=91049 AND `Comment`='PvP BG Quest Master - Orgrimmar');
