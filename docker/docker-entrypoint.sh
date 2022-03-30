#!/bin/bash

cp -r /config $ZEBEAT_HOME/config
$ZEBEAT_HOME/metricbeat run --e --path.config $ZEBEAT_HOME/config