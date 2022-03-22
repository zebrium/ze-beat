#!/bin/bash

cp -r /config /localconfig
/usr/local/bin/metricbeat run --e --path.config /localconfig
