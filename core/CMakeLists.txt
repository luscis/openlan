cmake_minimum_required(VERSION 2.8)
project(openlan C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_C_FLAGS "-std=gnu99 -g -DHAVE_CONFIG_H")
set(CMAKE_VERBOSE_MAKEFILE ON)

include_directories(../3rd/ovs)
include_directories(../3rd/ovs/lib)
include_directories(../3rd/ovs/include)
include_directories(.)
include_directories(idlc)
include_directories(tcp)
include_directories(udp)

link_directories(${CMAKE_SOURCE_DIR}/../build/obj/usr/lib)
link_directories(${CMAKE_SOURCE_DIR}/../build/obj/usr/lib64)

file(GLOB IDL_SOURCES "idlc/*.c")
file(GLOB UDP_SOURCES "udp/*.c")
file(GLOB TCP_SOURCES "tcp/*.c")

add_executable(openudp ${UDP_SOURCES} ${IDL_SOURCES})
target_link_libraries(openudp libopenvswitch.a)
target_link_libraries(openudp pthread ssl crypto rt m unbound)

add_executable(opentcp ${TCP_SOURCES})
target_link_libraries(opentcp pthread)
