# Guia de uso de questbg

## Que instala
Motor PvP independiente para AzerothCore 3.3.5a, NPC neutral 91049 y nueve misiones.
El modulo se instala con el nombre de carpeta mod-pvp-battleground-quests.
En Carlitos ya esta compilado e instalado. No necesitas volver a importar SQL para jugar.

## Uso por el jugador
1. Llegar a nivel 80.
2. Buscar al Comandante de Battlegrounds en Ventormenta o en Orgrimmar. Ambos spawns son neutrales.
3. Hablar con el NPC y aceptar las misiones que interesen antes de entrar a la BG.
4. Entrar por la cola normal a un Campo de Batalla.
5. Cumplir el objetivo. El servidor actualiza el registro de misiones; no entrega la recompensa al completar el objetivo.
6. Salir de la BG y volver al NPC para entregar cada mision completada.
7. Comprobar las pociones y emblemas en las bolsas, y el honor y los puntos de arena en los saldos PvP del personaje.
Si el inventario no tiene espacio, liberar espacio y volver a entregar. La entrega usa las comprobaciones nativas del core.

Actualmente cada mision se puede entregar una sola vez por personaje: repeat_mode=0.
Las bajas acumuladas no se reinician al terminar una partida. Las 50 bajas pueden conseguirse en varias BG mientras se conserve la mision.

## Misiones y resultados
Todas entregan adicionalmente 2 Pociones indestructibles (item 40093) y 3 Pociones de mana runicas (item 33448).
Los IDs se eligieron al azar durante la configuracion; no se sortea un objeto diferente en cada entrega.

| ID | Nombre | Objetivo y cuando se completa | Honor | Arena points | Emblemas de Escarcha |
|---|---|---|---:|---:|---:|
| 91050 | Dominio de los Campos de Batalla | Acumular 50 bajas validas en cualquiera de las seis BG admitidas. Completa al llegar a 50/50. | 5000 | 50 | 3 |
| 91051 | Carniceria en Ojo de la Tormenta | Acumular 50 bajas validas exclusivamente en Ojo. Otras BG no avanzan esta mision. | 6500 | 75 | 5 |
| 91052 | Victoria: Ojo de la Tormenta | Estar en el equipo ganador cuando el core cierre oficialmente una partida de Ojo. | 7500 | 100 | 5 |
| 91053 | Victoria: Garganta Grito de Guerra | Ganar una partida de Garganta. Una victoria en otra BG no cuenta. | 5000 | 50 | 3 |
| 91054 | Victoria: Cuenca de Arathi | Ganar una partida de Arathi. Una victoria en otra BG no cuenta. | 5000 | 50 | 3 |
| 91055 | Victoria: Valle de Alterac | Ganar una partida de Alterac. Una victoria en otra BG no cuenta. | 5000 | 50 | 3 |
| 91056 | Victoria: Playa de los Ancestros | Ganar una partida de Playa. Una victoria en otra BG no cuenta. | 5000 | 50 | 3 |
| 91057 | Victoria: Isla de la Conquista | Ganar una partida de Isla. Una victoria en otra BG no cuenta. | 5000 | 50 | 3 |
| 91058 | Victoria: Campos de Batalla | Ganar una partida en cualquiera de las seis BG admitidas. | 5000 | 50 | 3 |

Los emblemas usan el item 49426.
Las seis BG admitidas son Alterac, Garganta, Arathi, Ojo, Playa e Isla. Mundo abierto, Wintergrasp y arenas no son objetivos de estas misiones.

## Como se combinan los objetivos
Si aceptas 91050 y 91051, cada baja valida en Ojo avanza ambas. Una baja valida en Garganta avanza solamente 91050.
Si aceptas 91052 y 91058, ganar Ojo completa ambas; se entregan separadamente y cada una concede su recompensa.
Lo mismo ocurre entre 91058 y la mision especifica de cualquier otra BG.
Completar las dos misiones de 50 bajas no exige ganar la partida. Ganar una BG no exige conseguir 50 bajas.
Perder una BG no completa las misiones de victoria, pero no borra bajas ya obtenidas.
Las misiones de victoria se completan en el cierre oficial, no al capturar una bandera o torre individual.
La mision se debe aceptar antes del evento que da credito; no hay credito retroactivo.

