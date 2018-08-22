cmake_minimum_required(VERSION 2.8)
project(stellarstation-api CXX)

include(${CMAKE_BINARY_DIR}/conanbuildinfo.cmake)
conan_basic_setup()

include_directories(.)
file(GLOB_RECURSE SOURCES RELATIVE ${CMAKE_SOURCE_DIR} "stellarstation/api/v1/*.cc")
add_library(stellarstation ${SOURCES})
