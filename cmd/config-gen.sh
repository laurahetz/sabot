#!/bin/bash

# Read input 
#read -p "IP Address: " IP
read -p "Num. threads for multi client exeriment: " THREADS
#read -p "Config File Name (saved in $PWD/app/benchmarks/<NAME>.json) : " NAME
# Can be "F" = full benchmarks, "S" = small DB size benchmarks with low number of repetitions
#read -p 'Which benchmark configs to create "f" (full) or "s" (small for testing): ' BENCH

#THREADS="10"
IP="0.0.0.0"
#NAME="configs.json"

DBPATH=/app/db/
AUTHs="false true"
MULTI="false true"
BENCHs="f s t"


for BENCH in $BENCHs; do


    if [ "$BENCH" = "f" ]; then
        RATEs="1 5 10"
        DBEXPs="10 12 14 16 18"
        REP="16"
        NAME="configs_full.json"
    elif [ "$BENCH" = "s" ]; then
        RATEs="1 5"
        DBEXPs="10 12 14 16"
        REP="5"
        NAME="configs_small.json"
    elif [ "$BENCH" = "t" ]; then
        RATEs="1"
        DBEXPs="10"
        REP="1"
        NAME="configs_test.json"
    else
        echo 'Please choose valid benchmarking set "f" (full) or "s" (small benchmark set) or "t" for function tests.'
        exit 1
    fi

    out='{'$'\r'
    out+='"Addr1": "'"$IP"':50051",'$'\r'
    out+='"Addr2": "'"$IP"':50052",'$'\r'
    out+='"Configs": ['$'\r'

    first=true

    for dbexp in $DBEXPs; do
        for auth in $AUTHs; do
            reset="true"
            for multi in $MULTI; do
                if [ "$multi" = "false" ]; then
                    t=1
                else 
                    t=$THREADS
                fi
                for rate in $RATEs; do
                    if [ "$first" = false ]; then
                        out+=','$'\r'
                    else
                        first=false
                    fi
                        out+='{'$'\r'
                        out+='"Idx": 0,'$'\r'\
                        out+='"Dbfile": "'"$DBPATH"'db_'"$dbexp"'_32_32_'"$auth"'",'$'\r'
                        out+='"RateR": '"$rate"','$'\r'
                        out+='"RateS": '"$rate"','$'\r'
                        out+='"MultiClient": '"$multi"','$'\r'
                        out+='"NumThreads": '"$t"','$'\r'
                        out+='"ResetServer": '"$reset"','$'\r'
                        out+='"Repetitions": '"$REP"','$'\r'
                        out+='"DBType": 0'$'\r'
                        out+='}'
                        # only reset server once for each DB size and auth type
                        if [ "$reset" = "true" ]; then
                            reset="false"
                        fi
                done
            done
        done
    done

    out+=']'$'\r''}'
    OUTPATH="$PWD"/app/benchmarks/"$NAME"
    echo "$out" > "$OUTPATH"
done