## Logica de bajas y antifarming
Una baja debe llegar al evento PvP nativo con asesino y victima jugadores distintos, misma BG, misma instancia y mapa, equipos asignados enemigos y partida iniciada.
No cuentan criaturas, suicidios, arenas, mundo abierto, preparacion ni jugadores del propio equipo de BG.
block_same_account=1 bloquea las bajas entre personajes de la misma cuenta.

Valores actuales:
- victim_cooldown_seconds=60: la misma victima no da otro credito para esa mision hasta transcurrir 60 segundos desde el ultimo credito.
- victim_credit_limit=3: como maximo tres creditos de una misma victima por mision y partida.
- Cero desactiva la restriccion correspondiente.
- El control usa GUIDs del asesino y la victima, junto con quest e instancia. Cambiar apariencia o equipo no crea una victima nueva.
- Recargar el modulo o reconectar no limpia el estado de antifarming de una partida existente.
- Destruir la BG limpia el estado de esa instancia. Reiniciar worldserver reinicia ese estado en memoria.
No puedes completar las 50 bajas repitiendo indefinidamente contra una sola victima con la configuracion actual.
La prueba automatizada desactivo temporalmente estos limites para completar el recorrido de 50 bajas despues de haber probado sus bloqueos; luego los restauro a 60/3.

## Logica de victorias
El modulo toma el equipo ganador nativo y compara el equipo asignado al jugador dentro de la BG.
Solo da credito a las misiones aceptadas y pendientes que correspondan a esa BG o a cualquier BG.
Cada mision de victoria tiene un credito exclusivo: ganar Ojo no completa Arathi ni Garganta.
min_win_participation_seconds=0 no agrega un tiempo minimo propio. Un valor mayor exige haber permanecido ese tiempo en la BG antes de la victoria.
El personaje tiene que seguir participando cuando el core procesa la recompensa final; salir antes no produce una victoria retroactiva.
La entrega al NPC es independiente de la victoria: completar el objetivo y cobrar son pasos distintos.

## CrossFaction BG
CFBG esta activado en Carlitos.
GetBgTeamId exige equipos enemigos dentro de la BG, incluso cuando dos personajes eran originalmente de la misma faccion.
require_opposite_faction=1, actualmente configurado, agrega otra condicion: sus facciones originales deben ser distintas.
Por eso dos personajes originalmente Alianza, enfrentados por CFBG, no dan credito de bajas con el valor actual.
require_opposite_faction=0 permite que esos enemigos de CFBG den credito; los companeros del mismo equipo siguen excluidos.
Este filtro adicional afecta a las bajas. La victoria se decide por el equipo asignado, no por la faccion original.

Para habilitar el credito entre enemigos de la misma faccion original:
```sql
UPDATE pvp_bg_quest_config
SET require_opposite_faction=0
WHERE objective_type=2;
```
Luego usar .pvpbgq reload en el juego con una cuenta GM.
Se probo en vivo con dos personajes originalmente Alianza: filtro 1 bloqueo y filtro 0 dio un credito. Se restauro 1 al terminar.
No se ha ejecutado una victoria real con un ganador cambiado de faccion por CFBG; ese camino se reviso en el source.

## Verificacion de puntos de arena
No se perdieron. El personaje Pveleadkfmiq (GUID 481) conserva arenaPoints=225:
- Mision 91050: +50.
- Mision 91051: +75.
- Mision 91052: +100.
- Total: 225, con una entrega registrada de cada mision.

La mision 91058 tambien quedo completada (status=1, credito de victoria=1), pero no se entrego al NPC durante esa prueba. Sus 50 puntos de arena siguen pendientes de la entrega. Si se devuelve y no hay otros cambios en el saldo, pasaria de 225 a 275; no se ha ejecutado esa cuarta entrega.

