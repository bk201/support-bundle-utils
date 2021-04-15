# Harvester support bundle utils

This project contains support bundle scripts and utilities for Harvester.

- `bin/harvester-sb-collector.sh`: The script is used to collect k3os node logs. It can be run in a container with host log folder mapped or be run on host.
- `support-bundle-utils`: A program to generate and download Harvester support bundles. Users can get a bundle with the pre-build image:
  ```
  mkdir -p bundles
  docker run --rm -v "$(pwd)/bundles:/bundles" -w /bundles bk201z/support-bundle-utils:dev support-bundle-utils download https://HARVESTER_API_IP:30443 --user <user> --password <password> --insecure

  # the bundle will be stored in bundles
  ```
