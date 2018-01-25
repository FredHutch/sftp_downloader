#!/bin/bash

set -e

>&2 echo "line 1 to stderr"
echo "line 2 to stdout"
echo "line 3 to stdout"
>&2 echo "line 4 to stderr"
>&2 echo "line 5 to stderr"
echo line 6 to stdout
