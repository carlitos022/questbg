/*
 * PvP Battleground Quests for AzerothCore 3.3.5a.
 * Native quest/reward APIs belong to the AzerothCore Project.
 */
#include "AllBattlegroundScript.h"
#include "Battleground.h"
#include "Chat.h"
#include "CommandScript.h"
#include "DatabaseEnv.h"
#include "Log.h"
#include "ObjectMgr.h"
#include "Player.h"
#include "PlayerScript.h"
#include "QuestDef.h"
#include "WorldScript.h"
#include "WorldSession.h"

#include <chrono>
#include <map>
#include <memory>
#include <mutex>
#include <tuple>
#include <vector>

using namespace Acore::ChatCommands;

namespace PvPBGQ
{
using Clock = std::chrono::steady_clock;
enum ObjectiveType : uint8 { Win = 1, Kill = 2 };

struct Rule
{
    uint32 questId;
    uint8 objective;
    uint8 bgType;
    uint16 requiredCount;
    uint32 creditEntry;
    bool oppositeFaction;
    bool blockSameAccount;
    uint32 victimCooldown;
    uint16 victimLimit;
    uint32 minParticipation;
};

struct Config
{
    bool enabled = false;
    uint64 revision = 0;
    std::vector<Rule> rules;
};

struct VictimCredit
{
    Clock::time_point last;
    uint32 count = 0;
};

using JoinKey = std::pair<uint32, ObjectGuid>;
using VictimKey = std::tuple<uint32, uint32, ObjectGuid, ObjectGuid>;

std::mutex StateMutex;
std::shared_ptr<Config const> Current = std::make_shared<Config>();
std::map<JoinKey, Clock::time_point> Joined;
std::map<VictimKey, VictimCredit> Victims;

bool Supported(uint8 bgType)
{
    return bgType == BATTLEGROUND_AV || bgType == BATTLEGROUND_WS || bgType == BATTLEGROUND_AB ||
        bgType == BATTLEGROUND_EY || bgType == BATTLEGROUND_SA || bgType == BATTLEGROUND_IC;
}

bool Matches(Rule const& rule, Battleground* bg)
{
    return bg && !bg->isArena() && Supported(bg->GetBgTypeID(true)) &&
        (rule.bgType == 0 || rule.bgType == bg->GetBgTypeID(true));
}

std::shared_ptr<Config const> Snapshot()
{
    std::lock_guard<std::mutex> guard(StateMutex);
    return Current;
}

bool Reload()
{
    auto next = std::make_shared<Config>();
    QueryResult settings = WorldDatabase.Query(
        "SELECT enabled,sync_revision FROM pvp_bg_quest_settings WHERE id=1");
    if (!settings)
    {
        LOG_ERROR("module.pvpbgq", "Configuration unavailable; install the module SQL first.");
        return false;
    }

    next->enabled = settings->Fetch()[0].Get<bool>();
    uint64 previousRevision = settings->Fetch()[1].Get<uint64>();
    QueryResult result = WorldDatabase.Query(
        "SELECT quest_id,objective_type,battleground_type,required_count,"
        "IF(objective_type=1,IF(credit_entry=0,quest_id,credit_entry),0),"
        "require_opposite_faction,block_same_account,victim_cooldown_seconds,"
        "victim_credit_limit,min_win_participation_seconds "
        "FROM pvp_bg_quest_config WHERE enabled=1 ORDER BY quest_id");

    if (result)
    {
        do
        {
            Field* fields = result->Fetch();
            Rule rule{fields[0].Get<uint32>(), fields[1].Get<uint8>(), fields[2].Get<uint8>(),
                fields[3].Get<uint16>(), fields[4].Get<uint32>(), fields[5].Get<bool>(),
                fields[6].Get<bool>(), fields[7].Get<uint32>(), fields[8].Get<uint16>(),
                fields[9].Get<uint32>()};
            if ((rule.objective != Win && rule.objective != Kill) || !rule.requiredCount ||
                (rule.bgType && !Supported(rule.bgType)) ||
                (rule.objective == Kill && rule.requiredCount > 255))
            {
                LOG_ERROR("module.pvpbgq", "Invalid rule for quest {}; reload rejected.", rule.questId);
                return false;
            }
            next->rules.push_back(rule);
        } while (result->NextRow());
    }

    // Constant SQL: the procedure validates references and atomically projects native quest data.
    WorldDatabase.DirectExecute("CALL pvp_bg_quests_sync()");
    QueryResult revision = WorldDatabase.Query(
        "SELECT sync_revision FROM pvp_bg_quest_settings WHERE id=1");
    if (!revision || revision->Fetch()[0].Get<uint64>() <= previousRevision)
    {
        LOG_ERROR("module.pvpbgq", "Quest synchronization failed; previous configuration retained.");
        return false;
    }
    next->revision = revision->Fetch()[0].Get<uint64>();

    // Load only our exclusive victory credits, using the core prepared query.
    for (Rule const& rule : next->rules)
    {
        if (rule.objective != Win)
            continue;
        WorldDatabasePreparedStatement* statement =
            WorldDatabase.GetPreparedStatement(WORLD_SEL_CREATURE_TEMPLATE);
        statement->SetData(0, rule.creditEntry);
        PreparedQueryResult creature = WorldDatabase.Query(statement);
        if (!creature)
        {
            LOG_ERROR("module.pvpbgq", "Victory credit {} missing.", rule.creditEntry);
            return false;
        }
        sObjectMgr->LoadCreatureTemplate(creature->Fetch());
    }

    sObjectMgr->LoadQuests();
    sObjectMgr->LoadQuestStartersAndEnders();
    for (Rule const& rule : next->rules)
    {
        Quest const* quest = sObjectMgr->GetQuestTemplate(rule.questId);
        if (!quest || quest->GetZoneOrSort() != -QUEST_SORT_BATTLEGROUNDS ||
            (rule.objective == Kill && quest->GetPlayersSlain() != rule.requiredCount) ||
            (rule.objective == Win && (quest->RequiredNpcOrGo[0] != int32(rule.creditEntry) ||
                quest->RequiredNpcOrGoCount[0] != rule.requiredCount)))
        {
            LOG_ERROR("module.pvpbgq", "Native quest {} does not match its configuration.", rule.questId);
            return false;
        }
    }

    {
        std::lock_guard<std::mutex> guard(StateMutex);
        Current = next;
        // Do not clear antifarm state on reload: reconnect/reload cannot reset victim limits.
    }
    LOG_INFO("module.pvpbgq", "Loaded {} rules, revision {}, enabled={}.",
        next->rules.size(), next->revision, next->enabled);
    return true;
}

bool AllowVictim(Rule const& rule, Player* killer, Player* killed, Battleground* bg)
{
    auto now = Clock::now();
    VictimKey key{bg->GetInstanceID(), rule.questId, killer->GetGUID(), killed->GetGUID()};
    std::lock_guard<std::mutex> guard(StateMutex);
    auto found = Victims.find(key);
    if (found != Victims.end())
    {
        if (rule.victimLimit && found->second.count >= rule.victimLimit)
            return false;
        if (rule.victimCooldown &&
            now - found->second.last < std::chrono::seconds(rule.victimCooldown))
            return false;
        found->second.last = now;
        ++found->second.count;
    }
    else
        Victims.emplace(key, VictimCredit{now, 1});
    return true;
}

class PlayerEvents : public PlayerScript
{
public:
    PlayerEvents() : PlayerScript("PvPBGQPlayerEvents",
        {PLAYERHOOK_ON_PVP_KILL, PLAYERHOOK_ON_PLAYER_COMPLETE_QUEST}) { }

