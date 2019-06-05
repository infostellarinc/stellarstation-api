#!/bin/bash

if [ -z "$CI" ]; then

  # OS specific support (must be 'true' or 'false').
  windows=false
  darwin=false
  linux=false
  case "`uname`" in
    MSYS* )
      windows=true
      ;;
    CYGWIN* )
      windows=true
      ;;
    Darwin* )
      darwin=true
      ;;
    MINGW* )
      windows=true
      ;;
    Linux* )
      linux=true
      ;;
  esac

  GRADLE_HOME="${GRADLE_USER_HOME:-${HOME}/.gradle}"
  OPENJDK_DIR="$GRADLE_HOME/curiostack/openjdk"

  export JAVA_HOME="$OPENJDK_DIR/jdk-12.0.1"
  if "$darwin" = "true"; then
    export JAVA_HOME="$JAVA_HOME.jdk/Contents/Home"
  fi

  if "$linux" = "true"; then
    SRC="https://download.java.net/java/GA/jdk12.0.1/69cfe15208a647278a19ef0990eea691/12/GPL/openjdk-12.0.1_linux-x64_bin.tar.gz"
    DEST="$OPENJDK_DIR/openjdk-12.0.1_linux-x64_bin.tar.gz"
  fi

  if "$darwin" = "true"; then
    SRC="https://download.java.net/java/GA/jdk12.0.1/69cfe15208a647278a19ef0990eea691/12/GPL/openjdk-12.0.1_osx-x64_bin.tar.gz"
    DEST="$OPENJDK_DIR/openjdk-12.0.1_osx-x64_bin.tar.gz"
  fi

  if "$windows" = "true"; then
    SRC="https://download.java.net/java/GA/jdk12.0.1/69cfe15208a647278a19ef0990eea691/12/GPL/openjdk-12.0.1_windows-x64_bin.zip"
    DEST="$OPENJDK_DIR/openjdk-12.0.1_windows-x64_bin.zip"
  fi

  if [ ! -d "$JAVA_HOME" ]; then
    mkdir -p "$OPENJDK_DIR"

    echo "Downloading OpenJDK"
    curl "$SRC" -o "$DEST"

    cd "$OPENJDK_DIR"

    if "$windows" = "true"; then
      unzip "$DEST"
    else
      tar -xf "$DEST"
    fi
  fi

fi
