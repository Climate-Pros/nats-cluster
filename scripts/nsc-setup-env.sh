#!/bin/sh

[ -z "$NKEYS_PATH" ] && {
    export NKEYS_PATH=$(pwd)/nsc/nkeys
}

[ -z "$NSC_HOME" ] && {
    export NSC_HOME=$(pwd)/nsc/accounts
}

[ -z "$XDG_DATA_HOME" ] && {
    export XDG_DATA_HOME=$(pwd)/nsc/data
}

[ -z "$XDG_CONFIG_HOME" ] && {
    export XDG_CONFIG_HOME=$(pwd)/nsc/config
}

if [ ! -f .nsc.env ]; then
  echo '
# NSC Environment Setup
export NKEYS_PATH=$(pwd)/nsc/nkeys
export NSC_HOME=$(pwd)/nsc/accounts
export XDG_DATA_HOME=$(pwd)/nsc/data
export XDG_CONFIG_HOME=$(pwd)/nsc/config
' > .nsc.env
fi

mkdir -p "$NKEYS_PATH"
mkdir -p "$NSC_HOME"
mkdir -p "$XDG_DATA_HOME"
mkdir -p "$XDG_CONFIG_HOME"
