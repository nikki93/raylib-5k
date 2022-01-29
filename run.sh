#!/bin/bash

export PROJECT_NAME="raylib-5k"
export WSLENV="PROJECT_NAME"

set -e
set -o pipefail

PLATFORM="macOS"
EXE=""
TIME="time"
TIME_TOTAL="time"

if [[ -f /proc/version ]]; then
  if grep -q Linux /proc/version; then
    PLATFORM="lin"
    TIME="time --format=%es\n"
    TIME_TOTAL="time --format=total\t%es\n"
  fi
  if grep -q Microsoft /proc/version; then
    EXE=".exe"
    PLATFORM="win"
  fi
fi
CMAKE="cmake$EXE"
CLANG_FORMAT="clang-format$EXE"
GO="go$EXE"
GX="gx$EXE"
CMAKE="$TIME $CMAKE"
GO="$TIME $GO"

validate-glsl() {
  #cd assets/
  #for f in *.frag; do
  #  glslangValidator$EXE $f | sed "s/^ERROR: 0/assets\/$f/g" | sed "/\.frag$/d"
  #done
  #cd - > /dev/null
  :
}

run-gx() {
  if [ ../gx/gx.go -nt ../gx/$GX ] || [ ../gx/gx.hh -nt ../gx/$GX ]; then
    cd ../gx
    $GO build gx.go
    cd - > /dev/null
  fi
  mkdir -p build/gx
  $TIME ../gx/$GX ./game build/gx/game
}

case "$1" in
  # Compile commands DB (used by editor plugins)
  db)
    run-gx
    $CMAKE -DCMAKE_EXPORT_COMPILE_COMMANDS=ON -DCMAKE_BUILD_TYPE=Debug -H. -Bbuild/db -GNinja
    cp ./build/db/compile_commands.json .
    ;;

  # Format
  format)
    $CLANG_FORMAT -i -style=file $(find core/ game/ -name '*.hh' -o -name '*.cc' -type f)
    ;;

  # Count lines of code
  cloc)
    cloc core/ game/ --by-file --exclude_list_file=.cloc_exclude_list
    ;;

  # Desktop
  release)
    validate-glsl
    run-gx
    $CMAKE -H. -Bbuild/release -GNinja
    $CMAKE --build build/release
    case $PLATFORM in
      lin|macOS)
        if [[ -z "$VALGRIND" ]]; then
          ./build/release/$PROJECT_NAME $2
        else
          SUPPRESSIONS="
          {
            ignore_versioned_system_libs
            Memcheck:Leak
            ...
            obj:*/lib*/lib*.so.*
          }
          {
            ignore_decompose_rpath
            Memcheck:Addr8
            ...
            fun:decompose_rpath
            ...
          }
          {
            ignore_iris_dri
            Memcheck:Leak
            ...
            obj:*/dri/iris_dri.so
            ...
          }
          {
            ignore_dl_open
            Memcheck:Leak
            ...
            fun:_dl_open
            ...
          }
          {
            ignore_dl_open
            Memcheck:Addr16
            ...
            fun:_dl_open
            ...
          }
          {
            ignore_dl_close
            Memcheck:Leak
            ...
            fun:_dl_close
            ...
          }
          "
          valgrind \
            --log-file="./build/valgrind.log" \
            --suppressions=<(echo "$SUPPRESSIONS") \
            --gen-suppressions=all \
            --leak-check=full \
            --show-leak-kinds=all \
            -s \
            ./build/release/$PROJECT_NAME $2
          cat build/valgrind.log
        fi
        ;;
      win)
        ./build/release/$PROJECT_NAME.exe $2
        ;;
    esac
    ;;
  debug)
    validate-glsl
    run-gx
    $CMAKE -DCMAKE_BUILD_TYPE=Debug -H. -Bbuild/debug -GNinja
    $CMAKE --build build/debug
    case $PLATFORM in
      lin|macOS)
        ./build/debug/$PROJECT_NAME $2
        ;;
      win)
        ./build/debug/$PROJECT_NAME.exe $2
        ;;
    esac
    ;;

  # Web
  web-init)
    case $PLATFORM in
      lin|macOS)
        cd vendor/emsdk
        ./emsdk install latest
        ./emsdk activate latest
        ;;
      win)
        cd vendor/emsdk
        cmd.exe /c emsdk install latest
        cmd.exe /c emsdk activate latest
        ;;
    esac
    ;;
  web-release)
    if [[ ! -f "vendor/emsdk/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake" ]]; then
      ./run.sh web-init
    fi
    validate-glsl
    run-gx
    if [[ ! -d "build/web-release" ]]; then
      $CMAKE -DWEB=ON -H. -Bbuild/web-release -GNinja
    fi
    $CMAKE --build build/web-release
    touch build/web-release/reload-trigger
    ;;
  web-release-fast)
    if [[ ! -f "vendor/emsdk/upstream/emscripten/cmake/Modules/Platform/Emscripten.cmake" ]]; then
      ./run.sh web-init
    fi
    validate-glsl
    run-gx
    if [[ ! -d "build/web-release-fast" ]]; then
      $CMAKE -DWEB=ON -DWEB_BUNDLE_ASSETS=ON -DRELEASE_FAST=ON -H. -Bbuild/web-release-fast -GNinja
    fi
    $CMAKE --build build/web-release-fast
    cp build/web-release-fast/{index.*,$PROJECT_NAME.*} app/web-release-fast
    rm app/web-release-fast/raylib-5k.data
    touch app/web-release-fast/raylib-5k.data
    ;;
  web-watch-release)
    find CMakeLists.txt core game assets web -type f | entr $TIME_TOTAL ./run.sh web-release
    ;;
  web-serve-release)
    npx http-server -p 9002 -c-1 build/web-release
    ;;
  itch-publish)
    ./run.sh web-release-fast
    cd build/web-release-fast
    gsed -i 's/new URLSearch/true || new URLSearch/g' index.js
    zip itch-publish.zip -r index.* $PROJECT_NAME.*
    mv itch-publish.zip ../..
    ;;
  web-publish)
    ./run.sh web-release-fast
    cd build/web-release-fast
    gsed -i 's/new URLSearch/true || new URLSearch/g' index.js
    cd -
    rm -rf docs/*
    cp build/web-release-fast/{index.*,$PROJECT_NAME.*} docs/
    git add docs
    git commit -m 'web publish'
    git push origin master
    ;;

  # Electron
  electron)
    cd app
    npm i
    npx electron-packager . --overwrite
    ;;
esac
