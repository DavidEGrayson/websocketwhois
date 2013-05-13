#!/bin/sh

zcat data/org.zone.gz | zonefile_process/zonefile_process org > data/org.domains
