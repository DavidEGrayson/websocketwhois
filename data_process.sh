#!/bin/sh

# TODO: find a smooth way to replace data/org.domains while the go process is still running.
# It should be able to use the old file for a while longer, but detect the new file.
zcat data/org.zone.gz | zonefile_process/zonefile_process org > data/org.domains
