# Rescate y pruebas PvP/BG — 2026-11-10

Fuente: /home/azerothcore/modules/mod-pvp-battleground-quests, VPS Astravane.

## Pruebas documentadas (históricas)
- 2026-10-04: E2E AzerothGhost con NPC neutral, nueve quests 91050–91058, reglas de bajas, recompensas nativas, anti-farming y CFBG. Ver VERIFICATION.md.
- El E2E inicial documenta victorias reales en Eye of the Storm, bajas en Warsong y validación de facciones; no demuestra victorias en todas las BG.
- 2026-11-08: TestPvPBGQ_QueueMapCoverage terminó FAIL tras un timeout al intentar dejar la BG. Evidencia original: /home/astravane-bg-lab/logs/pvpbgq-map-coverage-20261108.log.

## Pendiente
Completar entradas, victorias, entregas y limpieza en AV, AB, WSG, SotA, IoC y reproducir con Playerbots. No marcar fallos como pruebas aprobadas.

## Integridad
Los archivos de código y pruebas del VPS se cotejaron por blob SHA Git; el repositorio conserva GUIA_DE_USO.md y otros documentos adicionales de GitHub. Los cambios del source AzerothCore y módulos adicionales están en el respaldo privado de carlitos022/endgame, rama backup/astravane-2026-11-10.
