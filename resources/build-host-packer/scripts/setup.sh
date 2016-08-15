#!/usr/bin/env bash

# disable unattended updates
echo 'APT::Periodic::Enable "0";' >> /etc/apt/apt.conf.d/10periodic

# wait for apt-get auto-update to complete
echo "Waiting for automatic apt-get upgrade to finish ..."

{
	while ps -C apt-get,dpkg ;
	do
		sleep 60
	done
} > /dev/null 2>&1
