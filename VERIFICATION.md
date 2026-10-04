# Verificacion del modulo PvP
Fecha: 2026-10-04. Instalacion: Carlitos, AzerothCore 3.3.5a y MySQL 8.4.

## Resultado funcional
Cinco pruebas de funcionalidad pasaron con clientes automatizados AzerothGhost contra authserver/worldserver reales.
Se usaron personajes de pruebas de nivel 80. Las colas, invitaciones, entrada, progreso, victorias y entregas usan caminos nativos.
El harness usa comandos GM para posicionar personajes y aplicar dano; la victoria de Ojo se obtuvo capturando cuatro torres y alcanzando 1600 puntos.
Solo la prueba aislada de entregas usa completar una quest por GM como fixture.

| Prueba | Resultado |
|---|---|
| NeutralNPC_NativeRewards_NoReplay | PASS, 15.53 s; Alianza y Horda, sin pago anticipado ni duplicados, persistencia tras reconexion |
| SQL_RewardSlotCompaction | PASS, 0.09 s; frost solo, item1 solo, item2 solo, ambos y ninguno |
| EotS_RealKills_AntiFarm_RealVictory | PASS, 628.03 s; 50 bajas, cooldown, limite por victima, reload, preparacion/mundo excluidos, victoria real y entregas |
| Warsong_HordeKill_ExcludesEotS | PASS, 123.56 s; baja Horda, mision general +1 y exclusiva de Ojo 0 |
| CFBG_SameOriginalFaction_ConfigurableCredit | PASS, 124.13 s; dos personajes originalmente Alianza en equipos enemigos |

Las pruebas se ejecutaron por grupos. Las primeras esperas de la victoria de Ojo fallaron por deteccion de bandera y despues por un timeout insuficiente. La version final espera hasta 360 segundos despues de capturar las cuatro torres y paso.
No se han ejecutado victorias reales de AV, AB, WSG, SotA ni IoC; sus reglas estan configuradas y usan el mismo motor.

## Recompensas obtenidas jugando
Personaje Pveleadkfmiq, GUID 481. Las quests 91050, 91051 y 91052 se ganaron durante la BG y se devolvieron al NPC neutral.

| Quest | Honor otorgado | Arena points otorgados | Item 40093 | Item 33448 | Frost 49426 |
|---|---:|---:|---:|---:|---:|
| 91050 | 5000 | 50 | 2 | 3 | 3 |
| 91051 | 6500 | 75 | 2 | 3 | 5 |
| 91052 | 7500 | 100 | 2 | 3 | 5 |

Auditoria SQL posterior al reinicio: inventario con 40093 x6, 33448 x9 y 49426 x13; tres registros nativos de entrega, uno por quest; arena points 225.
Honor total observado: 26564, incluyendo honor de la BG ademas de los 19000 de las tres quests.
Los dos IDs de pociones se seleccionaron al azar entre items existentes durante la configuracion; no cambian al azar en cada entrega.

## CrossFaction BG
CFBG.Enable=1 y CFBG.Battlefield.Enable=1 durante todas las pruebas.
En la prueba dedicada, el servidor registro GUIDs 480/479 con equipos asignados 1/0 y facciones originales 0/0, misma BG, mismo mapa y partida activa.
require_opposite_faction=1 bloqueo esa baja; require_opposite_faction=0 permitio un credito.
GetBgTeamId comprueba siempre que los equipos asignados sean enemigos. GetTeamId(true) aplica solamente el filtro opcional de facciones originales.
Las victorias consultan el equipo asignado antes de salir de la BG. Ese camino se reviso en el source; no se ha probado una victoria de un personaje cambiado de faccion por CFBG.

## Estado final
- Nueve quests administradas, nueve relaciones de inicio y nueve de entrega del NPC.
- Dos spawns neutrales, Ventormenta y Orgrimmar.
- Recompensas iniciales de las nueve quests: 40093 x2 y 33448 x3, mas honor, arena points y frost de cada regla.
- Configuracion restaurada: require_opposite_faction=1, cooldown=60 segundos, limite=3 por victima/quest/partida.
- Para dar credito a cualquier enemigo asignado por CFBG, poner require_opposite_faction=0 en las reglas de bajas y ejecutar .pvpbgq reload; ver README.
- Reglas activas recargadas: revision 50, enabled=true.
- Logger del modulo restaurado a INFO.
- Console.Enable=0 en los servidores para permitir su ejecucion en segundo plano sin cierre por EOF.
- authserver activo, puerto 3724; worldserver activo, puerto 8085.

Compilacion final Release: PASS.
SHA256 del worldserver compilado e instalado:
9DB8EADEE93CC5D5D7115FBAA66051BF78E7BAF4AA166B22AE432AB58BDDE054

## Comprobaciones de codigo y SQL
Linter C++ del repositorio: PASS. Diff sin errores de whitespace.
SQL importado y procedimiento ejecutado contra MySQL 8.4, sin errores en la version final; reimportacion conserva configuracion y no duplica spawns.
Se ejecutaron las funciones del checker SQL del core directamente sobre ambos archivos del modulo. Pasaron whitespace, semicolons y motor InnoDB.
El checker generico no dio PASS global: exige DELETE antes de cada INSERT, incluso INSERT IGNORE y upserts usados para conservar ajustes del operador; tambien identifica MESSAGE_TEXT y COUNT como identificadores sin backticks aunque son sintaxis/funcion de MySQL.
Estas excepciones estan registradas; no se cambiaron el checker ni las tablas base del core para ocultarlas.
El source SQL tiene identificadores delimitados y operaciones acotadas a las quests administradas. No se borra el historial nativo de entregas.

## Archivos
src/mod_pvp_battleground_quests.cpp contiene un unico loader del modulo.
data/sql/db-world/base.sql contiene configuracion, recompensas y sincronizacion nativa.
data/sql/db-world/quests_and_npc.sql contiene el NPC y spawns idempotentes.
tests/pvp_bg_quests_test.go contiene la suite reproducible; copiarla a e2e/local/legends/pvp_bg_quests del core.
La suite localiza el spawn del NPC por entrada/mapa, sin depender del GUID de Carlitos.

Destino de publicacion designado: https://github.com/carlitos022/questbg. Ver GUIA_DE_USO.md para la comprobacion adicional del saldo de 225 puntos de arena y las instrucciones por mision.
