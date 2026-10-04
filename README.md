# questbg

Modulo instalado como mod-pvp-battleground-quests. Repositorio: https://github.com/carlitos022/questbg

La [GUIA_DE_USO.md](GUIA_DE_USO.md) detalla cada mision, sus recompensas, CrossFaction BG y la verificacion de puntos de arena.

Motor de misiones para AzerothCore 3.3.5a. Codigo independiente en modules; no modifica archivos del core.

## Misiones y NPC

NPC neutral 91049, Comandante de Battlegrounds, con spawns en Ventormenta y Orgrimmar. Ambas facciones pueden aceptar y entregar sus misiones mediante el sistema nativo.

| Quest | Objetivo |
|---|---|
| 91050 | 50 bajas en cualquier BG |
| 91051 | 50 bajas exclusivamente en Ojo de la Tormenta |
| 91052 | Ganar Ojo de la Tormenta |
| 91053 | Ganar Garganta Grito de Guerra |
| 91054 | Ganar Cuenca de Arathi |
| 91055 | Ganar Valle de Alterac |
| 91056 | Ganar Playa de los Ancestros |
| 91057 | Ganar Isla de la Conquista |
| 91058 | Ganar cualquier BG |

## Configuracion sin recompilar

Fuente de objetivos: pvp_bg_quest_config. Fuente de recompensas: pvp_bg_quest_rewards, enlazada por quest_id. pvp_bg_quest_settings controla enabled y npc_entry. El procedimiento pvp_bg_quests_sync proyecta la configuracion en tablas nativas y protege IDs ocupados fuera del modulo.

Ejemplo para modificar la cantidad y las recompensas:

```sql
UPDATE pvp_bg_quest_config SET required_count=50 WHERE quest_id=91051;
UPDATE pvp_bg_quest_rewards
SET honor_points=6500, arena_points=75, frost_emblems=5,
    item1=40093, item1_count=2, item2=33448, item2_count=3
WHERE quest_id=91051;
```

Ejecutar dentro del juego: .pvpbgq reload. .pvpbgq status informa revision y numero de reglas. El servidor recarga objetivos, recompensas, relaciones del NPC y templates de credito. El cliente puede conservar textos en su cache de quests; borrar esa cache si muestra textos anteriores.

Las nueve misiones incluyen dos items seleccionados al azar entre IDs existentes: 40093 (Pocion indestructible), cantidad 2; 33448 (Pocion de mana runica), cantidad 3. La seleccion se hace durante la configuracion; cada entrega usa esas cantidades fijas. Frost usa item 49426. Se compactan los items automaticamente en slots consecutivos, incluido frost cuando no hay otros items.

objective_type: 1=victorias, 2=bajas. battleground_type: 0=cualquier BG; 1=AV, 2=WSG, 3=AB, 7=EotS, 9=SotA, 30=IoC. required_count de bajas: 1..255, limite del campo nativo. min_level: 1..80. repeat_mode: 0=una vez, 1=diaria, 2=semanal. Las misiones iniciales son de una sola entrega.

Para nuevas victorias usar credit_entry=0: el ID de la propia quest se usa como criatura de credito exclusiva. Cada victoria necesita un credito diferente. Agregar una fila de configuracion y su fila de recompensas y ejecutar reload.

## Progreso y recompensas

Las bajas requieren una partida activa, misma instancia y mapa, jugadores diferentes y equipos enemigos. No cuentan mundo abierto, arenas, preparacion, suicidios ni criaturas. require_opposite_faction y block_same_account permiten restricciones adicionales.

## Compatibilidad con CrossFaction BG

Los equipos enemigos se comprueban siempre mediante GetBgTeamId; una baja contra un jugador del mismo equipo de BG nunca cuenta. Las victorias tambien usan el equipo asignado a la BG, antes de que CFBG restaure la faccion al salir.

require_opposite_faction=1 (valor inicial) agrega el requisito de facciones originales distintas mediante GetTeamId(true). Con este filtro, dos personajes Alianza enfrentados por CFBG no dan credito. Para contar cualquier enemigo del equipo contrario en CFBG, usar require_opposite_faction=0 y ejecutar .pvpbgq reload. Este cambio conserva la comprobacion obligatoria de equipos enemigos, instancia y partida activa. El antifarming usa GUIDs, por lo que cambiar de raza o faccion aparente no reinicia sus limites.

```sql
UPDATE pvp_bg_quest_config
SET require_opposite_faction=0
WHERE objective_type=2;
```


victim_cooldown_seconds=60 y victim_credit_limit=3 por defecto; limite por victima, quest y partida. Cero desactiva cada restriccion. El estado se mantiene en memoria entre reloads y reconexiones y se limpia al destruir la BG. Un reinicio del proceso reinicia esa memoria. min_win_participation_seconds configura la estancia minima antes de una victoria.

Las victorias usan el equipo ganador nativo y creditos exclusivos por quest; una victoria de Ojo no completa la mision de Warsong. Las recompensas se entregan exclusivamente al devolver la quest al NPC. Honor, arena points e items usan RewardQuest nativo, con controles de inventario, repetibilidad y persistencia del core. OnPlayerCompleteQuest solo registra auditoria; no vuelve a otorgar recompensas.

## Instalacion

1. Colocar el modulo en modules/mod-pvp-battleground-quests.
2. Configurar CMake y compilar worldserver.
3. Importar data/sql/db-world/base.sql y despues quests_and_npc.sql en acore_world.
4. Arrancar worldserver y comprobar el mensaje Loaded 9 rules.
5. Usar .pvpbgq status y comprobar ambas ubicaciones del NPC.

No se necesita una tabla de recompensas en acore_characters. El historico de entregas, inventario y progreso utilizan tablas nativas.

## Pruebas

Suite del cliente real AzerothGhost en e2e/local/legends/pvp_bg_quests; copia de la suite en tests. Ejecutar desde e2e con el entorno de pruebas configurado:

```powershell
.\run_legends_e2e.ps1 -Package ./local/legends/pvp_bg_quests -Runtime med -Timeout 20m
```

La prueba de recompensas usa una quest completada por GM exclusivamente como fixture de turn-in. La prueba de BG entra por los paquetes normales de cola e invitacion, aplica dano por el camino nativo permitido por el harness y obtiene la victoria con capturas reales. Los cambios temporales de configuracion se restauran al terminar.