    void OnPlayerPVPKill(Player* killer, Player* killed) override
    {
        auto config = Snapshot();
        if (!config->enabled || !killer || !killed || killer == killed)
            return;
        Battleground* bg = killer->GetBattleground();
        LOG_DEBUG("module.pvpbgq", "PvP kill: killer={} victim={} bg={} status={} type={} teams={}/{} "
            "original={}/{} sameMap={} sameBG={}.",
            killer->GetGUID().ToString(), killed->GetGUID().ToString(),
            bg ? bg->GetInstanceID() : 0, bg ? uint32(bg->GetStatus()) : 0,
            bg ? uint32(bg->GetBgTypeID(true)) : 0, uint32(killer->GetBgTeamId()),
            uint32(killed->GetBgTeamId()), uint32(killer->GetTeamId(true)),
            uint32(killed->GetTeamId(true)), killer->GetMap() == killed->GetMap(),
            killed->GetBattleground() == bg);
        if (!bg || bg->isArena() || bg->GetStatus() != STATUS_IN_PROGRESS ||
            killed->GetBattleground() != bg || killer->GetMap() != killed->GetMap() ||
            killer->GetBgTeamId() == killed->GetBgTeamId())
            return;

        for (Rule const& rule : config->rules)
        {
            if (rule.objective != Kill || !Matches(rule, bg) ||
                killer->GetQuestStatus(rule.questId) != QUEST_STATUS_INCOMPLETE ||
                (rule.oppositeFaction && killer->GetTeamId(true) == killed->GetTeamId(true)) ||
                (rule.blockSameAccount && killer->GetSession()->GetAccountId() == killed->GetSession()->GetAccountId()))
                continue;
            if (!AllowVictim(rule, killer, killed, bg))
                continue;
            if (Quest const* quest = sObjectMgr->GetQuestTemplate(rule.questId))
                killer->KilledPlayerCreditForQuest(1, quest);
        }
    }

