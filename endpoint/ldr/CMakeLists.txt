cmake_minimum_required(VERSION 3.15)

project(Kratos LANGUAGES CXX)

# Default to a Debug build when no build type is selected (single-config generators).
if(NOT CMAKE_BUILD_TYPE AND NOT CMAKE_CONFIGURATION_TYPES)
    set(CMAKE_BUILD_TYPE Debug CACHE STRING "Build type" FORCE)
endif()

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

file(GLOB_RECURSE SOURCES CONFIGURE_DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/src/*.cc")

add_executable(Kratos ${SOURCES})

target_include_directories(Kratos PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/includes")

# Define DEBUG for every source of the target in Debug builds.
target_compile_definitions(Kratos PRIVATE $<$<CONFIG:Debug>:DEBUG>)

target_link_options(Kratos PRIVATE -static -static-libgcc -static-libstdc++)
