//go:build e2e

package pvpbgquests_test

import (
	"database/sql"
	"encoding/binary"
	"github.com/azerothcore/AzerothGhost/e2e/e2eharness"
	"github.com/azerothcore/azerothcore-wotlk/e2e/internal/meta"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"testing"
	"time"
)

func wait(t *testing.T, label string, timeout time.Duration, f func() bool) {
	t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		if f() {
			return
		}
		select {
		case <-timer.C:
			t.Fatalf("timeout: %s", label)
		case <-tick.C:
		}
	}
}
func dbWorld(t *testing.T) *sql.DB {
	t.Helper()
	db, e := e2eharness.OpenWorldDB()
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func execSQL(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, e := db.Exec(q, args...); e != nil {
		t.Fatal(e)
	}
}
func npc(t *testing.T, b *e2eharness.ScenarioBot) uint64 {
	t.Helper()
	var spawn uint32
	db := dbWorld(t)
	if e := db.QueryRow("SELECT guid FROM creature WHERE id=91049 AND map=0 ORDER BY guid LIMIT 1").Scan(&spawn); e != nil {
		t.Fatal(e)
	}
	b.GoCreatureGUID(t, spawn)
	var g uint64
	wait(t, "neutral NPC visible", 10*time.Second, func() bool {
		u := e2eharness.UnitsByEntry(b.World, 20, 91049)
		if len(u) > 0 {
			g = u[0].GUID
		}
		return g != 0
	})
	return g
}
func accept(t *testing.T, b *e2eharness.ScenarioBot, g uint64, q uint32) {
	t.Helper()
	if e := b.World.QuestgiverHello(g); e != nil {
		t.Fatal(e)
	}
	if e := b.World.QuestgiverAcceptQuest(g, q); e != nil {
		t.Fatal(e)
	}
	b.FlushWorld(t)
	b.AssertQuestStatus(t, q, 3)
}
func rewardState(t *testing.T, b *e2eharness.ScenarioBot) [5]int {
	t.Helper()
	if e := b.World.SetTarget(b.GUID); e != nil {
		t.Fatal(e)
	}
	b.Save(t)
	var s [5]int
	if e := b.CharDB.QueryRow("SELECT totalHonorPoints,arenaPoints FROM characters WHERE guid=?", b.GUID).Scan(&s[0], &s[1]); e != nil {
		t.Fatal(e)
	}
	s[2] = b.InventoryCount(t, 33447)
	s[3] = b.InventoryCount(t, 33448)
	s[4] = b.InventoryCount(t, 49426)
	return s
}
func TestPvPBGQ_NeutralNPC_NativeRewards_NoReplay(t *testing.T) {
	meta.Begin(t, meta.TestMeta{Tags: []string{"med", "legends", "pvpbgq", "serial"}, Runtime: "med", Category: "legends/pvp_bg_quests"})
	db := dbWorld(t)
	bots := e2eharness.NewScenario(t, e2eharness.ScenarioOpts{Prefix: "PvRew", Bots: []e2eharness.BotSpec{{Role: "alliance", Race: 1, Class: 1, Level: 80}, {Role: "horde", Race: 2, Class: 1, Level: 80}}})
	var old [7]int
	if e := db.QueryRow("SELECT honor_points,arena_points,item1,item1_count,item2,item2_count,frost_emblems FROM pvp_bg_quest_rewards WHERE quest_id=91052").Scan(&old[0], &old[1], &old[2], &old[3], &old[4], &old[5], &old[6]); e != nil {
		t.Fatal(e)
	}
	defer func() {
		execSQL(t, db, "UPDATE pvp_bg_quest_rewards SET honor_points=?,arena_points=?,item1=?,item1_count=?,item2=?,item2_count=?,frost_emblems=? WHERE quest_id=91052", old[0], old[1], old[2], old[3], old[4], old[5], old[6])
		bots[0].GM(t, ".pvpbgq reload")
		bots[0].FlushWorld(t)
	}()
	execSQL(t, db, "UPDATE pvp_bg_quest_rewards SET honor_points=23,arena_points=17,item1=33447,item1_count=2,item2=33448,item2_count=3,frost_emblems=2 WHERE quest_id=91052")
	bots[0].GM(t, ".pvpbgq reload")
	bots[0].FlushWorld(t)
	for _, b := range bots {
		g := npc(t, b)
		accept(t, b, g, 91052)
		before := rewardState(t, b)
		// Setup for the isolated turn-in test; the BG test below earns victory through gameplay.
		b.GM(t, ".quest complete 91052")
		b.FlushWorld(t)
		b.AssertQuestStatus(t, 91052, 1)
		if got := rewardState(t, b); got != before {
			t.Fatalf("objective completion paid rewards: before=%v got=%v", before, got)
		}
		if e := b.World.QuestgiverCompleteQuest(g, 91052); e != nil {
			t.Fatal(e)
		}
		if e := b.World.QuestgiverChooseReward(g, 91052, 0); e != nil {
			t.Fatal(e)
		}
		b.FlushWorld(t)
		want := before
		want[0] += 23
		want[1] += 17
		want[2] += 2
		want[3] += 3
		want[4] += 2
		wait(t, "native rewards saved", 8*time.Second, func() bool { return rewardState(t, b) == want })
		if e := b.World.QuestgiverChooseReward(g, 91052, 0); e != nil {
			t.Fatal(e)
		}
		b.FlushWorld(t)
		if got := rewardState(t, b); got != want {
			t.Fatalf("reward replay: %v want %v", got, want)
		}
		b.Relog(t)
		if got := rewardState(t, b); got != want {
			t.Fatalf("relog rewards: %v want %v", got, want)
		}
		var rewarded int
		if e := b.CharDB.QueryRow("SELECT COUNT(*) FROM character_queststatus_rewarded WHERE guid=? AND quest=91052", b.GUID).Scan(&rewarded); e != nil || rewarded != 1 {
			t.Fatalf("rewarded ledger count=%d err=%v", rewarded, e)
		}
		t.Logf("PASS %s: neutral NPC, no early payout, honor+arena+two items+frost, no replay after relog", b.Role)
	}
}
func progress(t *testing.T, b *e2eharness.ScenarioBot, q uint32) (int, int, int) {
	t.Helper()
	if e := b.World.SetTarget(b.GUID); e != nil {
		t.Fatal(e)
	}
	b.Save(t)
	var status, kills, mobs int
	if e := b.CharDB.QueryRow("SELECT status,playercount,mobcount1 FROM character_queststatus WHERE guid=? AND quest=?", b.GUID, q).Scan(&status, &kills, &mobs); e != nil {
		t.Fatal(e)
	}
	return status, kills, mobs
}
func packet(t *testing.T, b *e2eharness.ScenarioBot, op uint16, data []byte) {
	t.Helper()
	if e := b.World.SendPacketRaw(op, data); e != nil {
		t.Fatal(e)
	}
}
func queue(t *testing.T, b *e2eharness.ScenarioBot, kind uint32) <-chan uint32 {
	ch := make(chan uint32, 64)
	cancel := b.World.AddPacketHook(func(op uint16, d []byte) {
		if op == 0x2D4 && len(d) >= 23 {
			select {
			case ch <- binary.LittleEndian.Uint32(d[19:23]):
			default:
			}
		}
	})
	t.Cleanup(cancel)
	d := make([]byte, 17)
	binary.LittleEndian.PutUint32(d[8:], kind)
	packet(t, b, 0x2EE, d)
	return ch
}
func awaitStatus(t *testing.T, ch <-chan uint32, want uint32) {
	t.Helper()
	timer := time.NewTimer(90 * time.Second)
	defer timer.Stop()
	for {
		select {
		case s := <-ch:
			if s == want {
				return
			}
		case <-timer.C:
			t.Fatalf("BG status %d not received", want)
		}
	}
}
func enter(t *testing.T, b *e2eharness.ScenarioBot, kind uint32) {
	d := make([]byte, 9)
	binary.LittleEndian.PutUint32(d[2:], kind)
	binary.LittleEndian.PutUint16(d[6:], 0x1F90)
	d[8] = 1
	packet(t, b, 0x2D5, d)
}
func kill(t *testing.T, a, v *e2eharness.ScenarioBot) {
	t.Helper()
	v.GM(t, ".cheat god off")
	v.GM(t, ".gm off")
	a.GM(t, ".gm off")
	v.Teleport(t, 2178, 1569, 1160.4, 566)
	a.Teleport(t, 2174, 1569, 1160.4, 566)
	v.FlushWorld(t)
	a.FlushWorld(t)
	a.Damage(t, v.GUID, 10000000)
	v.WaitDead(t, 8*time.Second)
}

func TestPvPBGQ_EotS_RealKills_AntiFarm_RealVictory(t *testing.T) {
	meta.Begin(t, meta.TestMeta{Tags: []string{"med", "legends", "pvpbgq", "serial"}, Runtime: "med", Category: "legends/pvp_bg_quests"})
	db := dbWorld(t)
	bots := e2eharness.NewScenario(t, e2eharness.ScenarioOpts{Prefix: "PvEY", Bots: []e2eharness.BotSpec{{Role: "alliance", Race: 1, Class: 1, Level: 80}, {Role: "horde", Race: 2, Class: 1, Level: 80}}})
	a, v := bots[0], bots[1]

	var old [2][2]int
	for i, q := range []int{91050, 91051} {
		if e := db.QueryRow("SELECT victim_cooldown_seconds,victim_credit_limit FROM pvp_bg_quest_config WHERE quest_id=?", q).Scan(&old[i][0], &old[i][1]); e != nil {
			t.Fatal(e)
		}
	}
	defer func() {
		for i, q := range []int{91050, 91051} {
			execSQL(t, db, "UPDATE pvp_bg_quest_config SET victim_cooldown_seconds=?,victim_credit_limit=? WHERE quest_id=?", old[i][0], old[i][1], q)
		}
		a.GM(t, ".pvpbgq reload")
		a.FlushWorld(t)
	}()
	a.GM(t, ".pvpbgq reload")
	a.FlushWorld(t)
	for _, b := range bots {
		g := npc(t, b)
		for q := uint32(91050); q <= 91058; q++ {
			accept(t, b, g, q)
		}
		b.GM(t, ".cheat god off")
		b.GM(t, ".gm visible on")
		b.GM(t, ".gm off")
		b.FlushWorld(t)
	}
	e2eharness.EnableHostilePvP(t, a, v)
	a.Teleport(t, -8778.65, -1538.79, 262.44, 0)
	v.Teleport(t, -8777, -1538.79, 262.44, 0)
	t.Logf("world PvP fixture: attacker=%X victim=%X victimHP=%d", a.GUID, v.GUID, v.World.Health())
	a.Damage(t, v.GUID, 10000000)
	v.WaitDead(t, 8*time.Second)
	for _, q := range []uint32{91050, 91051} {
		_, n, _ := progress(t, a, q)
		if n != 0 {
			t.Fatalf("outside BG quest %d count=%d", q, n)
		}
	}
	v.GM(t, ".revive")
	v.WaitAlive(t, 8*time.Second)
	a.GM(t, ".debug bg")
	a.FlushWorld(t)
	defer func() { a.GM(t, ".debug bg"); a.FlushWorld(t) }()
	ca := queue(t, a, 7)
	cv := queue(t, v, 7)
	awaitStatus(t, ca, 2)
	awaitStatus(t, cv, 2)
	enter(t, a, 7)
	enter(t, v, 7)
	awaitStatus(t, ca, 3)
	awaitStatus(t, cv, 3)
	for _, b := range bots {
		wait(t, "EotS map", 15*time.Second, func() bool { _, _, _, _, m := b.World.Position(); return m == 566 })
		wait(t, "BG preparation", 10*time.Second, func() bool { return b.HasAura(44521) })
	}
	defer func() {
		for _, b := range bots {
			if _, _, _, _, m := b.World.Position(); m == 566 || m == 489 {
				b.GM(t, ".combatstop")
				b.FlushWorld(t)
				leave(t, b, 0)
			}
		}
	}()
	kill(t, a, v)
	for _, q := range []uint32{91050, 91051} {
		_, n, _ := progress(t, a, q)
		if n != 0 {
			t.Fatalf("preparation kill quest %d count=%d", q, n)
		}
	}
	v.GM(t, ".revive")
	v.WaitAlive(t, 8*time.Second)
	t.Log("Waiting for native BG start (120-second preparation)")
	wait(t, "BG preparation removed", 150*time.Second, func() bool { return !a.HasAura(44521) && !v.HasAura(44521) })
	kill(t, a, v)
	for _, q := range []uint32{91050, 91051} {
		_, n, _ := progress(t, a, q)
		if n != 1 {
			t.Fatalf("first BG kill quest %d count=%d", q, n)
		}
	}
	v.GM(t, ".revive")
	v.WaitAlive(t, 8*time.Second)
	kill(t, a, v)
	for _, q := range []uint32{91050, 91051} {
		_, n, _ := progress(t, a, q)
		if n != 1 {
			t.Fatalf("cooldown failed quest %d count=%d", q, n)
		}
	}
	a.GM(t, ".pvpbgq reload")
	a.FlushWorld(t)
	v.GM(t, ".revive")
	v.WaitAlive(t, 8*time.Second)
	kill(t, a, v)
	for _, q := range []uint32{91050, 91051} {
		_, n, _ := progress(t, a, q)
		if n != 1 {
			t.Fatalf("reload reset antifarm quest %d count=%d", q, n)
		}
	}
	execSQL(t, db, "UPDATE pvp_bg_quest_config SET victim_cooldown_seconds=0,victim_credit_limit=2 WHERE quest_id IN (91050,91051)")
	a.GM(t, ".pvpbgq reload")
	a.FlushWorld(t)
	v.GM(t, ".revive")
	v.WaitAlive(t, 8*time.Second)
	kill(t, a, v)
	v.GM(t, ".revive")
	v.WaitAlive(t, 8*time.Second)
	kill(t, a, v)
	for _, q := range []uint32{91050, 91051} {
		_, n, _ := progress(t, a, q)
		if n != 2 {
			t.Fatalf("victim limit failed quest %d count=%d", q, n)
		}
	}
	execSQL(t, db, "UPDATE pvp_bg_quest_config SET victim_cooldown_seconds=0,victim_credit_limit=0 WHERE quest_id IN (91050,91051)")
	a.GM(t, ".pvpbgq reload")
	a.FlushWorld(t)
	for i := 3; i <= 50; i++ {
		v.GM(t, ".revive")
		v.WaitAlive(t, 8*time.Second)
		kill(t, a, v)
		for _, q := range []uint32{91050, 91051} {
			st, n, _ := progress(t, a, q)
			want := 3
			if i == 50 {
				want = 1
			}
			if n != i || st != want {
				t.Fatalf("kill %d quest %d status=%d count=%d", i, q, st, n)
			}
		}
		if i%10 == 0 {
			t.Logf("Verified native PvP kill progress %d/50", i)
		}
	}
	v.GM(t, ".revive")
	v.WaitAlive(t, 8*time.Second)
	v.Teleport(t, 1805, 1540, 1260, 566)
	var mu sync.Mutex
	bases := uint32(0)
	cancel := a.World.AddPacketHook(func(op uint16, d []byte) {
		if op == 0x2C3 && len(d) >= 8 && binary.LittleEndian.Uint32(d[:4]) == 2752 {
			mu.Lock()
			bases = binary.LittleEndian.Uint32(d[4:8])
			mu.Unlock()
		}
	})
	defer cancel()
	points := [][3]float32{{2044.28, 1729.68, 1189.96}, {2048.83, 1393.65, 1194.49}, {2286.56, 1402.36, 1197.11}, {2284.48, 1731.23, 1189.99}}
	for i, p := range points {
		a.Teleport(t, p[0], p[1], p[2], 566)
		wait(t, "native tower capture", 65*time.Second, func() bool { mu.Lock(); defer mu.Unlock(); return bases >= uint32(i+1) })
		t.Logf("Native EotS towers captured: %d", i+1)
	}
	wait(t, "native EotS victory from four captured towers", 360*time.Second, func() bool { st, _, _ := progress(t, a, 91052); return st == 1 })
	for _, q := range []uint32{91052, 91058} {
		st, _, n := progress(t, a, q)
		if st != 1 || n != 1 {
			t.Fatalf("winner quest %d status=%d credit=%d", q, st, n)
		}
		st, _, n = progress(t, v, q)
		if st != 3 || n != 0 {
			t.Fatalf("loser quest %d status=%d credit=%d", q, st, n)
		}
	}
	for q := uint32(91053); q <= 91057; q++ {
		st, _, n := progress(t, a, q)
		if st != 3 || n != 0 {
			t.Fatalf("EotS advanced another BG quest %d", q)
		}
	}
	t.Log("PASS real EotS victory, exclusive victory credits, loser excluded, 50 native kills, outside/preparation excluded, cooldown+limit+reload verified")
	leave(t, a, 0)
	g := npc(t, a)
	for _, q := range []uint32{91050, 91051, 91052} {
		nativePayout(t, db, a, g, q)
	}

}

func TestPvPBGQ_Warsong_HordeKill_ExcludesEotS(t *testing.T) {
	meta.Begin(t, meta.TestMeta{Tags: []string{"med", "legends", "pvpbgq", "serial"}, Runtime: "med", Category: "legends/pvp_bg_quests"})
	bots := e2eharness.NewScenario(t, e2eharness.ScenarioOpts{Prefix: "PvWS", Bots: []e2eharness.BotSpec{{Race: 1, Class: 1, Level: 80}, {Race: 2, Class: 1, Level: 80}}})
	a, h := bots[0], bots[1]
	g := npc(t, h)
	accept(t, h, g, 91050)
	accept(t, h, g, 91051)
	for _, b := range bots {
		b.GM(t, ".gm visible on")
		b.GM(t, ".gm off")
		b.GM(t, ".cheat god off")
		b.FlushWorld(t)
	}
	a.GM(t, ".debug bg")
	a.FlushWorld(t)
	defer func() { a.GM(t, ".debug bg"); a.FlushWorld(t) }()
	ca := queue(t, a, 2)
	ch := queue(t, h, 2)
	awaitStatus(t, ca, 2)
	awaitStatus(t, ch, 2)
	enter(t, a, 2)
	enter(t, h, 2)
	awaitStatus(t, ca, 3)
	awaitStatus(t, ch, 3)
	defer func() {
		for _, b := range bots {
			if _, _, _, _, m := b.World.Position(); m == 566 || m == 489 {
				b.GM(t, ".combatstop")
				b.FlushWorld(t)
				leave(t, b, 0)
			}
		}
	}()
	for _, b := range bots {
		wait(t, "WSG map", 15*time.Second, func() bool { _, _, _, _, m := b.World.Position(); return m == 489 })
		wait(t, "WSG preparation", 10*time.Second, func() bool { return b.HasAura(44521) })
	}
	wait(t, "native WSG start", 150*time.Second, func() bool { return !a.HasAura(44521) && !h.HasAura(44521) })
	h.Teleport(t, 1235, 1450, 310, 489)
	a.Teleport(t, 1238, 1450, 310, 489)
	h.FlushWorld(t)
	a.FlushWorld(t)
	h.Damage(t, a.GUID, 10000000)
	a.WaitDead(t, 8*time.Second)
	_, global, _ := progress(t, h, 91050)
	_, eye, _ := progress(t, h, 91051)
	if global != 1 || eye != 0 {
		t.Fatalf("Horde WSG kill: global=%d EotS=%d", global, eye)
	}
	t.Log("PASS Horde native WSG kill: any-BG=1, EotS-only=0")
}
func TestPvPBGQ_SQL_RewardSlotCompaction(t *testing.T) {
	meta.Begin(t, meta.TestMeta{Tags: []string{"med", "legends", "pvpbgq", "serial"}, Runtime: "med", Category: "legends/pvp_bg_quests"})
	db := dbWorld(t)
	var old [5]int
	if e := db.QueryRow("SELECT item1,item1_count,item2,item2_count,frost_emblems FROM pvp_bg_quest_rewards WHERE quest_id=91052").Scan(&old[0], &old[1], &old[2], &old[3], &old[4]); e != nil {
		t.Fatal(e)
	}
	defer func() {
		execSQL(t, db, "UPDATE pvp_bg_quest_rewards SET item1=?,item1_count=?,item2=?,item2_count=?,frost_emblems=? WHERE quest_id=91052", old[0], old[1], old[2], old[3], old[4])
		execSQL(t, db, "CALL pvp_bg_quests_sync()")
	}()
	cases := []struct {
		in   [5]int
		want [6]int
	}{
		{in: [5]int{0, 0, 0, 0, 2}, want: [6]int{49426, 2, 0, 0, 0, 0}},
		{in: [5]int{33447, 2, 0, 0, 3}, want: [6]int{33447, 2, 49426, 3, 0, 0}},
		{in: [5]int{0, 0, 33448, 3, 2}, want: [6]int{33448, 3, 49426, 2, 0, 0}},
		{in: [5]int{33447, 2, 33448, 3, 4}, want: [6]int{33447, 2, 33448, 3, 49426, 4}},
		{in: [5]int{0, 0, 0, 0, 0}, want: [6]int{}},
	}
	for _, c := range cases {
		execSQL(t, db, "UPDATE pvp_bg_quest_rewards SET item1=?,item1_count=?,item2=?,item2_count=?,frost_emblems=? WHERE quest_id=91052", c.in[0], c.in[1], c.in[2], c.in[3], c.in[4])
		execSQL(t, db, "CALL pvp_bg_quests_sync()")
		var got [6]int
		if e := db.QueryRow("SELECT RewardItem1,RewardAmount1,RewardItem2,RewardAmount2,RewardItem3,RewardAmount3 FROM quest_template WHERE ID=91052").Scan(&got[0], &got[1], &got[2], &got[3], &got[4], &got[5]); e != nil {
			t.Fatal(e)
		}
		if got != c.want {
			t.Fatalf("compact %v: got=%v want=%v", c.in, got, c.want)
		}
	}
	t.Log("PASS native reward slots: frost only, item1 only, item2 only, both items, no items")
}

func leave(t *testing.T, b *e2eharness.ScenarioBot, outsideMap uint32) {
	t.Helper()
	b.GM(t, ".combatstop")
	b.FlushWorld(t)
	packet(t, b, 0x2E1, make([]byte, 8))
	wait(t, "leave native BG", 15*time.Second, func() bool { _, _, _, _, m := b.World.Position(); return m == outsideMap })
	b.WaitInWorld(t, 15*time.Second)
}
func nativePayout(t *testing.T, db *sql.DB, b *e2eharness.ScenarioBot, g uint64, q uint32) {
	t.Helper()
	var honor, arena, item1, n1, item2, n2, frost int
	if e := db.QueryRow("SELECT honor_points,arena_points,item1,item1_count,item2,item2_count,frost_emblems FROM pvp_bg_quest_rewards WHERE quest_id=?", q).Scan(&honor, &arena, &item1, &n1, &item2, &n2, &frost); e != nil {
		t.Fatal(e)
	}
	if e := b.World.SetTarget(b.GUID); e != nil {
		t.Fatal(e)
	}
	b.Save(t)
	var beforeH, beforeA int
	if e := b.CharDB.QueryRow("SELECT totalHonorPoints,arenaPoints FROM characters WHERE guid=?", b.GUID).Scan(&beforeH, &beforeA); e != nil {
		t.Fatal(e)
	}
	before1, before2, beforeF := b.InventoryCount(t, uint32(item1)), b.InventoryCount(t, uint32(item2)), b.InventoryCount(t, 49426)
	if e := b.World.QuestgiverCompleteQuest(g, q); e != nil {
		t.Fatal(e)
	}
	if e := b.World.QuestgiverChooseReward(g, q, 0); e != nil {
		t.Fatal(e)
	}
	b.FlushWorld(t)
	wait(t, "real BG quest payout", 10*time.Second, func() bool {
		b.Save(t)
		var h, a int
		if e := b.CharDB.QueryRow("SELECT totalHonorPoints,arenaPoints FROM characters WHERE guid=?", b.GUID).Scan(&h, &a); e != nil {
			t.Fatal(e)
		}
		return h == beforeH+honor && a == beforeA+arena && b.InventoryCount(t, uint32(item1)) == before1+n1 && b.InventoryCount(t, uint32(item2)) == before2+n2 && b.InventoryCount(t, 49426) == beforeF+frost
	})
	t.Logf("PASS real earned quest %d paid honor=%d arena=%d item=%dx%d item=%dx%d frost=%d", q, honor, arena, item1, n1, item2, n2, frost)
}

func TestPvPBGQ_CFBG_SameOriginalFaction_ConfigurableCredit(t *testing.T) {
	meta.Begin(t, meta.TestMeta{Tags: []string{"med", "legends", "pvpbgq", "cfbg", "serial"}, Runtime: "med", Category: "legends/pvp_bg_quests"})
	db := dbWorld(t)
	bots := e2eharness.NewScenario(t, e2eharness.ScenarioOpts{Prefix: "PvCF", Count: 2, Race: 1, Class: 1, Level: 80})
	a, v := bots[0], bots[1]
	var opposite, cooldown, limit int
	if e := db.QueryRow("SELECT require_opposite_faction,victim_cooldown_seconds,victim_credit_limit FROM pvp_bg_quest_config WHERE quest_id=91050").Scan(&opposite, &cooldown, &limit); e != nil {
		t.Fatal(e)
	}
	defer func() {
		execSQL(t, db, "UPDATE pvp_bg_quest_config SET require_opposite_faction=?,victim_cooldown_seconds=?,victim_credit_limit=? WHERE quest_id=91050", opposite, cooldown, limit)
		a.GM(t, ".pvpbgq reload")
		a.FlushWorld(t)
	}()
	execSQL(t, db, "UPDATE pvp_bg_quest_config SET require_opposite_faction=1,victim_cooldown_seconds=0,victim_credit_limit=0 WHERE quest_id=91050")
	a.GM(t, ".pvpbgq reload")
	a.FlushWorld(t)
	g := npc(t, a)
	accept(t, a, g, 91050)
	npc(t, v)
	for _, b := range bots {
		b.GM(t, ".gm visible on")
		b.GM(t, ".gm off")
		b.GM(t, ".cheat god off")
		b.FlushWorld(t)
	}
	a.GM(t, ".debug bg")
	a.FlushWorld(t)
	defer func() { a.GM(t, ".debug bg"); a.FlushWorld(t) }()
	ca := queue(t, a, 7)
	cv := queue(t, v, 7)
	awaitStatus(t, ca, 2)
	awaitStatus(t, cv, 2)
	enter(t, a, 7)
	enter(t, v, 7)
	awaitStatus(t, ca, 3)
	awaitStatus(t, cv, 3)
	defer func() {
		for _, b := range bots {
			if _, _, _, _, m := b.World.Position(); m == 566 {
				leave(t, b, 0)
			}
		}
	}()
	for _, b := range bots {
		wait(t, "CFBG Eye map", 15*time.Second, func() bool { _, _, _, _, m := b.World.Position(); return m == 566 })
		wait(t, "CFBG preparation", 10*time.Second, func() bool { return b.HasAura(44521) })
	}
	wait(t, "native CFBG start", 150*time.Second, func() bool { return !a.HasAura(44521) && !v.HasAura(44521) })
	kill(t, a, v)
	_, n, _ := progress(t, a, 91050)
	if n != 0 {
		t.Fatalf("same original faction with strict rule count=%d", n)
	}
	execSQL(t, db, "UPDATE pvp_bg_quest_config SET require_opposite_faction=0 WHERE quest_id=91050")
	a.GM(t, ".pvpbgq reload")
	a.FlushWorld(t)
	v.GM(t, ".revive")
	v.WaitAlive(t, 8*time.Second)
	kill(t, a, v)
	_, n, _ = progress(t, a, 91050)
	if n != 1 {
		t.Fatalf("opposite CFBG assigned teams did not count: %d", n)
	}
	t.Log("PASS CFBG native assignment: two original Alliance players in opposing BG teams; strict faction blocked, assigned-team mode credited")
}