    void OnPlayerCompleteQuest(Player* player, Quest const* quest) override
    {
        if (!player || !quest)
            return;
        auto config = Snapshot();
        for (Rule const& rule : config->rules)
        {
            if (rule.questId != quest->GetQuestId())
                continue;
            // RewardQuest has already granted the native reward. Never grant it a second time here.
            LOG_INFO("module.pvpbgq", "Native turn-in: player={} quest={} honor={} arena={}.",
                player->GetGUID().ToString(), rule.questId,
                quest->GetRewHonorAddition(), quest->GetRewArenaPoints());
            break;
        }
    }
};

class BGEvents : public AllBattlegroundScript
{
public:
    BGEvents() : AllBattlegroundScript("PvPBGQBGEvents",
        {ALLBATTLEGROUNDHOOK_ON_BATTLEGROUND_ADD_PLAYER,
         ALLBATTLEGROUNDHOOK_ON_BATTLEGROUND_REMOVE_PLAYER_AT_LEAVE,
         ALLBATTLEGROUNDHOOK_ON_BATTLEGROUND_END_REWARD,
         ALLBATTLEGROUNDHOOK_ON_BATTLEGROUND_DESTROY}) { }

    void OnBattlegroundAddPlayer(Battleground* bg, Player* player) override
    {
        if (!bg || !player || bg->isArena())
            return;
        std::lock_guard<std::mutex> guard(StateMutex);
        Joined[{bg->GetInstanceID(), player->GetGUID()}] = Clock::now();
    }

    void OnBattlegroundRemovePlayerAtLeave(Battleground* bg, Player* player) override
    {
        if (!bg || !player)
            return;
        std::lock_guard<std::mutex> guard(StateMutex);
        Joined.erase({bg->GetInstanceID(), player->GetGUID()});
    }

    void OnBattlegroundDestroy(Battleground* bg) override
    {
        if (!bg)
            return;
        std::lock_guard<std::mutex> guard(StateMutex);
        for (auto it = Joined.begin(); it != Joined.end();)
        {
            if (it->first.first == bg->GetInstanceID())
                it = Joined.erase(it);
            else
                ++it;
        }
        for (auto it = Victims.begin(); it != Victims.end();)
        {
            if (std::get<0>(it->first) == bg->GetInstanceID())
                it = Victims.erase(it);
            else
                ++it;
        }
    }

    void OnBattlegroundEndReward(Battleground* bg, Player* player, TeamId winner) override
    {
        auto config = Snapshot();
        if (!config->enabled || !bg || !player || bg->isArena() ||
            (winner != TEAM_ALLIANCE && winner != TEAM_HORDE) ||
            player->GetBattleground() != bg || player->GetBgTeamId() != winner)
            return;

        uint32 participation = 0;
        {
            std::lock_guard<std::mutex> guard(StateMutex);
            auto joined = Joined.find({bg->GetInstanceID(), player->GetGUID()});
            if (joined != Joined.end())
                participation = uint32(std::chrono::duration_cast<std::chrono::seconds>(
                    Clock::now() - joined->second).count());
        }
        for (Rule const& rule : config->rules)
        {
            if (rule.objective != Win || !Matches(rule, bg) ||
                player->GetQuestStatus(rule.questId) != QUEST_STATUS_INCOMPLETE ||
                participation < rule.minParticipation)
                continue;
            // Unique credit per quest: winning EotS cannot complete a Warsong quest.
            player->KilledMonsterCredit(rule.creditEntry);
        }
    }
};

class Commands : public CommandScript
{
public:
    Commands() : CommandScript("PvPBGQCommands") { }

    ChatCommandTable GetCommands() const override
    {
        static ChatCommandTable actions =
        {
            {"reload", HandleReload, SEC_GAMEMASTER, Console::Yes},
            {"status", HandleStatus, SEC_GAMEMASTER, Console::Yes}
        };
        static ChatCommandTable commands = {{"pvpbgq", actions}};
        return commands;
    }

    static bool HandleReload(ChatHandler* handler)
    {
        if (!Reload())
        {
            handler->SendSysMessage("PvPBGQ: recarga rechazada; revisa el log y los datos SQL.");
            return false;
        }
        handler->SendSysMessage("PvPBGQ: objetivos, textos y recompensas recargados sin recompilar.");
        return true;
    }

    static bool HandleStatus(ChatHandler* handler)
    {
        auto config = Snapshot();
        handler->PSendSysMessage("PvPBGQ: enabled={} quests={} revision={}",
            config->enabled, config->rules.size(), config->revision);
        return true;
    }
};

class Startup : public WorldScript
{
public:
    Startup() : WorldScript("PvPBGQStartup", {WORLDHOOK_ON_STARTUP}) { }
    void OnStartup() override { Reload(); }
};
}

void Addmod_pvp_battleground_questsScripts()
{
    new PvPBGQ::PlayerEvents();
    new PvPBGQ::BGEvents();
    new PvPBGQ::Commands();
    new PvPBGQ::Startup();
}
