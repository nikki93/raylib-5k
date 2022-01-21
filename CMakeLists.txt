cmake_minimum_required(VERSION 3.16)
set(PROJECT_NAME $ENV{PROJECT_NAME})

# Make CMake a little quieter...
function(message)
  list(GET ARGV 0 MessageType)
  if(MessageType STREQUAL FATAL_ERROR OR
     MessageType STREQUAL SEND_ERROR OR
     MessageType STREQUAL WARNING OR
     MessageType STREQUAL AUTHOR_WARNING)
    list(REMOVE_AT ARGV 0)
    _message(${MessageType} "${ARGV}")
  endif()
endfunction()


#
# Toolchain
#

if(WEB)
  # Emscripten
  set(CMAKE_TOOLCHAIN_FILE vendor/emsdk/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake)
elseif(WIN32)
  set(CMAKE_C_COMPILER "C:/Program Files/LLVM/bin/clang.exe")
  set(CMAKE_CXX_COMPILER "C:/Program Files/LLVM/bin/clang++.exe")
  add_compile_options(-Wno-deprecated-declarations)
else()
  set(CMAKE_C_COMPILER clang)
  set(CMAKE_CXX_COMPILER clang++)
endif()


#
# Project
#

project(${PROJECT_NAME})
if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE Release) # Release build by default
endif()


#
# Target
#

add_executable(${PROJECT_NAME})

