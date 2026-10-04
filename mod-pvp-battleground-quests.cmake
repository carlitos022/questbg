# This module owns exactly one translation unit and one native loader.
# Explicit selection also removes stale or accidentally restored prototype loaders.
if(TARGET modules)
  get_target_property(PVPBGQ_MODULE_SOURCES modules SOURCES)
  list(FILTER PVPBGQ_MODULE_SOURCES EXCLUDE REGEX "mod-pvp-battleground-quests")
  list(APPEND PVPBGQ_MODULE_SOURCES
    "${CMAKE_SOURCE_DIR}/modules/mod-pvp-battleground-quests/src/mod_pvp_battleground_quests.cpp")
  set_property(TARGET modules PROPERTY SOURCES "${PVPBGQ_MODULE_SOURCES}")
endif()
