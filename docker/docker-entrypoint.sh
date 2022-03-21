#!/bin/bash

cp -r /config /localconfig
cd /localconfig
/usr/local/bin/metricbeat run --e --path.config /localconfig