# Sources
file(GLOB MY_SOURCES CONFIGURE_DEPENDS core/*.hh core/*.cc game/*.hh game/*.cc build/gx/*.hh build/gx/*.cc)
target_sources(${PROJECT_NAME} PRIVATE  ${MY_SOURCES})
target_include_directories(${PROJECT_NAME} PRIVATE . build/gx)

# Standard
set_target_properties(${PROJECT_NAME} PROPERTIES CXX_STANDARD 20)

# Warnings
target_compile_options(${PROJECT_NAME} PRIVATE -fcolor-diagnostics -fdiagnostics-absolute-paths
  -pedantic-errors -Wall -Wextra 
  -Wno-pragma-once-outside-header -Wno-pragma-pack -Wno-unused-const-variable
  -Wno-gnu-zero-variadic-macro-arguments -Wno-unused-parameter -Wno-unused-function
  #-ftime-trace
  )
target_link_options(${PROJECT_NAME} PRIVATE -fcolor-diagnostics -fdiagnostics-absolute-paths
  -pedantic-errors -Wall -Wextra)

# Disable exceptions and rtti
target_compile_options(${PROJECT_NAME} PRIVATE -fno-exceptions -fno-rtti)

# Precompiled header
if(NOT CMAKE_EXPORT_COMPILE_COMMANDS)
  target_precompile_headers(${PROJECT_NAME} PRIVATE core/precomp.hh)
endif()

# Web flags
if(WEB)
  target_compile_definitions(${PROJECT_NAME} PRIVATE EMSCRIPTEN_HAS_UNBOUND_TYPE_NAMES=0)
  target_link_options(${PROJECT_NAME} PRIVATE "SHELL:--bind")
  target_link_options(${PROJECT_NAME} PRIVATE "SHELL:-s WASM=1")
  target_link_options(${PROJECT_NAME} PRIVATE "SHELL:-s MAX_WEBGL_VERSION=2")
  target_link_options(${PROJECT_NAME} PRIVATE "SHELL:-s INITIAL_MEMORY=512MB")
  target_link_options(${PROJECT_NAME} PRIVATE "SHELL:-lnodefs.js")
  target_link_options(${PROJECT_NAME} PRIVATE "SHELL:-s ENVIRONMENT=web,webview")
  if(RELEASE_FAST)
  else()
    target_compile_options(${PROJECT_NAME} PRIVATE -O0) # Makes build much faster
    target_link_options(${PROJECT_NAME} PRIVATE -O0) # Makes build much faster
    #target_compile_options(${PROJECT_NAME} PRIVATE "SHELL:-g")
    #target_link_options(${PROJECT_NAME} PRIVATE "SHELL:-g")
    #target_compile_options(${PROJECT_NAME} PRIVATE
    #  "SHELL:-gsource-map --source-map-base file://${CMAKE_BINARY_DIR}/ -fsanitize=address -s ALLOW_MEMORY_GROWTH")
    #target_link_options(${PROJECT_NAME} PRIVATE
    #  "SHELL:-gsource-map --source-map-base file://${CMAKE_BINARY_DIR}/ -fsanitize=address -s ALLOW_MEMORY_GROWTH")
    target_compile_options(${PROJECT_NAME} PRIVATE "SHELL:--profiling")
    target_link_options(${PROJECT_NAME} PRIVATE "SHELL:--profiling")
  endif()
  if(WEB_BUNDLE_ASSETS)
    file(GLOB MY_ASSETS CONFIGURE_DEPENDS ${PROJECT_SOURCE_DIR}/assets/*)
    set_target_properties(${PROJECT_NAME} PROPERTIES LINK_DEPENDS "${MY_ASSETS}")
  endif()
endif()

# Always add debug info in desktop builds
if(NOT WEB)
  target_compile_options(${PROJECT_NAME} PRIVATE -g)
  target_link_options(${PROJECT_NAME} PRIVATE -g)
  add_compile_options(-g)
  add_link_options(-g)
  if(WIN32)
    target_link_options(${PROJECT_NAME} PRIVATE -Xlinker /debug:full)
    add_link_options(-Xlinker /debug:full)
  endif()
endif()


#
# Dependencies
#

# raylib
if(WEB)
  target_link_options(${PROJECT_NAME} PRIVATE "SHELL:-s USE_GLFW=3")
  set(PLATFORM "Web" CACHE STRING "" FORCE)
else()
  set(OpenGL_GL_PREFERENCE "GLVND" CACHE STRING "" FORCE)
endif()
set(BUILD_EXAMPLES OFF CACHE BOOL "" FORCE)
add_subdirectory(vendor/raylib)
target_compile_definitions(raylib PUBLIC SUPPORT_FILEFORMAT_TGA=1)
target_link_libraries(${PROJECT_NAME} PRIVATE raylib)

# entt
target_compile_definitions(${PROJECT_NAME} PRIVATE ENTT_NO_ETO)
target_include_directories(${PROJECT_NAME} PRIVATE vendor/entt/src)

# cJSON
add_library(cjson STATIC)
target_sources(cjson PRIVATE vendor/cJSON/cJSON.c)
target_include_directories(cjson PUBLIC vendor/cJSON)
target_link_libraries(${PROJECT_NAME} PRIVATE cjson)

# Chipmunk2D
set(BUILD_DEMOS OFF CACHE BOOL "" FORCE)
set(INSTALL_DEMOS OFF CACHE BOOL "" FORCE)
set(BUILD_SHARED OFF CACHE BOOL "" FORCE)
set(BUILD_STATIC ON CACHE BOOL "" FORCE)
set(INSTALL_STATIC OFF CACHE BOOL "" FORCE)
add_subdirectory(vendor/Chipmunk2D EXCLUDE_FROM_ALL)
target_compile_definitions(chipmunk_static PUBLIC CP_USE_DOUBLES=0 CP_USE_CGTYPES=0)
target_include_directories(${PROJECT_NAME} PRIVATE vendor/Chipmunk2D/include)
target_link_libraries(${PROJECT_NAME} PRIVATE chipmunk_static)


#
# Assets
#

# Bundle 'assets/'
if(WEB)
  if(WEB_BUNDLE_ASSETS)
    target_link_options(${PROJECT_NAME} PRIVATE "SHELL:--preload-file ${PROJECT_SOURCE_DIR}/assets@assets")
  endif()
elseif("${CMAKE_SYSTEM_NAME}" STREQUAL "iOS")
  target_sources(${PROJECT_NAME} PRIVATE "assets")
  set_target_properties(${PROJECT_NAME} PROPERTIES RESOURCE "assets")
endif()

# Static web files
if(WEB)
  configure_file(web/index.html index.html)
  configure_file(web/index.css index.css COPYONLY)
  configure_file(web/index.js index.js)
endif()
