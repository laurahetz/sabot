#!/bin/bash
mkdir app/db
DBPATH=$PWD/app/db
AUTHs="false true"
EXPs="10 12 14 16 18"
LEN="32"
DBTYPE="0"

for exp in $EXPs; do
    for auth in $AUTHs; do
        out="db_"$exp"_"$LEN"_"$LEN"_"$auth"";
        echo $out;
        podman run -v $DBPATH:/app/db sabot /app/dbgen -sizeExp=$exp -keyLen=$LEN -valLen=$LEN -auth=$auth -path=/app/db/$out -dbtype=$DBTYPE;
    done
done