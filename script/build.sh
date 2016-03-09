#!/bin/bash
cd build
rm -rf .*
gox github.com/evolsnow/robot
mkdir -p out
for FILE in *
do
  DIR="${FILE%.*}"_v1.1
  mkdir -p "$DIR"
  cp "${FILE}" "$DIR"
  cp ../../config.json "$DIR"
  zip -r "${DIR%.*}.zip" "$DIR"
  mv "${DIR%.*}.zip" out/
  tar -cvzf "${DIR%.*}.tar.gz" "$DIR"
  mv "${DIR%.*}.tar.gz" out/
  rm -rf "$DIR"
  rm out/out*
done