Los valores arena_points de la configuracion coinciden con RewardArenaPoints en quest_template.
Player::RewardQuest llama a ModifyArenaPoints cuando la quest tiene esa recompensa.
SetArenaPoints actualiza PLAYER_FIELD_ARENA_CURRENCY y registra la moneda conocida del personaje.
Los puntos no son un objeto en las bolsas y no requieren jugar una partida de arena para cobrar esta recompensa.
Se acreditan al entregar la mision al NPC. No se acreditan solamente por dejar el objetivo como completado.
La prueba tambien verifico que repetir el paquete de entrega y reconectar no vuelve a pagar.

MaxArenaPoints de Carlitos es 10000. El core limita el saldo a ese tope; una entrega que lo exceda no puede dejar un saldo mayor.
Se comprobaron entregas por debajo del tope. No se ha hecho una prueba especifica de entrega estando ya en 10000.
El honor y los puntos de arena usan los saldos nativos; los emblemas y pociones se guardan como items.

Para consultar un personaje concreto en characters:
```sql
SELECT guid,name,arenaPoints,totalHonorPoints
FROM characters
WHERE name='Pveleadkfmiq';
```

## Administracion y cambios
Comandos GM:
- .pvpbgq status: muestra estado, reglas cargadas y revision.
- .pvpbgq reload: sincroniza la configuracion y recarga las quests, recompensas y relaciones del NPC.

Ejemplo: cambiar recompensas de 91051 en world:
```sql
UPDATE pvp_bg_quest_rewards
SET honor_points=6500,arena_points=75,
    item1=40093,item1_count=2,
    item2=33448,item2_count=3,frost_emblems=5
WHERE quest_id=91051;
```
Ejecutar .pvpbgq reload despues del cambio.

repeat_mode: 0=una sola entrega, 1=diaria, 2=semanal. Los resets y la elegibilidad posterior los controla el core.
enabled de pvp_bg_quest_config activa una regla; enabled de pvp_bg_quest_settings controla el motor.
battleground_type: 0=cualquiera; 1=Alterac, 2=Garganta, 3=Arathi, 7=Ojo, 9=Playa, 30=Isla.
objective_type: 1=victoria, 2=bajas.
required_count de bajas admite 1..255 por el campo nativo.
Para agregar una nueva mision crear una fila de configuracion y su fila de recompensas con un quest_id disponible, y ejecutar reload.
Para victorias usar credit_entry=0 y un quest_id exclusivo; el procedimiento protege colisiones con contenido ajeno al modulo.

El motor no concede un segundo lote manual de recompensas. OnPlayerCompleteQuest registra auditoria; RewardQuest nativo controla items, monedas, inventario, historial y repetibilidad.
La compactacion de slots permite frost solo, un item, dos items o ninguna recompensa de item sin dejar huecos invalidos.
Si el cliente muestra textos antiguos despues de un cambio, revisar su cache de quests; el estado y los pagos del servidor se comprueban con datos nativos.

## Instalacion en otro servidor
```bash
git clone https://github.com/carlitos022/questbg.git modules/mod-pvp-battleground-quests
```
Configurar CMake y compilar worldserver; importar primero data/sql/db-world/base.sql y despues quests_and_npc.sql en la base world.
Arrancar worldserver, comprobar Loaded 9 rules y usar .pvpbgq status.
Los SQL iniciales conservan la configuracion existente mediante INSERT IGNORE y evitan spawns duplicados.
No se importa una tabla custom de entregas en characters: se utilizan inventario, progreso e historial nativos.

## Cobertura ejecutada
Pasaron cinco pruebas funcionales: NPC y recompensas para Alianza/Horda sin duplicados, compactacion SQL, 50 bajas y victoria real de Ojo con pagos, baja Horda en Garganta, y enemigos CFBG de la misma faccion original.
No se han ejecutado victorias reales de Alterac, Arathi, Garganta, Playa o Isla. Estan configuradas y usan el mismo motor.
Ver VERIFICATION.md para los resultados y las excepciones del checker SQL del core.
