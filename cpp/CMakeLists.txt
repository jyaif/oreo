cmake_minimum_required (VERSION 2.6)
project(oreo)

# for tests:
include_directories("src")

set(CMAKE_CXX_FLAGS "-std=c++17 -Werror")

add_executable(
	oreo_test_bin
	test/test.cpp
    src/oreo.h
)