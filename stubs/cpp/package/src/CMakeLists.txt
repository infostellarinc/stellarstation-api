cmake_minimum_required(VERSION 2.8)
project(stellarstation-api CXX)

include(${CMAKE_BINARY_DIR}/conanbuildinfo.cmake)
conan_basic_setup()

include_directories(.)
add_library(stellarstation stellarstation/api/v1/stellarstation.grpc.pb.cc stellarstation/api/v1/stellarstation.pb.cc)
